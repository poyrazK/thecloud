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
		WithArgs(apiKey.ID, apiKey.UserID, apiKey.TenantID, apiKey.Key, apiKey.KeyHash, apiKey.Name, apiKey.CreatedAt, apiKey.ExpiresAt, apiKey.DefaultTenantID).
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

	cases := []struct {
		name     string
		hash     string
		setup    func()
		wantErr  bool
		checkKey func(*domain.APIKey)
	}{
		{
			name: "found",
			hash: keyHash,
			setup: func() {
				mock.ExpectQuery(`SELECT id, user_id, tenant_id, key, name, created_at, last_used, default_tenant_id, expires_at FROM api_keys WHERE key_hash = \$1`).
					WithArgs(keyHash).
					WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "tenant_id", "key", "name", "created_at", "last_used", "default_tenant_id", "expires_at"}).
						AddRow(id, userID, tenantID, "secret-key", "test-key", now, lastUsed, nil, nil))
			},
			wantErr: false,
			checkKey: func(k *domain.APIKey) {
				assert.Equal(t, id, k.ID)
				assert.Equal(t, tenantID, k.TenantID)
			},
		},
		{
			name: "not_found",
			hash: "notfoundhash",
			setup: func() {
				mock.ExpectQuery(`SELECT id, user_id, tenant_id, key, name, created_at, last_used, default_tenant_id, expires_at FROM api_keys WHERE key_hash = \$1`).
					WithArgs("notfoundhash").
					WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "tenant_id", "key", "name", "created_at", "last_used", "default_tenant_id", "expires_at"}))
			},
			wantErr:  true,
			checkKey: func(k *domain.APIKey) {},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup()
			apiKey, err := repo.GetAPIKeyByHash(context.Background(), tc.hash)
			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, apiKey)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, apiKey)
				tc.checkKey(apiKey)
			}
		})
	}
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
