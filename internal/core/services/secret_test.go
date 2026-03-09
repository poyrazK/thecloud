package services_test

import (
	"context"
	"log/slog"
	"testing"

	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupSecretServiceIntegrationTest(t *testing.T) (ports.SecretService, ports.SecretRepository, context.Context) {
	t.Helper()
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewSecretRepository(db)
	rbacSvc := new(MockRBACService)
	rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	eventRepo := postgres.NewEventRepository(db)
	eventSvc := services.NewEventService(services.EventServiceParams{
		Repo:    eventRepo,
		RBACSvc: rbacSvc,
	})
	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(services.AuditServiceParams{
		Repo:    auditRepo,
		RBACSvc: rbacSvc,
	})

	key := "test-master-key-32-chars-long-!!!"
	logger := slog.Default()

	svc, err := services.NewSecretService(services.SecretServiceParams{
		Repo:        repo,
		RBACSvc:     rbacSvc,
		EventSvc:    eventSvc,
		AuditSvc:    auditSvc,
		Logger:      logger,
		MasterKey:   key,
		Environment: "test",
	})
	require.NoError(t, err)

	return svc, repo, ctx
}

func TestSecretService_Integration(t *testing.T) {
	svc, repo, ctx := setupSecretServiceIntegrationTest(t)
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	t.Run("CreateAndRetrieve", func(t *testing.T) {
		name := "db-password"
		value := "super-secret-123"

		sec, err := svc.CreateSecret(ctx, name, value, "password")
		require.NoError(t, err)
		assert.Equal(t, name, sec.Name)
		assert.Equal(t, userID, sec.UserID)
		assert.Equal(t, tenantID, sec.TenantID)

		// Retrieve
		fetched, err := svc.GetSecret(ctx, sec.ID)
		require.NoError(t, err)
		assert.Equal(t, value, fetched.EncryptedValue)

		// Get by Name
		fetchedByName, err := svc.GetSecretByName(ctx, name)
		require.NoError(t, err)
		assert.Equal(t, value, fetchedByName.EncryptedValue)
	})

	t.Run("ListSecrets", func(t *testing.T) {
		_, _ = svc.CreateSecret(ctx, "s1", "v1", "test")
		_, _ = svc.CreateSecret(ctx, "s2", "v2", "test")

		secrets, err := svc.ListSecrets(ctx)
		require.NoError(t, err)
		// Including previous test secret, should be 3
		assert.GreaterOrEqual(t, len(secrets), 2)
	})

	t.Run("Delete", func(t *testing.T) {
		sec, _ := svc.CreateSecret(ctx, "to-delete", "val", "test")

		err := svc.DeleteSecret(ctx, sec.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, sec.ID)
		require.Error(t, err)
	})
}
