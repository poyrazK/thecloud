package services_test

import (
	"context"
	"log/slog"
	"testing"

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

	t.Run("RevokeKey", func(t *testing.T) {
		keyID := uuid.New()
		apiKey := &domain.APIKey{ID: keyID, UserID: userID, Name: "old-key"}
		mockRepo.On("GetAPIKeyByID", mock.Anything, keyID).Return(apiKey, nil).Once()
		mockRepo.On("DeleteAPIKey", mock.Anything, keyID).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "api_key.revoke", "api_key", keyID.String(), mock.Anything).Return(nil).Once()

		err := svc.RevokeKey(ctx, userID, keyID)
		require.NoError(t, err)
	})
}
