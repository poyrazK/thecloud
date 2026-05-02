// Package postgres provides PostgreSQL-backed repository implementations.
package postgres

import (
	"context"
	stdlib_errors "errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/errors"
)

// NATGatewayRepository provides PostgreSQL-backed persistence for NAT Gateways.
type NATGatewayRepository struct {
	db DB
}

// NewNATGatewayRepository creates a NATGatewayRepository using the provided DB.
func NewNATGatewayRepository(db DB) *NATGatewayRepository {
	return &NATGatewayRepository{db: db}
}

// Create inserts a new NAT Gateway record into the database.
func (r *NATGatewayRepository) Create(ctx context.Context, nat *domain.NATGateway) error {
	query := `
		INSERT INTO nat_gateways (id, vpc_id, subnet_id, elastic_ip_id, user_id, tenant_id, status, private_ip, arn, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Exec(ctx, query, nat.ID, nat.VPCID, nat.SubnetID, nat.ElasticIPID, nat.UserID, nat.TenantID, nat.Status, nat.PrivateIP, nat.ARN, nat.CreatedAt)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to create NAT gateway", err)
	}
	return nil
}

// GetByID retrieves a single NAT Gateway by its UUID.
func (r *NATGatewayRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.NATGateway, error) {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `SELECT id, vpc_id, subnet_id, elastic_ip_id, user_id, tenant_id, status, private_ip, arn, created_at FROM nat_gateways WHERE id = $1 AND tenant_id = $2`
	return r.scanNATGateway(r.db.QueryRow(ctx, query, id, tenantID))
}

// ListBySubnet returns all NAT Gateways in a specific subnet.
func (r *NATGatewayRepository) ListBySubnet(ctx context.Context, subnetID uuid.UUID) ([]*domain.NATGateway, error) {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `SELECT id, vpc_id, subnet_id, elastic_ip_id, user_id, tenant_id, status, private_ip, arn, created_at FROM nat_gateways WHERE subnet_id = $1 AND tenant_id = $2`
	rows, err := r.db.Query(ctx, query, subnetID, tenantID)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to list NAT gateways by subnet", err)
	}
	return r.scanNATGateways(rows)
}

// ListByVPC returns all NAT Gateways for a given VPC.
func (r *NATGatewayRepository) ListByVPC(ctx context.Context, vpcID uuid.UUID) ([]*domain.NATGateway, error) {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `SELECT id, vpc_id, subnet_id, elastic_ip_id, user_id, tenant_id, status, private_ip, arn, created_at FROM nat_gateways WHERE vpc_id = $1 AND tenant_id = $2 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, vpcID, tenantID)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to list NAT gateways by VPC", err)
	}
	return r.scanNATGateways(rows)
}

// Update modifies an existing NAT Gateway.
func (r *NATGatewayRepository) Update(ctx context.Context, nat *domain.NATGateway) error {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `
		UPDATE nat_gateways SET status = $1, private_ip = $2
		WHERE id = $3 AND tenant_id = $4
	`
	cmd, err := r.db.Exec(ctx, query, nat.Status, nat.PrivateIP, nat.ID, tenantID)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to update NAT gateway", err)
	}
	if cmd.RowsAffected() == 0 {
		return errors.New(errors.NotFound, "NAT gateway not found")
	}
	return nil
}

// Delete removes a NAT Gateway from the database.
func (r *NATGatewayRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `DELETE FROM nat_gateways WHERE id = $1 AND tenant_id = $2`
	cmd, err := r.db.Exec(ctx, query, id, tenantID)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to delete NAT gateway", err)
	}
	if cmd.RowsAffected() == 0 {
		return errors.New(errors.NotFound, "NAT gateway not found")
	}
	return nil
}

func (r *NATGatewayRepository) scanNATGateway(row pgx.Row) (*domain.NATGateway, error) {
	var nat domain.NATGateway
	err := row.Scan(&nat.ID, &nat.VPCID, &nat.SubnetID, &nat.ElasticIPID, &nat.UserID, &nat.TenantID, &nat.Status, &nat.PrivateIP, &nat.ARN, &nat.CreatedAt)
	if err != nil {
		if stdlib_errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(errors.NotFound, "NAT gateway not found")
		}
		return nil, errors.Wrap(errors.Internal, "failed to scan NAT gateway", err)
	}
	return &nat, nil
}

func (r *NATGatewayRepository) scanNATGateways(rows pgx.Rows) ([]*domain.NATGateway, error) {
	defer rows.Close()
	var nats []*domain.NATGateway
	for rows.Next() {
		nat, err := r.scanNATGateway(rows)
		if err != nil {
			return nil, err
		}
		nats = append(nats, nat)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to iterate NAT gateways", err)
	}
	return nats, nil
}
