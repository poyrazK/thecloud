package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupIdentityServiceTest(_ *testing.T) (*MockIdentityRepo, *MockAuditService, ports.IdentityService) {
	repo := new(MockIdentityRepo)
	audit := new(MockAuditService)
	svc := services.NewIdentityService(repo, audit)
	return repo, audit, svc
}

const (
	apiKeyCreateAction = "api_key.create"
	apiKeyType         = "api_key"
)

func TestIdentityServiceCreateKeySuccess(t *testing.T) {
	repo, audit, svc := setupIdentityServiceTest(t)
	defer repo.AssertExpectations(t)
	defer audit.AssertExpectations(t)

	ctx := context.Background()
	userID := uuid.New()

	repo.On("CreateAPIKey", ctx, mock.MatchedBy(func(k *domain.APIKey) bool {
		return k.UserID == userID && len(k.Key) > 10 && k.Name == "Test Key"
	})).Return(nil)
	audit.On("Log", ctx, userID, apiKeyCreateAction, apiKeyType, mock.Anything, mock.Anything).Return(nil)

	key, err := svc.CreateKey(ctx, userID, "Test Key")

	assert.NoError(t, err)
	assert.NotNil(t, key)
	assert.Contains(t, key.Key, "thecloud_")
	assert.Equal(t, userID, key.UserID)
}

func TestIdentityServiceValidateAPIKeySuccess(t *testing.T) {
	repo, _, svc := setupIdentityServiceTest(t)
	defer repo.AssertExpectations(t)

	ctx := context.Background()

	keyStr := "thecloud_abc123"
	userID := uuid.New()
	apiKey := &domain.APIKey{ID: uuid.New(), UserID: userID, Key: keyStr}

	repo.On("GetAPIKeyByKey", ctx, keyStr).Return(apiKey, nil)

	result, err := svc.ValidateAPIKey(ctx, keyStr)

	assert.NoError(t, err)
	assert.Equal(t, apiKey.ID, result.ID)
	assert.Equal(t, userID, result.UserID)
}

func TestIdentityServiceValidateAPIKeyNotFound(t *testing.T) {
	repo, _, svc := setupIdentityServiceTest(t)
	defer repo.AssertExpectations(t)

	ctx := context.Background()

	repo.On("GetAPIKeyByKey", ctx, "invalid-key").Return(nil, assert.AnError)

	result, err := svc.ValidateAPIKey(ctx, "invalid-key")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestIdentityServiceListKeys(t *testing.T) {
	repo, _, svc := setupIdentityServiceTest(t)
	defer repo.AssertExpectations(t)

	ctx := context.Background()
	userID := uuid.New()
	keys := []*domain.APIKey{{ID: uuid.New(), UserID: userID}}

	repo.On("ListAPIKeysByUserID", ctx, userID).Return(keys, nil)

	result, err := svc.ListKeys(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, keys, result)
}

func TestIdentityServiceRevokeKey(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	keyID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		repo, audit, svc := setupIdentityServiceTest(t)
		apiKey := &domain.APIKey{ID: keyID, UserID: userID, Name: "Test"}

		repo.On("GetAPIKeyByID", ctx, keyID).Return(apiKey, nil)
		repo.On("DeleteAPIKey", ctx, keyID).Return(nil)
		audit.On("Log", ctx, userID, "api_key.revoke", apiKeyType, keyID.String(), mock.Anything).Return(nil)

		err := svc.RevokeKey(ctx, userID, keyID)
		assert.NoError(t, err)
	})

	t.Run("WrongUser", func(t *testing.T) {
		repo, _, svc := setupIdentityServiceTest(t)
		apiKey := &domain.APIKey{ID: keyID, UserID: uuid.New(), Name: "Test"}

		repo.On("GetAPIKeyByID", ctx, keyID).Return(apiKey, nil)

		err := svc.RevokeKey(ctx, userID, keyID)
		assert.Error(t, err)
	})
}

func TestIdentityServiceRotateKey(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	keyID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		repo, audit, svc := setupIdentityServiceTest(t)
		apiKey := &domain.APIKey{ID: keyID, UserID: userID, Name: "Test"}

		repo.On("GetAPIKeyByID", ctx, keyID).Return(apiKey, nil)
		repo.On("CreateAPIKey", ctx, mock.Anything).Return(nil)
		repo.On("DeleteAPIKey", ctx, keyID).Return(nil)
		audit.On("Log", ctx, userID, apiKeyCreateAction, apiKeyType, mock.Anything, mock.Anything).Return(nil)
		audit.On("Log", ctx, userID, "api_key.rotate", apiKeyType, keyID.String(), mock.Anything).Return(nil)

		newKey, err := svc.RotateKey(ctx, userID, keyID)
		assert.NoError(t, err)
		assert.NotNil(t, newKey)
	})
}

func TestIdentityServiceErrorPaths(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	keyID := uuid.New()

	t.Run("CreateKey_RepoError", func(t *testing.T) {
		repo, _, svc := setupIdentityServiceTest(t)
		repo.On("CreateAPIKey", ctx, mock.Anything).Return(assert.AnError)
		_, err := svc.CreateKey(ctx, userID, "Test")
		assert.Error(t, err)
	})

	t.Run("RevokeKey_GetError", func(t *testing.T) {
		repo, _, svc := setupIdentityServiceTest(t)
		repo.On("GetAPIKeyByID", ctx, keyID).Return(nil, assert.AnError)
		err := svc.RevokeKey(ctx, userID, keyID)
		assert.Error(t, err)
	})

	t.Run("RevokeKey_DeleteError", func(t *testing.T) {
		repo, _, svc := setupIdentityServiceTest(t)
		apiKey := &domain.APIKey{ID: keyID, UserID: userID}
		repo.On("GetAPIKeyByID", ctx, keyID).Return(apiKey, nil)
		repo.On("DeleteAPIKey", ctx, keyID).Return(assert.AnError)
		err := svc.RevokeKey(ctx, userID, keyID)
		assert.Error(t, err)
	})

	t.Run("RotateKey_GetError", func(t *testing.T) {
		repo, _, svc := setupIdentityServiceTest(t)
		repo.On("GetAPIKeyByID", ctx, keyID).Return(nil, assert.AnError)
		_, err := svc.RotateKey(ctx, userID, keyID)
		assert.Error(t, err)
	})

	t.Run("RotateKey_WrongUser", func(t *testing.T) {
		repo, _, svc := setupIdentityServiceTest(t)
		apiKey := &domain.APIKey{ID: keyID, UserID: uuid.New()}
		repo.On("GetAPIKeyByID", ctx, keyID).Return(apiKey, nil)
		_, err := svc.RotateKey(ctx, userID, keyID)
		assert.Error(t, err)
	})

	t.Run("RotateKey_DeleteOldError", func(t *testing.T) {
		repo, audit, svc := setupIdentityServiceTest(t)
		apiKey := &domain.APIKey{ID: keyID, UserID: userID}
		repo.On("GetAPIKeyByID", ctx, keyID).Return(apiKey, nil)
		repo.On("CreateAPIKey", ctx, mock.Anything).Return(nil)
		repo.On("DeleteAPIKey", ctx, keyID).Return(assert.AnError)
		audit.On("Log", ctx, userID, apiKeyCreateAction, apiKeyType, mock.Anything, mock.Anything).Return(nil)

		newKey, err := svc.RotateKey(ctx, userID, keyID)
		assert.NoError(t, err) // Should succeed anyway
		assert.NotNil(t, newKey)
	})
}

func TestIdentityServiceRotateKeyCreateError(t *testing.T) {
	repo, _, svc := setupIdentityServiceTest(t)
	ctx := context.Background()
	userID := uuid.New()
	keyID := uuid.New()
	apiKey := &domain.APIKey{ID: keyID, UserID: userID}

	repo.On("GetAPIKeyByID", ctx, keyID).Return(apiKey, nil)
	repo.On("CreateAPIKey", ctx, mock.Anything).Return(assert.AnError)

	_, err := svc.RotateKey(ctx, userID, keyID)
	assert.Error(t, err)
}
