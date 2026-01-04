package domain

import (
	"time"

	"github.com/google/uuid"
)

type CronStatus string

const (
	CronStatusActive  CronStatus = "ACTIVE"
	CronStatusPaused  CronStatus = "PAUSED"
	CronStatusDeleted CronStatus = "DELETED"
)

type CronJob struct {
	ID            uuid.UUID  `json:"id"`
	UserID        uuid.UUID  `json:"user_id"`
	Name          string     `json:"name"`
	Schedule      string     `json:"schedule"` // Cron expression
	TargetURL     string     `json:"target_url"`
	TargetMethod  string     `json:"target_method"`
	TargetPayload string     `json:"target_payload"`
	Status        CronStatus `json:"status"`
	LastRunAt     *time.Time `json:"last_run_at"`
	NextRunAt     *time.Time `json:"next_run_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type CronJobRun struct {
	ID         uuid.UUID `json:"id"`
	JobID      uuid.UUID `json:"job_id"`
	Status     string    `json:"status"` // SUCCESS, FAILED
	StatusCode int       `json:"status_code"`
	Response   string    `json:"response"`
	DurationMs int64     `json:"duration_ms"`
	StartedAt  time.Time `json:"started_at"`
}
