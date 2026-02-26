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

func TestVpcService_Unit(t *testing.T) {
	repo := new(MockVpcRepo)
	lbRepo := new(MockLBRepo)
	network := new(MockNetworkBackend)
	auditSvc := new(MockAuditService)
	peeringRepo := new(MockVPCPeeringRepo)
	svc := services.NewVpcService(repo, lbRepo, peeringRepo, network, auditSvc, slog.Default(), "10.0.0.0/16")

	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)

	t.Run("CreateVPC", func(t *testing.T) {
		network.On("CreateBridge", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		repo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "vpc.create", "vpc", mock.Anything, mock.Anything).Return(nil).Once()

		vpc, err := svc.CreateVPC(ctx, "test-vpc", "10.1.0.0/16")
		require.NoError(t, err)
		assert.NotNil(t, vpc)
		assert.Equal(t, "10.1.0.0/16", vpc.CIDRBlock)
		repo.AssertExpectations(t)
	})

	t.Run("DeleteVPC_Success", func(t *testing.T) {
		vpcID := uuid.New()
		vpc := &domain.VPC{ID: vpcID, UserID: userID, NetworkID: "br-1"}

		repo.On("GetByID", mock.Anything, vpcID).Return(vpc, nil).Once()
		lbRepo.On("ListAll", mock.Anything).Return([]*domain.LoadBalancer{}, nil).Once()
		peeringRepo.On("ListByVPC", mock.Anything, vpcID).Return([]*domain.VPCPeering{}, nil).Once()
		network.On("DeleteBridge", mock.Anything, "br-1").Return(nil).Once()
		repo.On("Delete", mock.Anything, vpcID).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "vpc.delete", "vpc", vpcID.String(), mock.Anything).Return(nil).Once()

		err := svc.DeleteVPC(ctx, vpcID.String())
		require.NoError(t, err)
	})

	t.Run("DeleteVPC_WithActivePeering", func(t *testing.T) {
		vpcID := uuid.New()
		vpc := &domain.VPC{ID: vpcID, UserID: userID, NetworkID: "br-1"}

		repo.On("GetByID", mock.Anything, vpcID).Return(vpc, nil).Once()
		lbRepo.On("ListAll", mock.Anything).Return([]*domain.LoadBalancer{}, nil).Once()
		peeringRepo.On("ListByVPC", mock.Anything, vpcID).Return([]*domain.VPCPeering{
			{ID: uuid.New(), Status: domain.PeeringStatusActive},
		}, nil).Once()

		err := svc.DeleteVPC(ctx, vpcID.String())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "active peering connections")
	})
}
