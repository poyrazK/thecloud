package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v3"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

func TestSecretRepository_Create(t *testing.T) {
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
}

func TestSecretRepository_GetByID(t *testing.T) {
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
}

func TestSecretRepository_List(t *testing.T) {
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
}

func TestSecretRepository_Update(t *testing.T) {
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
}

func TestSecretRepository_Delete(t *testing.T) {
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
}
