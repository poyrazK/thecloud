// Package ports defines service and repository interfaces.
package ports

import (
	"context"
)

// StorageBackend abstracts block storage operations (e.g., LVM, Ceph, EBS-like) to decouple virtual disk management from infrastructure provider details.
type StorageBackend interface {
	// CreateVolume provisions a new block storage device of a specified size.
	CreateVolume(ctx context.Context, name string, sizeGB int) (string, error)
	// DeleteVolume permanently removes a block storage device (only if not currently attached).
	DeleteVolume(ctx context.Context, name string) error
	// AttachVolume connects a block storage device to a specific compute instance.
	AttachVolume(ctx context.Context, volumeName, instanceID string) error
	// DetachVolume disconnects a block storage device from a compute instance.
	DetachVolume(ctx context.Context, volumeName, instanceID string) error
	// CreateSnapshot establishes a point-in-time copy of a block volume on the backend.
	CreateSnapshot(ctx context.Context, volumeName, snapshotName string) error
	// RestoreSnapshot rolls back or initializes a volume from an existing backend snapshot.
	RestoreSnapshot(ctx context.Context, volumeName, snapshotName string) error
	// DeleteSnapshot removes a point-in-time backup from the backend.
	DeleteSnapshot(ctx context.Context, snapshotName string) error
	// Ping verifies the connectivity and responsiveness of the storage backend service.
	Ping(ctx context.Context) error
	// Type returns the identifier of the storage implementation (e.g., "ceph", "lvm").
	Type() string
}
