// Package domain defines core business entities.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// APIKey represents a long-lived credential for programmatic access.
type APIKey struct {
	ID              uuid.UUID  `json:"id"`
	UserID          uuid.UUID  `json:"user_id"`
	TenantID        uuid.UUID  `json:"tenant_id"`
	DefaultTenantID *uuid.UUID `json:"default_tenant_id,omitempty"`
	Key             string     `json:"key,omitempty"` // plaintext shown only at create/rotate; empty when listed
	KeyHash         string     `json:"-"`             // stored in DB, never serialized to JSON
	Name            string     `json:"name"`
	CreatedAt       time.Time  `json:"created_at"`
	LastUsed        time.Time  `json:"last_used"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
}
