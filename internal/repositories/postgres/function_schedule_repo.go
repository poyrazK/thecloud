// Package postgres provides PostgreSQL-backed repository implementations.
package postgres

import (
	"context"
	"fmt"
	stdlib_errors "errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
)

// PostgresFunctionScheduleRepository provides PostgreSQL-backed function schedule persistence.
type PostgresFunctionScheduleRepository struct {
	db DB
}

// NewPostgresFunctionScheduleRepository creates a function schedule repository using the provided DB.
func NewPostgresFunctionScheduleRepository(db DB) ports.FunctionScheduleRepository {
	return &PostgresFunctionScheduleRepository{db: db}
}

func (r *PostgresFunctionScheduleRepository) Create(ctx context.Context, schedule *domain.FunctionSchedule) error {
	query := `
		INSERT INTO function_schedules (id, user_id, tenant_id, function_id, name, schedule, payload, status, next_run_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Exec(ctx, query,
		schedule.ID,
		schedule.UserID,
		schedule.TenantID,
		schedule.FunctionID,
		schedule.Name,
		schedule.Schedule,
		schedule.Payload,
		schedule.Status,
		schedule.NextRunAt,
		schedule.CreatedAt,
		schedule.UpdatedAt,
	)
	return err
}

func (r *PostgresFunctionScheduleRepository) GetByID(ctx context.Context, id, userID, tenantID uuid.UUID) (*domain.FunctionSchedule, error) {
	query := `SELECT id, user_id, tenant_id, function_id, name, schedule, payload, status, last_run_at, next_run_at, claimed_until, created_at, updated_at FROM function_schedules WHERE id = $1 AND user_id = $2 AND tenant_id = $3`
	sched, err := r.scanFunctionSchedule(r.db.QueryRow(ctx, query, id, userID, tenantID))
	if err != nil {
		if stdlib_errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(errors.NotFound, fmt.Sprintf("function schedule not found: %s", id))
		}
		return nil, errors.Wrap(errors.Internal, "failed to get function schedule", err)
	}
	return sched, nil
}

func (r *PostgresFunctionScheduleRepository) List(ctx context.Context, userID, tenantID uuid.UUID) ([]*domain.FunctionSchedule, error) {
	query := `SELECT id, user_id, tenant_id, function_id, name, schedule, payload, status, last_run_at, next_run_at, claimed_until, created_at, updated_at FROM function_schedules WHERE user_id = $1 AND tenant_id = $2 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, userID, tenantID)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to list function schedules", err)
	}
	return r.scanFunctionSchedules(rows)
}

func (r *PostgresFunctionScheduleRepository) Update(ctx context.Context, schedule *domain.FunctionSchedule) error {
	query := `
		UPDATE function_schedules
		SET status = $1, last_run_at = $2, next_run_at = $3, claimed_until = $4, updated_at = NOW()
		WHERE id = $5
	`
	_, err := r.db.Exec(ctx, query, schedule.Status, schedule.LastRunAt, schedule.NextRunAt, schedule.ClaimedUntil, schedule.ID)
	return err
}

func (r *PostgresFunctionScheduleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM function_schedules WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *PostgresFunctionScheduleRepository) ClaimNextSchedulesToRun(ctx context.Context, lockTimeout time.Duration) ([]*domain.FunctionSchedule, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	now := time.Now()
	claimExpiry := now.Add(lockTimeout)
	farFuture := now.Add(365 * 24 * time.Hour)

	selectQuery := `
		SELECT id, user_id, tenant_id, function_id, name, schedule, payload, status, last_run_at, next_run_at, claimed_until, created_at, updated_at
		FROM function_schedules
		WHERE status = 'ACTIVE'
		  AND next_run_at <= NOW()
		  AND (claimed_until IS NULL OR claimed_until <= NOW())
		ORDER BY next_run_at ASC
		LIMIT 10
		FOR UPDATE SKIP LOCKED
	`
	rows, err := tx.Query(ctx, selectQuery)
	if err != nil {
		return nil, err
	}

	schedules, err := r.scanFunctionSchedules(rows)
	if err != nil {
		return nil, err
	}

	if len(schedules) == 0 {
		return nil, tx.Commit(ctx)
	}

	ids := make([]uuid.UUID, len(schedules))
	for i, sched := range schedules {
		ids[i] = sched.ID
	}

	updateQuery := `
		UPDATE function_schedules
		SET next_run_at = $1, claimed_until = $2, updated_at = NOW()
		WHERE id = ANY($3)
	`
	_, err = tx.Exec(ctx, updateQuery, farFuture, claimExpiry, ids)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	for _, sched := range schedules {
		sched.NextRunAt = &farFuture
		sched.ClaimedUntil = &claimExpiry
	}
	return schedules, nil
}

func (r *PostgresFunctionScheduleRepository) CompleteScheduleRun(ctx context.Context, run *domain.FunctionScheduleRun, schedule *domain.FunctionSchedule, nextRunAt time.Time) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	insertQuery := `
		INSERT INTO function_schedule_runs (id, schedule_id, invocation_id, status, status_code, duration_ms, error_message, started_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = tx.Exec(ctx, insertQuery, run.ID, run.ScheduleID, run.InvocationID, run.Status, run.StatusCode, run.DurationMs, run.ErrorMessage, run.StartedAt)
	if err != nil {
		return err
	}

	updateQuery := `
		UPDATE function_schedules
		SET last_run_at = $1, next_run_at = $2, claimed_until = NULL, updated_at = $3
		WHERE id = $4
	`
	_, err = tx.Exec(ctx, updateQuery, run.StartedAt, nextRunAt, time.Now(), schedule.ID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *PostgresFunctionScheduleRepository) ReapStaleClaims(ctx context.Context) (int, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	query := `
		UPDATE function_schedules
		SET claimed_until = NULL,
		    next_run_at = NOW(),
		    updated_at = NOW()
		WHERE claimed_until IS NOT NULL
		  AND claimed_until < NOW()
		  AND next_run_at > NOW() + INTERVAL '1 day'
		RETURNING id
	`
	rows, err := tx.Query(ctx, query)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	return count, tx.Commit(ctx)
}

func (r *PostgresFunctionScheduleRepository) GetScheduleRuns(ctx context.Context, scheduleID uuid.UUID, limit int) ([]*domain.FunctionScheduleRun, error) {
	query := `SELECT id, schedule_id, invocation_id, status, status_code, duration_ms, error_message, started_at FROM function_schedule_runs WHERE schedule_id = $1 ORDER BY started_at DESC LIMIT $2`
	rows, err := r.db.Query(ctx, query, scheduleID, limit)
	if err != nil {
		return nil, err
	}
	return r.scanFunctionScheduleRuns(rows)
}

func (r *PostgresFunctionScheduleRepository) scanFunctionSchedule(row pgx.Row) (*domain.FunctionSchedule, error) {
	var sched domain.FunctionSchedule
	var status string
	var payload []byte
	err := row.Scan(
		&sched.ID,
		&sched.UserID,
		&sched.TenantID,
		&sched.FunctionID,
		&sched.Name,
		&sched.Schedule,
		&payload,
		&status,
		&sched.LastRunAt,
		&sched.NextRunAt,
		&sched.ClaimedUntil,
		&sched.CreatedAt,
		&sched.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	sched.Payload = payload
	sched.Status = domain.FunctionScheduleStatus(status)
	return &sched, nil
}

func (r *PostgresFunctionScheduleRepository) scanFunctionSchedules(rows pgx.Rows) ([]*domain.FunctionSchedule, error) {
	defer rows.Close()
	var schedules []*domain.FunctionSchedule
	for rows.Next() {
		sched, err := r.scanFunctionSchedule(rows)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, sched)
	}
	return schedules, rows.Err()
}

func (r *PostgresFunctionScheduleRepository) scanFunctionScheduleRun(row pgx.Row) (*domain.FunctionScheduleRun, error) {
	var run domain.FunctionScheduleRun
	err := row.Scan(
		&run.ID,
		&run.ScheduleID,
		&run.InvocationID,
		&run.Status,
		&run.StatusCode,
		&run.DurationMs,
		&run.ErrorMessage,
		&run.StartedAt,
	)
	if err != nil {
		return nil, err
	}
	return &run, nil
}

func (r *PostgresFunctionScheduleRepository) scanFunctionScheduleRuns(rows pgx.Rows) ([]*domain.FunctionScheduleRun, error) {
	defer rows.Close()
	var runs []*domain.FunctionScheduleRun
	for rows.Next() {
		run, err := r.scanFunctionScheduleRun(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, rows.Err()
}