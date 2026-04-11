package services_test

import (
	"context"
	"fmt"
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

func TestIdentityService_Unit(t *testing.T) {
	mockRepo := new(MockIdentityRepo)
	mockAuditSvc := new(MockAuditService)
	rbacSvc := new(MockRBACService)
	svc := services.NewIdentityService(services.IdentityServiceParams{
		Repo: mockRepo, RbacSvc: rbacSvc, AuditSvc: mockAuditSvc,
	})

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	t.Run("CreateKey", func(t *testing.T) {
		mockRepo.On("CreateAPIKey", mock.Anything, mock.MatchedBy(func(apiKey *domain.APIKey) bool {
			return apiKey.TenantID == tenantID &&
				apiKey.DefaultTenantID != nil && *apiKey.DefaultTenantID == tenantID
		})).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "api_key.create", "api_key", mock.Anything, mock.Anything).Return(nil).Once()

		key, err := svc.CreateKey(ctx, userID, "my-key")
		require.NoError(t, err)
		assert.NotNil(t, key)
		assert.Contains(t, key.Key, "thecloud_")
		assert.Equal(t, tenantID, key.TenantID)
		assert.NotNil(t, key.DefaultTenantID)
		assert.Equal(t, tenantID, *key.DefaultTenantID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("CreateKey_NoTenant", func(t *testing.T) {
		noTenantCtx := appcontext.WithUserID(context.Background(), userID)
		noTenantRepo := new(MockIdentityRepo)
		noTenantAuditSvc := new(MockAuditService)
		noTenantRBAC := new(MockRBACService)
		noTenantSvc := services.NewIdentityService(services.IdentityServiceParams{
			Repo: noTenantRepo, AuditSvc: noTenantAuditSvc, RbacSvc: noTenantRBAC,
		})

		noTenantRepo.On("CreateAPIKey", mock.Anything, mock.MatchedBy(func(apiKey *domain.APIKey) bool {
			return apiKey.TenantID == uuid.Nil && apiKey.DefaultTenantID == nil
		})).Return(nil).Once()
		noTenantAuditSvc.On("Log", mock.Anything, userID, "api_key.create", "api_key", mock.Anything, mock.Anything).Return(nil).Once()

		key, err := noTenantSvc.CreateKey(noTenantCtx, userID, "no-tenant-key")
		require.NoError(t, err)
		assert.NotNil(t, key)
		assert.Equal(t, uuid.Nil, key.TenantID)
		assert.Nil(t, key.DefaultTenantID)
		noTenantRepo.AssertExpectations(t)
	})

	t.Run("ValidateAPIKey", func(t *testing.T) {
		keyStr := "thecloud_123"
		apiKey := &domain.APIKey{Key: keyStr, UserID: userID}
		mockRepo.On("GetAPIKeyByHash", mock.Anything, mock.Anything).Return(apiKey, nil).Once()

		res, err := svc.ValidateAPIKey(ctx, keyStr)
		require.NoError(t, err)
		assert.Equal(t, userID, res.UserID)
	})

	t.Run("ValidateAPIKey_Expired", func(t *testing.T) {
		keyStr := "expired"
		past := time.Now().Add(-1 * time.Hour)
		apiKey := &domain.APIKey{Key: keyStr, ExpiresAt: &past}
		mockRepo.On("GetAPIKeyByHash", mock.Anything, mock.Anything).Return(apiKey, nil).Once()

		_, err := svc.ValidateAPIKey(ctx, keyStr)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})

	t.Run("ListKeys", func(t *testing.T) {
		rbacSvc.On("Authorize", mock.Anything, userID, tenantID, domain.PermissionIdentityRead, "*").Return(nil).Once()
		mockRepo.On("ListAPIKeysByUserID", mock.Anything, userID).Return([]*domain.APIKey{{ID: uuid.New()}}, nil).Once()
		res, err := svc.ListKeys(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("RevokeKey", func(t *testing.T) {
		keyID := uuid.New()
		apiKey := &domain.APIKey{ID: keyID, UserID: userID, Name: "old-key"}
		rbacSvc.On("Authorize", mock.Anything, userID, tenantID, domain.PermissionIdentityDelete, keyID.String()).Return(nil).Once()
		mockRepo.On("GetAPIKeyByID", mock.Anything, keyID).Return(apiKey, nil).Once()
		mockRepo.On("DeleteAPIKey", mock.Anything, keyID).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "api_key.revoke", "api_key", keyID.String(), mock.Anything).Return(nil).Once()

		err := svc.RevokeKey(ctx, userID, keyID)
		require.NoError(t, err)
	})

	t.Run("RevokeKey_Unauthorized", func(t *testing.T) {
		keyID := uuid.New()
		apiKey := &domain.APIKey{ID: keyID, UserID: uuid.New()} // different user
		unauthRepo := new(MockIdentityRepo)
		unauthAuditSvc := new(MockAuditService)
		unauthRBAC := new(MockRBACService)
		unauthRBAC.On("Authorize", mock.Anything, userID, tenantID, domain.PermissionIdentityDelete, keyID.String()).Return(fmt.Errorf("forbidden")).Once()
		unauthRepo.On("GetAPIKeyByID", mock.Anything, keyID).Return(apiKey, nil).Once()
		unauthSvc := services.NewIdentityService(services.IdentityServiceParams{
			Repo: unauthRepo, AuditSvc: unauthAuditSvc, RbacSvc: unauthRBAC, Logger: slog.Default(),
		})

		err := unauthSvc.RevokeKey(ctx, userID, keyID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
		unauthRBAC.AssertExpectations(t)
		unauthRepo.AssertExpectations(t)
		unauthAuditSvc.AssertExpectations(t)
	})

	t.Run("RotateKey", func(t *testing.T) {
		keyID := uuid.New()
		apiKey := &domain.APIKey{ID: keyID, UserID: userID, Name: "rotate-me"}
		rbacSvc.On("Authorize", mock.Anything, userID, tenantID, domain.PermissionIdentityDelete, keyID.String()).Return(nil).Once()
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
		var createdKeyID uuid.UUID
		rbacSvc.On("Authorize", mock.Anything, userID, tenantID, domain.PermissionIdentityDelete, keyID.String()).Return(nil).Once()
		mockRepo.On("GetAPIKeyByID", mock.Anything, keyID).Return(apiKey, nil).Once()
		mockRepo.On("CreateAPIKey", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			createdKeyID = args.Get(1).(*domain.APIKey).ID
		}).Return(nil).Once()
		mockRepo.On("DeleteAPIKey", mock.Anything, keyID).Return(fmt.Errorf("delete fail")).Once()
		mockRepo.On("DeleteAPIKey", mock.Anything, mock.MatchedBy(func(id uuid.UUID) bool {
			return createdKeyID != uuid.Nil && id == createdKeyID
		})).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "api_key.create", "api_key", mock.Anything, mock.Anything).Return(nil).Once()

		newKey, err := svc.RotateKey(ctx, userID, keyID)
		require.Error(t, err)
		assert.Nil(t, newKey)
		assert.Contains(t, err.Error(), "failed to delete old key")
	})
}
