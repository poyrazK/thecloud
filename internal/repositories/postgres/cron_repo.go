// Package postgres provides PostgreSQL-backed repository implementations.
package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

// PostgresCronRepository provides PostgreSQL-backed cron job persistence.
type PostgresCronRepository struct {
	db DB
}

// NewPostgresCronRepository creates a cron repository using the provided DB.
func NewPostgresCronRepository(db DB) ports.CronRepository {
	return &PostgresCronRepository{db: db}
}

func (r *PostgresCronRepository) CreateJob(ctx context.Context, job *domain.CronJob) error {
	query := `
		INSERT INTO cron_jobs (id, user_id, name, schedule, target_url, target_method, target_payload, status, next_run_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Exec(ctx, query,
		job.ID,
		job.UserID,
		job.Name,
		job.Schedule,
		job.TargetURL,
		job.TargetMethod,
		job.TargetPayload,
		job.Status,
		job.NextRunAt,
		job.CreatedAt,
		job.UpdatedAt,
	)
	return err
}

func (r *PostgresCronRepository) GetJobByID(ctx context.Context, id, userID uuid.UUID) (*domain.CronJob, error) {
	query := `SELECT id, user_id, name, schedule, target_url, target_method, target_payload, status, last_run_at, next_run_at, created_at, updated_at FROM cron_jobs WHERE id = $1 AND user_id = $2`
	return r.scanCronJob(r.db.QueryRow(ctx, query, id, userID))
}

func (r *PostgresCronRepository) ListJobs(ctx context.Context, userID uuid.UUID) ([]*domain.CronJob, error) {
	query := `SELECT id, user_id, name, schedule, target_url, target_method, target_payload, status, last_run_at, next_run_at, created_at, updated_at FROM cron_jobs WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	return r.scanCronJobs(rows)
}

func (r *PostgresCronRepository) UpdateJob(ctx context.Context, job *domain.CronJob) error {
	query := `
		UPDATE cron_jobs 
		SET status = $1, last_run_at = $2, next_run_at = $3, updated_at = NOW()
		WHERE id = $4
	`
	_, err := r.db.Exec(ctx, query, job.Status, job.LastRunAt, job.NextRunAt, job.ID)
	return err
}

func (r *PostgresCronRepository) DeleteJob(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM cron_jobs WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *PostgresCronRepository) GetNextJobsToRun(ctx context.Context) ([]*domain.CronJob, error) {
	query := `SELECT id, user_id, name, schedule, target_url, target_method, target_payload, status, last_run_at, next_run_at, created_at, updated_at FROM cron_jobs WHERE status = 'ACTIVE' AND next_run_at <= NOW() FOR UPDATE SKIP LOCKED`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	return r.scanCronJobs(rows)
}

func (r *PostgresCronRepository) scanCronJob(row pgx.Row) (*domain.CronJob, error) {
	var job domain.CronJob
	var status string
	err := row.Scan(
		&job.ID,
		&job.UserID,
		&job.Name,
		&job.Schedule,
		&job.TargetURL,
		&job.TargetMethod,
		&job.TargetPayload,
		&status,
		&job.LastRunAt,
		&job.NextRunAt,
		&job.CreatedAt,
		&job.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	job.Status = domain.CronStatus(status)
	return &job, nil
}

func (r *PostgresCronRepository) scanCronJobs(rows pgx.Rows) ([]*domain.CronJob, error) {
	defer rows.Close()
	var jobs []*domain.CronJob
	for rows.Next() {
		job, err := r.scanCronJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (r *PostgresCronRepository) SaveJobRun(ctx context.Context, run *domain.CronJobRun) error {
	query := `INSERT INTO cron_job_runs (id, job_id, status, status_code, response, duration_ms, started_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(ctx, query, run.ID, run.JobID, run.Status, run.StatusCode, run.Response, run.DurationMs, run.StartedAt)
	return err
}
