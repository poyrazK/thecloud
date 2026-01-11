package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/errors"
)

type FunctionRepository struct {
	db DB
}

func NewFunctionRepository(db DB) *FunctionRepository {
	return &FunctionRepository{db: db}
}

func (r *FunctionRepository) Create(ctx context.Context, f *domain.Function) error {
	query := `
		INSERT INTO functions (id, user_id, name, runtime, handler, code_path, timeout_seconds, memory_mb, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Exec(ctx, query,
		f.ID, f.UserID, f.Name, f.Runtime, f.Handler, f.CodePath, f.Timeout, f.MemoryMB, f.Status, f.CreatedAt, f.UpdatedAt,
	)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to create function", err)
	}
	return nil
}

func (r *FunctionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Function, error) {
	query := `SELECT id, user_id, name, runtime, handler, code_path, timeout_seconds, memory_mb, status, created_at, updated_at FROM functions WHERE id = $1`
	return r.scanFunction(r.db.QueryRow(ctx, query, id))
}

func (r *FunctionRepository) GetByName(ctx context.Context, userID uuid.UUID, name string) (*domain.Function, error) {
	query := `SELECT id, user_id, name, runtime, handler, code_path, timeout_seconds, memory_mb, status, created_at, updated_at FROM functions WHERE user_id = $1 AND name = $2`
	return r.scanFunction(r.db.QueryRow(ctx, query, userID, name))
}

func (r *FunctionRepository) List(ctx context.Context, userID uuid.UUID) ([]*domain.Function, error) {
	query := `SELECT id, user_id, name, runtime, handler, code_path, timeout_seconds, memory_mb, status, created_at, updated_at FROM functions WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to list functions", err)
	}
	return r.scanFunctions(rows)
}

func (r *FunctionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM functions WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to delete function", err)
	}
	return nil
}

func (r *FunctionRepository) CreateInvocation(ctx context.Context, i *domain.Invocation) error {
	query := `
		INSERT INTO invocations (id, function_id, status, started_at, ended_at, duration_ms, status_code, logs)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(ctx, query,
		i.ID, i.FunctionID, i.Status, i.StartedAt, i.EndedAt, i.DurationMs, i.StatusCode, i.Logs,
	)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to create invocation", err)
	}
	return nil
}

func (r *FunctionRepository) GetInvocations(ctx context.Context, functionID uuid.UUID, limit int) ([]*domain.Invocation, error) {
	query := `SELECT id, function_id, status, started_at, ended_at, duration_ms, status_code, logs FROM invocations WHERE function_id = $1 ORDER BY started_at DESC LIMIT $2`
	rows, err := r.db.Query(ctx, query, functionID, limit)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to get invocations", err)
	}
	return r.scanInvocations(rows)
}

func (r *FunctionRepository) scanFunction(row pgx.Row) (*domain.Function, error) {
	f := &domain.Function{}
	err := row.Scan(&f.ID, &f.UserID, &f.Name, &f.Runtime, &f.Handler, &f.CodePath, &f.Timeout, &f.MemoryMB, &f.Status, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return nil, errors.Wrap(errors.NotFound, "function not found", err)
	}
	return f, nil
}

func (r *FunctionRepository) scanFunctions(rows pgx.Rows) ([]*domain.Function, error) {
	defer rows.Close()
	var functions []*domain.Function
	for rows.Next() {
		f, err := r.scanFunction(rows)
		if err != nil {
			return nil, err
		}
		functions = append(functions, f)
	}
	return functions, nil
}

func (r *FunctionRepository) scanInvocation(row pgx.Row) (*domain.Invocation, error) {
	i := &domain.Invocation{}
	err := row.Scan(&i.ID, &i.FunctionID, &i.Status, &i.StartedAt, &i.EndedAt, &i.DurationMs, &i.StatusCode, &i.Logs)
	if err != nil {
		return nil, err
	}
	return i, nil
}

func (r *FunctionRepository) scanInvocations(rows pgx.Rows) ([]*domain.Invocation, error) {
	defer rows.Close()
	var invocations []*domain.Invocation
	for rows.Next() {
		i, err := r.scanInvocation(rows)
		if err != nil {
			return nil, err
		}
		invocations = append(invocations, i)
	}
	return invocations, nil
}
