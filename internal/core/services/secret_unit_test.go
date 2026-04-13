package services_test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	domain "github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSecretService_Unit(t *testing.T) {
	mockRepo := new(MockSecretRepo)
	mockRBAC := new(MockRBACService)
	mockRBAC.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockEvent := new(MockEventService)
	mockAudit := new(MockAuditService)
	mockAudit.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockEvent.On("RecordEvent", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	svc, err := services.NewSecretService(services.SecretServiceParams{
		Repo:        mockRepo,
		RBACSvc:     mockRBAC,
		EventSvc:    mockEvent,
		AuditSvc:    mockAudit,
		Logger:      slog.Default(),
		MasterKey:   "test-master-key-32-chars-long-!!!",
		Environment: "test",
	})
	require.NoError(t, err)

	ctx := context.Background()
	tenantID := uuid.New()
	userID := uuid.New()
	ctx = appcontext.WithTenantID(ctx, tenantID)
	ctx = appcontext.WithUserID(ctx, userID)

	// Set up Create mock before computing validEncrypted (consumed here)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
	createdSecret, err := svc.CreateSecret(ctx, "temp", "tempvalue", "tempdesc")
	require.NoError(t, err)
	validEncrypted := createdSecret.EncryptedValue

	t.Run("CreateSecret_Success", func(t *testing.T) {
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()

		secret, err := svc.CreateSecret(ctx, "db-password", "supersecret", "my db password")
		require.NoError(t, err)
		assert.NotNil(t, secret)
		assert.Equal(t, "db-password", secret.Name)
		assert.Equal(t, userID, secret.UserID)
		assert.Equal(t, tenantID, secret.TenantID)
		assert.NotEmpty(t, secret.EncryptedValue)
	})

	t.Run("CreateSecret_RepoError", func(t *testing.T) {
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(fmt.Errorf("db error")).Once()

		_, err := svc.CreateSecret(ctx, "db-password", "supersecret", "desc")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})

	t.Run("GetSecret_Success", func(t *testing.T) {
		id := uuid.New()
		secret := &domain.Secret{
			ID:             id,
			UserID:         userID,
			TenantID:       tenantID,
			Name:           "my-secret",
			EncryptedValue: validEncrypted,
		}
		mockRepo.On("GetByID", mock.Anything, id).Return(secret, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()

		result, err := svc.GetSecret(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, id, result.ID)
		assert.Equal(t, "tempvalue", result.EncryptedValue) // decrypted value
	})

	t.Run("GetSecret_NotFound", func(t *testing.T) {
		id := uuid.New()
		mockRepo.On("GetByID", mock.Anything, id).Return(nil, errors.New(errors.NotFound, "not found")).Once()

		_, err := svc.GetSecret(ctx, id)
		require.Error(t, err)
	})

	t.Run("GetSecret_TenantMismatch", func(t *testing.T) {
		id := uuid.New()
		otherTenantID := uuid.New()
		secret := &domain.Secret{
			ID:       id,
			UserID:   userID,
			TenantID: otherTenantID,
			Name:     "other-tenant-secret",
		}
		mockRepo.On("GetByID", mock.Anything, id).Return(secret, nil).Once()

		_, err := svc.GetSecret(ctx, id)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("GetSecret_RepoError", func(t *testing.T) {
		id := uuid.New()
		mockRepo.On("GetByID", mock.Anything, id).Return(nil, fmt.Errorf("db error")).Once()

		_, err := svc.GetSecret(ctx, id)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})

	t.Run("GetSecretByName_Success", func(t *testing.T) {
		secret := &domain.Secret{
			ID:             uuid.New(),
			UserID:         userID,
			TenantID:       tenantID,
			Name:           "named-secret",
			EncryptedValue: validEncrypted,
		}
		mockRepo.On("GetByName", mock.Anything, "named-secret").Return(secret, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()

		result, err := svc.GetSecretByName(ctx, "named-secret")
		require.NoError(t, err)
		assert.Equal(t, secret.ID, result.ID)
		assert.Equal(t, "tempvalue", result.EncryptedValue)
	})

	t.Run("GetSecretByName_NotFound", func(t *testing.T) {
		mockRepo.On("GetByName", mock.Anything, "unknown").Return(nil, errors.New(errors.NotFound, "not found")).Once()

		_, err := svc.GetSecretByName(ctx, "unknown")
		require.Error(t, err)
	})

	t.Run("GetSecretByName_TenantMismatch", func(t *testing.T) {
		otherTenantID := uuid.New()
		secret := &domain.Secret{
			ID:       uuid.New(),
			UserID:   userID,
			TenantID: otherTenantID,
			Name:     "other-secret",
		}
		mockRepo.On("GetByName", mock.Anything, "other-secret").Return(secret, nil).Once()

		_, err := svc.GetSecretByName(ctx, "other-secret")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("GetSecretByName_RepoError", func(t *testing.T) {
		mockRepo.On("GetByName", mock.Anything, "some-name").Return(nil, fmt.Errorf("db error")).Once()

		_, err := svc.GetSecretByName(ctx, "some-name")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})

	t.Run("ListSecrets_Success", func(t *testing.T) {
		secrets := []*domain.Secret{
			{ID: uuid.New(), TenantID: tenantID, Name: "s1"},
			{ID: uuid.New(), TenantID: tenantID, Name: "s2"},
		}
		mockRepo.On("List", mock.Anything).Return(secrets, nil).Once()

		result, err := svc.ListSecrets(ctx)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		for _, s := range result {
			assert.Equal(t, "[REDACTED]", s.EncryptedValue)
		}
	})

	t.Run("ListSecrets_FiltersOtherTenant", func(t *testing.T) {
		otherTenantID := uuid.New()
		secrets := []*domain.Secret{
			{ID: uuid.New(), TenantID: tenantID, Name: "mine"},
			{ID: uuid.New(), TenantID: otherTenantID, Name: "other"},
		}
		mockRepo.On("List", mock.Anything).Return(secrets, nil).Once()

		result, err := svc.ListSecrets(ctx)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "mine", result[0].Name)
	})

	t.Run("ListSecrets_RepoError", func(t *testing.T) {
		mockRepo.On("List", mock.Anything).Return(nil, fmt.Errorf("db error")).Once()

		_, err := svc.ListSecrets(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})

	t.Run("DeleteSecret_Success", func(t *testing.T) {
		id := uuid.New()
		secret := &domain.Secret{ID: id, TenantID: tenantID, Name: "to-delete"}
		mockRepo.On("GetByID", mock.Anything, id).Return(secret, nil).Once()
		mockRepo.On("Delete", mock.Anything, id).Return(nil).Once()

		err := svc.DeleteSecret(ctx, id)
		require.NoError(t, err)
	})

	t.Run("DeleteSecret_NotFound", func(t *testing.T) {
		id := uuid.New()
		mockRepo.On("GetByID", mock.Anything, id).Return(nil, errors.New(errors.NotFound, "not found")).Once()

		err := svc.DeleteSecret(ctx, id)
		require.Error(t, err)
	})

	t.Run("DeleteSecret_TenantMismatch", func(t *testing.T) {
		id := uuid.New()
		otherTenantID := uuid.New()
		secret := &domain.Secret{ID: id, TenantID: otherTenantID}
		mockRepo.On("GetByID", mock.Anything, id).Return(secret, nil).Once()

		err := svc.DeleteSecret(ctx, id)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot delete")
	})

	t.Run("DeleteSecret_RepoDeleteError", func(t *testing.T) {
		id := uuid.New()
		secret := &domain.Secret{ID: id, TenantID: tenantID}
		mockRepo.On("GetByID", mock.Anything, id).Return(secret, nil).Once()
		mockRepo.On("Delete", mock.Anything, id).Return(fmt.Errorf("db error")).Once()

		err := svc.DeleteSecret(ctx, id)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})

	t.Run("Encrypt_Success", func(t *testing.T) {
		plainText := "hello world"
		result, err := svc.Encrypt(ctx, userID, plainText)
		require.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.NotEqual(t, plainText, result)
	})

	t.Run("Decrypt_Success", func(t *testing.T) {
		plainText := "hello world"
		encrypted, _ := svc.Encrypt(ctx, userID, plainText)

		result, err := svc.Decrypt(ctx, userID, encrypted)
		require.NoError(t, err)
		assert.Equal(t, plainText, result)
	})

	t.Run("Decrypt_InvalidCipherText", func(t *testing.T) {
		_, err := svc.Decrypt(ctx, userID, "not-valid-base64-encrypted")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "decrypt")
	})
}

func TestNewSecretService_Errors(t *testing.T) {
	logger := slog.Default()
	masterKey := "test-master-key-32-chars-long-!!!"

	t.Run("NilLogger", func(t *testing.T) {
		_, err := services.NewSecretService(services.SecretServiceParams{
			Logger:    nil,
			Repo:      new(MockSecretRepo),
			RBACSvc:   new(MockRBACService),
			EventSvc:  new(MockEventService),
			AuditSvc:  new(MockAuditService),
			MasterKey: masterKey,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "logger is required")
	})

	t.Run("NilRepo", func(t *testing.T) {
		_, err := services.NewSecretService(services.SecretServiceParams{
			Logger:    logger,
			Repo:      nil,
			RBACSvc:   new(MockRBACService),
			EventSvc:  new(MockEventService),
			AuditSvc:  new(MockAuditService),
			MasterKey: masterKey,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "repository is required")
	})

	t.Run("NilRBAC", func(t *testing.T) {
		_, err := services.NewSecretService(services.SecretServiceParams{
			Logger:    logger,
			Repo:      new(MockSecretRepo),
			RBACSvc:   nil,
			EventSvc:  new(MockEventService),
			AuditSvc:  new(MockAuditService),
			MasterKey: masterKey,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "rbac service is required")
	})

	t.Run("NilEventSvc", func(t *testing.T) {
		_, err := services.NewSecretService(services.SecretServiceParams{
			Logger:    logger,
			Repo:      new(MockSecretRepo),
			RBACSvc:   new(MockRBACService),
			EventSvc:  nil,
			AuditSvc:  new(MockAuditService),
			MasterKey: masterKey,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "event service is required")
	})

	t.Run("NilAuditSvc", func(t *testing.T) {
		_, err := services.NewSecretService(services.SecretServiceParams{
			Logger:    logger,
			Repo:      new(MockSecretRepo),
			RBACSvc:   new(MockRBACService),
			EventSvc:  new(MockEventService),
			AuditSvc:  nil,
			MasterKey: masterKey,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "audit service is required")
	})

	t.Run("EmptyMasterKeyInProduction", func(t *testing.T) {
		_, err := services.NewSecretService(services.SecretServiceParams{
			Logger:      logger,
			Repo:        new(MockSecretRepo),
			RBACSvc:     new(MockRBACService),
			EventSvc:    new(MockEventService),
			AuditSvc:    new(MockAuditService),
			MasterKey:   "",
			Environment: "production",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "SECRETS_ENCRYPTION_KEY is required in production")
	})

	t.Run("ShortMasterKey_Warns", func(t *testing.T) {
		svc, err := services.NewSecretService(services.SecretServiceParams{
			Logger:    logger,
			Repo:      new(MockSecretRepo),
			RBACSvc:   new(MockRBACService),
			EventSvc:  new(MockEventService),
			AuditSvc:  new(MockAuditService),
			MasterKey: "short",
		})
		require.NoError(t, err)
		assert.NotNil(t, svc)
	})

	t.Run("DefaultKeyUsedInDev", func(t *testing.T) {
		svc, err := services.NewSecretService(services.SecretServiceParams{
			Logger:    logger,
			Repo:      new(MockSecretRepo),
			RBACSvc:   new(MockRBACService),
			EventSvc:  new(MockEventService),
			AuditSvc:  new(MockAuditService),
			MasterKey: "",
		})
		require.NoError(t, err)
		assert.NotNil(t, svc)
	})
}

func TestSecretService_AuthorizeErrors(t *testing.T) {
	mockRepo := new(MockSecretRepo)
	mockRBAC := new(MockRBACService)
	mockEvent := new(MockEventService)
	mockAudit := new(MockAuditService)

	svc, err := services.NewSecretService(services.SecretServiceParams{
		Repo:        mockRepo,
		RBACSvc:     mockRBAC,
		EventSvc:    mockEvent,
		AuditSvc:    mockAudit,
		Logger:      slog.Default(),
		MasterKey:   "test-master-key-32-chars-long-!!!",
		Environment: "test",
	})
	require.NoError(t, err)

	ctx := context.Background()
	tenantID := uuid.New()
	userID := uuid.New()
	ctx = appcontext.WithTenantID(ctx, tenantID)
	ctx = appcontext.WithUserID(ctx, userID)

	t.Run("CreateSecret_Unauthorized", func(t *testing.T) {
		authErr := errors.New(errors.Forbidden, "permission denied")
		mockRBAC.On("Authorize", mock.Anything, userID, tenantID, domain.PermissionSecretWrite, "*").Return(authErr).Once()

		_, err := svc.CreateSecret(ctx, "name", "value", "desc")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "permission denied")
	})

	t.Run("GetSecret_Unauthorized", func(t *testing.T) {
		id := uuid.New()
		authErr := errors.New(errors.Forbidden, "permission denied")
		mockRBAC.On("Authorize", mock.Anything, userID, tenantID, domain.PermissionSecretRead, id.String()).Return(authErr).Once()

		_, err := svc.GetSecret(ctx, id)
		require.Error(t, err)
	})

	t.Run("ListSecrets_Unauthorized", func(t *testing.T) {
		authErr := errors.New(errors.Forbidden, "permission denied")
		mockRBAC.On("Authorize", mock.Anything, userID, tenantID, domain.PermissionSecretRead, "*").Return(authErr).Once()

		_, err := svc.ListSecrets(ctx)
		require.Error(t, err)
	})

	t.Run("DeleteSecret_Unauthorized", func(t *testing.T) {
		id := uuid.New()
		authErr := errors.New(errors.Forbidden, "permission denied")
		mockRBAC.On("Authorize", mock.Anything, userID, tenantID, domain.PermissionSecretDelete, id.String()).Return(authErr).Once()

		err := svc.DeleteSecret(ctx, id)
		require.Error(t, err)
	})
}
