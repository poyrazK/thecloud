package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIdentityRepository_CreateAPIKey(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewIdentityRepository(mock)
	tenantID := uuid.New()
	apiKey := &domain.APIKey{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TenantID:  tenantID,
		Key:       "secret-key",
		KeyHash:   "hash-of-secret-key",
		Name:      "test-key",
		CreatedAt: time.Now(),
	}

	mock.ExpectExec("INSERT INTO api_keys").
		WithArgs(apiKey.ID, apiKey.UserID, apiKey.TenantID, apiKey.Key, apiKey.KeyHash, apiKey.Name, apiKey.CreatedAt, apiKey.ExpiresAt).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.CreateAPIKey(context.Background(), apiKey)
	require.NoError(t, err)
}

func TestIdentityRepository_GetAPIKeyByHash(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewIdentityRepository(mock)
	id := uuid.New()
	userID := uuid.New()
	tenantID := uuid.New()
	keyHash := "hash-of-secret-key"
	now := time.Now()
	var lastUsed *time.Time = nil

	mock.ExpectQuery("SELECT id, user_id, tenant_id, key, name, created_at, last_used, default_tenant_id, expires_at FROM api_keys").
		WithArgs(keyHash).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "tenant_id", "key", "name", "created_at", "last_used", "default_tenant_id", "expires_at"}).
			AddRow(id, userID, tenantID, "secret-key", "test-key", now, lastUsed, nil, nil))

	apiKey, err := repo.GetAPIKeyByHash(context.Background(), keyHash)
	require.NoError(t, err)
	assert.NotNil(t, apiKey)
	assert.Equal(t, id, apiKey.ID)
	assert.Equal(t, tenantID, apiKey.TenantID)
}

func TestIdentityRepository_ListAPIKeysByUserID(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewIdentityRepository(mock)
	userID := uuid.New()
	tenantID := uuid.New()
	now := time.Now()
	var lastUsed *time.Time = nil

	mock.ExpectQuery("SELECT id, user_id, tenant_id, key, name, created_at, last_used, default_tenant_id, expires_at FROM api_keys").
		WithArgs(userID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "tenant_id", "key", "name", "created_at", "last_used", "default_tenant_id", "expires_at"}).
			AddRow(uuid.New(), userID, tenantID, "secret-key", "test-key", now, lastUsed, nil, nil))

	keys, err := repo.ListAPIKeysByUserID(context.Background(), userID)
	require.NoError(t, err)
	assert.Len(t, keys, 1)
}
