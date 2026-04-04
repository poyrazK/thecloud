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

func TestLBService_Unit(t *testing.T) {
	mockRepo := new(MockLBRepo)
	mockVpcRepo := new(MockVpcRepo)
	mockInstRepo := new(MockInstanceRepo)
	mockAuditSvc := new(MockAuditService)
	rbacSvc := new(MockRBACService)
	rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	svc := services.NewLBService(mockRepo, rbacSvc, mockVpcRepo, mockInstRepo, mockAuditSvc, slog.Default())

	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)

	t.Run("CreateLB", func(t *testing.T) {
		vpcID := uuid.New()
		mockVpcRepo.On("GetByID", mock.Anything, vpcID).Return(&domain.VPC{ID: vpcID}, nil).Once()
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "lb.create", "loadbalancer", mock.Anything, mock.Anything).Return(nil).Once()

		lb, err := svc.Create(ctx, "test-lb", vpcID, 80, "round-robin", "")
		require.NoError(t, err)
		assert.NotNil(t, lb)
		assert.Equal(t, "test-lb", lb.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("CreateLB_Idempotency", func(t *testing.T) {
		key := "key-123"
		vpcID := uuid.New()
		expected := &domain.LoadBalancer{ID: uuid.New(), Name: "existing", IdempotencyKey: key}
		mockRepo.On("GetByIdempotencyKey", mock.Anything, key).Return(expected, nil).Once()

		lb, err := svc.Create(ctx, "ignored", vpcID, 80, "", key)
		require.NoError(t, err)
		assert.Equal(t, expected, lb)
	})

	t.Run("CreateLB_VPCNotFound", func(t *testing.T) {
		vpcID := uuid.New()
		mockVpcRepo.On("GetByID", mock.Anything, vpcID).Return(nil, fmt.Errorf("not found")).Once()

		_, err := svc.Create(ctx, "fail", vpcID, 80, "", "")
		require.Error(t, err)
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
		require.NoError(t, err)
	})

	t.Run("AddTarget_CrossVPC", func(t *testing.T) {
		lbID := uuid.New()
		vpc1 := uuid.New()
		vpc2 := uuid.New()
		instID := uuid.New()

		lb := &domain.LoadBalancer{ID: lbID, VpcID: vpc1}
		inst := &domain.Instance{ID: instID, VpcID: &vpc2}

		mockRepo.On("GetByID", mock.Anything, lbID).Return(lb, nil).Once()
		mockInstRepo.On("GetByID", mock.Anything, instID).Return(inst, nil).Once()

		err := svc.AddTarget(ctx, lbID, instID, 80, 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "target must be in same VPC as LB")
	})

	t.Run("DeleteLB", func(t *testing.T) {
		lbID := uuid.New()
		lb := &domain.LoadBalancer{ID: lbID, UserID: userID, Name: "to-delete"}
		mockRepo.On("GetByID", mock.Anything, lbID).Return(lb, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(l *domain.LoadBalancer) bool {
			return l.Status == domain.LBStatusDeleted
		})).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "lb.delete", "loadbalancer", lbID.String(), mock.Anything).Return(nil).Once()

		err := svc.Delete(ctx, lbID.String())
		require.NoError(t, err)
	})

	t.Run("RemoveTarget", func(t *testing.T) {
		lbID := uuid.New()
		instID := uuid.New()
		lb := &domain.LoadBalancer{ID: lbID, UserID: userID}
		mockRepo.On("GetByID", mock.Anything, lbID).Return(lb, nil).Once()
		mockRepo.On("RemoveTarget", mock.Anything, lbID, instID).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "lb.target_remove", "loadbalancer", lbID.String(), mock.Anything).Return(nil).Once()

		err := svc.RemoveTarget(ctx, lbID, instID)
		require.NoError(t, err)
	})
}
