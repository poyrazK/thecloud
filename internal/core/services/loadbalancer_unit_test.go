package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLBService_Unit(t *testing.T) {
	mockRepo := new(MockLBRepo)
	mockVpcRepo := new(MockVpcRepo)
	mockInstRepo := new(MockInstanceRepo)
	mockAuditSvc := new(MockAuditService)
	svc := services.NewLBService(mockRepo, mockVpcRepo, mockInstRepo, mockAuditSvc)

	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)

	t.Run("CreateLB", func(t *testing.T) {
		vpcID := uuid.New()
		mockVpcRepo.On("GetByID", mock.Anything, vpcID).Return(&domain.VPC{ID: vpcID}, nil).Once()
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "lb.create", "loadbalancer", mock.Anything, mock.Anything).Return(nil).Once()

		lb, err := svc.Create(ctx, "test-lb", vpcID, 80, "round-robin", "")
		assert.NoError(t, err)
		assert.NotNil(t, lb)
		assert.Equal(t, "test-lb", lb.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("AddTarget", func(t *testing.T) {
		lbID := uuid.New()
		vpcID := uuid.New()
		instID := uuid.New()
		
		lb := &domain.LoadBalancer{ID: lbID, VpcID: vpcID, UserID: userID}
		inst := &domain.Instance{ID: instID, VpcID: &vpcID}
		
		mockRepo.On("GetByID", mock.Anything, lbID).Return(lb, nil).Once()
		mockInstRepo.On("GetByID", mock.Anything, instID).Return(inst, nil).Once()
		mockRepo.On("AddTarget", mock.Anything, mock.Anything).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "lb.target_add", "loadbalancer", lbID.String(), mock.Anything).Return(nil).Once()

		err := svc.AddTarget(ctx, lbID, instID, 8080, 1)
		assert.NoError(t, err)
	})
}
