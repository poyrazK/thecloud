// Package postgres provides PostgreSQL-backed repository implementations.
package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	stdlib_errors "errors"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/errors"
)

// SubnetRepository provides PostgreSQL-backed subnet persistence.
type SubnetRepository struct {
	db DB
}

// NewSubnetRepository creates a SubnetRepository using the provided DB.
func NewSubnetRepository(db DB) *SubnetRepository {
	return &SubnetRepository{db: db}
}

func (r *SubnetRepository) Create(ctx context.Context, subnet *domain.Subnet) error {
	query := `
		INSERT INTO subnets (id, user_id, vpc_id, name, cidr_block, availability_zone, gateway_ip, arn, status, created_at)
		VALUES ($1, $2, $3, $4, NULLIF($5, '')::cidr, $6, NULLIF($7, '')::inet, $8, $9, $10)
	`
	_, err := r.db.Exec(ctx, query, subnet.ID, subnet.UserID, subnet.VPCID, subnet.Name, subnet.CIDRBlock, subnet.AvailabilityZone, subnet.GatewayIP, subnet.ARN, subnet.Status, subnet.CreatedAt)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to create subnet", err)
	}
	return nil
}

func (r *SubnetRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subnet, error) {
	userID := appcontext.UserIDFromContext(ctx)
	query := `SELECT id, user_id, vpc_id, name, cidr_block::text, availability_zone, COALESCE(gateway_ip::text, ''), arn, status, created_at FROM subnets WHERE id = $1 AND user_id = $2`
	return r.scanSubnet(r.db.QueryRow(ctx, query, id, userID))
}

func (r *SubnetRepository) GetByName(ctx context.Context, vpcID uuid.UUID, name string) (*domain.Subnet, error) {
	userID := appcontext.UserIDFromContext(ctx)
	query := `SELECT id, user_id, vpc_id, name, cidr_block::text, availability_zone, COALESCE(gateway_ip::text, ''), arn, status, created_at FROM subnets WHERE vpc_id = $1 AND name = $2 AND user_id = $3`
	return r.scanSubnet(r.db.QueryRow(ctx, query, vpcID, name, userID))
}

func (r *SubnetRepository) ListByVPC(ctx context.Context, vpcID uuid.UUID) ([]*domain.Subnet, error) {
	userID := appcontext.UserIDFromContext(ctx)
	query := `SELECT id, user_id, vpc_id, name, cidr_block::text, availability_zone, COALESCE(gateway_ip::text, ''), arn, status, created_at FROM subnets WHERE vpc_id = $1 AND user_id = $2 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, vpcID, userID)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to list subnets", err)
	}
	return r.scanSubnets(rows)
}

func (r *SubnetRepository) scanSubnet(row pgx.Row) (*domain.Subnet, error) {
	var s domain.Subnet
	err := row.Scan(&s.ID, &s.UserID, &s.VPCID, &s.Name, &s.CIDRBlock, &s.AvailabilityZone, &s.GatewayIP, &s.ARN, &s.Status, &s.CreatedAt)
	if err != nil {
		if stdlib_errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(errors.NotFound, "subnet not found")
		}
		return nil, errors.Wrap(errors.Internal, "failed to scan subnet", err)
	}
	return &s, nil
}

func (r *SubnetRepository) scanSubnets(rows pgx.Rows) ([]*domain.Subnet, error) {
	defer rows.Close()
	var subnets []*domain.Subnet
	for rows.Next() {
		s, err := r.scanSubnet(rows)
		if err != nil {
			return nil, err
		}
		subnets = append(subnets, s)
	}
	return subnets, nil
}

func (r *SubnetRepository) Delete(ctx context.Context, id uuid.UUID) error {
	userID := appcontext.UserIDFromContext(ctx)
	query := `DELETE FROM subnets WHERE id = $1 AND user_id = $2`
	cmd, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to delete subnet", err)
	}
	if cmd.RowsAffected() == 0 {
		return errors.New(errors.NotFound, "subnet not found")
	}
	return nil
}
