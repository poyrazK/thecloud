// Package postgres provides PostgreSQL-backed repository implementations.
package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

type stackRepository struct {
	db DB
}

// NewStackRepository creates a stack repository using the provided DB.
func NewStackRepository(db DB) *stackRepository {
	return &stackRepository{db: db}
}

func (r *stackRepository) Create(ctx context.Context, s *domain.Stack) error {
	_, err := r.db.Exec(ctx,
		"INSERT INTO stacks (id, user_id, name, template, parameters, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		s.ID, s.UserID, s.Name, s.Template, s.Parameters, string(s.Status), s.CreatedAt, s.UpdatedAt)
	return err
}

func (r *stackRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Stack, error) {
	s, err := r.scanStack(r.db.QueryRow(ctx,
		"SELECT id, user_id, name, template, parameters, status, status_reason, created_at, updated_at FROM stacks WHERE id = $1",
		id))
	if err != nil {
		return nil, err
	}

	resources, err := r.ListResources(ctx, id)
	if err == nil {
		s.Resources = resources
	}

	return s, nil
}

func (r *stackRepository) GetByName(ctx context.Context, userID uuid.UUID, name string) (*domain.Stack, error) {
	return r.scanStack(r.db.QueryRow(ctx,
		"SELECT id, user_id, name, template, parameters, status, status_reason, created_at, updated_at FROM stacks WHERE user_id = $1 AND name = $2",
		userID, name))
}

func (r *stackRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Stack, error) {
	rows, err := r.db.Query(ctx,
		"SELECT id, user_id, name, template, parameters, status, status_reason, created_at, updated_at FROM stacks WHERE user_id = $1 ORDER BY created_at DESC",
		userID)
	if err != nil {
		return nil, err
	}
	return r.scanStacks(rows)
}

func (r *stackRepository) scanStack(row pgx.Row) (*domain.Stack, error) {
	s := &domain.Stack{}
	var status string
	err := row.Scan(&s.ID, &s.UserID, &s.Name, &s.Template, &s.Parameters, &status, &s.StatusReason, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	s.Status = domain.StackStatus(status)
	return s, nil
}

func (r *stackRepository) scanStacks(rows pgx.Rows) ([]*domain.Stack, error) {
	defer rows.Close()
	var stacks []*domain.Stack
	for rows.Next() {
		s, err := r.scanStack(rows)
		if err != nil {
			return nil, err
		}
		stacks = append(stacks, s)
	}
	return stacks, nil
}

func (r *stackRepository) Update(ctx context.Context, s *domain.Stack) error {
	_, err := r.db.Exec(ctx,
		"UPDATE stacks SET status = $1, status_reason = $2, updated_at = $3 WHERE id = $4",
		string(s.Status), s.StatusReason, s.UpdatedAt, s.ID)
	return err
}

func (r *stackRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM stacks WHERE id = $1", id)
	return err
}

func (r *stackRepository) AddResource(ctx context.Context, res *domain.StackResource) error {
	_, err := r.db.Exec(ctx,
		"INSERT INTO stack_resources (id, stack_id, logical_id, physical_id, resource_type, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		res.ID, res.StackID, res.LogicalID, res.PhysicalID, res.ResourceType, res.Status, res.CreatedAt)
	return err
}

func (r *stackRepository) ListResources(ctx context.Context, stackID uuid.UUID) ([]domain.StackResource, error) {
	rows, err := r.db.Query(ctx,
		"SELECT id, stack_id, logical_id, physical_id, resource_type, status, created_at FROM stack_resources WHERE stack_id = $1",
		stackID)
	if err != nil {
		return nil, err
	}
	return r.scanStackResources(rows)
}

func (r *stackRepository) scanStackResource(row pgx.Row) (domain.StackResource, error) {
	res := domain.StackResource{}
	err := row.Scan(&res.ID, &res.StackID, &res.LogicalID, &res.PhysicalID, &res.ResourceType, &res.Status, &res.CreatedAt)
	if err != nil {
		return domain.StackResource{}, err
	}
	return res, nil
}

func (r *stackRepository) scanStackResources(rows pgx.Rows) ([]domain.StackResource, error) {
	defer rows.Close()
	var resources []domain.StackResource
	for rows.Next() {
		res, err := r.scanStackResource(rows)
		if err != nil {
			return nil, err
		}
		resources = append(resources, res)
	}
	return resources, nil
}

func (r *stackRepository) DeleteResources(ctx context.Context, stackID uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM stack_resources WHERE stack_id = $1", stackID)
	return err
}
