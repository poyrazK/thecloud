package services_test

import (
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/noop"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVPCPeeringServiceIntegration(t *testing.T) {
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)
	tenantID := appcontext.TenantIDFromContext(ctx)

	repo := postgres.NewVPCPeeringRepository(db)
	vpcRepo := postgres.NewVpcRepository(db)
	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(auditRepo)
	network := noop.NewNoopNetworkAdapter(slog.Default())

	svc := services.NewVPCPeeringService(services.VPCPeeringServiceParams{
		Repo:     repo,
		VpcRepo:  vpcRepo,
		Network:  network,
		AuditSvc: auditSvc,
		Logger:   slog.Default(),
	})

	// Helper to create a VPC
	createVPC := func(name, cidr string) *domain.VPC {
		vpc := &domain.VPC{
			ID:        uuid.New(),
			UserID:    appcontext.UserIDFromContext(ctx),
			TenantID:  tenantID,
			Name:      name,
			CIDRBlock: cidr,
			Status:    "available",
			CreatedAt: time.Now(),
		}
		err := vpcRepo.Create(ctx, vpc)
		require.NoError(t, err)
		return vpc
	}

	t.Run("Lifecycle_Success", func(t *testing.T) {
		vpc1 := createVPC("vpc-1", "10.10.0.0/16")
		vpc2 := createVPC("vpc-2", "10.20.0.0/16")

		// 1. Create Peering
		peering, err := svc.CreatePeering(ctx, vpc1.ID, vpc2.ID)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, peering.ID)
		assert.Equal(t, domain.PeeringStatusPendingAcceptance, peering.Status)

		// 2. Accept Peering
		accepted, err := svc.AcceptPeering(ctx, peering.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.PeeringStatusActive, accepted.Status)

		// 3. Get Peering
		got, err := svc.GetPeering(ctx, peering.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.PeeringStatusActive, got.Status)

		// 4. List Peerings
		list, err := svc.ListPeerings(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, list)

		// 5. Delete Peering
		err = svc.DeletePeering(ctx, peering.ID)
		require.NoError(t, err)

		// 6. Verify Deleted
		_, err = svc.GetPeering(ctx, peering.ID)
		require.Error(t, err)
	})

	t.Run("CreatePeering_OverlappingCIDRs", func(t *testing.T) {
		vpc1 := createVPC("vpc-overlap-1", "10.30.0.0/16")
		vpc2 := createVPC("vpc-overlap-2", "10.30.10.0/24") // Overlaps

		_, err := svc.CreatePeering(ctx, vpc1.ID, vpc2.ID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "overlap")
	})

	t.Run("RejectPeering_Success", func(t *testing.T) {
		vpc1 := createVPC("vpc-reject-1", "10.40.0.0/16")
		vpc2 := createVPC("vpc-reject-2", "10.50.0.0/16")

		peering, _ := svc.CreatePeering(ctx, vpc1.ID, vpc2.ID)
		
		err := svc.RejectPeering(ctx, peering.ID)
		require.NoError(t, err)

		got, _ := svc.GetPeering(ctx, peering.ID)
		assert.Equal(t, domain.PeeringStatusRejected, got.Status)
	})

	t.Run("MultiTenancy_Isolation", func(t *testing.T) {
		vpc1 := createVPC("vpc-t1-1", "10.60.0.0/16")
		vpc2 := createVPC("vpc-t1-2", "10.70.0.0/16")
		peering, _ := svc.CreatePeering(ctx, vpc1.ID, vpc2.ID)

		// Create another user/tenant
		ctx2 := setupTestUser(t, db)
		
		// Should not be able to see peering from first tenant
		list, err := svc.ListPeerings(ctx2)
		require.NoError(t, err)
		for _, p := range list {
			assert.NotEqual(t, peering.ID, p.ID)
		}

		// Should not be able to reject peering from first tenant
		err = svc.RejectPeering(ctx2, peering.ID)
		require.Error(t, err)
	})
}
