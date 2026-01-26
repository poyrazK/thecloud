// Package domain defines core business entities.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// VolumeStatus represents the lifecycle state of a storage volume.
type VolumeStatus string

const (
	// VolumeStatusAvailable indicates the volume is ready for attachment.
	VolumeStatusAvailable VolumeStatus = "AVAILABLE"
	// VolumeStatusInUse indicates the volume is attached to an instance.
	VolumeStatusInUse VolumeStatus = "IN-USE"
	// VolumeStatusDeleting indicates the volume is being removed.
	VolumeStatusDeleting VolumeStatus = "DELETING"
)

// Volume represents a block storage device.
// Volumes provide persistent storage independent of instance lifecycle.
type Volume struct {
	ID          uuid.UUID    `json:"id"`
	UserID      uuid.UUID    `json:"user_id"`
	TenantID    uuid.UUID    `json:"tenant_id"`
	Name        string       `json:"name"`
	SizeGB      int          `json:"size_gb"`
	Status      VolumeStatus `json:"status"`
	InstanceID  *uuid.UUID   `json:"instance_id,omitempty"`  // Attached instance
	BackendPath string       `json:"backend_path,omitempty"` // Physical path on host/storage
	MountPath   string       `json:"mount_path,omitempty"`   // Internal mount point
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}
