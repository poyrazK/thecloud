// Package domain defines core business entities.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// AuditLog represents a security or operational event recorded for compliance and troubleshooting.
// It tracks who did what, when, and from where.
type AuditLog struct {
	ID           uuid.UUID              `json:"id"`
	UserID       uuid.UUID              `json:"user_id"`       // The user who performed the action
	Action       string                 `json:"action"`        // The action performed (e.g., "instance:launch")
	ResourceType string                 `json:"resource_type"` // Type of affected resource (e.g., "INSTANCE")
	ResourceID   string                 `json:"resource_id"`   // ID of the affected resource
	Details      map[string]interface{} `json:"details"`       // Additional context about the action
	IPAddress    string                 `json:"ip_address"`    // Originating IP of the request
	UserAgent    string                 `json:"user_agent"`    // Browser or CLI identifier
	CreatedAt    time.Time              `json:"created_at"`
}
