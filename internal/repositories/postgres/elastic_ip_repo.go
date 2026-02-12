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

// ElasticIPRepository provides a PostgreSQL implementation for managing Elastic IP metadata.
type ElasticIPRepository struct {
	db DB
}

// NewElasticIPRepository creates a new ElasticIPRepository with the given database pool.
func NewElasticIPRepository(db DB) *ElasticIPRepository {
	return &ElasticIPRepository{db: db}
}

// Create inserts a new Elastic IP record into the database.
func (r *ElasticIPRepository) Create(ctx context.Context, eip *domain.ElasticIP) error {
	query := `
		INSERT INTO elastic_ips (id, user_id, tenant_id, public_ip, instance_id, vpc_id, status, arn, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Exec(ctx, query,
		eip.ID, eip.UserID, eip.TenantID, eip.PublicIP,
		eip.InstanceID, eip.VpcID, eip.Status, eip.ARN,
		eip.CreatedAt, eip.UpdatedAt,
	)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to create elastic ip", err)
	}
	return nil
}

// GetByID retrieves a single Elastic IP by its UUID.
func (r *ElasticIPRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ElasticIP, error) {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `
		SELECT id, user_id, tenant_id, public_ip, instance_id, vpc_id, status, arn, created_at, updated_at 
		FROM elastic_ips 
		WHERE id = $1 AND tenant_id = $2
	`
	return r.scanElasticIP(r.db.QueryRow(ctx, query, id, tenantID))
}

// GetByPublicIP retrieves a single Elastic IP by its public address.
func (r *ElasticIPRepository) GetByPublicIP(ctx context.Context, publicIP string) (*domain.ElasticIP, error) {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `
		SELECT id, user_id, tenant_id, public_ip, instance_id, vpc_id, status, arn, created_at, updated_at 
		FROM elastic_ips 
		WHERE public_ip = $1 AND tenant_id = $2
	`
	return r.scanElasticIP(r.db.QueryRow(ctx, query, publicIP, tenantID))
}

// GetByInstanceID retrieves the Elastic IP associated with a specific instance.
func (r *ElasticIPRepository) GetByInstanceID(ctx context.Context, instanceID uuid.UUID) (*domain.ElasticIP, error) {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `
		SELECT id, user_id, tenant_id, public_ip, instance_id, vpc_id, status, arn, created_at, updated_at 
		FROM elastic_ips 
		WHERE instance_id = $1 AND tenant_id = $2
	`
	return r.scanElasticIP(r.db.QueryRow(ctx, query, instanceID, tenantID))
}

// List returns all Elastic IPs belonging to the authenticated tenant.
func (r *ElasticIPRepository) List(ctx context.Context) ([]*domain.ElasticIP, error) {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `
		SELECT id, user_id, tenant_id, public_ip, instance_id, vpc_id, status, arn, created_at, updated_at 
		FROM elastic_ips 
		WHERE tenant_id = $1 
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to list elastic ips", err)
	}
	return r.scanElasticIPs(rows)
}

// Update modifies an existing Elastic IP's metadata or status.
func (r *ElasticIPRepository) Update(ctx context.Context, eip *domain.ElasticIP) error {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `
		UPDATE elastic_ips 
		SET instance_id = $1, vpc_id = $2, status = $3, updated_at = $4 
		WHERE id = $5 AND tenant_id = $6
	`
	cmd, err := r.db.Exec(ctx, query,
		eip.InstanceID, eip.VpcID, eip.Status, eip.UpdatedAt,
		eip.ID, tenantID,
	)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to update elastic ip", err)
	}
	if cmd.RowsAffected() == 0 {
		return errors.New(errors.NotFound, "elastic ip not found")
	}
	return nil
}

// Delete removes an Elastic IP record from the database.
func (r *ElasticIPRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `DELETE FROM elastic_ips WHERE id = $1 AND tenant_id = $2`
	cmd, err := r.db.Exec(ctx, query, id, tenantID)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to delete elastic ip", err)
	}
	if cmd.RowsAffected() == 0 {
		return errors.New(errors.NotFound, "elastic ip not found")
	}
	return nil
}

func (r *ElasticIPRepository) scanElasticIP(row pgx.Row) (*domain.ElasticIP, error) {
	var eip domain.ElasticIP
	err := row.Scan(
		&eip.ID, &eip.UserID, &eip.TenantID, &eip.PublicIP,
		&eip.InstanceID, &eip.VpcID, &eip.Status, &eip.ARN,
		&eip.CreatedAt, &eip.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New(errors.NotFound, "elastic ip not found")
		}
		return nil, errors.Wrap(errors.Internal, "failed to scan elastic ip", err)
	}
	return &eip, nil
}

func (r *ElasticIPRepository) scanElasticIPs(rows pgx.Rows) ([]*domain.ElasticIP, error) {
	defer rows.Close()
	var eips []*domain.ElasticIP
	for rows.Next() {
		eip, err := r.scanElasticIP(rows)
		if err != nil {
			return nil, err
		}
		eips = append(eips, eip)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to list elastic ips", err)
	}
	return eips, nil
}
