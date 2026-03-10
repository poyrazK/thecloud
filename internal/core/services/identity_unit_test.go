package services_test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestIdentityService_Unit(t *testing.T) {
	mockRepo := new(MockIdentityRepo)
	mockAuditSvc := new(MockAuditService)
	svc := services.NewIdentityService(mockRepo, mockAuditSvc, slog.Default())

	ctx := context.Background()
	userID := uuid.New()

	t.Run("CreateKey", func(t *testing.T) {
		mockRepo.On("CreateAPIKey", mock.Anything, mock.Anything).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "api_key.create", "api_key", mock.Anything, mock.Anything).Return(nil).Once()

		key, err := svc.CreateKey(ctx, userID, "my-key")
		require.NoError(t, err)
		assert.NotNil(t, key)
		assert.Contains(t, key.Key, "thecloud_")
		mockRepo.AssertExpectations(t)
	})

	t.Run("ValidateAPIKey", func(t *testing.T) {
		keyStr := "thecloud_123"
		apiKey := &domain.APIKey{Key: keyStr, UserID: userID}
		mockRepo.On("GetAPIKeyByKey", mock.Anything, keyStr).Return(apiKey, nil).Once()

		res, err := svc.ValidateAPIKey(ctx, keyStr)
		require.NoError(t, err)
		assert.Equal(t, userID, res.UserID)
	})

	t.Run("ValidateAPIKey_Expired", func(t *testing.T) {
		keyStr := "expired"
		past := time.Now().Add(-1 * time.Hour)
		apiKey := &domain.APIKey{Key: keyStr, ExpiresAt: &past}
		mockRepo.On("GetAPIKeyByKey", mock.Anything, keyStr).Return(apiKey, nil).Once()

		_, err := svc.ValidateAPIKey(ctx, keyStr)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})

	t.Run("ListKeys", func(t *testing.T) {
		mockRepo.On("ListAPIKeysByUserID", mock.Anything, userID).Return([]*domain.APIKey{{ID: uuid.New()}}, nil).Once()
		res, err := svc.ListKeys(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("RevokeKey", func(t *testing.T) {
		keyID := uuid.New()
		apiKey := &domain.APIKey{ID: keyID, UserID: userID, Name: "old-key"}
		mockRepo.On("GetAPIKeyByID", mock.Anything, keyID).Return(apiKey, nil).Once()
		mockRepo.On("DeleteAPIKey", mock.Anything, keyID).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "api_key.revoke", "api_key", keyID.String(), mock.Anything).Return(nil).Once()

		err := svc.RevokeKey(ctx, userID, keyID)
		require.NoError(t, err)
	})

	t.Run("RevokeKey_Unauthorized", func(t *testing.T) {
		keyID := uuid.New()
		apiKey := &domain.APIKey{ID: keyID, UserID: uuid.New()} // different user
		mockRepo.On("GetAPIKeyByID", mock.Anything, keyID).Return(apiKey, nil).Once()

		err := svc.RevokeKey(ctx, userID, keyID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("RotateKey", func(t *testing.T) {
		keyID := uuid.New()
		apiKey := &domain.APIKey{ID: keyID, UserID: userID, Name: "rotate-me"}
		mockRepo.On("GetAPIKeyByID", mock.Anything, keyID).Return(apiKey, nil).Once()
		mockRepo.On("CreateAPIKey", mock.Anything, mock.Anything).Return(nil).Once()
		mockRepo.On("DeleteAPIKey", mock.Anything, keyID).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "api_key.create", "api_key", mock.Anything, mock.Anything).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "api_key.rotate", "api_key", keyID.String(), mock.Anything).Return(nil).Once()

		newKey, err := svc.RotateKey(ctx, userID, keyID)
		require.NoError(t, err)
		assert.NotNil(t, newKey)
		assert.NotEqual(t, keyID, newKey.ID)
	})

	t.Run("RotateKey_DeleteFail", func(t *testing.T) {
		keyID := uuid.New()
		apiKey := &domain.APIKey{ID: keyID, UserID: userID, Name: "rotate-me"}
		mockRepo.On("GetAPIKeyByID", mock.Anything, keyID).Return(apiKey, nil).Once()
		mockRepo.On("CreateAPIKey", mock.Anything, mock.Anything).Return(nil).Once()
		mockRepo.On("DeleteAPIKey", mock.Anything, keyID).Return(fmt.Errorf("delete fail")).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "api_key.create", "api_key", mock.Anything, mock.Anything).Return(nil).Once()

		// Should still return new key
		newKey, err := svc.RotateKey(ctx, userID, keyID)
		require.NoError(t, err)
		assert.NotNil(t, newKey)
	})
}
