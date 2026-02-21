package services_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIdentityServiceCreateKeySuccess(t *testing.T) {
	db := setupDB(t)
	defer db.Close()
	cleanDB(t, db)
	ctx := setupTestUser(t, db)
	userID := appcontext.UserIDFromContext(ctx)

	identityRepo := postgres.NewIdentityRepository(db)
	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(auditRepo)
	svc := services.NewIdentityService(identityRepo, auditSvc, slog.Default())

	key, err := svc.CreateKey(ctx, userID, "Test Key")
	require.NoError(t, err)
	assert.NotNil(t, key)
	assert.Contains(t, key.Key, "thecloud_")
	assert.Equal(t, userID, key.UserID)

	// Verify stored
	fetched, err := identityRepo.GetAPIKeyByID(ctx, key.ID)
	assert.NoError(t, err)
	assert.Equal(t, key.ID, fetched.ID)

	// Verify audit log
	logs, err := auditRepo.ListByUserID(ctx, userID, 10)
	assert.NoError(t, err)
	assert.NotEmpty(t, logs)
	assert.Equal(t, "api_key.create", logs[0].Action)
}

func TestIdentityServiceValidateAPIKeySuccess(t *testing.T) {
	db := setupDB(t)
	defer db.Close()
	cleanDB(t, db)
	ctx := setupTestUser(t, db)
	userID := appcontext.UserIDFromContext(ctx)

	identityRepo := postgres.NewIdentityRepository(db)
	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(auditRepo)
	svc := services.NewIdentityService(identityRepo, auditSvc, slog.Default())

	key, err := svc.CreateKey(ctx, userID, "Val Key")
	require.NoError(t, err)

	result, err := svc.ValidateAPIKey(ctx, key.Key)
	assert.NoError(t, err)
	assert.Equal(t, key.ID, result.ID)
	assert.Equal(t, userID, result.UserID)
}

func TestIdentityServiceValidateAPIKeyNotFound(t *testing.T) {
	db := setupDB(t)
	defer db.Close()
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	identityRepo := postgres.NewIdentityRepository(db)
	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(auditRepo)
	svc := services.NewIdentityService(identityRepo, auditSvc, slog.Default())

	result, err := svc.ValidateAPIKey(ctx, "invalid-key")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestIdentityServiceListKeys(t *testing.T) {
	db := setupDB(t)
	defer db.Close()
	cleanDB(t, db)
	ctx := setupTestUser(t, db)
	userID := appcontext.UserIDFromContext(ctx)

	identityRepo := postgres.NewIdentityRepository(db)
	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(auditRepo)
	svc := services.NewIdentityService(identityRepo, auditSvc, slog.Default())

	_, err := svc.CreateKey(ctx, userID, "Key 1")
	require.NoError(t, err)
	_, err = svc.CreateKey(ctx, userID, "Key 2")
	require.NoError(t, err)

	result, err := svc.ListKeys(ctx, userID)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestIdentityServiceRevokeKey(t *testing.T) {
	db := setupDB(t)
	defer db.Close()
	cleanDB(t, db)
	ctx := setupTestUser(t, db)
	userID := appcontext.UserIDFromContext(ctx)

	identityRepo := postgres.NewIdentityRepository(db)
	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(auditRepo)
	svc := services.NewIdentityService(identityRepo, auditSvc, slog.Default())

	t.Run("Success", func(t *testing.T) {
		key, err := svc.CreateKey(ctx, userID, "To Revoke")
		require.NoError(t, err)

		err = svc.RevokeKey(ctx, userID, key.ID)
		assert.NoError(t, err)

		// Verify deletion
		_, err = identityRepo.GetAPIKeyByID(ctx, key.ID)
		assert.Error(t, err)
	})

	t.Run("WrongUser", func(t *testing.T) {
		// Create a key for another user
		otherUserID := uuid.New()
		otherUser := &domain.User{
			ID:           otherUserID,
			Email:        "other@test.com",
			PasswordHash: "hash",
			Name:         "Other User",
			Role:         "user",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		userRepo := postgres.NewUserRepo(db)
		err := userRepo.Create(context.Background(), otherUser)
		require.NoError(t, err)

		otherKey := &domain.APIKey{
			ID:        uuid.New(),
			UserID:    otherUserID,
			Name:      "Other Key",
			Key:       "thecloud_other",
			CreatedAt: time.Now(),
		}
		err = identityRepo.CreateAPIKey(context.Background(), otherKey)
		require.NoError(t, err)

		err = svc.RevokeKey(ctx, userID, otherKey.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot revoke key owned by another user")
	})
}

func TestIdentityServiceRotateKey(t *testing.T) {
	db := setupDB(t)
	defer db.Close()
	cleanDB(t, db)
	ctx := setupTestUser(t, db)
	userID := appcontext.UserIDFromContext(ctx)

	identityRepo := postgres.NewIdentityRepository(db)
	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(auditRepo)
	svc := services.NewIdentityService(identityRepo, auditSvc, slog.Default())

	key, err := svc.CreateKey(ctx, userID, "To Rotate")
	require.NoError(t, err)

	newKey, err := svc.RotateKey(ctx, userID, key.ID)
	assert.NoError(t, err)
	assert.NotNil(t, newKey)
	assert.NotEqual(t, key.Key, newKey.Key)
	assert.Contains(t, newKey.Name, "(rotated)")

	// Old key should be gone
	_, err = identityRepo.GetAPIKeyByID(ctx, key.ID)
	assert.Error(t, err)

	// New key should exist
	fetched, err := identityRepo.GetAPIKeyByID(ctx, newKey.ID)
	assert.NoError(t, err)
	assert.Equal(t, newKey.ID, fetched.ID)
}
