package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
)

func TestInstanceTypeService_List(t *testing.T) {
	mockRepo := new(MockInstanceTypeRepo)
	service := services.NewInstanceTypeService(mockRepo)

	ctx := context.Background()
	expectedTypes := []*domain.InstanceType{
		{ID: "t2.micro", Name: "Micro Instance", VCPUs: 1, MemoryMB: 1024},
		{ID: "t2.small", Name: "Small Instance", VCPUs: 1, MemoryMB: 2048},
	}

	t.Run("success", func(t *testing.T) {
		mockRepo.On("List", ctx).Return(expectedTypes, nil).Once()

		result, err := service.List(ctx)

		assert.NoError(t, err)
		assert.Equal(t, expectedTypes, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		repoErr := errors.New("database error")
		mockRepo.On("List", ctx).Return(nil, repoErr).Once()

		result, err := service.List(ctx)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, repoErr, err)
		mockRepo.AssertExpectations(t)
	})
}
