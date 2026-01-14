// Package ports defines service and repository interfaces.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

// CronRepository facilitates the storage and scheduling of background cron tasks.
type CronRepository interface {
	// CreateJob persists a new cron job configuration.
	CreateJob(ctx context.Context, job *domain.CronJob) error
	// GetJobByID retrieves a specific cron job for an authorized user.
	GetJobByID(ctx context.Context, id, userID uuid.UUID) (*domain.CronJob, error)
	// ListJobs returns all cron jobs owned by a user.
	ListJobs(ctx context.Context, userID uuid.UUID) ([]*domain.CronJob, error)
	// UpdateJob modifies an existing cron job's schedule, target, or status.
	UpdateJob(ctx context.Context, job *domain.CronJob) error
	// DeleteJob removes a cron job from persistent storage.
	DeleteJob(ctx context.Context, id uuid.UUID) error

	// For the scheduler worker

	// GetNextJobsToRun returns a slice of jobs that are due for execution according to their schedule.
	GetNextJobsToRun(ctx context.Context) ([]*domain.CronJob, error)
	// SaveJobRun records the execution result and metadata of a completed cron task.
	SaveJobRun(ctx context.Context, run *domain.CronJobRun) error
}

// CronService provides business logic for managing scheduled background tasks.
type CronService interface {
	// CreateJob schedules a new recurring task.
	CreateJob(ctx context.Context, name, schedule, targetURL, targetMethod, targetPayload string) (*domain.CronJob, error)
	// ListJobs returns all tasks for the current user.
	ListJobs(ctx context.Context) ([]*domain.CronJob, error)
	// GetJob fetches details for a specific scheduled task.
	GetJob(ctx context.Context, id uuid.UUID) (*domain.CronJob, error)
	// PauseJob temporarily disables a scheduled task.
	PauseJob(ctx context.Context, id uuid.UUID) error
	// ResumeJob re-enables a previously paused task.
	ResumeJob(ctx context.Context, id uuid.UUID) error
	// DeleteJob permanently removes a scheduled task.
	DeleteJob(ctx context.Context, id uuid.UUID) error
}
