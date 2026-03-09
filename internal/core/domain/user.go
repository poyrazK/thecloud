// Package domain defines core business entities.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents an authenticated entity in the system.
type User struct {
	ID              uuid.UUID  `json:"id"`
	TenantID        uuid.UUID  `json:"tenant_id"`
	DefaultTenantID *uuid.UUID `json:"default_tenant_id,omitempty"`
	Email           string     `json:"email"`
	PasswordHash    string     `json:"-"` // Never serialize password
	Name            string     `json:"name"`
	Role            string     `json:"role"` // "admin" or "user"
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
