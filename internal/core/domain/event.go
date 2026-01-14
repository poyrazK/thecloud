// Package domain defines core business entities.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Event represents a significant system occurrence or user action.
// Unlike AuditLog, Events are primarily used for system coordination and observability.
type Event struct {
	ID           uuid.UUID   `json:"id"`
	UserID       uuid.UUID   `json:"user_id"`
	Action       string      `json:"action"`        // High-level action identifier (e.g., "INSTANCE_LAUNCH")
	ResourceID   string      `json:"resource_id"`   // Uniquely identifies the affected resource
	ResourceType string      `json:"resource_type"` // Classification of the resource (e.g., "INSTANCE", "VPC")
	Metadata     interface{} `json:"metadata"`      // Arbitrary event-specific JSON data
	CreatedAt    time.Time   `json:"created_at"`
}
