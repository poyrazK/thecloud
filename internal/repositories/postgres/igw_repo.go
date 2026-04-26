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

// IGWRepository provides PostgreSQL-backed persistence for Internet Gateways.
type IGWRepository struct {
	db DB
}

// NewIGWRepository creates an IGWRepository using the provided DB.
func NewIGWRepository(db DB) *IGWRepository {
	return &IGWRepository{db: db}
}

// Create inserts a new Internet Gateway record into the database.
func (r *IGWRepository) Create(ctx context.Context, igw *domain.InternetGateway) error {
	query := `
		INSERT INTO internet_gateways (id, vpc_id, user_id, tenant_id, status, arn, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query, igw.ID, igw.VPCID, igw.UserID, igw.TenantID, igw.Status, igw.ARN, igw.CreatedAt)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to create internet gateway", err)
	}
	return nil
}

// GetByID retrieves a single Internet Gateway by its UUID.
func (r *IGWRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.InternetGateway, error) {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `SELECT id, vpc_id, user_id, tenant_id, status, arn, created_at FROM internet_gateways WHERE id = $1 AND tenant_id = $2`
	return r.scanIGW(r.db.QueryRow(ctx, query, id, tenantID))
}

// GetByVPC returns the Internet Gateway attached to a specific VPC (if any).
func (r *IGWRepository) GetByVPC(ctx context.Context, vpcID uuid.UUID) (*domain.InternetGateway, error) {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `SELECT id, vpc_id, user_id, tenant_id, status, arn, created_at FROM internet_gateways WHERE vpc_id = $1 AND tenant_id = $2`
	return r.scanIGW(r.db.QueryRow(ctx, query, vpcID, tenantID))
}

// Update modifies an existing Internet Gateway.
func (r *IGWRepository) Update(ctx context.Context, igw *domain.InternetGateway) error {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `
		UPDATE internet_gateways SET vpc_id = $1, status = $2
		WHERE id = $3 AND tenant_id = $4
	`
	cmd, err := r.db.Exec(ctx, query, igw.VPCID, igw.Status, igw.ID, tenantID)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to update internet gateway", err)
	}
	if cmd.RowsAffected() == 0 {
		return errors.New(errors.NotFound, "internet gateway not found")
	}
	return nil
}

// Delete removes an Internet Gateway from the database.
func (r *IGWRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `DELETE FROM internet_gateways WHERE id = $1 AND tenant_id = $2`
	cmd, err := r.db.Exec(ctx, query, id, tenantID)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to delete internet gateway", err)
	}
	if cmd.RowsAffected() == 0 {
		return errors.New(errors.NotFound, "internet gateway not found")
	}
	return nil
}

// ListAll returns all Internet Gateways for the current tenant.
func (r *IGWRepository) ListAll(ctx context.Context) ([]*domain.InternetGateway, error) {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `SELECT id, vpc_id, user_id, tenant_id, status, arn, created_at FROM internet_gateways WHERE tenant_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to list internet gateways", err)
	}
	return r.scanIGWs(rows)
}

func (r *IGWRepository) scanIGWs(rows pgx.Rows) ([]*domain.InternetGateway, error) {
	defer rows.Close()
	var igws []*domain.InternetGateway
	for rows.Next() {
		igw, err := r.scanIGW(rows)
		if err != nil {
			return nil, err
		}
		igws = append(igws, igw)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to iterate internet gateways", err)
	}
	return igws, nil
}

func (r *IGWRepository) scanIGW(row pgx.Row) (*domain.InternetGateway, error) {
	var igw domain.InternetGateway
	err := row.Scan(&igw.ID, &igw.VPCID, &igw.UserID, &igw.TenantID, &igw.Status, &igw.ARN, &igw.CreatedAt)
	if err != nil {
		if stdlib_errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(errors.NotFound, "internet gateway not found")
		}
		return nil, errors.Wrap(errors.Internal, "failed to scan internet gateway", err)
	}
	return &igw, nil
}