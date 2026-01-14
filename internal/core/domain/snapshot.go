package domain

import (
	"time"

	"github.com/google/uuid"
)

// SnapshotStatus represents the lifecycle state of a storage snapshot.
type SnapshotStatus string

const (
	// SnapshotStatusCreating indicates the snapshot process is ongoing.
	SnapshotStatusCreating SnapshotStatus = "CREATING"
	// SnapshotStatusAvailable indicates the snapshot is ready for storage or restoration.
	SnapshotStatusAvailable SnapshotStatus = "AVAILABLE"
	// SnapshotStatusDeleting indicates the snapshot is being removed.
	SnapshotStatusDeleting SnapshotStatus = "DELETING"
	// SnapshotStatusError indicates the snapshot process failed.
	SnapshotStatusError SnapshotStatus = "ERROR"
)

// Snapshot represents a point-in-time copy of a storage volume.
// Snapshots act as backups and can be used to restore data to a new volume.
type Snapshot struct {
	ID          uuid.UUID      `json:"id"`
	UserID      uuid.UUID      `json:"user_id"`
	VolumeID    uuid.UUID      `json:"volume_id"`   // Source volume ID
	VolumeName  string         `json:"volume_name"` // Source volume name (snapshot time)
	SizeGB      int            `json:"size_gb"`     // Size at snapshot time
	Status      SnapshotStatus `json:"status"`
	Description string         `json:"description,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}
