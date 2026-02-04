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
)

func TestSecretRepository_Create(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSecretRepository(mock)
		secret := &domain.Secret{
			ID:             uuid.New(),
			UserID:         uuid.New(),
			Name:           "test-secret",
			EncryptedValue: "encrypted",
			Description:    "desc",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		mock.ExpectExec("INSERT INTO secrets").
			WithArgs(secret.ID, secret.UserID, secret.Name, secret.EncryptedValue, secret.Description, secret.CreatedAt, secret.UpdatedAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = repo.Create(context.Background(), secret)
		assert.NoError(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSecretRepository(mock)
		secret := &domain.Secret{
			ID: uuid.New(),
		}

		mock.ExpectExec("INSERT INTO secrets").
			WillReturnError(errors.New("db error"))

		err = repo.Create(context.Background(), secret)
		assert.Error(t, err)
	})
}

func TestSecretRepository_GetByID(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSecretRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		now := time.Now()
		var lastAccessedAt *time.Time = nil

		mock.ExpectQuery("SELECT id, user_id, name, encrypted_value, description, created_at, updated_at, last_accessed_at FROM secrets").
			WithArgs(id, userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "encrypted_value", "description", "created_at", "updated_at", "last_accessed_at"}).
				AddRow(id, userID, "test-secret", "encrypted", "desc", now, now, lastAccessedAt))

		secret, err := repo.GetByID(ctx, id)
		assert.NoError(t, err)
		assert.NotNil(t, secret)
		assert.Equal(t, id, secret.ID)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSecretRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectQuery("SELECT id, user_id, name, encrypted_value, description, created_at, updated_at, last_accessed_at FROM secrets").
			WithArgs(id, userID).
			WillReturnError(pgx.ErrNoRows)

		secret, err := repo.GetByID(ctx, id)
		assert.Error(t, err)
		assert.Nil(t, secret)
		theCloudErr, ok := err.(*theclouderrors.Error)
		if ok {
			assert.Equal(t, theclouderrors.NotFound, theCloudErr.Type)
		}
	})
}

func TestSecretRepository_List(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSecretRepository(mock)
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		now := time.Now()
		var lastAccessedAt *time.Time = nil

		mock.ExpectQuery("SELECT id, user_id, name, encrypted_value, description, created_at, updated_at, last_accessed_at FROM secrets").
			WithArgs(userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "encrypted_value", "description", "created_at", "updated_at", "last_accessed_at"}).
				AddRow(uuid.New(), userID, "test-secret", "encrypted", "desc", now, now, lastAccessedAt))

		secrets, err := repo.List(ctx)
		assert.NoError(t, err)
		assert.Len(t, secrets, 1)
	})

	t.Run("db error", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSecretRepository(mock)
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectQuery("SELECT id, user_id, name, encrypted_value, description, created_at, updated_at, last_accessed_at FROM secrets").
			WithArgs(userID).
			WillReturnError(errors.New("db error"))

		secrets, err := repo.List(ctx)
		assert.Error(t, err)
		assert.Nil(t, secrets)
	})
}

func TestSecretRepository_Update(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSecretRepository(mock)
		secret := &domain.Secret{
			ID:             uuid.New(),
			UserID:         uuid.New(),
			EncryptedValue: "new-encrypted",
			Description:    "new-desc",
			UpdatedAt:      time.Now(),
		}

		mock.ExpectExec("UPDATE secrets").
			WithArgs(secret.EncryptedValue, secret.Description, pgxmock.AnyArg(), secret.LastAccessedAt, secret.ID, secret.UserID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err = repo.Update(context.Background(), secret)
		assert.NoError(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSecretRepository(mock)
		secret := &domain.Secret{
			ID: uuid.New(),
		}

		mock.ExpectExec("UPDATE secrets").
			WillReturnError(errors.New("db error"))

		err = repo.Update(context.Background(), secret)
		assert.Error(t, err)
	})
}

func TestSecretRepository_Delete(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSecretRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectExec("DELETE FROM secrets").
			WithArgs(id, userID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = repo.Delete(ctx, id)
		assert.NoError(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSecretRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectExec("DELETE FROM secrets").
			WithArgs(id, userID).
			WillReturnError(errors.New("db error"))

		err = repo.Delete(ctx, id)
		assert.Error(t, err)
	})
}
