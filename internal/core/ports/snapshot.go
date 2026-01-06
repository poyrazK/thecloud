package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

type SnapshotRepository interface {
	Create(ctx context.Context, s *domain.Snapshot) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Snapshot, error)
	ListByVolumeID(ctx context.Context, volumeID uuid.UUID) ([]*domain.Snapshot, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Snapshot, error)
	Update(ctx context.Context, s *domain.Snapshot) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type SnapshotService interface {
	CreateSnapshot(ctx context.Context, volumeID uuid.UUID, description string) (*domain.Snapshot, error)
	ListSnapshots(ctx context.Context) ([]*domain.Snapshot, error)
	GetSnapshot(ctx context.Context, id uuid.UUID) (*domain.Snapshot, error)
	DeleteSnapshot(ctx context.Context, id uuid.UUID) error
	RestoreSnapshot(ctx context.Context, snapshotID uuid.UUID, newVolumeName string) (*domain.Volume, error)
}
