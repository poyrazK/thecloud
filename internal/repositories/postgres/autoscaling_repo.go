package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/poyraz/cloud/internal/core/domain"
	errs "github.com/poyraz/cloud/internal/errors"
)

type AutoScalingRepo struct {
	db *pgxpool.Pool
}

func NewAutoScalingRepo(db *pgxpool.Pool) *AutoScalingRepo {
	return &AutoScalingRepo{db: db}
}

// Scaling Groups

func (r *AutoScalingRepo) CreateGroup(ctx context.Context, group *domain.ScalingGroup) error {
	query := `
		INSERT INTO scaling_groups (
			id, idempotency_key, name, vpc_id, load_balancer_id, image, ports,
			min_instances, max_instances, desired_count, current_count, status, version, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`
	var idempotencyKey interface{}
	if group.IdempotencyKey != "" {
		idempotencyKey = group.IdempotencyKey
	}

	_, err := r.db.Exec(ctx, query,
		group.ID, idempotencyKey, group.Name, group.VpcID, group.LoadBalancerID,
		group.Image, group.Ports, group.MinInstances, group.MaxInstances,
		group.DesiredCount, group.CurrentCount, group.Status, group.Version,
		group.CreatedAt, group.UpdatedAt,
	)
	return err
}

func (r *AutoScalingRepo) GetGroupByID(ctx context.Context, id uuid.UUID) (*domain.ScalingGroup, error) {
	query := `
		SELECT id, idempotency_key, name, vpc_id, load_balancer_id, image, ports,
			   min_instances, max_instances, desired_count, current_count, status, version, created_at, updated_at
		FROM scaling_groups WHERE id = $1
	`
	var g domain.ScalingGroup
	var lbID *uuid.UUID
	var ports sql.NullString
	var idk sql.NullString

	err := r.db.QueryRow(ctx, query, id).Scan(
		&g.ID, &idk, &g.Name, &g.VpcID, &lbID, &g.Image, &ports,
		&g.MinInstances, &g.MaxInstances, &g.DesiredCount, &g.CurrentCount,
		&g.Status, &g.Version, &g.CreatedAt, &g.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, errs.New(errs.NotFound, "scaling group not found")
	}
	if err != nil {
		return nil, err
	}
	g.LoadBalancerID = lbID
	if ports.Valid {
		g.Ports = ports.String
	}
	if idk.Valid {
		g.IdempotencyKey = idk.String
	}
	return &g, nil
}

func (r *AutoScalingRepo) GetGroupByIdempotencyKey(ctx context.Context, key string) (*domain.ScalingGroup, error) {
	query := `
		SELECT id, idempotency_key, name, vpc_id, load_balancer_id, image, ports,
			   min_instances, max_instances, desired_count, current_count, status, version, created_at, updated_at
		FROM scaling_groups WHERE idempotency_key = $1
	`
	var g domain.ScalingGroup
	var lbID *uuid.UUID
	var ports sql.NullString
	var idk sql.NullString // Should match key

	err := r.db.QueryRow(ctx, query, key).Scan(
		&g.ID, &idk, &g.Name, &g.VpcID, &lbID, &g.Image, &ports,
		&g.MinInstances, &g.MaxInstances, &g.DesiredCount, &g.CurrentCount,
		&g.Status, &g.Version, &g.CreatedAt, &g.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, errs.New(errs.NotFound, "scaling group not found")
	}
	if err != nil {
		return nil, err
	}
	g.LoadBalancerID = lbID
	if ports.Valid {
		g.Ports = ports.String
	}
	if idk.Valid {
		g.IdempotencyKey = idk.String
	}
	return &g, nil
}

func (r *AutoScalingRepo) ListGroups(ctx context.Context) ([]*domain.ScalingGroup, error) {
	query := `
		SELECT id, idempotency_key, name, vpc_id, load_balancer_id, image, ports,
			   min_instances, max_instances, desired_count, current_count, status, version, created_at, updated_at
		FROM scaling_groups
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*domain.ScalingGroup
	for rows.Next() {
		var g domain.ScalingGroup
		var lbID *uuid.UUID
		var ports sql.NullString
		var idk sql.NullString

		if err := rows.Scan(
			&g.ID, &idk, &g.Name, &g.VpcID, &lbID, &g.Image, &ports,
			&g.MinInstances, &g.MaxInstances, &g.DesiredCount, &g.CurrentCount,
			&g.Status, &g.Version, &g.CreatedAt, &g.UpdatedAt,
		); err != nil {
			return nil, err
		}
		g.LoadBalancerID = lbID
		if ports.Valid {
			g.Ports = ports.String
		}
		if idk.Valid {
			g.IdempotencyKey = idk.String
		}
		groups = append(groups, &g)
	}
	return groups, nil
}

func (r *AutoScalingRepo) CountGroupsByVPC(ctx context.Context, vpcID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM scaling_groups WHERE vpc_id = $1", vpcID).Scan(&count)
	return count, err
}

func (r *AutoScalingRepo) UpdateGroup(ctx context.Context, group *domain.ScalingGroup) error {
	query := `
		UPDATE scaling_groups
		SET desired_count = $1, current_count = $2, status = $3, 
			min_instances = $4, max_instances = $5,
			version = version + 1, updated_at = NOW()
		WHERE id = $6
	`
	_, err := r.db.Exec(ctx, query,
		group.DesiredCount, group.CurrentCount, group.Status,
		group.MinInstances, group.MaxInstances,
		group.ID,
	)
	return err
}

func (r *AutoScalingRepo) DeleteGroup(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM scaling_groups WHERE id = $1", id)
	return err
}

// Policies

func (r *AutoScalingRepo) CreatePolicy(ctx context.Context, policy *domain.ScalingPolicy) error {
	query := `
		INSERT INTO scaling_policies (
			id, scaling_group_id, name, metric_type, target_value,
			scale_out_step, scale_in_step, cooldown_sec, last_scaled_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query,
		policy.ID, policy.ScalingGroupID, policy.Name, policy.MetricType, policy.TargetValue,
		policy.ScaleOutStep, policy.ScaleInStep, policy.CooldownSec, policy.LastScaledAt,
	)
	return err
}

func (r *AutoScalingRepo) GetPoliciesForGroup(ctx context.Context, groupID uuid.UUID) ([]*domain.ScalingPolicy, error) {
	query := `SELECT id, scaling_group_id, name, metric_type, target_value, scale_out_step, scale_in_step, cooldown_sec, last_scaled_at FROM scaling_policies WHERE scaling_group_id = $1`
	rows, err := r.db.Query(ctx, query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []*domain.ScalingPolicy
	for rows.Next() {
		var p domain.ScalingPolicy
		var lastScaledAt sql.NullTime
		if err := rows.Scan(
			&p.ID, &p.ScalingGroupID, &p.Name, &p.MetricType, &p.TargetValue,
			&p.ScaleOutStep, &p.ScaleInStep, &p.CooldownSec, &lastScaledAt,
		); err != nil {
			return nil, err
		}
		if lastScaledAt.Valid {
			t := lastScaledAt.Time
			p.LastScaledAt = &t
		}
		policies = append(policies, &p)
	}
	return policies, nil
}

func (r *AutoScalingRepo) GetAllPolicies(ctx context.Context, groupIDs []uuid.UUID) (map[uuid.UUID][]*domain.ScalingPolicy, error) {
	if len(groupIDs) == 0 {
		return make(map[uuid.UUID][]*domain.ScalingPolicy), nil
	}

	query := `
		SELECT id, scaling_group_id, name, metric_type, target_value, 
			   scale_out_step, scale_in_step, cooldown_sec, last_scaled_at 
		FROM scaling_policies WHERE scaling_group_id = ANY($1)
	`
	rows, err := r.db.Query(ctx, query, groupIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[uuid.UUID][]*domain.ScalingPolicy)
	for rows.Next() {
		var p domain.ScalingPolicy
		var lastScaledAt sql.NullTime
		if err := rows.Scan(
			&p.ID, &p.ScalingGroupID, &p.Name, &p.MetricType, &p.TargetValue,
			&p.ScaleOutStep, &p.ScaleInStep, &p.CooldownSec, &lastScaledAt,
		); err != nil {
			return nil, err
		}
		if lastScaledAt.Valid {
			t := lastScaledAt.Time
			p.LastScaledAt = &t
		}
		result[p.ScalingGroupID] = append(result[p.ScalingGroupID], &p)
	}
	return result, nil
}

func (r *AutoScalingRepo) UpdatePolicyLastScaled(ctx context.Context, policyID uuid.UUID, t time.Time) error {
	_, err := r.db.Exec(ctx, "UPDATE scaling_policies SET last_scaled_at = $1 WHERE id = $2", t, policyID)
	return err
}

func (r *AutoScalingRepo) DeletePolicy(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM scaling_policies WHERE id = $1", id)
	return err
}

// Group Instances

func (r *AutoScalingRepo) AddInstanceToGroup(ctx context.Context, groupID, instanceID uuid.UUID) error {
	_, err := r.db.Exec(ctx, "INSERT INTO scaling_group_instances (scaling_group_id, instance_id) VALUES ($1, $2)", groupID, instanceID)
	return err
}

func (r *AutoScalingRepo) RemoveInstanceFromGroup(ctx context.Context, groupID, instanceID uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM scaling_group_instances WHERE scaling_group_id = $1 AND instance_id = $2", groupID, instanceID)
	return err
}

func (r *AutoScalingRepo) GetInstancesInGroup(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.db.Query(ctx, "SELECT instance_id FROM scaling_group_instances WHERE scaling_group_id = $1", groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func (r *AutoScalingRepo) GetAllScalingGroupInstances(ctx context.Context, groupIDs []uuid.UUID) (map[uuid.UUID][]uuid.UUID, error) {
	if len(groupIDs) == 0 {
		return make(map[uuid.UUID][]uuid.UUID), nil
	}

	query := `
		SELECT scaling_group_id, instance_id
		FROM scaling_group_instances
		WHERE scaling_group_id = ANY($1)
	`
	rows, err := r.db.Query(ctx, query, groupIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[uuid.UUID][]uuid.UUID)
	for rows.Next() {
		var sgID, instID uuid.UUID
		if err := rows.Scan(&sgID, &instID); err != nil {
			return nil, err
		}
		result[sgID] = append(result[sgID], instID)
	}
	return result, nil
}

// Metrics

func (r *AutoScalingRepo) GetAverageCPU(ctx context.Context, instanceIDs []uuid.UUID, since time.Time) (float64, error) {
	if len(instanceIDs) == 0 {
		return 0, nil
	}

	query := `
		SELECT COALESCE(AVG(cpu_percent), 0)
		FROM metrics_history
		WHERE instance_id = ANY($1) AND recorded_at >= $2
	`
	var avg float64
	err := r.db.QueryRow(ctx, query, instanceIDs, since).Scan(&avg)
	return avg, err
}
