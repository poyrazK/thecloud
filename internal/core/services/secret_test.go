package services_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupSecretServiceIntegrationTest(t *testing.T) (ports.SecretService, ports.SecretRepository, context.Context) {
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewSecretRepository(db)
	rbacSvc := new(MockRBACService)
	rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	eventRepo := postgres.NewEventRepository(db)
	eventSvc := services.NewEventService(eventRepo, rbacSvc, nil, slog.Default())

	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(auditRepo, rbacSvc)

	logger := slog.Default()
	key := "test-key-must-be-32-bytes-long---"
	svc, _ := services.NewSecretService(services.SecretServiceParams{
		Repo:        repo,
		RBACSvc:     rbacSvc,
		EventSvc:    eventSvc,
		AuditSvc:    auditSvc,
		Logger:      logger,
		MasterKey:   key,
		Environment: "test",
	})

	return svc, repo, ctx
}

func TestSecretService_Integration(t *testing.T) {
	svc, _, ctx := setupSecretServiceIntegrationTest(t)

	t.Run("CreateSecret", func(t *testing.T) {
		name := "test-secret"
		value := "secret-value"
		desc := "test description"

		secret, err := svc.CreateSecret(ctx, name, value, desc)
		require.NoError(t, err)
		assert.NotNil(t, secret)
		assert.Equal(t, name, secret.Name)
		assert.Equal(t, desc, secret.Description)
		assert.NotEmpty(t, secret.EncryptedValue)
	})

	t.Run("GetSecret", func(t *testing.T) {
		name := "get-test-secret"
		value := "get-secret-value"
		secret, err := svc.CreateSecret(ctx, name, value, "")
		require.NoError(t, err)

		fetched, err := svc.GetSecret(ctx, secret.ID)
		require.NoError(t, err)
		assert.Equal(t, value, fetched.EncryptedValue)
	})

	t.Run("GetSecretByName", func(t *testing.T) {
		name := "get-by-name-test"
		value := "name-value"
		_, err := svc.CreateSecret(ctx, name, value, "")
		require.NoError(t, err)

		fetched, err := svc.GetSecretByName(ctx, name)
		require.NoError(t, err)
		assert.Equal(t, value, fetched.EncryptedValue)
	})

	t.Run("ListSecrets", func(t *testing.T) {
		_, _ = svc.CreateSecret(ctx, "sec1", "v1", "")
		_, _ = svc.CreateSecret(ctx, "sec2", "v2", "")

		list, err := svc.ListSecrets(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 2)
		// Value should be redacted in list
		assert.Equal(t, "[REDACTED]", list[0].EncryptedValue)
	})

	t.Run("DeleteSecret", func(t *testing.T) {
		secret, _ := svc.CreateSecret(ctx, "to-delete", "val", "")
		err := svc.DeleteSecret(ctx, secret.ID)
		require.NoError(t, err)

		_, err = svc.GetSecret(ctx, secret.ID)
		assert.Error(t, err)
	})

	t.Run("EncryptDecrypt", func(t *testing.T) {
		userID := uuid.New()
		plain := "hello-world"
		cipher, err := svc.Encrypt(ctx, userID, plain)
		require.NoError(t, err)
		assert.NotEmpty(t, cipher)

		decrypted, err := svc.Decrypt(ctx, userID, cipher)
		require.NoError(t, err)
		assert.Equal(t, plain, decrypted)
	})
}
