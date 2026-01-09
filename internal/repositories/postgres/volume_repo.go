package postgres

import (
	"context"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/errors"
)

type VolumeRepository struct {
	db DB
}

func NewVolumeRepository(db DB) *VolumeRepository {
	return &VolumeRepository{db: db}
}

func (r *VolumeRepository) Create(ctx context.Context, v *domain.Volume) error {
	query := `INSERT INTO volumes (id, user_id, name, size_gb, status, instance_id, mount_path, created_at, updated_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.Exec(ctx, query, v.ID, v.UserID, v.Name, v.SizeGB, string(v.Status), v.InstanceID, v.MountPath, v.CreatedAt, v.UpdatedAt)
	return err
}

func (r *VolumeRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Volume, error) {
	userID := appcontext.UserIDFromContext(ctx)
	query := `SELECT id, user_id, name, size_gb, status, instance_id, mount_path, created_at, updated_at FROM volumes WHERE id = $1 AND user_id = $2`
	v := &domain.Volume{}
	var status string
	err := r.db.QueryRow(ctx, query, id, userID).Scan(&v.ID, &v.UserID, &v.Name, &v.SizeGB, &status, &v.InstanceID, &v.MountPath, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, err
	}
	v.Status = domain.VolumeStatus(status)
	return v, nil
}

func (r *VolumeRepository) GetByName(ctx context.Context, name string) (*domain.Volume, error) {
	userID := appcontext.UserIDFromContext(ctx)
	query := `SELECT id, user_id, name, size_gb, status, instance_id, mount_path, created_at, updated_at FROM volumes WHERE name = $1 AND user_id = $2`
	v := &domain.Volume{}
	var status string
	err := r.db.QueryRow(ctx, query, name, userID).Scan(&v.ID, &v.UserID, &v.Name, &v.SizeGB, &status, &v.InstanceID, &v.MountPath, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, err
	}
	v.Status = domain.VolumeStatus(status)
	return v, nil
}

func (r *VolumeRepository) List(ctx context.Context) ([]*domain.Volume, error) {
	userID := appcontext.UserIDFromContext(ctx)
	query := `SELECT id, user_id, name, size_gb, status, instance_id, mount_path, created_at, updated_at FROM volumes WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var volumes []*domain.Volume
	for rows.Next() {
		v := &domain.Volume{}
		var status string
		if err := rows.Scan(&v.ID, &v.UserID, &v.Name, &v.SizeGB, &status, &v.InstanceID, &v.MountPath, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		v.Status = domain.VolumeStatus(status)
		volumes = append(volumes, v)
	}
	return volumes, nil
}

func (r *VolumeRepository) ListByInstanceID(ctx context.Context, instanceID uuid.UUID) ([]*domain.Volume, error) {
	userID := appcontext.UserIDFromContext(ctx)
	query := `SELECT id, user_id, name, size_gb, status, instance_id, mount_path, created_at, updated_at FROM volumes WHERE instance_id = $1 AND user_id = $2`
	rows, err := r.db.Query(ctx, query, instanceID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var volumes []*domain.Volume
	for rows.Next() {
		v := &domain.Volume{}
		var status string
		if err := rows.Scan(&v.ID, &v.UserID, &v.Name, &v.SizeGB, &status, &v.InstanceID, &v.MountPath, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		v.Status = domain.VolumeStatus(status)
		volumes = append(volumes, v)
	}
	return volumes, nil
}

func (r *VolumeRepository) Update(ctx context.Context, v *domain.Volume) error {
	query := `UPDATE volumes SET status = $1, instance_id = $2, mount_path = $3, updated_at = $4 WHERE id = $5 AND user_id = $6`
	_, err := r.db.Exec(ctx, query, string(v.Status), v.InstanceID, v.MountPath, v.UpdatedAt, v.ID, v.UserID)
	return err
}

func (r *VolumeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	userID := appcontext.UserIDFromContext(ctx)
	query := `DELETE FROM volumes WHERE id = $1 AND user_id = $2`
	cmd, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errors.New(errors.NotFound, "volume not found")
	}
	return nil
}
