package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupIdentityServiceTest(t *testing.T) (*services.IdentityService, *MockRBACService, context.Context) {
	t.Helper()
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewIdentityRepository(db)
	rbacSvc := new(MockRBACService)

	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(services.AuditServiceParams{
		Repo:    auditRepo,
		RBACSvc: rbacSvc,
	})

	svc := services.NewIdentityService(services.IdentityServiceParams{
		Repo:     repo,
		RbacSvc:  rbacSvc,
		AuditSvc: auditSvc,
	})
	return svc, rbacSvc, ctx
}

func TestIdentityService_CreateKey(t *testing.T) {
	// Build context without DB setup — this test uses mocks only.
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockIdentityRepo)
		mockAuditSvc := new(MockAuditService)
		mockRBAC := new(MockRBACService)
		mockSvc := services.NewIdentityService(services.IdentityServiceParams{
			Repo: mockRepo, RbacSvc: mockRBAC, AuditSvc: mockAuditSvc,
		})

		mockRepo.On("CreateAPIKey", mock.Anything, mock.MatchedBy(func(apiKey *domain.APIKey) bool {
			return apiKey.TenantID == tenantID &&
				apiKey.DefaultTenantID != nil && *apiKey.DefaultTenantID == tenantID
		})).Return(nil).Once()
		mockRBAC.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "api_key.create", "api_key", mock.Anything, mock.Anything).Return(nil).Once()

		name := "test-key"
		key, err := mockSvc.CreateKey(ctx, userID, name)
		require.NoError(t, err)
		assert.NotNil(t, key)
		assert.Equal(t, name, key.Name)
		assert.Equal(t, userID, key.UserID)
		assert.Contains(t, key.Key, "thecloud_")
		assert.Equal(t, tenantID, key.TenantID)
		assert.NotNil(t, key.DefaultTenantID)
		assert.Equal(t, tenantID, *key.DefaultTenantID)
		mockRepo.AssertExpectations(t)
	})
}

func TestIdentityService_ValidateAPIKey(t *testing.T) {
	svc, _, ctx := setupIdentityServiceTest(t)
	userID := appcontext.UserIDFromContext(ctx)

	key, _ := svc.CreateKey(ctx, userID, "session")

	t.Run("Valid", func(t *testing.T) {
		validated, err := svc.ValidateAPIKey(ctx, key.Key)
		require.NoError(t, err)
		assert.Equal(t, userID, validated.UserID)
	})

	t.Run("Invalid", func(t *testing.T) {
		_, err := svc.ValidateAPIKey(ctx, "invalid-key")
		require.Error(t, err)
	})
}

func TestIdentityService_RevokeKey(t *testing.T) {
	svc, rbacSvc, ctx := setupIdentityServiceTest(t)
	userID := appcontext.UserIDFromContext(ctx)

	key, _ := svc.CreateKey(ctx, userID, "to-revoke")

	t.Run("Success", func(t *testing.T) {
		rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		err := svc.RevokeKey(ctx, userID, key.ID)
		require.NoError(t, err)

		// Should no longer validate
		_, err = svc.ValidateAPIKey(ctx, key.Key)
		require.Error(t, err)
	})

	t.Run("WrongUser", func(t *testing.T) {
		rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New(errors.Forbidden, "forbidden")).Once()
		otherUserID := uuid.New()
		key2, _ := svc.CreateKey(ctx, userID, "other")

		err := svc.RevokeKey(ctx, otherUserID, key2.ID)
		require.Error(t, err) // Forbidden if context user is not owner/admin
	})
}

func TestIdentityService_RotateKey(t *testing.T) {
	svc, rbacSvc, ctx := setupIdentityServiceTest(t)
	userID := appcontext.UserIDFromContext(ctx)

	oldKey, _ := svc.CreateKey(ctx, userID, "original")

	t.Run("Success", func(t *testing.T) {
		rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		newKey, err := svc.RotateKey(ctx, userID, oldKey.ID)
		require.NoError(t, err)
		assert.NotEqual(t, oldKey.Key, newKey.Key)
		assert.NotEqual(t, oldKey.ID, newKey.ID)

		// Old should be invalid
		_, err = svc.ValidateAPIKey(ctx, oldKey.Key)
		require.Error(t, err)

		// New should be valid
		validated, err := svc.ValidateAPIKey(ctx, newKey.Key)
		require.NoError(t, err)
		assert.Equal(t, userID, validated.UserID)
	})
}
