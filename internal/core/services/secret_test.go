package services_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
)

func setupSecretServiceIntegrationTest(t *testing.T) (ports.SecretService, ports.SecretRepository, context.Context) {
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
	svc, _, ctx := setupSecretServiceIntegrationTest(t)
	userID := appcontext.UserIDFromContext(ctx)

	t.Run("CreateAndGet", func(t *testing.T) {
		name := "MY_SECRET"
		value := "super-secret-value"

		secret, err := svc.CreateSecret(ctx, name, value, "test secret")
		assert.NoError(t, err)
		assert.NotNil(t, secret)
		assert.NotEqual(t, value, secret.EncryptedValue)

		// Get by ID
		fetched, err := svc.GetSecret(ctx, secret.ID)
		assert.NoError(t, err)
		assert.Equal(t, value, fetched.EncryptedValue) // Should be decrypted
		assert.Equal(t, userID, fetched.UserID)

		// Get by Name
		fetchedByName, err := svc.GetSecretByName(ctx, name)
		assert.NoError(t, err)
		assert.Equal(t, value, fetchedByName.EncryptedValue)
	})

	t.Run("List", func(t *testing.T) {
		_, _ = svc.CreateSecret(ctx, "S1", "V1", "")
		_, _ = svc.CreateSecret(ctx, "S2", "V2", "")

		secrets, err := svc.ListSecrets(ctx)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(secrets), 2)
	})

	t.Run("EncryptDecrypt", func(t *testing.T) {
		plain := "hello secret"
		cipher, err := svc.Encrypt(ctx, userID, plain)
		assert.NoError(t, err)
		assert.NotEmpty(t, cipher)
		assert.NotEqual(t, plain, cipher)

		decrypted, err := svc.Decrypt(ctx, userID, cipher)
		assert.NoError(t, err)
		assert.Equal(t, plain, decrypted)
	})

	t.Run("Delete", func(t *testing.T) {
		secret, _ := svc.CreateSecret(ctx, "DEL_ME", "val", "")

		err := svc.DeleteSecret(ctx, secret.ID)
		assert.NoError(t, err)

		_, err = svc.GetSecret(ctx, secret.ID)
		assert.Error(t, err)
	})
}
