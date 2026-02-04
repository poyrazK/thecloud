package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

func TestIdentityRepository_CreateAPIKey(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewIdentityRepository(mock)
	apiKey := &domain.APIKey{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Key:       "secret-key",
		Name:      "test-key",
		CreatedAt: time.Now(),
	}

	mock.ExpectExec("INSERT INTO api_keys").
		WithArgs(apiKey.ID, apiKey.UserID, apiKey.Key, apiKey.Name, apiKey.CreatedAt, apiKey.ExpiresAt).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.CreateAPIKey(context.Background(), apiKey)
	assert.NoError(t, err)
}

func TestIdentityRepository_GetAPIKeyByKey(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewIdentityRepository(mock)
	id := uuid.New()
	userID := uuid.New()
	key := "secret-key"
	now := time.Now()
	var lastUsed *time.Time = nil

	mock.ExpectQuery("SELECT id, user_id, key, name, created_at, last_used, default_tenant_id, expires_at FROM api_keys").
		WithArgs(key).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "key", "name", "created_at", "last_used", "default_tenant_id", "expires_at"}).
			AddRow(id, userID, key, "test-key", now, lastUsed, nil, nil))

	apiKey, err := repo.GetAPIKeyByKey(context.Background(), key)
	assert.NoError(t, err)
	assert.NotNil(t, apiKey)
	assert.Equal(t, id, apiKey.ID)
}

func TestIdentityRepository_ListAPIKeysByUserID(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewIdentityRepository(mock)
	userID := uuid.New()
	now := time.Now()
	var lastUsed *time.Time = nil

	mock.ExpectQuery("SELECT id, user_id, key, name, created_at, last_used, default_tenant_id, expires_at FROM api_keys").
		WithArgs(userID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "key", "name", "created_at", "last_used", "default_tenant_id", "expires_at"}).
			AddRow(uuid.New(), userID, "secret-key", "test-key", now, lastUsed, nil, nil))

	keys, err := repo.ListAPIKeysByUserID(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, keys, 1)
}
