// Package domain defines core business entities.
package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// FunctionScheduleStatus represents the state of a function schedule.
type FunctionScheduleStatus string

const (
	FunctionScheduleStatusActive  FunctionScheduleStatus = "ACTIVE"
	FunctionScheduleStatusPaused  FunctionScheduleStatus = "PAUSED"
	FunctionScheduleStatusDeleted FunctionScheduleStatus = "DELETED"
)

// FunctionSchedule represents a scheduled invocation of a serverless function.
type FunctionSchedule struct {
	ID         uuid.UUID              `json:"id"`
	UserID     uuid.UUID              `json:"user_id"`
	TenantID   uuid.UUID              `json:"tenant_id"`
	FunctionID uuid.UUID              `json:"function_id"`
	Name       string                 `json:"name"`
	Schedule   string                 `json:"schedule"` // Cron expression (e.g. "*/5 * * * *")
	Payload    json.RawMessage       `json:"payload"`
	Status     FunctionScheduleStatus `json:"status"`
	LastRunAt  *time.Time             `json:"last_run_at,omitempty"`
	NextRunAt  *time.Time             `json:"next_run_at,omitempty"`
	ClaimedUntil *time.Time           `json:"claimed_until,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// FunctionScheduleRun records a single execution of a FunctionSchedule.
type FunctionScheduleRun struct {
	ID           uuid.UUID `json:"id"`
	ScheduleID   uuid.UUID `json:"schedule_id"`
	InvocationID uuid.UUID `json:"invocation_id"`
	Status       string    `json:"status"`        // "SUCCESS" or "FAILED"
	StatusCode   int       `json:"status_code"`   // Exit code from function
	DurationMs   int64     `json:"duration_ms"`   // Execution time in milliseconds
	ErrorMessage string    `json:"error_message,omitempty"`
	StartedAt    time.Time `json:"started_at"`
}