package services_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStackService_Unit(t *testing.T) {
	mockRepo := new(MockStackRepo)
	mockInstSvc := new(MockInstanceService)
	mockVpcSvc := new(MockVpcService)
	mockVolSvc := new(MockVolumeService)
	mockSnapSvc := new(MockSnapshotService)
	svc := services.NewStackService(mockRepo, mockInstSvc, mockVpcSvc, mockVolSvc, mockSnapSvc, slog.Default())

	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)

	t.Run("CreateStack", func(t *testing.T) {
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		// Allow background updates
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Maybe()
		
		stack, err := svc.CreateStack(ctx, "my-stack", "Resources: {}", nil)
		assert.NoError(t, err)
		assert.NotNil(t, stack)
		assert.Equal(t, "my-stack", stack.Name)
		
		// Give background processing a tiny bit of time
		time.Sleep(10 * time.Millisecond)
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
		assert.NoError(t, err)
		assert.True(t, res.Valid)
	})

	t.Run("ValidateTemplate_Invalid", func(t *testing.T) {
		res, err := svc.ValidateTemplate(ctx, "invalid: yaml: :")
		assert.NoError(t, err)
		assert.False(t, res.Valid)
		assert.NotEmpty(t, res.Errors)
	})
}
