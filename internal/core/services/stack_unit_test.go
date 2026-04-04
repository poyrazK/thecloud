package services_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestStackServiceUnit(t *testing.T) {
	mockRepo := new(MockStackRepo)
	mockInstSvc := new(MockInstanceService)
	mockVpcSvc := new(MockVpcService)
	mockVolSvc := new(MockVolumeService)
	mockSnapSvc := new(MockSnapshotService)
	rbacSvc := new(MockRBACService)
	rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	svc := services.NewStackService(mockRepo, rbacSvc, mockInstSvc, mockVpcSvc, mockVolSvc, mockSnapSvc, slog.Default())

	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)

	t.Run("CreateStack", func(t *testing.T) {
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		// Allow background updates
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Maybe()

		stack, err := svc.CreateStack(ctx, "my-stack", "Resources: {}", nil)
		require.NoError(t, err)
		assert.NotNil(t, stack)
		assert.Equal(t, "my-stack", stack.Name)

		// Give background processing a tiny bit of time
		time.Sleep(10 * time.Millisecond)
	})

	t.Run("GetStack", func(t *testing.T) {
		id := uuid.New()
		mockRepo.On("GetByID", mock.Anything, id).Return(&domain.Stack{ID: id}, nil).Once()
		res, err := svc.GetStack(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, id, res.ID)
	})

	t.Run("ListStacks", func(t *testing.T) {
		mockRepo.On("ListByUserID", mock.Anything, userID).Return([]*domain.Stack{{ID: uuid.New()}}, nil).Once()
		res, err := svc.ListStacks(ctx)
		require.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("DeleteStack", func(t *testing.T) {
		id := uuid.New()
		stack := &domain.Stack{ID: id, UserID: userID}
		mockRepo.On("GetByID", mock.Anything, id).Return(stack, nil).Once()
		mockRepo.On("ListResources", mock.Anything, id).Return([]domain.StackResource{}, nil).Once()
		mockRepo.On("Delete", mock.Anything, id).Return(nil).Once()

		err := svc.DeleteStack(ctx, id)
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond) // wait for goroutine
	})

	t.Run("ValidateTemplate_Valid", func(t *testing.T) {
		template := `
Resources:
  MyVPC:
    Type: VPC
    Properties:
      CIDRBlock: 10.0.0.0/16
`
		res, err := svc.ValidateTemplate(ctx, template)
		require.NoError(t, err)
		assert.True(t, res.Valid)
	})

	t.Run("ValidateTemplate_Invalid", func(t *testing.T) {
		res, err := svc.ValidateTemplate(ctx, "invalid: yaml: :")
		require.NoError(t, err)
		assert.False(t, res.Valid)
		assert.NotEmpty(t, res.Errors)
	})

	t.Run("ValidateTemplate_Empty", func(t *testing.T) {
		res, err := svc.ValidateTemplate(ctx, "Resources: {}")
		require.NoError(t, err)
		assert.False(t, res.Valid)
		assert.Contains(t, res.Errors[0], "at least one resource")
	})
}
