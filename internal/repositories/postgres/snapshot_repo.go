package postgres

import (
	"context"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/errors"
)

type SnapshotRepository struct {
	db DB
}

func NewSnapshotRepository(db DB) *SnapshotRepository {
	return &SnapshotRepository{db: db}
}

func (r *SnapshotRepository) Create(ctx context.Context, s *domain.Snapshot) error {
	query := `INSERT INTO snapshots (id, user_id, volume_id, volume_name, size_gb, status, description, created_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.Exec(ctx, query, s.ID, s.UserID, s.VolumeID, s.VolumeName, s.SizeGB, string(s.Status), s.Description, s.CreatedAt)
	return err
}

func (r *SnapshotRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Snapshot, error) {
	userID := appcontext.UserIDFromContext(ctx)
	query := `SELECT id, user_id, volume_id, volume_name, size_gb, status, description, created_at FROM snapshots WHERE id = $1 AND user_id = $2`
	s := &domain.Snapshot{}
	var status string
	err := r.db.QueryRow(ctx, query, id, userID).Scan(&s.ID, &s.UserID, &s.VolumeID, &s.VolumeName, &s.SizeGB, &status, &s.Description, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	s.Status = domain.SnapshotStatus(status)
	return s, nil
}

func (r *SnapshotRepository) ListByVolumeID(ctx context.Context, volumeID uuid.UUID) ([]*domain.Snapshot, error) {
	userID := appcontext.UserIDFromContext(ctx)
	query := `SELECT id, user_id, volume_id, volume_name, size_gb, status, description, created_at FROM snapshots WHERE volume_id = $1 AND user_id = $2 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, volumeID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []*domain.Snapshot
	for rows.Next() {
		s := &domain.Snapshot{}
		var status string
		if err := rows.Scan(&s.ID, &s.UserID, &s.VolumeID, &s.VolumeName, &s.SizeGB, &status, &s.Description, &s.CreatedAt); err != nil {
			return nil, err
		}
		s.Status = domain.SnapshotStatus(status)
		snapshots = append(snapshots, s)
	}
	return snapshots, nil
}

func (r *SnapshotRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Snapshot, error) {
	query := `SELECT id, user_id, volume_id, volume_name, size_gb, status, description, created_at FROM snapshots WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []*domain.Snapshot
	for rows.Next() {
		s := &domain.Snapshot{}
		var status string
		if err := rows.Scan(&s.ID, &s.UserID, &s.VolumeID, &s.VolumeName, &s.SizeGB, &status, &s.Description, &s.CreatedAt); err != nil {
			return nil, err
		}
		s.Status = domain.SnapshotStatus(status)
		snapshots = append(snapshots, s)
	}
	return snapshots, nil
}

func (r *SnapshotRepository) Update(ctx context.Context, s *domain.Snapshot) error {
	query := `UPDATE snapshots SET status = $1, description = $2 WHERE id = $3 AND user_id = $4`
	_, err := r.db.Exec(ctx, query, string(s.Status), s.Description, s.ID, s.UserID)
	return err
}

func (r *SnapshotRepository) Delete(ctx context.Context, id uuid.UUID) error {
	userID := appcontext.UserIDFromContext(ctx)
	query := `DELETE FROM snapshots WHERE id = $1 AND user_id = $2`
	cmd, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errors.New(errors.NotFound, "snapshot not found")
	}
	return nil
}
