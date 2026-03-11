package services_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupIdentityServiceTest(t *testing.T) (*services.IdentityService, *postgres.IdentityRepository, context.Context) {
	t.Helper()
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewIdentityRepository(db)
	rbacSvc := new(MockRBACService)
	rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(services.AuditServiceParams{
		Repo:    auditRepo,
		RBACSvc: rbacSvc, Logger: slog.Default(),
	})

	svc := services.NewIdentityService(services.IdentityServiceParams{
		Repo:     repo,
		RbacSvc:  rbacSvc,
		AuditSvc: auditSvc, Logger: slog.Default(),
	})
	return svc, repo, ctx
}

func TestIdentityService_CreateKey(t *testing.T) {
	svc, repo, ctx := setupIdentityServiceTest(t)
	userID := appcontext.UserIDFromContext(ctx)

	t.Run("Success", func(t *testing.T) {
		name := "test-key"
		key, err := svc.CreateKey(ctx, userID, name)
		require.NoError(t, err)
		assert.NotNil(t, key)
		assert.Equal(t, name, key.Name)
		assert.Equal(t, userID, key.UserID)
		assert.Contains(t, key.Key, "thecloud_")

		// Verify in DB
		dbKey, err := repo.GetAPIKeyByID(ctx, key.ID)
		require.NoError(t, err)
		assert.Equal(t, key.Key, dbKey.Key)
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
	svc, _, ctx := setupIdentityServiceTest(t)
	userID := appcontext.UserIDFromContext(ctx)

	key, _ := svc.CreateKey(ctx, userID, "to-revoke")

	t.Run("Success", func(t *testing.T) {
		err := svc.RevokeKey(ctx, userID, key.ID)
		require.NoError(t, err)

		// Should no longer validate
		_, err = svc.ValidateAPIKey(ctx, key.Key)
		require.Error(t, err)
	})

	t.Run("WrongUser", func(t *testing.T) {
		otherUserID := uuid.New()
		key2, _ := svc.CreateKey(ctx, userID, "other")

		err := svc.RevokeKey(ctx, otherUserID, key2.ID)
		require.Error(t, err) // Forbidden if context user is not owner/admin
	})
}

func TestIdentityService_RotateKey(t *testing.T) {
	svc, _, ctx := setupIdentityServiceTest(t)
	userID := appcontext.UserIDFromContext(ctx)

	oldKey, _ := svc.CreateKey(ctx, userID, "original")

	t.Run("Success", func(t *testing.T) {
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
