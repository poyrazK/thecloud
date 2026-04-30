// Package ports defines service and repository interfaces.
package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

// FunctionScheduleRepository manages the persistence of function schedules.
type FunctionScheduleRepository interface {
	Create(ctx context.Context, schedule *domain.FunctionSchedule) error
	GetByID(ctx context.Context, id, userID, tenantID uuid.UUID) (*domain.FunctionSchedule, error)
	List(ctx context.Context, userID, tenantID uuid.UUID) ([]*domain.FunctionSchedule, error)
	Update(ctx context.Context, schedule *domain.FunctionSchedule) error
	Delete(ctx context.Context, id uuid.UUID) error

	// ClaimNextSchedulesToRun atomically selects due schedules and claims them.
	// Uses FOR UPDATE SKIP LOCKED for distributed safety.
	ClaimNextSchedulesToRun(ctx context.Context, lockTimeout time.Duration) ([]*domain.FunctionSchedule, error)
	// CompleteScheduleRun atomically records run result and advances next_run_at.
	CompleteScheduleRun(ctx context.Context, run *domain.FunctionScheduleRun, schedule *domain.FunctionSchedule, nextRunAt time.Time) error
	// ReapStaleClaims resets schedules whose claim expired.
	ReapStaleClaims(ctx context.Context) (int, error)
	// GetScheduleRuns returns run history for a schedule.
	GetScheduleRuns(ctx context.Context, scheduleID uuid.UUID, limit int) ([]*domain.FunctionScheduleRun, error)
}

// FunctionScheduleService provides business logic for scheduled function invocations.
type FunctionScheduleService interface {
	CreateSchedule(ctx context.Context, functionID uuid.UUID, name, schedule string, payload []byte) (*domain.FunctionSchedule, error)
	ListSchedules(ctx context.Context) ([]*domain.FunctionSchedule, error)
	GetSchedule(ctx context.Context, id uuid.UUID) (*domain.FunctionSchedule, error)
	DeleteSchedule(ctx context.Context, id uuid.UUID) error
	PauseSchedule(ctx context.Context, id uuid.UUID) error
	ResumeSchedule(ctx context.Context, id uuid.UUID) error
	GetScheduleRuns(ctx context.Context, id uuid.UUID, limit int) ([]*domain.FunctionScheduleRun, error)
}
