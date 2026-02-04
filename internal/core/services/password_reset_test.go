package services_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupPasswordResetServiceIntegrationTest(t *testing.T) (ports.PasswordResetService, ports.PasswordResetRepository, ports.UserRepository, context.Context) {
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewPasswordResetRepository(db)
	userRepo := postgres.NewUserRepo(db)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := services.NewPasswordResetService(repo, userRepo, logger)

	return svc, repo, userRepo, ctx
}

func TestPasswordResetService_Integration(t *testing.T) {
	svc, repo, userRepo, ctx := setupPasswordResetServiceIntegrationTest(t)

	t.Run("RequestAndReset", func(t *testing.T) {
		email := "reset@test.com"
		userID := uuid.New()
		err := userRepo.Create(ctx, &domain.User{ID: userID, Email: email, PasswordHash: "old"})
		require.NoError(t, err)

		// 1. Request Reset
		err = svc.RequestReset(ctx, email)
		assert.NoError(t, err)

		// In an integration environment, we verify the token creation indirectly via success status,
		// as email interception is outside the scope of this test.
		// For the reset phase, we seed a known token directly into the repository to ensure
		// we have a valid reference for verification.
		tokenStr := "my-secret-token"
		hash := sha256.Sum256([]byte(tokenStr))
		tokenHash := hex.EncodeToString(hash[:])

		resetToken := &domain.PasswordResetToken{
			ID:        uuid.New(),
			UserID:    userID,
			TokenHash: tokenHash,
			ExpiresAt: time.Now().Add(time.Hour),
			Used:      false,
			CreatedAt: time.Now(),
		}
		err = repo.Create(ctx, resetToken)
		require.NoError(t, err)

		// 2. Reset Password
		newPass := "new-secure-password"
		err = svc.ResetPassword(ctx, tokenStr, newPass)
		assert.NoError(t, err)

		// Verify user password updated
		updatedUser, err := userRepo.GetByID(ctx, userID)
		assert.NoError(t, err)
		assert.NotEqual(t, "old", updatedUser.PasswordHash)

		// Verify token used
		usedToken, err := repo.GetByTokenHash(ctx, tokenHash)
		assert.NoError(t, err)
		assert.True(t, usedToken.Used)
	})

	t.Run("ExpiredToken", func(t *testing.T) {
		userID := uuid.New()
		_ = userRepo.Create(ctx, &domain.User{ID: userID, Email: "expired@test.com"})

		tokenStr := "expired-token"
		hash := sha256.Sum256([]byte(tokenStr))
		tokenHash := hex.EncodeToString(hash[:])

		resetToken := &domain.PasswordResetToken{
			ID:        uuid.New(),
			UserID:    userID,
			TokenHash: tokenHash,
			ExpiresAt: time.Now().Add(-time.Hour), // Expired
			Used:      false,
			CreatedAt: time.Now().Add(-2 * time.Hour),
		}
		_ = repo.Create(ctx, resetToken)

		err := svc.ResetPassword(ctx, tokenStr, "new-pass")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})
}
