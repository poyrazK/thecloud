// Package domain defines core business entities.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Secret represents a sensitive configuration value (e.g., API key, certificate) stored securely.
// Values are encrypted at rest and only decrypted during authorized retrieval.
type Secret struct {
	ID             uuid.UUID  `json:"id"`
	UserID         uuid.UUID  `json:"user_id"`
	Name           string     `json:"name"`
	EncryptedValue string     `json:"encrypted_value,omitempty"` // AES-encrypted representation of the secret content
	Description    string     `json:"description"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	LastAccessedAt *time.Time `json:"last_accessed_at,omitempty"`
}
