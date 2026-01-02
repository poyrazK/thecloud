package domain

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID           uuid.UUID   `json:"id"`
	UserID       uuid.UUID   `json:"user_id"`
	Action       string      `json:"action"`        // e.g. INSTANCE_LAUNCH
	ResourceID   string      `json:"resource_id"`   // e.g. UUID of instance
	ResourceType string      `json:"resource_type"` // e.g. INSTANCE, VPC
	Metadata     interface{} `json:"metadata"`      // JSON details
	CreatedAt    time.Time   `json:"created_at"`
}
