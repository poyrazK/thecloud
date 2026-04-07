package services_test

import (
	"context"
	"testing"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestInstanceTypeService_Unit(t *testing.T) {
	mockRepo := new(MockInstanceTypeRepo)
	rbacSvc := new(MockRBACService)
	rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	svc := services.NewInstanceTypeService(mockRepo, rbacSvc)

	t.Run("List", func(t *testing.T) {
		expected := []*domain.InstanceType{{ID: "t2.micro"}}
		mockRepo.On("List", mock.Anything).Return(expected, nil).Once()

		res, err := svc.List(context.Background())
		require.NoError(t, err)
		assert.Equal(t, expected, res)
		mockRepo.AssertExpectations(t)
	})
}
