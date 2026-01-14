// Package postgres provides PostgreSQL-backed repository implementations.
package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/errors"
)

// LBRepository provides PostgreSQL-backed load balancer persistence.
type LBRepository struct {
	db DB
}

// NewLBRepository creates an LBRepository using the provided DB.
func NewLBRepository(db DB) *LBRepository {
	return &LBRepository{db: db}
}

func (r *LBRepository) Create(ctx context.Context, lb *domain.LoadBalancer) error {
	query := `
		INSERT INTO load_balancers (id, user_id, idempotency_key, name, vpc_id, port, algorithm, status, version, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Exec(ctx, query,
		lb.ID, lb.UserID, lb.IdempotencyKey, lb.Name, lb.VpcID, lb.Port, lb.Algorithm, lb.Status, lb.Version, lb.CreatedAt,
	)
	if err != nil {
		// Check for unique constraint violation on idempotency_key
		return errors.Wrap(errors.Internal, "failed to create load balancer", err)
	}
	return nil
}

func (r *LBRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.LoadBalancer, error) {
	userID := appcontext.UserIDFromContext(ctx)
	query := `
		SELECT id, user_id, COALESCE(idempotency_key, ''), name, vpc_id, port, algorithm, status, version, created_at
		FROM load_balancers
		WHERE id = $1 AND user_id = $2
	`
	return r.scanLB(r.db.QueryRow(ctx, query, id, userID))
}

func (r *LBRepository) GetByIdempotencyKey(ctx context.Context, key string) (*domain.LoadBalancer, error) {
	if key == "" {
		return nil, errors.New(errors.NotFound, "idempotency key empty")
	}
	userID := appcontext.UserIDFromContext(ctx)
	query := `
		SELECT id, user_id, idempotency_key, name, vpc_id, port, algorithm, status, version, created_at
		FROM load_balancers
		WHERE idempotency_key = $1 AND user_id = $2
	`
	return r.scanLB(r.db.QueryRow(ctx, query, key, userID))
}

func (r *LBRepository) List(ctx context.Context) ([]*domain.LoadBalancer, error) {
	userID := appcontext.UserIDFromContext(ctx)
	query := `
		SELECT id, user_id, COALESCE(idempotency_key, ''), name, vpc_id, port, algorithm, status, version, created_at
		FROM load_balancers
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to list load balancers", err)
	}
	return r.scanLBs(rows)
}

func (r *LBRepository) ListAll(ctx context.Context) ([]*domain.LoadBalancer, error) {
	query := `
		SELECT id, user_id, COALESCE(idempotency_key, ''), name, vpc_id, port, algorithm, status, version, created_at
		FROM load_balancers
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to list all load balancers", err)
	}
	return r.scanLBs(rows)
}

func (r *LBRepository) scanLB(row pgx.Row) (*domain.LoadBalancer, error) {
	var lb domain.LoadBalancer
	var status string
	err := row.Scan(
		&lb.ID, &lb.UserID, &lb.IdempotencyKey, &lb.Name, &lb.VpcID, &lb.Port, &lb.Algorithm, &status, &lb.Version, &lb.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrLBNotFound
		}
		return nil, errors.Wrap(errors.Internal, "failed to scan load balancer", err)
	}
	lb.Status = domain.LBStatus(status)
	return &lb, nil
}

func (r *LBRepository) scanLBs(rows pgx.Rows) ([]*domain.LoadBalancer, error) {
	defer rows.Close()
	var lbs []*domain.LoadBalancer
	for rows.Next() {
		lb, err := r.scanLB(rows)
		if err != nil {
			return nil, err
		}
		lbs = append(lbs, lb)
	}
	return lbs, nil
}

func (r *LBRepository) Update(ctx context.Context, lb *domain.LoadBalancer) error {
	query := `
		UPDATE load_balancers
		SET name = $1, port = $2, algorithm = $3, status = $4, version = version + 1
		WHERE id = $5 AND version = $6 AND user_id = $7
	`
	cmd, err := r.db.Exec(ctx, query, lb.Name, lb.Port, lb.Algorithm, lb.Status, lb.ID, lb.Version, lb.UserID)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to update load balancer", err)
	}
	if cmd.RowsAffected() == 0 {
		return errors.New(errors.Conflict, "update conflict: load balancer was modified or not found")
	}
	lb.Version++
	return nil
}

func (r *LBRepository) Delete(ctx context.Context, id uuid.UUID) error {
	userID := appcontext.UserIDFromContext(ctx)
	query := `DELETE FROM load_balancers WHERE id = $1 AND user_id = $2`
	cmd, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to delete load balancer", err)
	}
	if cmd.RowsAffected() == 0 {
		return errors.ErrLBNotFound
	}
	return nil
}

func (r *LBRepository) AddTarget(ctx context.Context, target *domain.LBTarget) error {
	query := `
		INSERT INTO lb_targets (id, lb_id, instance_id, port, weight, health)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query,
		target.ID, target.LBID, target.InstanceID, target.Port, target.Weight, target.Health,
	)
	if err != nil {
		// Handle unique constraint on (lb_id, instance_id)
		return errors.Wrap(errors.Internal, "failed to add load balancer target", err)
	}
	return nil
}

func (r *LBRepository) RemoveTarget(ctx context.Context, lbID, instanceID uuid.UUID) error {
	query := `DELETE FROM lb_targets WHERE lb_id = $1 AND instance_id = $2`
	cmd, err := r.db.Exec(ctx, query, lbID, instanceID)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to remove load balancer target", err)
	}
	if cmd.RowsAffected() == 0 {
		return errors.New(errors.NotFound, "target not found")
	}
	return nil
}

func (r *LBRepository) ListTargets(ctx context.Context, lbID uuid.UUID) ([]*domain.LBTarget, error) {
	query := `
		SELECT id, lb_id, instance_id, port, weight, health
		FROM lb_targets
		WHERE lb_id = $1
	`
	rows, err := r.db.Query(ctx, query, lbID)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to list load balancer targets", err)
	}
	return r.scanTargets(rows)
}

func (r *LBRepository) UpdateTargetHealth(ctx context.Context, lbID, instanceID uuid.UUID, health string) error {
	query := `
		UPDATE lb_targets
		SET health = $1
		WHERE lb_id = $2 AND instance_id = $3
	`
	_, err := r.db.Exec(ctx, query, health, lbID, instanceID)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to update target health", err)
	}
	return nil
}

func (r *LBRepository) GetTargetsForInstance(ctx context.Context, instanceID uuid.UUID) ([]*domain.LBTarget, error) {
	query := `
		SELECT id, lb_id, instance_id, port, weight, health
		FROM lb_targets
		WHERE instance_id = $1
	`
	rows, err := r.db.Query(ctx, query, instanceID)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to get targets for instance", err)
	}
	return r.scanTargets(rows)
}

func (r *LBRepository) scanTarget(row pgx.Row) (*domain.LBTarget, error) {
	var t domain.LBTarget
	err := row.Scan(&t.ID, &t.LBID, &t.InstanceID, &t.Port, &t.Weight, &t.Health)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to scan load balancer target", err)
	}
	return &t, nil
}

func (r *LBRepository) scanTargets(rows pgx.Rows) ([]*domain.LBTarget, error) {
	defer rows.Close()
	var targets []*domain.LBTarget
	for rows.Next() {
		t, err := r.scanTarget(rows)
		if err != nil {
			return nil, err
		}
		targets = append(targets, t)
	}
	return targets, nil
}
