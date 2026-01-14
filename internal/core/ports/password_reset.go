// Package ports defines service and repository interfaces.
package ports

import (
	"context"

	"github.com/poyrazk/thecloud/internal/core/domain"
)

// PasswordResetRepository manages the lifecycle of temporary authentication tokens used for account recovery.
type PasswordResetRepository interface {
	// Create persists a new password reset token.
	Create(ctx context.Context, token *domain.PasswordResetToken) error
	// GetByTokenHash retrieves a token based on its cryptographic hash.
	GetByTokenHash(ctx context.Context, hash string) (*domain.PasswordResetToken, error)
	// MarkAsUsed invalidates a token after it has been successfully used.
	MarkAsUsed(ctx context.Context, tokenID string) error
	// DeleteExpired removes tokens that have passed their expiration date from storage.
	DeleteExpired(ctx context.Context) error
}

// PasswordResetService provides business logic for secure user account recovery workflows.
type PasswordResetService interface {
	// RequestReset initiates the password recovery process for a user identified by email.
	// It generates a secure token and facilitates notification (e.g., via email).
	RequestReset(ctx context.Context, email string) error

	// ResetPassword validates a recovery token and updates the user's password if authorized.
	ResetPassword(ctx context.Context, token, newPassword string) error
}
