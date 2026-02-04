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
	t.Parallel()
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

func TestListRules(t *testing.T) {
	t.Parallel()
	mockRepo := new(MockLifecycleRepository)
	mockStorageRepo := new(MockStorageRepo)
	svc := services.NewLifecycleService(mockRepo, mockStorageRepo)

	userID := uuid.New()
	bucketName := "test-bucket"
	ctx := appcontext.WithUserID(context.Background(), userID)

	t.Run("Success", func(t *testing.T) {
		mockStorageRepo.On("GetBucket", ctx, bucketName).Return(&domain.Bucket{
			Name:   bucketName,
			UserID: userID,
		}, nil).Once()

		rules := []*domain.LifecycleRule{{ID: uuid.New(), BucketName: bucketName}}
		mockRepo.On("List", ctx, bucketName).Return(rules, nil).Once()

		result, err := svc.ListRules(ctx, bucketName)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(result))
		mockStorageRepo.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Forbidden", func(t *testing.T) {
		otherUserID := uuid.New()
		ctxOther := appcontext.WithUserID(context.Background(), otherUserID)

		mockStorageRepo.On("GetBucket", ctxOther, bucketName).Return(&domain.Bucket{
			Name:   bucketName,
			UserID: userID,
		}, nil).Once()

		_, err := svc.ListRules(ctxOther, bucketName)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "you don't own this bucket")
	})
}

func TestDeleteRule(t *testing.T) {
	t.Parallel()
	mockRepo := new(MockLifecycleRepository)
	mockStorageRepo := new(MockStorageRepo)
	svc := services.NewLifecycleService(mockRepo, mockStorageRepo)

	userID := uuid.New()
	bucketName := "test-bucket"
	ctx := appcontext.WithUserID(context.Background(), userID)
	ruleID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mockStorageRepo.On("GetBucket", ctx, bucketName).Return(&domain.Bucket{
			Name:   bucketName,
			UserID: userID,
		}, nil).Once()

		mockRepo.On("Get", ctx, ruleID).Return(&domain.LifecycleRule{
			ID:         ruleID,
			BucketName: bucketName,
		}, nil).Once()

		mockRepo.On("Delete", ctx, ruleID).Return(nil).Once()

		err := svc.DeleteRule(ctx, bucketName, ruleID.String())
		assert.NoError(t, err)
		mockStorageRepo.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("WrongBucket", func(t *testing.T) {
		mockStorageRepo.On("GetBucket", ctx, bucketName).Return(&domain.Bucket{
			Name:   bucketName,
			UserID: userID,
		}, nil).Once()

		mockRepo.On("Get", ctx, ruleID).Return(&domain.LifecycleRule{
			ID:         ruleID,
			BucketName: "other-bucket",
		}, nil).Once()

		err := svc.DeleteRule(ctx, bucketName, ruleID.String())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not belong to the specified bucket")
	})

	t.Run("Forbidden", func(t *testing.T) {
		otherUserID := uuid.New()
		ctxOther := appcontext.WithUserID(context.Background(), otherUserID)

		mockStorageRepo.On("GetBucket", ctxOther, bucketName).Return(&domain.Bucket{
			Name:   bucketName,
			UserID: userID,
		}, nil).Once()

		err := svc.DeleteRule(ctxOther, bucketName, ruleID.String())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "you don't own this bucket")
	})
}
