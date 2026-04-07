package services_test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"

	"github.com/google/uuid"
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

	t.Run("CreatePeering_SelfPeering", func(t *testing.T) {
		id := uuid.New()
		_, err := svc.CreatePeering(ctx, id, id)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot peer a VPC with itself")
	})

	t.Run("CreatePeering_VPCNotFound", func(t *testing.T) {
		vpcID := uuid.New()
		mockVpcRepo.On("GetByID", mock.Anything, vpcID).Return(nil, fmt.Errorf("not found")).Once()

		_, err := svc.CreatePeering(ctx, vpcID, uuid.New())
		require.Error(t, err)
	})

	t.Run("AcceptPeering_WrongStatus", func(t *testing.T) {
		peeringID := uuid.New()
		peering := &domain.VPCPeering{ID: peeringID, Status: domain.PeeringStatusActive}
		mockRepo.On("GetByID", mock.Anything, peeringID).Return(peering, nil).Once()

		_, err := svc.AcceptPeering(ctx, peeringID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "only pending")
	})

	t.Run("RejectPeering_NotFound", func(t *testing.T) {
		peeringID := uuid.New()
		mockRepo.On("GetByID", mock.Anything, peeringID).Return(nil, fmt.Errorf("not found")).Once()

		err := svc.RejectPeering(ctx, peeringID)
		require.Error(t, err)
	})
}
