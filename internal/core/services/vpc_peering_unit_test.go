package services_test

import (
	"context"
	"fmt"
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

func TestVPCPeeringService_Unit(t *testing.T) {
	mockRepo := new(MockVPCPeeringRepo)
	mockVpcRepo := new(MockVpcRepo)
	mockNetwork := new(MockNetworkBackend)
	mockAuditSvc := new(MockAuditService)
	svc := services.NewVPCPeeringService(services.VPCPeeringServiceParams{
		Repo:     mockRepo,
		VpcRepo:  mockVpcRepo,
		Network:  mockNetwork,
		AuditSvc: mockAuditSvc,
		Logger:   slog.Default(),
	})

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	t.Run("CreatePeering_SelfPeering", func(t *testing.T) {
		id := uuid.New()
		_, err := svc.CreatePeering(ctx, id, id)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot peer a VPC with itself")
	})

	t.Run("CreatePeering_RequesterVPCNotFound", func(t *testing.T) {
		vpcID := uuid.New()
		mockVpcRepo.On("GetByID", mock.Anything, vpcID).Return(nil, fmt.Errorf("not found")).Once()

		_, err := svc.CreatePeering(ctx, vpcID, uuid.New())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "requester VPC not found")
	})

	t.Run("CreatePeering_AccepterVPCNotFound", func(t *testing.T) {
		requesterVPCID := uuid.New()
		accepterVPCID := uuid.New()
		requesterVPC := &domain.VPC{ID: requesterVPCID, Name: "req", CIDRBlock: "10.0.0.0/16"}
		mockVpcRepo.On("GetByID", mock.Anything, requesterVPCID).Return(requesterVPC, nil).Once()
		mockVpcRepo.On("GetByID", mock.Anything, accepterVPCID).Return(nil, fmt.Errorf("not found")).Once()

		_, err := svc.CreatePeering(ctx, requesterVPCID, accepterVPCID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "accepter VPC not found")
	})

	t.Run("CreatePeering_CIDROverlap", func(t *testing.T) {
		requesterVPCID := uuid.New()
		accepterVPCID := uuid.New()
		requesterVPC := &domain.VPC{ID: requesterVPCID, Name: "req", CIDRBlock: "10.0.0.0/16"}
		accepterVPC := &domain.VPC{ID: accepterVPCID, Name: "acc", CIDRBlock: "10.0.0.0/24"} // Overlaps
		mockVpcRepo.On("GetByID", mock.Anything, requesterVPCID).Return(requesterVPC, nil).Once()
		mockVpcRepo.On("GetByID", mock.Anything, accepterVPCID).Return(accepterVPC, nil).Once()

		_, err := svc.CreatePeering(ctx, requesterVPCID, accepterVPCID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "overlap")
	})

	t.Run("CreatePeering_DuplicatePeering", func(t *testing.T) {
		requesterVPCID := uuid.New()
		accepterVPCID := uuid.New()
		requesterVPC := &domain.VPC{ID: requesterVPCID, Name: "req", CIDRBlock: "10.0.0.0/16"}
		accepterVPC := &domain.VPC{ID: accepterVPCID, Name: "acc", CIDRBlock: "11.0.0.0/16"}
		mockVpcRepo.On("GetByID", mock.Anything, requesterVPCID).Return(requesterVPC, nil).Once()
		mockVpcRepo.On("GetByID", mock.Anything, accepterVPCID).Return(accepterVPC, nil).Once()
		mockRepo.On("GetActiveByVPCPair", mock.Anything, requesterVPCID, accepterVPCID).Return(&domain.VPCPeering{ID: uuid.New()}, nil).Once()

		_, err := svc.CreatePeering(ctx, requesterVPCID, accepterVPCID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("CreatePeering_Success", func(t *testing.T) {
		requesterVPCID := uuid.New()
		accepterVPCID := uuid.New()
		requesterVPC := &domain.VPC{ID: requesterVPCID, Name: "req", CIDRBlock: "10.0.0.0/16"}
		accepterVPC := &domain.VPC{ID: accepterVPCID, Name: "acc", CIDRBlock: "11.0.0.0/16"}
		mockVpcRepo.On("GetByID", mock.Anything, requesterVPCID).Return(requesterVPC, nil).Once()
		mockVpcRepo.On("GetByID", mock.Anything, accepterVPCID).Return(accepterVPC, nil).Once()
		mockRepo.On("GetActiveByVPCPair", mock.Anything, requesterVPCID, accepterVPCID).Return(nil, nil).Once()
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "vpc_peering.create", "vpc_peering", mock.Anything, mock.Anything).Return(nil).Once()

		peering, err := svc.CreatePeering(ctx, requesterVPCID, accepterVPCID)
		require.NoError(t, err)
		assert.NotNil(t, peering)
		assert.Equal(t, domain.PeeringStatusPendingAcceptance, peering.Status)
		assert.Equal(t, requesterVPCID, peering.RequesterVPCID)
		assert.Equal(t, accepterVPCID, peering.AccepterVPCID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("AcceptPeering_WrongStatus", func(t *testing.T) {
		peeringID := uuid.New()
		peering := &domain.VPCPeering{ID: peeringID, Status: domain.PeeringStatusActive}
		mockRepo.On("GetByID", mock.Anything, peeringID).Return(peering, nil).Once()

		_, err := svc.AcceptPeering(ctx, peeringID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "only pending")
	})

	t.Run("AcceptPeering_NotFound", func(t *testing.T) {
		peeringID := uuid.New()
		mockRepo.On("GetByID", mock.Anything, peeringID).Return(nil, fmt.Errorf("not found")).Once()

		_, err := svc.AcceptPeering(ctx, peeringID)
		require.Error(t, err)
	})

	t.Run("AcceptPeering_FlowProgrammingFailure", func(t *testing.T) {
		peeringID := uuid.New()
		requesterVPCID := uuid.New()
		accepterVPCID := uuid.New()
		peering := &domain.VPCPeering{
			ID:             peeringID,
			RequesterVPCID: requesterVPCID,
			AccepterVPCID:  accepterVPCID,
			Status:         domain.PeeringStatusPendingAcceptance,
		}
		requesterVPC := &domain.VPC{ID: requesterVPCID, Name: "req", NetworkID: "net-req", CIDRBlock: "10.0.0.0/16"}
		accepterVPC := &domain.VPC{ID: accepterVPCID, Name: "acc", NetworkID: "net-acc", CIDRBlock: "11.0.0.0/16"}

		mockRepo.On("GetByID", mock.Anything, peeringID).Return(peering, nil).Once()
		mockVpcRepo.On("GetByID", mock.Anything, requesterVPCID).Return(requesterVPC, nil).Once()
		mockVpcRepo.On("GetByID", mock.Anything, accepterVPCID).Return(accepterVPC, nil).Once()
		mockNetwork.On("AddFlowRule", mock.Anything, "net-req", mock.Anything).Return(fmt.Errorf("flow error")).Once()
		mockRepo.On("UpdateStatus", mock.Anything, peeringID, domain.PeeringStatusFailed).Return(nil).Once()

		_, err := svc.AcceptPeering(ctx, peeringID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to establish network peering")
	})

	t.Run("AcceptPeering_Success", func(t *testing.T) {
		peeringID := uuid.New()
		requesterVPCID := uuid.New()
		accepterVPCID := uuid.New()
		peering := &domain.VPCPeering{
			ID:             peeringID,
			RequesterVPCID: requesterVPCID,
			AccepterVPCID:  accepterVPCID,
			Status:         domain.PeeringStatusPendingAcceptance,
		}
		requesterVPC := &domain.VPC{ID: requesterVPCID, Name: "req", NetworkID: "net-req", CIDRBlock: "10.0.0.0/16"}
		accepterVPC := &domain.VPC{ID: accepterVPCID, Name: "acc", NetworkID: "net-acc", CIDRBlock: "11.0.0.0/16"}

		mockRepo.On("GetByID", mock.Anything, peeringID).Return(peering, nil).Once()
		mockVpcRepo.On("GetByID", mock.Anything, requesterVPCID).Return(requesterVPC, nil).Once()
		mockVpcRepo.On("GetByID", mock.Anything, accepterVPCID).Return(accepterVPC, nil).Once()
		mockNetwork.On("AddFlowRule", mock.Anything, "net-req", mock.Anything).Return(nil).Once()
		mockNetwork.On("AddFlowRule", mock.Anything, "net-acc", mock.Anything).Return(nil).Once()
		mockRepo.On("UpdateStatus", mock.Anything, peeringID, domain.PeeringStatusActive).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "vpc_peering.accept", "vpc_peering", mock.Anything, mock.Anything).Return(nil).Once()

		result, err := svc.AcceptPeering(ctx, peeringID)
		require.NoError(t, err)
		assert.Equal(t, domain.PeeringStatusActive, result.Status)
		mockRepo.AssertExpectations(t)
	})

	t.Run("RejectPeering_NotFound", func(t *testing.T) {
		peeringID := uuid.New()
		mockRepo.On("GetByID", mock.Anything, peeringID).Return(nil, fmt.Errorf("not found")).Once()

		err := svc.RejectPeering(ctx, peeringID)
		require.Error(t, err)
	})

	t.Run("RejectPeering_WrongStatus", func(t *testing.T) {
		peeringID := uuid.New()
		peering := &domain.VPCPeering{ID: peeringID, Status: domain.PeeringStatusActive}
		mockRepo.On("GetByID", mock.Anything, peeringID).Return(peering, nil).Once()

		err := svc.RejectPeering(ctx, peeringID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "only pending")
	})

	t.Run("RejectPeering_Success", func(t *testing.T) {
		peeringID := uuid.New()
		peering := &domain.VPCPeering{ID: peeringID, Status: domain.PeeringStatusPendingAcceptance}
		mockRepo.On("GetByID", mock.Anything, peeringID).Return(peering, nil).Once()
		mockRepo.On("UpdateStatus", mock.Anything, peeringID, domain.PeeringStatusRejected).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "vpc_peering.reject", "vpc_peering", mock.Anything, mock.Anything).Return(nil).Once()

		err := svc.RejectPeering(ctx, peeringID)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("DeletePeering_NotFound", func(t *testing.T) {
		peeringID := uuid.New()
		mockRepo.On("GetByID", mock.Anything, peeringID).Return(nil, fmt.Errorf("not found")).Once()

		err := svc.DeletePeering(ctx, peeringID)
		require.Error(t, err)
	})

	t.Run("DeletePeering_NonActiveStatus", func(t *testing.T) {
		peeringID := uuid.New()
		peering := &domain.VPCPeering{ID: peeringID, Status: domain.PeeringStatusPendingAcceptance}
		mockRepo.On("GetByID", mock.Anything, peeringID).Return(peering, nil).Once()
		mockRepo.On("Delete", mock.Anything, peeringID).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "vpc_peering.delete", "vpc_peering", mock.Anything, mock.Anything).Return(nil).Once()

		err := svc.DeletePeering(ctx, peeringID)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("DeletePeering_ActiveWithOVSCleanup", func(t *testing.T) {
		peeringID := uuid.New()
		requesterVPCID := uuid.New()
		accepterVPCID := uuid.New()
		peering := &domain.VPCPeering{
			ID:             peeringID,
			RequesterVPCID: requesterVPCID,
			AccepterVPCID:  accepterVPCID,
			Status:         domain.PeeringStatusActive,
		}
		requesterVPC := &domain.VPC{ID: requesterVPCID, Name: "req", NetworkID: "net-req", CIDRBlock: "10.0.0.0/16"}
		accepterVPC := &domain.VPC{ID: accepterVPCID, Name: "acc", NetworkID: "net-acc", CIDRBlock: "11.0.0.0/16"}

		mockRepo.On("GetByID", mock.Anything, peeringID).Return(peering, nil).Once()
		mockVpcRepo.On("GetByID", mock.Anything, requesterVPCID).Return(requesterVPC, nil).Once()
		mockVpcRepo.On("GetByID", mock.Anything, accepterVPCID).Return(accepterVPC, nil).Once()
		mockNetwork.On("DeleteFlowRule", mock.Anything, "net-req", mock.Anything).Return(nil).Once()
		mockNetwork.On("DeleteFlowRule", mock.Anything, "net-acc", mock.Anything).Return(nil).Once()
		mockRepo.On("Delete", mock.Anything, peeringID).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "vpc_peering.delete", "vpc_peering", mock.Anything, mock.Anything).Return(nil).Once()

		err := svc.DeletePeering(ctx, peeringID)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("GetPeering_Success", func(t *testing.T) {
		peeringID := uuid.New()
		expectedPeering := &domain.VPCPeering{ID: peeringID, Status: domain.PeeringStatusActive}
		mockRepo.On("GetByID", mock.Anything, peeringID).Return(expectedPeering, nil).Once()

		result, err := svc.GetPeering(ctx, peeringID)
		require.NoError(t, err)
		assert.Equal(t, peeringID, result.ID)
		assert.Equal(t, domain.PeeringStatusActive, result.Status)
	})

	t.Run("GetPeering_NotFound", func(t *testing.T) {
		peeringID := uuid.New()
		mockRepo.On("GetByID", mock.Anything, peeringID).Return(nil, fmt.Errorf("not found")).Once()

		_, err := svc.GetPeering(ctx, peeringID)
		require.Error(t, err)
	})

	t.Run("ListPeerings_Success", func(t *testing.T) {
		peerings := []*domain.VPCPeering{
			{ID: uuid.New(), Status: domain.PeeringStatusActive},
			{ID: uuid.New(), Status: domain.PeeringStatusPendingAcceptance},
		}
		mockRepo.On("List", mock.Anything, tenantID).Return(peerings, nil).Once()

		result, err := svc.ListPeerings(ctx)
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("ListPeerings_Error", func(t *testing.T) {
		mockRepo.On("List", mock.Anything, tenantID).Return(nil, fmt.Errorf("db error")).Once()

		_, err := svc.ListPeerings(ctx)
		require.Error(t, err)
	})
}
