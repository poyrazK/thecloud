// Package ports defines service and repository interfaces.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

// SnapshotRepository handles the persistent state of point-in-time storage backups (snapshots).
type SnapshotRepository interface {
	// Create saves a new snapshot record.
	Create(ctx context.Context, s *domain.Snapshot) error
	// GetByID retrieves a specific snapshot by its Unique ID.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Snapshot, error)
	// ListByVolumeID returns all snapshots created from a specific block volume.
	ListByVolumeID(ctx context.Context, volumeID uuid.UUID) ([]*domain.Snapshot, error)
	// ListByUserID returns all snapshots owned by a specific user.
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Snapshot, error)
	// Update modifies an existing snapshot's metadata or operational status.
	Update(ctx context.Context, s *domain.Snapshot) error
	// Delete removes a snapshot record from persistent storage.
	Delete(ctx context.Context, id uuid.UUID) error
}

// SnapshotService provides business logic for performing point-in-time backups and restores of block storage.
type SnapshotService interface {
	// CreateSnapshot initiates a point-in-time copy of a block volume.
	CreateSnapshot(ctx context.Context, volumeID uuid.UUID, description string) (*domain.Snapshot, error)
	// ListSnapshots returns all snapshots belonging to the current authorized user.
	ListSnapshots(ctx context.Context) ([]*domain.Snapshot, error)
	// GetSnapshot retrieves details for a specific point-in-time backup.
	GetSnapshot(ctx context.Context, id uuid.UUID) (*domain.Snapshot, error)
	// DeleteSnapshot decommissioning a point-in-time backup.
	DeleteSnapshot(ctx context.Context, id uuid.UUID) error
	// RestoreSnapshot creates a new usable block volume from an existing point-in-time backup.
	RestoreSnapshot(ctx context.Context, snapshotID uuid.UUID, newVolumeName string) (*domain.Volume, error)
}
