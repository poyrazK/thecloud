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

func TestLifecycleService_Unit(t *testing.T) {
	mockRepo := new(MockLifecycleRepository)
	mockStorageRepo := new(MockStorageRepo)
	svc := services.NewLifecycleService(mockRepo, mockStorageRepo)

	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)

	t.Run("CreateRule_Success", func(t *testing.T) {
		bucket := &domain.Bucket{Name: "my-bucket", UserID: userID}
		mockStorageRepo.On("GetBucket", mock.Anything, "my-bucket").Return(bucket, nil).Once()
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()

		rule, err := svc.CreateRule(ctx, "my-bucket", "logs/", 30, true)
		assert.NoError(t, err)
		assert.NotNil(t, rule)
		assert.Equal(t, 30, rule.ExpirationDays)
		mockRepo.AssertExpectations(t)
	})

	t.Run("CreateRule_Forbidden", func(t *testing.T) {
		bucket := &domain.Bucket{Name: "other-bucket", UserID: uuid.New()} // Different owner
		mockStorageRepo.On("GetBucket", mock.Anything, "other-bucket").Return(bucket, nil).Once()

		rule, err := svc.CreateRule(ctx, "other-bucket", "", 10, true)
		assert.Error(t, err)
		assert.Nil(t, rule)
		// Fix: Check for uppercase "FORBIDDEN" as returned by the internal/errors package
		assert.Contains(t, err.Error(), "FORBIDDEN")
	})
}
