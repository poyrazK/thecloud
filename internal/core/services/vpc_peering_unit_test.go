package services_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testCIDR1 = "10.0.0.0/16"
	testCIDR2 = "10.1.0.0/16"
	brReq     = "br-req"
	brAcc     = "br-acc"
)

func TestVPCPeeringServiceUnit(t *testing.T) {
	repo := new(MockVPCPeeringRepo)
	vpcRepo := new(MockVpcRepo)
	network := new(MockNetworkBackend)
	auditSvc := new(MockAuditService)
	logger := slog.Default()

	svc := services.NewVPCPeeringService(services.VPCPeeringServiceParams{
		Repo:     repo,
		VpcRepo:  vpcRepo,
		Network:  network,
		AuditSvc: auditSvc,
		Logger:   logger,
	})

	ctx := context.Background()
	tenantID := uuid.New()
	userID := uuid.New()
	ctx = appcontext.WithTenantID(ctx, tenantID)
	ctx = appcontext.WithUserID(ctx, userID)

	t.Run("CreatePeering_Success", func(t *testing.T) {
		reqVPCID := uuid.New()
		accVPCID := uuid.New()

		reqVPC := &domain.VPC{ID: reqVPCID, TenantID: tenantID, CIDRBlock: "10.0.0.0/16"}
		accVPC := &domain.VPC{ID: accVPCID, TenantID: tenantID, CIDRBlock: "10.1.0.0/16"}

		vpcRepo.On("GetByID", mock.Anything, reqVPCID).Return(reqVPC, nil).Once()
		vpcRepo.On("GetByID", mock.Anything, accVPCID).Return(accVPC, nil).Once()
		repo.On("GetActiveByVPCPair", mock.Anything, reqVPCID, accVPCID).Return(nil, nil).Once()
		repo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "vpc_peering.create", "vpc_peering", mock.Anything, mock.Anything).Return(nil).Once()

		peering, err := svc.CreatePeering(ctx, reqVPCID, accVPCID)
		require.NoError(t, err)
		assert.NotNil(t, peering)
		assert.Equal(t, domain.PeeringStatusPendingAcceptance, peering.Status)

		repo.AssertExpectations(t)
	})

	t.Run("CreatePeering_OverlappingCIDRs", func(t *testing.T) {
		reqVPCID := uuid.New()
		accVPCID := uuid.New()

		reqVPC := &domain.VPC{ID: reqVPCID, TenantID: tenantID, CIDRBlock: "10.0.0.0/16"}
		accVPC := &domain.VPC{ID: accVPCID, TenantID: tenantID, CIDRBlock: "10.0.1.0/24"} // Overlaps

		vpcRepo.On("GetByID", mock.Anything, reqVPCID).Return(reqVPC, nil).Once()
		vpcRepo.On("GetByID", mock.Anything, accVPCID).Return(accVPC, nil).Once()

		_, err := svc.CreatePeering(ctx, reqVPCID, accVPCID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "VPC CIDR blocks overlap")
	})

	t.Run("AcceptPeering_Success", func(t *testing.T) {
		peeringID := uuid.New()
		reqVPCID := uuid.New()
		accVPCID := uuid.New()

		peering := &domain.VPCPeering{
			ID:             peeringID,
			RequesterVPCID: reqVPCID,
			AccepterVPCID:  accVPCID,
			Status:         domain.PeeringStatusPendingAcceptance,
			TenantID:       tenantID,
		}

		reqVPC := &domain.VPC{ID: reqVPCID, NetworkID: "br-req", CIDRBlock: "10.0.0.0/16"}
		accVPC := &domain.VPC{ID: accVPCID, NetworkID: "br-acc", CIDRBlock: "10.1.0.0/16"}

		repo.On("GetByID", mock.Anything, peeringID).Return(peering, nil).Once()
		vpcRepo.On("GetByID", mock.Anything, reqVPCID).Return(reqVPC, nil).Once()
		vpcRepo.On("GetByID", mock.Anything, accVPCID).Return(accVPC, nil).Once()

		// OVS flows
		network.On("AddFlowRule", mock.Anything, "br-req", mock.Anything).Return(nil).Once()
		network.On("AddFlowRule", mock.Anything, "br-acc", mock.Anything).Return(nil).Once()

		repo.On("UpdateStatus", mock.Anything, peeringID, domain.PeeringStatusActive).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "vpc_peering.accept", "vpc_peering", peeringID.String(), mock.Anything).Return(nil).Once()

		updated, err := svc.AcceptPeering(ctx, peeringID)
		require.NoError(t, err)
		assert.Equal(t, domain.PeeringStatusActive, updated.Status)
	})

	t.Run("DeletePeering_Success", func(t *testing.T) {
		peeringID := uuid.New()
		reqVPCID := uuid.New()
		accVPCID := uuid.New()

		peering := &domain.VPCPeering{
			ID:             peeringID,
			RequesterVPCID: reqVPCID,
			AccepterVPCID:  accVPCID,
			Status:         domain.PeeringStatusActive,
			TenantID:       tenantID,
		}

		reqVPC := &domain.VPC{ID: reqVPCID, NetworkID: "br-req", CIDRBlock: "10.0.0.0/16"}
		accVPC := &domain.VPC{ID: accVPCID, NetworkID: "br-acc", CIDRBlock: "10.1.0.0/16"}

		repo.On("GetByID", mock.Anything, peeringID).Return(peering, nil).Once()
		vpcRepo.On("GetByID", mock.Anything, reqVPCID).Return(reqVPC, nil).Once()
		vpcRepo.On("GetByID", mock.Anything, accVPCID).Return(accVPC, nil).Once()

		// OVS flows removal
		network.On("DeleteFlowRule", mock.Anything, "br-req", mock.Anything).Return(nil).Once()
		network.On("DeleteFlowRule", mock.Anything, "br-acc", mock.Anything).Return(nil).Once()

		repo.On("Delete", mock.Anything, peeringID).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "vpc_peering.delete", "vpc_peering", peeringID.String(), mock.Anything).Return(nil).Once()

		err := svc.DeletePeering(ctx, peeringID)
		require.NoError(t, err)
	})

	t.Run("GetPeering_Success", func(t *testing.T) {
		peeringID := uuid.New()
		peering := &domain.VPCPeering{ID: peeringID, TenantID: tenantID}
		repo.On("GetByID", mock.Anything, peeringID).Return(peering, nil).Once()

		res, err := svc.GetPeering(ctx, peeringID)
		require.NoError(t, err)
		assert.Equal(t, peeringID, res.ID)
	})

	t.Run("ListPeerings_Success", func(t *testing.T) {
		peerings := []*domain.VPCPeering{{ID: uuid.New()}, {ID: uuid.New()}}
		repo.On("List", mock.Anything, tenantID).Return(peerings, nil).Once()

		res, err := svc.ListPeerings(ctx)
		require.NoError(t, err)
		assert.Len(t, res, 2)
	})

	t.Run("RejectPeering_Success", func(t *testing.T) {
		peeringID := uuid.New()
		peering := &domain.VPCPeering{ID: peeringID, Status: domain.PeeringStatusPendingAcceptance, TenantID: tenantID}
		repo.On("GetByID", mock.Anything, peeringID).Return(peering, nil).Once()
		repo.On("UpdateStatus", mock.Anything, peeringID, domain.PeeringStatusRejected).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "vpc_peering.reject", "vpc_peering", peeringID.String(), mock.Anything).Return(nil).Once()

		err := svc.RejectPeering(ctx, peeringID)
		require.NoError(t, err)
	})
}
