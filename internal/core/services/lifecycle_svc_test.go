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

func TestCreateRule(t *testing.T) {
	mockRepo := new(MockLifecycleRepository)
	mockStorageRepo := new(MockStorageRepo)
	svc := services.NewLifecycleService(mockRepo, mockStorageRepo)

	userID := uuid.New()
	bucketName := "test-bucket"

	ctx := appcontext.WithUserID(context.Background(), userID)

	// Success case
	t.Run("Success", func(t *testing.T) {
		mockStorageRepo.On("GetBucket", mock.Anything, bucketName).Return(&domain.Bucket{
			Name:   bucketName,
			UserID: userID,
		}, nil).Once()

		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(r *domain.LifecycleRule) bool {
			return r.BucketName == bucketName && r.Prefix == "logs/" && r.ExpirationDays == 30
		})).Return(nil).Once()

		rule, err := svc.CreateRule(ctx, bucketName, "logs/", 30, true)

		assert.NoError(t, err)
		assert.NotNil(t, rule)
		assert.Equal(t, "logs/", rule.Prefix)

		mockStorageRepo.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	// Forbidden case
	t.Run("Forbidden", func(t *testing.T) {
		otherUserID := uuid.New()
		ctxOther := appcontext.WithUserID(context.Background(), otherUserID)

		mockStorageRepo.On("GetBucket", mock.Anything, bucketName).Return(&domain.Bucket{
			Name:   bucketName,
			UserID: userID,
		}, nil).Once()

		rule, err := svc.CreateRule(ctxOther, bucketName, "logs/", 30, true)

		assert.Error(t, err)
		assert.Nil(t, rule)
		assert.Contains(t, err.Error(), "you don't own this bucket")

		mockStorageRepo.AssertExpectations(t)
	})
}
