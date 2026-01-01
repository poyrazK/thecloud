package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID           uuid.UUID       `json:"id"`
	Action       string          `json:"action"`        // e.g. INSTANCE_LAUNCH
	ResourceID   string          `json:"resource_id"`   // e.g. UUID of instance
	ResourceType string          `json:"resource_type"` // e.g. INSTANCE, VPC
	Metadata     json.RawMessage `json:"metadata"`      // JSON details
	CreatedAt    time.Time       `json:"created_at"`
}
