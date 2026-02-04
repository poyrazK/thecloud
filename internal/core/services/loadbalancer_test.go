package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupLBServiceIntegrationTest(t *testing.T) (ports.LBService, *postgres.LBRepository, *postgres.VpcRepository, *postgres.InstanceRepository, context.Context) {
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	lbRepo := postgres.NewLBRepository(db)
	vpcRepo := postgres.NewVpcRepository(db)
	instRepo := postgres.NewInstanceRepository(db)
	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(auditRepo)

	svc := services.NewLBService(lbRepo, vpcRepo, instRepo, auditSvc)
	return svc, lbRepo, vpcRepo, instRepo, ctx
}

func TestLBService_Integration(t *testing.T) {
	svc, lbRepo, vpcRepo, instRepo, ctx := setupLBServiceIntegrationTest(t)
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	// Setup a VPC
	vpc := &domain.VPC{
		ID:       uuid.New(),
		UserID:   userID,
		TenantID: tenantID,
		Name:     "lb-vpc",
	}
	err := vpcRepo.Create(ctx, vpc)
	require.NoError(t, err)

	t.Run("CreateLB", func(t *testing.T) {
		name := "test-lb"
		port := 80
		algo := "round-robin"
		key1 := "idempotency-key-1"

		lb, err := svc.Create(ctx, name, vpc.ID, port, algo, key1)
		assert.NoError(t, err)
		assert.NotNil(t, lb)
		assert.Equal(t, name, lb.Name)
		assert.Equal(t, userID, lb.UserID)

		// Idempotency check
		lb2, err := svc.Create(ctx, name, vpc.ID, port, algo, key1)
		assert.NoError(t, err)
		assert.Equal(t, lb.ID, lb2.ID)
	})

	t.Run("AddTarget", func(t *testing.T) {
		lb, _ := svc.Create(ctx, "target-lb", vpc.ID, 8080, "least-conn", "key-target")

		// Setup an instance in the same VPC
		inst := &domain.Instance{
			ID:       uuid.New(),
			UserID:   userID,
			TenantID: tenantID,
			Name:     "inst-1",
			VpcID:    &vpc.ID,
		}
		err := instRepo.Create(ctx, inst)
		require.NoError(t, err)

		err = svc.AddTarget(ctx, lb.ID, inst.ID, 80, 1)
		assert.NoError(t, err)

		// Verify target exists
		targets, err := lbRepo.ListTargets(ctx, lb.ID)
		assert.NoError(t, err)
		assert.Len(t, targets, 1)
		assert.Equal(t, inst.ID, targets[0].InstanceID)
	})

	t.Run("ListAndGet", func(t *testing.T) {
		lbs, err := svc.List(ctx)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(lbs), 1)

		lbID := lbs[0].ID
		fetched, err := svc.Get(ctx, lbID)
		assert.NoError(t, err)
		assert.Equal(t, lbID, fetched.ID)
	})

	t.Run("Delete", func(t *testing.T) {
		lb, _ := svc.Create(ctx, "to-delete", vpc.ID, 9000, "round-robin", "key-del")

		err := svc.Delete(ctx, lb.ID)
		assert.NoError(t, err)

		// Soft delete check - should be updated to StatusDeleted
		fetched, err := svc.Get(ctx, lb.ID)
		assert.NoError(t, err)
		assert.Equal(t, domain.LBStatusDeleted, fetched.Status)
	})
}
