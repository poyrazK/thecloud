package services_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupContainerServiceIntegrationTest(t *testing.T) (ports.ContainerService, ports.ContainerRepository, *pgxpool.Pool, context.Context) {
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewPostgresContainerRepository(db)
	eventRepo := postgres.NewEventRepository(db)
	eventSvc := services.NewEventService(eventRepo, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(auditRepo)

	svc := services.NewContainerService(repo, eventSvc, auditSvc)

	return svc, repo, db, ctx
}

func TestContainerService_Integration(t *testing.T) {
	svc, repo, db, ctx := setupContainerServiceIntegrationTest(t)
	userID := appcontext.UserIDFromContext(ctx)

	t.Run("DeploymentLifecycle", func(t *testing.T) {
		name := "web-deployment"
		dep, err := svc.CreateDeployment(ctx, name, "nginx:latest", 3, "80:80")
		assert.NoError(t, err)
		assert.NotNil(t, dep)
		assert.Equal(t, name, dep.Name)
		assert.Equal(t, 3, dep.Replicas)

		// Get
		fetched, err := svc.GetDeployment(ctx, dep.ID)
		assert.NoError(t, err)
		assert.Equal(t, dep.ID, fetched.ID)

		// List
		deps, err := svc.ListDeployments(ctx)
		assert.NoError(t, err)
		assert.Len(t, deps, 1)

		// Scale
		err = svc.ScaleDeployment(ctx, dep.ID, 5)
		assert.NoError(t, err)

		scaled, _ := repo.GetDeploymentByID(ctx, dep.ID, userID)
		assert.Equal(t, 5, scaled.Replicas)
		assert.Equal(t, domain.DeploymentStatusScaling, scaled.Status)

		// Delete
		err = svc.DeleteDeployment(ctx, dep.ID)
		assert.NoError(t, err)

		deleting, _ := repo.GetDeploymentByID(ctx, dep.ID, userID)
		assert.Equal(t, domain.DeploymentStatusDeleting, deleting.Status)
	})

	t.Run("ContainerManagement", func(t *testing.T) {
		tenantID := appcontext.TenantIDFromContext(ctx)
		dep, _ := svc.CreateDeployment(ctx, "cnt-test", "alpine", 1, "")

		// Create instance first to satisfy FK constraint
		instRepo := postgres.NewInstanceRepository(db)
		instID := uuid.New()
		err := instRepo.Create(ctx, &domain.Instance{
			ID:       instID,
			UserID:   userID,
			TenantID: tenantID,
			Name:     "cnt-inst",
			Status:   domain.StatusStarting,
			Image:    "alpine",
		})
		require.NoError(t, err)

		err = repo.AddContainer(ctx, dep.ID, instID)
		require.NoError(t, err)

		containers, err := repo.GetContainers(ctx, dep.ID)
		assert.NoError(t, err)
		require.Len(t, containers, 1)
		assert.Equal(t, instID, containers[0])

		err = repo.RemoveContainer(ctx, dep.ID, instID)
		assert.NoError(t, err)

		containers, _ = repo.GetContainers(ctx, dep.ID)
		assert.Empty(t, containers)
	})
}

func TestContainer_ChaosRestart(t *testing.T) {
	// 1. Setup
	db, instSvc, compute, instRepo, _, _, ctx := setupInstanceServiceTest(t)

	containerRepo := postgres.NewPostgresContainerRepository(db)
	eventSvc := services.NewEventService(postgres.NewEventRepository(db), nil, slog.Default())
	auditSvc := services.NewAuditService(postgres.NewAuditRepository(db))

	containerSvc := services.NewContainerService(containerRepo, eventSvc, auditSvc)
	worker := services.NewContainerWorker(containerRepo, instSvc, eventSvc)

	// 2. Create Deployment
	dep, err := containerSvc.CreateDeployment(ctx, "chaos-web", "alpine:latest", 1, "")
	require.NoError(t, err)

	// 3. Reconcile (Should launch 1 instance)
	worker.Reconcile(ctx)

	// Verify 1 association
	ids, err := containerRepo.GetContainers(ctx, dep.ID)
	require.NoError(t, err)
	require.Len(t, ids, 1)
	instID := ids[0]

	// Manually provision it (simulating worker)
	err = instSvc.Provision(ctx, domain.ProvisionJob{InstanceID: instID})
	require.NoError(t, err)

	// Get container ID
	inst, err := instRepo.GetByID(ctx, instID)
	require.NoError(t, err)
	containerID := inst.ContainerID
	require.NotEmpty(t, containerID)

	// 4. CHAOS: Set status to ERROR in DB
	// This simulates the worker detecting a failure or a monitor updating the status
	_, err = db.Exec(ctx, "UPDATE instances SET status = 'ERROR' WHERE id = $1", instID)
	require.NoError(t, err)

	// 5. Reconcile Again (Should replace)
	worker.Reconcile(ctx)

	// 6. Verify New Instance
	newIds, err := containerRepo.GetContainers(ctx, dep.ID)
	require.NoError(t, err)
	assert.Len(t, newIds, 1)
	assert.NotEqual(t, instID, newIds[0], "Should have replaced the unhealthy instance")

	// Cleanup
	for _, id := range newIds {
		inst, _ := instRepo.GetByID(ctx, id)
		if inst.ContainerID != "" {
			_ = compute.DeleteInstance(ctx, inst.ContainerID)
		}
	}
}
