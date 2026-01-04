package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

type CronRepository interface {
	CreateJob(ctx context.Context, job *domain.CronJob) error
	GetJobByID(ctx context.Context, id, userID uuid.UUID) (*domain.CronJob, error)
	ListJobs(ctx context.Context, userID uuid.UUID) ([]*domain.CronJob, error)
	UpdateJob(ctx context.Context, job *domain.CronJob) error
	DeleteJob(ctx context.Context, id uuid.UUID) error

	// For the scheduler worker
	GetNextJobsToRun(ctx context.Context) ([]*domain.CronJob, error)
	SaveJobRun(ctx context.Context, run *domain.CronJobRun) error
}

type CronService interface {
	CreateJob(ctx context.Context, name, schedule, targetURL, targetMethod, targetPayload string) (*domain.CronJob, error)
	ListJobs(ctx context.Context) ([]*domain.CronJob, error)
	GetJob(ctx context.Context, id uuid.UUID) (*domain.CronJob, error)
	PauseJob(ctx context.Context, id uuid.UUID) error
	ResumeJob(ctx context.Context, id uuid.UUID) error
	DeleteJob(ctx context.Context, id uuid.UUID) error
}
