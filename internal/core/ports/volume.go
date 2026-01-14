// Package ports defines service and repository interfaces.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

// VolumeRepository handles the persistence of block storage volume metadata.
type VolumeRepository interface {
	// Create saves a new block volume record.
	Create(ctx context.Context, v *domain.Volume) error
	// GetByID retrieves a block volume by its unique UUID.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Volume, error)
	// GetByName retrieves a block volume by its friendly name (scoped to owner).
	GetByName(ctx context.Context, name string) (*domain.Volume, error)
	// List returns all block volumes owned by the current authorized user.
	List(ctx context.Context) ([]*domain.Volume, error)
	// ListByInstanceID returns all block volumes currently attached to a specific instance.
	ListByInstanceID(ctx context.Context, instanceID uuid.UUID) ([]*domain.Volume, error)
	// Update modifies an existing volume's metadata or status.
	Update(ctx context.Context, v *domain.Volume) error
	// Delete removes a block volume record from storage.
	Delete(ctx context.Context, id uuid.UUID) error
}

// VolumeService provides business logic for managing the lifecycle of block storage resources (e.g., EBS-like).
type VolumeService interface {
	// CreateVolume provisions a new block storage device.
	CreateVolume(ctx context.Context, name string, sizeGB int) (*domain.Volume, error)
	// ListVolumes returns all block storage devices registered to the current user.
	ListVolumes(ctx context.Context) ([]*domain.Volume, error)
	// GetVolume retrieves detailed information for a specific block storage device.
	GetVolume(ctx context.Context, idOrName string) (*domain.Volume, error)
	// DeleteVolume decommissioning a block storage device.
	DeleteVolume(ctx context.Context, idOrName string) error
	// ReleaseVolumesForInstance detaches every volume currently linked to a specific instance.
	ReleaseVolumesForInstance(ctx context.Context, instanceID uuid.UUID) error
}
