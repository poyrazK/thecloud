// Package domain defines core business entities.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// CronStatus represents the current state of a scheduled cron job.
type CronStatus string

const (
	// CronStatusActive indicates the job is scheduled and will run according to its expression.
	CronStatusActive CronStatus = "ACTIVE"
	// CronStatusPaused indicates the job is disabled and will not execute until resumed.
	CronStatusPaused CronStatus = "PAUSED"
	// CronStatusDeleted indicates the job has been removed from the scheduler.
	CronStatusDeleted CronStatus = "DELETED"
)

// CronJob represents a scheduled task that executes an HTTP request periodically.
type CronJob struct {
	ID            uuid.UUID  `json:"id"`
	UserID        uuid.UUID  `json:"user_id"`
	Name          string     `json:"name"`
	Schedule      string     `json:"schedule"`       // Standard cron expression (e.g., "*/5 * * * *")
	TargetURL     string     `json:"target_url"`     // The endpoint to call when the job triggers
	TargetMethod  string     `json:"target_method"`  // HTTP method (e.g., "POST", "GET")
	TargetPayload string     `json:"target_payload"` // Optional JSON payload for the request
	Status        CronStatus `json:"status"`
	LastRunAt     *time.Time `json:"last_run_at"`
	NextRunAt     *time.Time `json:"next_run_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// CronJobRun records the result of a single execution of a CronJob.
type CronJobRun struct {
	ID         uuid.UUID `json:"id"`
	JobID      uuid.UUID `json:"job_id"`
	Status     string    `json:"status"`      // Outcome of the run ("SUCCESS" or "FAILED")
	StatusCode int       `json:"status_code"` // HTTP response code from the target
	Response   string    `json:"response"`    // Raw response body from the target
	DurationMs int64     `json:"duration_ms"` // How long the execution took in milliseconds
	StartedAt  time.Time `json:"started_at"`
}
