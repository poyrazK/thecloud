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

// VolumeRepository provides PostgreSQL-backed volume persistence.
type VolumeRepository struct {
	db DB
}

// NewVolumeRepository creates a VolumeRepository using the provided DB.
func NewVolumeRepository(db DB) *VolumeRepository {
	return &VolumeRepository{db: db}
}

func (r *VolumeRepository) Create(ctx context.Context, v *domain.Volume) error {
	tenantID := appcontext.TenantIDFromContext(ctx)
	if tenantID == uuid.Nil {
		return errors.New(errors.Unauthorized, "tenant ID required in context")
	}
	query := `INSERT INTO volumes (id, user_id, tenant_id, name, size_gb, status, instance_id, backend_path, mount_path, created_at, updated_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := r.db.Exec(ctx, query, v.ID, v.UserID, tenantID, v.Name, v.SizeGB, string(v.Status), v.InstanceID, v.BackendPath, v.MountPath, v.CreatedAt, v.UpdatedAt)
	return err
}

func (r *VolumeRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Volume, error) {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `SELECT id, user_id, tenant_id, name, size_gb, status, instance_id, backend_path, mount_path, created_at, updated_at FROM volumes WHERE id = $1 AND tenant_id = $2`
	return r.scanVolume(r.db.QueryRow(ctx, query, id, tenantID))
}

func (r *VolumeRepository) GetByName(ctx context.Context, name string) (*domain.Volume, error) {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `SELECT id, user_id, tenant_id, name, size_gb, status, instance_id, backend_path, mount_path, created_at, updated_at FROM volumes WHERE name = $1 AND tenant_id = $2`
	return r.scanVolume(r.db.QueryRow(ctx, query, name, tenantID))
}

func (r *VolumeRepository) List(ctx context.Context) ([]*domain.Volume, error) {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `SELECT id, user_id, tenant_id, name, size_gb, status, instance_id, backend_path, mount_path, created_at, updated_at FROM volumes WHERE tenant_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	return r.scanVolumes(rows)
}

func (r *VolumeRepository) ListByInstanceID(ctx context.Context, instanceID uuid.UUID) ([]*domain.Volume, error) {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `SELECT id, user_id, tenant_id, name, size_gb, status, instance_id, backend_path, mount_path, created_at, updated_at FROM volumes WHERE instance_id = $1 AND tenant_id = $2`
	rows, err := r.db.Query(ctx, query, instanceID, tenantID)
	if err != nil {
		return nil, err
	}
	return r.scanVolumes(rows)
}

func (r *VolumeRepository) scanVolume(row pgx.Row) (*domain.Volume, error) {
	v := &domain.Volume{}
	var status string
	err := row.Scan(&v.ID, &v.UserID, &v.TenantID, &v.Name, &v.SizeGB, &status, &v.InstanceID, &v.BackendPath, &v.MountPath, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, err
	}
	v.Status = domain.VolumeStatus(status)
	return v, nil
}

func (r *VolumeRepository) scanVolumes(rows pgx.Rows) ([]*domain.Volume, error) {
	defer rows.Close()
	var volumes []*domain.Volume
	for rows.Next() {
		v, err := r.scanVolume(rows)
		if err != nil {
			return nil, err
		}
		volumes = append(volumes, v)
	}
	return volumes, nil
}

func (r *VolumeRepository) Update(ctx context.Context, v *domain.Volume) error {
	tenantID := appcontext.TenantIDFromContext(ctx)
	if tenantID == uuid.Nil {
		return errors.New(errors.Unauthorized, "tenant ID required in context")
	}
	query := `UPDATE volumes SET status = $1, instance_id = $2, backend_path = $3, mount_path = $4, updated_at = $5 WHERE id = $6 AND tenant_id = $7`
	_, err := r.db.Exec(ctx, query, string(v.Status), v.InstanceID, v.BackendPath, v.MountPath, v.UpdatedAt, v.ID, tenantID)
	return err
}

func (r *VolumeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `DELETE FROM volumes WHERE id = $1 AND tenant_id = $2`
	cmd, err := r.db.Exec(ctx, query, id, tenantID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errors.New(errors.NotFound, "volume not found")
	}
	return nil
}
