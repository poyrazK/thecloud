// Package domain defines core business entities.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// PasswordResetToken represents a temporary security credential for password recovery workflows.
type PasswordResetToken struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	TokenHash string    `json:"-"` // Cryptographic hash of the raw token sent to the user
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"` // Flag to prevent token reuse
	CreatedAt time.Time `json:"created_at"`
}
