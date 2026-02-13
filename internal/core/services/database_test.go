package services_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/docker"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testDBName = "test-db"
)

func setupDatabaseServiceTest(t *testing.T) (ports.DatabaseService, ports.DatabaseRepository, *docker.DockerAdapter, ports.VpcRepository, context.Context) {
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewDatabaseRepository(db)
	vpcRepo := postgres.NewVpcRepository(db)

	compute, err := docker.NewDockerAdapter(slog.Default())
	require.NoError(t, err)

	eventRepo := postgres.NewEventRepository(db)
	eventSvc := services.NewEventService(eventRepo, nil, slog.Default())

	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(auditRepo)

	logger := slog.Default()

	svc := services.NewDatabaseService(services.DatabaseServiceParams{
		Repo:     repo,
		Compute:  compute,
		VpcRepo:  vpcRepo,
		EventSvc: eventSvc,
		AuditSvc: auditSvc,
		Logger:   logger,
	})

	return svc, repo, compute, vpcRepo, ctx
}

func TestCreateDatabaseSuccess(t *testing.T) {
	svc, repo, compute, _, ctx := setupDatabaseServiceTest(t)

	db, err := svc.CreateDatabase(ctx, testDBName, "postgres", "16", nil)

	assert.NoError(t, err)
	assert.NotNil(t, db)
	assert.Equal(t, testDBName, db.Name)
	assert.Equal(t, domain.EnginePostgres, db.Engine)
	assert.NotEmpty(t, db.ContainerID)

	// Verify instance creation by checking connectivity
	ip, err := compute.GetInstanceIP(ctx, db.ContainerID)
	// Note: It might take a moment or fail if not yet ready, but Adapter retries.
	// For integration test with real docker, this should work eventually.
	assert.NoError(t, err)
	assert.NotEmpty(t, ip)

	// Clean up created container
	_ = compute.DeleteInstance(ctx, db.ContainerID)

	// Verify in DB
	fetched, err := repo.GetByID(ctx, db.ID)
	assert.NoError(t, err)
	assert.Equal(t, db.ID, fetched.ID)
}

func TestCreateDatabaseWithVpc(t *testing.T) {
	svc, _, compute, vpcRepo, ctx := setupDatabaseServiceTest(t)

	// Create a real VPC first
	// We need to create a VPC using repo, ensuring context has UserID
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	vpcID := uuid.New()
	// Create actual docker network
	networkName := "net-" + vpcID.String()
	netID, err := compute.CreateNetwork(ctx, networkName)
	require.NoError(t, err)
	defer func() { _ = compute.DeleteNetwork(ctx, netID) }()

	vpc := &domain.VPC{
		ID:        vpcID,
		UserID:    userID,
		TenantID:  tenantID,
		Name:      "test-vpc",
		CIDRBlock: "10.0.0.0/16",
		NetworkID: netID, // Use the real docker network ID
		Status:    "ACTIVE",
		CreatedAt: time.Now(),
	}
	err = vpcRepo.Create(ctx, vpc)
	require.NoError(t, err)

	// Now create DB in this VPC
	db, err := svc.CreateDatabase(ctx, testDBName, "postgres", "16", &vpcID)
	require.NoError(t, err)
	require.NotNil(t, db)
	assert.Equal(t, &vpcID, db.VpcID)

	// Cleanup
	_ = compute.DeleteInstance(ctx, db.ContainerID)
}
