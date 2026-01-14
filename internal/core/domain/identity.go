// Package domain defines core business entities.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// APIKey represents a long-lived credential for programmatic access.
type APIKey struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Key       string    `json:"key"` // The actual secret key
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	LastUsed  time.Time `json:"last_used"`
}
