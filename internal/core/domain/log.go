package domain

import (
	"time"

	"github.com/google/uuid"
)

// LogEntry represents a single log line from a resource.
type LogEntry struct {
	ID           uuid.UUID `json:"id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	ResourceID   string    `json:"resource_id"`
	ResourceType string    `json:"resource_type"`
	Level        string    `json:"level"`
	Message      string    `json:"message"`
	Timestamp    time.Time `json:"timestamp"`
	TraceID      string    `json:"trace_id,omitempty"`
}

// LogQuery defines filters for searching logs.
type LogQuery struct {
	TenantID     uuid.UUID  `json:"tenant_id"`
	ResourceID   string     `json:"resource_id,omitempty"`
	ResourceType string     `json:"resource_type,omitempty"`
	Level        string     `json:"level,omitempty"`
	Search       string     `json:"search,omitempty"`
	StartTime    *time.Time `json:"start_time,omitempty"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	Limit        int        `json:"limit,omitempty"`
	Offset       int        `json:"offset,omitempty"`
}
