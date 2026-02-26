package services_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSecretServiceIntegrationTest(t *testing.T) (ports.SecretService, ports.SecretRepository, context.Context) {
	t.Helper()
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewSecretRepository(db)
	eventRepo := postgres.NewEventRepository(db)
	auditRepo := postgres.NewAuditRepository(db)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	auditSvc := services.NewAuditService(auditRepo)
	eventSvc := services.NewEventService(eventRepo, nil, logger)

	key := "test-key-must-be-32-bytes-long---"
	svc := services.NewSecretService(repo, eventSvc, auditSvc, logger, key, "test")

	return svc, repo, ctx
}

func TestSecretService_Integration(t *testing.T) {
	svc, repo, ctx := setupSecretServiceIntegrationTest(t)
	userID := appcontext.UserIDFromContext(ctx)

	t.Run("CreateAndGet", func(t *testing.T) {
		name := "MY_SECRET"
		value := "super-secret-value"

		secret, err := svc.CreateSecret(ctx, name, value, "test secret")
		require.NoError(t, err)
		assert.NotNil(t, secret)
		assert.NotEqual(t, value, secret.EncryptedValue)

		// Get by ID
		fetched, err := svc.GetSecret(ctx, secret.ID)
		require.NoError(t, err)
		assert.Equal(t, secret.ID, fetched.ID)
		assert.Equal(t, value, fetched.EncryptedValue) // Should be decrypted into plaintext field
		assert.Equal(t, userID, fetched.UserID)

		// Get by Name
		fetchedByName, err := svc.GetSecretByName(ctx, name)
		require.NoError(t, err)
		assert.Equal(t, value, fetchedByName.EncryptedValue)
	})

	t.Run("DuplicateName", func(t *testing.T) {
		_, err := svc.CreateSecret(ctx, "DUP", "v1", "")
		require.NoError(t, err)
		_, err = svc.CreateSecret(ctx, "DUP", "v2", "")
		require.Error(t, err)
	})

	t.Run("GetNotFound", func(t *testing.T) {
		_, err := svc.GetSecret(ctx, uuid.New())
		require.Error(t, err)

		_, err = svc.GetSecretByName(ctx, "NON_EXISTENT")
		require.Error(t, err)
	})

	t.Run("List", func(t *testing.T) {
		_, _ = svc.CreateSecret(ctx, "S1", "V1", "")
		_, _ = svc.CreateSecret(ctx, "S2", "V2", "")

		secrets, err := svc.ListSecrets(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(secrets), 2)
		// Ensure redaction
		assert.Equal(t, "[REDACTED]", secrets[0].EncryptedValue)
	})

	t.Run("EncryptDecrypt", func(t *testing.T) {
		plain := "hello secret"
		cipher, err := svc.Encrypt(ctx, userID, plain)
		require.NoError(t, err)
		assert.NotEmpty(t, cipher)
		assert.NotEqual(t, plain, cipher)

		decrypted, err := svc.Decrypt(ctx, userID, cipher)
		require.NoError(t, err)
		assert.Equal(t, plain, decrypted)
	})

	t.Run("Delete", func(t *testing.T) {
		secret, _ := svc.CreateSecret(ctx, "DEL_ME", "val", "")

		err := svc.DeleteSecret(ctx, secret.ID)
		require.NoError(t, err)

		_, err = svc.GetSecret(ctx, secret.ID)
		require.Error(t, err)

		// Delete again - should fail because it's not found
		err = svc.DeleteSecret(ctx, secret.ID)
		require.Error(t, err)
	})

	t.Run("DecryptionFailure", func(t *testing.T) {
		secret := &domain.Secret{
			ID:             uuid.New(),
			UserID:         userID,
			Name:           "corrupt",
			EncryptedValue: "not-a-valid-cipher",
		}
		_ = repo.Create(ctx, secret)

		_, err := svc.GetSecret(ctx, secret.ID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decrypt")
	})
}

func TestNewSecretService_Variants(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("DefaultKeyDevelopment", func(t *testing.T) {
		svc := services.NewSecretService(nil, nil, nil, logger, "", "development")
		assert.NotNil(t, svc)
	})

	t.Run("ShortKeyWarning", func(t *testing.T) {
		svc := services.NewSecretService(nil, nil, nil, logger, "short", "development")
		assert.NotNil(t, svc)
	})
}
