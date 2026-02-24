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

func TestPasswordResetRepository_Create(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewPasswordResetRepository(mock)
	token := &domain.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: "hash",
		ExpiresAt: time.Now().Add(time.Hour),
		Used:      false,
		CreatedAt: time.Now(),
	}

	mock.ExpectExec("INSERT INTO password_reset_tokens").
		WithArgs(token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.Used, token.CreatedAt).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.Create(context.Background(), token)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPasswordResetRepository_GetByTokenHash(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewPasswordResetRepository(mock)
	hash := "somehash"
	userID := uuid.New()
	now := time.Now()

	mock.ExpectQuery("SELECT id, user_id, token_hash, expires_at, used, created_at FROM password_reset_tokens").
		WithArgs(hash).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at", "used", "created_at"}).
			AddRow(uuid.New(), userID, hash, now.Add(time.Hour), false, now))

	token, err := repo.GetByTokenHash(context.Background(), hash)
	require.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, hash, token.TokenHash)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPasswordResetRepository_MarkAsUsed(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewPasswordResetRepository(mock)
	tokenID := uuid.New()

	mock.ExpectExec("UPDATE password_reset_tokens SET used = true WHERE id = \\$1").
		WithArgs(tokenID.String()).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err = repo.MarkAsUsed(context.Background(), tokenID.String())
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPasswordResetRepository_DeleteExpired(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewPasswordResetRepository(mock)

	mock.ExpectExec("DELETE FROM password_reset_tokens WHERE expires_at < \\$1").
		WithArgs(pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("DELETE", 5))

	err = repo.DeleteExpired(context.Background())
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
