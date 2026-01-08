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

func setupIdentityServiceTest(t *testing.T) (*MockIdentityRepo, *MockAuditService, ports.IdentityService) {
	repo := new(MockIdentityRepo)
	audit := new(MockAuditService)
	svc := services.NewIdentityService(repo, audit)
	return repo, audit, svc
}

func TestIdentityService_CreateKey_Success(t *testing.T) {
	repo, audit, svc := setupIdentityServiceTest(t)
	defer repo.AssertExpectations(t)
	defer audit.AssertExpectations(t)

	ctx := context.Background()
	userID := uuid.New()

	repo.On("CreateAPIKey", ctx, mock.MatchedBy(func(k *domain.APIKey) bool {
		return k.UserID == userID && len(k.Key) > 10 && k.Name == "Test Key"
	})).Return(nil)
	audit.On("Log", ctx, userID, "api_key.create", "api_key", mock.Anything, mock.Anything).Return(nil)

	key, err := svc.CreateKey(ctx, userID, "Test Key")

	assert.NoError(t, err)
	assert.NotNil(t, key)
	assert.Contains(t, key.Key, "thecloud_")
	assert.Equal(t, userID, key.UserID)
}

func TestIdentityService_ValidateAPIKey_Success(t *testing.T) {
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

func TestIdentityService_ValidateAPIKey_NotFound(t *testing.T) {
	repo, _, svc := setupIdentityServiceTest(t)
	defer repo.AssertExpectations(t)

	ctx := context.Background()

	repo.On("GetAPIKeyByKey", ctx, "invalid-key").Return(nil, assert.AnError)

	result, err := svc.ValidateAPIKey(ctx, "invalid-key")

	assert.Error(t, err)
	assert.Nil(t, result)
}
