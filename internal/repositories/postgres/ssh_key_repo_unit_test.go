package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSHKeyRepo_Unit(t *testing.T) {
	t.Parallel()

	t.Run("Create", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewSSHKeyRepo(mock)
		ctx := context.Background()
		key := &domain.SSHKey{
			ID:          uuid.New(),
			UserID:      uuid.New(),
			TenantID:    uuid.New(),
			Name:        "test-key",
			PublicKey:   "ssh-rsa ...",
			Fingerprint: "aa:bb:cc",
			CreatedAt:   time.Now(),
		}

		mock.ExpectExec("INSERT INTO ssh_keys").
			WithArgs(key.ID, key.UserID, key.TenantID, key.Name, key.PublicKey, key.Fingerprint, key.CreatedAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = repo.Create(ctx, key)
		require.NoError(t, err)
	})

	t.Run("GetByID_NotFound", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewSSHKeyRepo(mock)
		ctx := context.Background()
		id := uuid.New()

		mock.ExpectQuery("SELECT .* FROM ssh_keys WHERE id = \\$1").
			WithArgs(id).
			WillReturnError(pgx.ErrNoRows)

		res, err := repo.GetByID(ctx, id)
		require.Error(t, err)
		assert.Nil(t, res)
	})
}
