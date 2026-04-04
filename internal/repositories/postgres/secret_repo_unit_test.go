package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	theclouderrors "github.com/poyrazk/thecloud/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecretRepositoryCreate(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewSecretRepository(mock)
		secret := &domain.Secret{
			ID:             uuid.New(),
			UserID:         uuid.New(),
			TenantID:       uuid.New(),
			Name:           "test-secret",
			EncryptedValue: "encrypted",
			Description:    "desc",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		mock.ExpectExec("INSERT INTO secrets").
			WithArgs(secret.ID, secret.UserID, secret.TenantID, secret.Name, secret.EncryptedValue, secret.Description, secret.CreatedAt, secret.UpdatedAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = repo.Create(context.Background(), secret)
		require.NoError(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewSecretRepository(mock)
		secret := &domain.Secret{
			ID: uuid.New(),
		}

		mock.ExpectExec("INSERT INTO secrets").
			WillReturnError(errors.New("db error"))

		err = repo.Create(context.Background(), secret)
		require.Error(t, err)
	})
}

func TestSecretRepositoryGetByID(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewSecretRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		tenantID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		ctx = appcontext.WithTenantID(ctx, tenantID)
		now := time.Now()
		var lastAccessedAt *time.Time = nil

		mock.ExpectQuery("SELECT id, user_id, tenant_id, name, encrypted_value, description, created_at, updated_at, last_accessed_at FROM secrets").
			WithArgs(id, tenantID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "tenant_id", "name", "encrypted_value", "description", "created_at", "updated_at", "last_accessed_at"}).
				AddRow(id, userID, tenantID, "test-secret", "encrypted", "desc", now, now, lastAccessedAt))

		secret, err := repo.GetByID(ctx, id)
		require.NoError(t, err)
		assert.NotNil(t, secret)
		assert.Equal(t, id, secret.ID)
		assert.Equal(t, tenantID, secret.TenantID)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewSecretRepository(mock)
		id := uuid.New()
		tenantID := uuid.New()
		ctx := appcontext.WithTenantID(context.Background(), tenantID)

		mock.ExpectQuery("SELECT id, user_id, tenant_id, name, encrypted_value, description, created_at, updated_at, last_accessed_at FROM secrets").
			WithArgs(id, tenantID).
			WillReturnError(pgx.ErrNoRows)

		secret, err := repo.GetByID(ctx, id)
		require.Error(t, err)
		assert.Nil(t, secret)
		var target *theclouderrors.Error
		if errors.As(err, &target) {
			assert.Equal(t, theclouderrors.NotFound, target.Type)
		}
	})
}

func TestSecretRepositoryList(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewSecretRepository(mock)
		userID := uuid.New()
		tenantID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		ctx = appcontext.WithTenantID(ctx, tenantID)
		now := time.Now()
		var lastAccessedAt *time.Time = nil

		mock.ExpectQuery("SELECT id, user_id, tenant_id, name, encrypted_value, description, created_at, updated_at, last_accessed_at FROM secrets").
			WithArgs(tenantID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "tenant_id", "name", "encrypted_value", "description", "created_at", "updated_at", "last_accessed_at"}).
				AddRow(uuid.New(), userID, tenantID, "test-secret", "encrypted", "desc", now, now, lastAccessedAt))

		secrets, err := repo.List(ctx)
		require.NoError(t, err)
		assert.Len(t, secrets, 1)
	})
}

func TestSecretRepositoryUpdate(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewSecretRepository(mock)
		secret := &domain.Secret{
			ID:             uuid.New(),
			UserID:         uuid.New(),
			TenantID:       uuid.New(),
			EncryptedValue: "new-encrypted",
			Description:    "new-desc",
			UpdatedAt:      time.Now(),
		}

		mock.ExpectExec("UPDATE secrets").
			WithArgs(secret.EncryptedValue, secret.Description, pgxmock.AnyArg(), secret.LastAccessedAt, secret.ID, secret.TenantID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err = repo.Update(context.Background(), secret)
		require.NoError(t, err)
	})
}

func TestSecretRepositoryDelete(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewSecretRepository(mock)
		id := uuid.New()
		tenantID := uuid.New()
		ctx := appcontext.WithTenantID(context.Background(), tenantID)

		mock.ExpectExec("DELETE FROM secrets").
			WithArgs(id, tenantID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = repo.Delete(ctx, id)
		require.NoError(t, err)
	})
}
