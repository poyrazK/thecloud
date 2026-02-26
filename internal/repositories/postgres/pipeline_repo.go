// Package postgres provides PostgreSQL-backed repository implementations.
package postgres

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	stdlib_errors "errors"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

// PipelineRepository provides PostgreSQL-backed pipeline persistence.
type PipelineRepository struct {
	db DB
}

// NewPipelineRepository creates a new pipeline repository.
func NewPipelineRepository(db DB) ports.PipelineRepository {
	return &PipelineRepository{db: db}
}

func (r *PipelineRepository) CreatePipeline(ctx context.Context, pipeline *domain.Pipeline) error {
	configJSON, err := json.Marshal(pipeline.Config)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO pipelines (id, user_id, name, repository_url, branch, webhook_secret, config, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err = r.db.Exec(ctx, query,
		pipeline.ID,
		pipeline.UserID,
		pipeline.Name,
		pipeline.RepositoryURL,
		pipeline.Branch,
		pipeline.WebhookSecret,
		configJSON,
		pipeline.Status,
		pipeline.CreatedAt,
		pipeline.UpdatedAt,
	)
	return err
}

func (r *PipelineRepository) GetPipelineByID(ctx context.Context, id uuid.UUID) (*domain.Pipeline, error) {
	query := `
		SELECT id, user_id, name, repository_url, branch, webhook_secret, config, status, created_at, updated_at
		FROM pipelines
		WHERE id = $1
	`
	return r.scanPipeline(r.db.QueryRow(ctx, query, id))
}

func (r *PipelineRepository) GetPipeline(ctx context.Context, id, userID uuid.UUID) (*domain.Pipeline, error) {
	query := `
		SELECT id, user_id, name, repository_url, branch, webhook_secret, config, status, created_at, updated_at
		FROM pipelines
		WHERE id = $1 AND user_id = $2
	`
	return r.scanPipeline(r.db.QueryRow(ctx, query, id, userID))
}

func (r *PipelineRepository) ListPipelines(ctx context.Context, userID uuid.UUID) ([]*domain.Pipeline, error) {
	query := `
		SELECT id, user_id, name, repository_url, branch, webhook_secret, config, status, created_at, updated_at
		FROM pipelines
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	return r.scanPipelines(rows)
}

func (r *PipelineRepository) UpdatePipeline(ctx context.Context, pipeline *domain.Pipeline) error {
	configJSON, err := json.Marshal(pipeline.Config)
	if err != nil {
		return err
	}

	query := `
		UPDATE pipelines
		SET name = $1,
			repository_url = $2,
			branch = $3,
			webhook_secret = $4,
			config = $5,
			status = $6,
			updated_at = NOW()
		WHERE id = $7 AND user_id = $8
	`
	_, err = r.db.Exec(ctx, query,
		pipeline.Name,
		pipeline.RepositoryURL,
		pipeline.Branch,
		pipeline.WebhookSecret,
		configJSON,
		pipeline.Status,
		pipeline.ID,
		pipeline.UserID,
	)
	return err
}

func (r *PipelineRepository) DeletePipeline(ctx context.Context, id, userID uuid.UUID) error {
	query := `DELETE FROM pipelines WHERE id = $1 AND user_id = $2`
	_, err := r.db.Exec(ctx, query, id, userID)
	return err
}

func (r *PipelineRepository) CreateBuild(ctx context.Context, build *domain.Build) error {
	query := `
		INSERT INTO builds (id, pipeline_id, user_id, commit_hash, trigger_type, status, started_at, finished_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Exec(ctx, query,
		build.ID,
		build.PipelineID,
		build.UserID,
		build.CommitHash,
		build.TriggerType,
		build.Status,
		build.StartedAt,
		build.FinishedAt,
		build.CreatedAt,
		build.UpdatedAt,
	)
	return err
}

func (r *PipelineRepository) GetBuild(ctx context.Context, id, userID uuid.UUID) (*domain.Build, error) {
	query := `
		SELECT id, pipeline_id, user_id, commit_hash, trigger_type, status, started_at, finished_at, created_at, updated_at
		FROM builds
		WHERE id = $1 AND user_id = $2
	`
	return r.scanBuild(r.db.QueryRow(ctx, query, id, userID))
}

func (r *PipelineRepository) ListBuildsByPipeline(ctx context.Context, pipelineID, userID uuid.UUID) ([]*domain.Build, error) {
	query := `
		SELECT id, pipeline_id, user_id, commit_hash, trigger_type, status, started_at, finished_at, created_at, updated_at
		FROM builds
		WHERE pipeline_id = $1 AND user_id = $2
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, pipelineID, userID)
	if err != nil {
		return nil, err
	}
	return r.scanBuilds(rows)
}

func (r *PipelineRepository) UpdateBuild(ctx context.Context, build *domain.Build) error {
	query := `
		UPDATE builds
		SET status = $1,
			started_at = $2,
			finished_at = $3,
			updated_at = NOW()
		WHERE id = $4 AND user_id = $5
	`
	_, err := r.db.Exec(ctx, query, build.Status, build.StartedAt, build.FinishedAt, build.ID, build.UserID)
	return err
}

func (r *PipelineRepository) CreateBuildStep(ctx context.Context, step *domain.BuildStep) error {
	query := `
		INSERT INTO build_steps (id, build_id, name, image, commands, status, exit_code, started_at, finished_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Exec(ctx, query,
		step.ID,
		step.BuildID,
		step.Name,
		step.Image,
		step.Commands,
		step.Status,
		step.ExitCode,
		step.StartedAt,
		step.FinishedAt,
		step.CreatedAt,
		step.UpdatedAt,
	)
	return err
}

func (r *PipelineRepository) ListBuildSteps(ctx context.Context, buildID, userID uuid.UUID) ([]*domain.BuildStep, error) {
	query := `
		SELECT s.id, s.build_id, s.name, s.image, s.commands, s.status, s.exit_code, s.started_at, s.finished_at, s.created_at, s.updated_at
		FROM build_steps s
		INNER JOIN builds b ON b.id = s.build_id
		WHERE s.build_id = $1 AND b.user_id = $2
		ORDER BY s.created_at ASC
	`
	rows, err := r.db.Query(ctx, query, buildID, userID)
	if err != nil {
		return nil, err
	}
	return r.scanBuildSteps(rows)
}

func (r *PipelineRepository) UpdateBuildStep(ctx context.Context, step *domain.BuildStep) error {
	query := `
		UPDATE build_steps
		SET status = $1,
			exit_code = $2,
			started_at = $3,
			finished_at = $4,
			updated_at = NOW()
		WHERE id = $5
	`
	_, err := r.db.Exec(ctx, query, step.Status, step.ExitCode, step.StartedAt, step.FinishedAt, step.ID)
	return err
}

func (r *PipelineRepository) AppendBuildLog(ctx context.Context, log *domain.BuildLog) error {
	query := `
		INSERT INTO build_logs (id, build_id, step_id, content, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(ctx, query, log.ID, log.BuildID, log.StepID, log.Content, log.CreatedAt)
	return err
}

func (r *PipelineRepository) ListBuildLogs(ctx context.Context, buildID, userID uuid.UUID, limit int) ([]*domain.BuildLog, error) {
	if limit <= 0 {
		limit = 200
	}

	query := `
		SELECT l.id, l.build_id, l.step_id, l.content, l.created_at
		FROM build_logs l
		INNER JOIN builds b ON b.id = l.build_id
		WHERE l.build_id = $1 AND b.user_id = $2
		ORDER BY l.created_at ASC
		LIMIT $3
	`
	rows, err := r.db.Query(ctx, query, buildID, userID, limit)
	if err != nil {
		return nil, err
	}
	return r.scanBuildLogs(rows)
}

func (r *PipelineRepository) ReserveWebhookDelivery(ctx context.Context, pipelineID uuid.UUID, provider, event, deliveryID string) (bool, error) {
	query := `
		INSERT INTO pipeline_webhook_deliveries (id, pipeline_id, provider, event, delivery_id, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`
	_, err := r.db.Exec(ctx, query, uuid.New(), pipelineID, provider, event, deliveryID)
	if err != nil {
		var pgErr *pgconn.PgError
		if stdlib_errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *PipelineRepository) scanPipeline(row pgx.Row) (*domain.Pipeline, error) {
	var pipeline domain.Pipeline
	var configJSON []byte
	var status string

	err := row.Scan(
		&pipeline.ID,
		&pipeline.UserID,
		&pipeline.Name,
		&pipeline.RepositoryURL,
		&pipeline.Branch,
		&pipeline.WebhookSecret,
		&configJSON,
		&status,
		&pipeline.CreatedAt,
		&pipeline.UpdatedAt,
	)
	if err != nil {
		if stdlib_errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	pipeline.Status = domain.PipelineStatus(status)
	if len(configJSON) > 0 {
		if err := json.Unmarshal(configJSON, &pipeline.Config); err != nil {
			return nil, err
		}
	}

	return &pipeline, nil
}

func (r *PipelineRepository) scanPipelines(rows pgx.Rows) ([]*domain.Pipeline, error) {
	defer rows.Close()
	var pipelines []*domain.Pipeline
	for rows.Next() {
		pipeline, err := r.scanPipeline(rows)
		if err != nil {
			return nil, err
		}
		pipelines = append(pipelines, pipeline)
	}
	return pipelines, nil
}

func (r *PipelineRepository) scanBuild(row pgx.Row) (*domain.Build, error) {
	var build domain.Build
	var triggerType string
	var status string

	err := row.Scan(
		&build.ID,
		&build.PipelineID,
		&build.UserID,
		&build.CommitHash,
		&triggerType,
		&status,
		&build.StartedAt,
		&build.FinishedAt,
		&build.CreatedAt,
		&build.UpdatedAt,
	)
	if err != nil {
		if stdlib_errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	build.TriggerType = domain.BuildTriggerType(triggerType)
	build.Status = domain.BuildStatus(status)
	return &build, nil
}

func (r *PipelineRepository) scanBuilds(rows pgx.Rows) ([]*domain.Build, error) {
	defer rows.Close()
	var builds []*domain.Build
	for rows.Next() {
		build, err := r.scanBuild(rows)
		if err != nil {
			return nil, err
		}
		builds = append(builds, build)
	}
	return builds, nil
}

func (r *PipelineRepository) scanBuildStep(row pgx.Row) (*domain.BuildStep, error) {
	var step domain.BuildStep
	var status string

	err := row.Scan(
		&step.ID,
		&step.BuildID,
		&step.Name,
		&step.Image,
		&step.Commands,
		&status,
		&step.ExitCode,
		&step.StartedAt,
		&step.FinishedAt,
		&step.CreatedAt,
		&step.UpdatedAt,
	)
	if err != nil {
		if stdlib_errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	step.Status = domain.BuildStatus(status)
	return &step, nil
}

func (r *PipelineRepository) scanBuildSteps(rows pgx.Rows) ([]*domain.BuildStep, error) {
	defer rows.Close()
	var steps []*domain.BuildStep
	for rows.Next() {
		step, err := r.scanBuildStep(rows)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}
	return steps, nil
}

func (r *PipelineRepository) scanBuildLog(row pgx.Row) (*domain.BuildLog, error) {
	var log domain.BuildLog
	err := row.Scan(&log.ID, &log.BuildID, &log.StepID, &log.Content, &log.CreatedAt)
	if err != nil {
		if stdlib_errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &log, nil
}

func (r *PipelineRepository) scanBuildLogs(rows pgx.Rows) ([]*domain.BuildLog, error) {
	defer rows.Close()
	var logs []*domain.BuildLog
	for rows.Next() {
		entry, err := r.scanBuildLog(rows)
		if err != nil {
			return nil, err
		}
		logs = append(logs, entry)
	}
	return logs, nil
}
