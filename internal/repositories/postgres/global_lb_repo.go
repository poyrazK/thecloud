// Package postgres provides PostgreSQL implementations of the platform's repositories.
package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/errors"
)

type globalLBRepository struct {
	db DB
}

// NewGlobalLBRepository returns a new instance of globalLBRepository using the provided DB.
func NewGlobalLBRepository(db DB) *globalLBRepository {
	return &globalLBRepository{db: db}
}

func (r *globalLBRepository) Create(ctx context.Context, glb *domain.GlobalLoadBalancer) error {
	query := `
		INSERT INTO global_load_balancers (
			id, user_id, tenant_id, name, hostname, policy, 
			health_check_protocol, health_check_port, health_check_path,
			health_check_interval, health_check_timeout, health_check_healthy_count, health_check_unhealthy_count,
			status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`
	_, err := r.db.Exec(ctx, query,
		glb.ID, glb.UserID, glb.TenantID, glb.Name, glb.Hostname, glb.Policy,
		glb.HealthCheck.Protocol, glb.HealthCheck.Port, glb.HealthCheck.Path,
		glb.HealthCheck.IntervalSec, glb.HealthCheck.TimeoutSec, glb.HealthCheck.HealthyCount, glb.HealthCheck.UnhealthyCount,
		glb.Status, glb.CreatedAt, glb.UpdatedAt,
	)
	return err
}

func (r *globalLBRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.GlobalLoadBalancer, error) {
	query := `
		SELECT id, user_id, tenant_id, name, hostname, policy, 
		health_check_protocol, health_check_port, health_check_path,
		health_check_interval, health_check_timeout, health_check_healthy_count, health_check_unhealthy_count,
		status, created_at, updated_at
		FROM global_load_balancers WHERE id = $1
	`

	// Implementation Note: Manual field mapping is utilized here to ensure
	// schema robustness and decoupled domain-to-storage mapping.
	row := r.db.QueryRow(ctx, query, id)
	return scanGlobalLB(row)
}

func (r *globalLBRepository) GetByHostname(ctx context.Context, hostname string) (*domain.GlobalLoadBalancer, error) {
	query := `
		SELECT id, user_id, tenant_id, name, hostname, policy, 
		health_check_protocol, health_check_port, health_check_path,
		health_check_interval, health_check_timeout, health_check_healthy_count, health_check_unhealthy_count,
		status, created_at, updated_at
		FROM global_load_balancers WHERE hostname = $1
	`
	row := r.db.QueryRow(ctx, query, hostname)
	return scanGlobalLB(row)
}

func (r *globalLBRepository) List(ctx context.Context, userID uuid.UUID) ([]*domain.GlobalLoadBalancer, error) {
	query := `
		SELECT id, user_id, tenant_id, name, hostname, policy, 
		health_check_protocol, health_check_port, health_check_path,
		health_check_interval, health_check_timeout, health_check_healthy_count, health_check_unhealthy_count,
		status, created_at, updated_at
		FROM global_load_balancers WHERE user_id = $1
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.GlobalLoadBalancer
	for rows.Next() {
		glb, err := scanGlobalLB(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, glb)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(errors.Internal, "iteration error", err)
	}

	return list, nil
}

func (r *globalLBRepository) Update(ctx context.Context, glb *domain.GlobalLoadBalancer) error {
	query := `
		UPDATE global_load_balancers SET
			name = $1, policy = $2, 
			health_check_protocol = $3, health_check_port = $4, health_check_path = $5,
			health_check_interval = $6, health_check_timeout = $7, 
			health_check_healthy_count = $8, health_check_unhealthy_count = $9,
			status = $10, updated_at = $11
		WHERE id = $12
	`
	_, err := r.db.Exec(ctx, query,
		glb.Name, glb.Policy,
		glb.HealthCheck.Protocol, glb.HealthCheck.Port, glb.HealthCheck.Path,
		glb.HealthCheck.IntervalSec, glb.HealthCheck.TimeoutSec,
		glb.HealthCheck.HealthyCount, glb.HealthCheck.UnhealthyCount,
		glb.Status, glb.UpdatedAt, glb.ID,
	)
	return err
}

func (r *globalLBRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	query := `DELETE FROM global_load_balancers WHERE id = $1 AND user_id = $2`
	_, err := r.db.Exec(ctx, query, id, userID)
	return err
}

func (r *globalLBRepository) AddEndpoint(ctx context.Context, ep *domain.GlobalEndpoint) error {
	query := `
		INSERT INTO global_lb_endpoints (
			id, global_lb_id, region, target_type, target_id, target_ip,
			weight, priority, healthy, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Exec(ctx, query,
		ep.ID, ep.GlobalLBID, ep.Region, ep.TargetType, ep.TargetID, ep.TargetIP,
		ep.Weight, ep.Priority, ep.Healthy, ep.CreatedAt,
	)
	return err
}

func (r *globalLBRepository) RemoveEndpoint(ctx context.Context, endpointID uuid.UUID) error {
	query := `DELETE FROM global_lb_endpoints WHERE id = $1`
	_, err := r.db.Exec(ctx, query, endpointID)
	return err
}

func (r *globalLBRepository) GetEndpointByID(ctx context.Context, endpointID uuid.UUID) (*domain.GlobalEndpoint, error) {
	query := `
		SELECT id, global_lb_id, region, target_type, target_id, HOST(target_ip),
		       weight, priority, healthy, last_health_check, created_at
		FROM global_lb_endpoints WHERE id = $1
	`
	row := r.db.QueryRow(ctx, query, endpointID)
	return scanEndpoint(row)
}

func (r *globalLBRepository) ListEndpoints(ctx context.Context, glbID uuid.UUID) ([]*domain.GlobalEndpoint, error) {
	query := `
		SELECT id, global_lb_id, region, target_type, target_id, HOST(target_ip),
		       weight, priority, healthy, last_health_check, created_at
		FROM global_lb_endpoints WHERE global_lb_id = $1
	`
	rows, err := r.db.Query(ctx, query, glbID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.GlobalEndpoint
	for rows.Next() {
		ep, err := scanEndpoint(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, ep)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(errors.Internal, "iteration error", err)
	}

	return list, nil
}

func (r *globalLBRepository) UpdateEndpointHealth(ctx context.Context, epID uuid.UUID, healthy bool) error {
	query := `UPDATE global_lb_endpoints SET healthy = $1, last_health_check = $2 WHERE id = $3`
	_, err := r.db.Exec(ctx, query, healthy, time.Now(), epID)
	return err
}

// Helpers

type scanner interface {
	Scan(dest ...interface{}) error
}

func scanGlobalLB(s scanner) (*domain.GlobalLoadBalancer, error) {
	var glb domain.GlobalLoadBalancer
	var hc domain.GlobalHealthCheckConfig
	// Explicit field mapping is required to ensure consistent row scanning.
	// Column order is assumed based on the current table definition.
	err := s.Scan(
		&glb.ID, &glb.UserID, &glb.TenantID, &glb.Name, &glb.Hostname, &glb.Policy,
		&hc.Protocol, &hc.Port, &hc.Path, &hc.IntervalSec, &hc.TimeoutSec, &hc.HealthyCount, &hc.UnhealthyCount,
		&glb.Status, &glb.CreatedAt, &glb.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New(errors.NotFound, "global load balancer not found")
		}
		return nil, errors.Wrap(errors.Internal, "failed to scan global load balancer", err)
	}
	glb.HealthCheck = hc
	return &glb, nil
}

func scanEndpoint(s scanner) (*domain.GlobalEndpoint, error) {
	var ep domain.GlobalEndpoint
	var lastHealthCheck *time.Time
	// id, glb_id, region, type, tid, tip, weight, prio, healthy, last_hc, created
	err := s.Scan(
		&ep.ID, &ep.GlobalLBID, &ep.Region, &ep.TargetType, &ep.TargetID, &ep.TargetIP,
		&ep.Weight, &ep.Priority, &ep.Healthy, &lastHealthCheck, &ep.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New(errors.NotFound, "endpoint not found")
		}
		return nil, errors.Wrap(errors.Internal, "failed to scan endpoint", err)
	}
	if lastHealthCheck != nil {
		ep.LastHealthCheck = *lastHealthCheck
	}
	return &ep, nil
}
