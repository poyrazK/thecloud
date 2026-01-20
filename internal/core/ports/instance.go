// Package ports defines service and repository interfaces.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

// InstanceRepository handles the persistence and retrieval of compute instance metadata.
type InstanceRepository interface {
	// Create saves a new compute instance record.
	Create(ctx context.Context, instance *domain.Instance) error
	// GetByID retrieves an instance by its unique UUID.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Instance, error)
	// GetByName retrieves an instance by its friendly name (typically scoped to owner).
	GetByName(ctx context.Context, name string) (*domain.Instance, error)
	// List returns instances authorized for the current operational context.
	List(ctx context.Context) ([]*domain.Instance, error)
	// ListAll returns every instance in the system (for administrative or global tasks).
	ListAll(ctx context.Context) ([]*domain.Instance, error)
	// ListBySubnet returns all instances residing within a specific subnet.
	ListBySubnet(ctx context.Context, subnetID uuid.UUID) ([]*domain.Instance, error)
	// Update modifies an existing instance's metadata or status.
	Update(ctx context.Context, instance *domain.Instance) error
	// Delete removes an instance record from persistent storage.
	Delete(ctx context.Context, id uuid.UUID) error
}

// InstanceService defines the business logic for managing the lifecycle of compute instances.
type InstanceService interface {
	// LaunchInstance provisions and starts a new compute resource with requested networking and storage.
	LaunchInstance(ctx context.Context, name, image, ports string, vpcID, subnetID *uuid.UUID, volumes []domain.VolumeAttachment) (*domain.Instance, error)
	// StopInstance gracefully shuts down or halts a running compute resource.
	StopInstance(ctx context.Context, idOrName string) error
	// ListInstances returns a slice of all compute resources accessible to the caller.
	ListInstances(ctx context.Context) ([]*domain.Instance, error)
	// GetInstance retrieves detailed information about a specific compute resource.
	GetInstance(ctx context.Context, idOrName string) (*domain.Instance, error)
	// GetInstanceLogs retrieves recent console/system output from the instance.
	GetInstanceLogs(ctx context.Context, idOrName string) (string, error)
	// GetInstanceStats retrieves real-time performance metrics for the instance.
	GetInstanceStats(ctx context.Context, idOrName string) (*domain.InstanceStats, error)
	// GetConsoleURL provides a secure endpoint for remote serial/VNC console access.
	GetConsoleURL(ctx context.Context, idOrName string) (string, error)
	// TerminateInstance permanently removes a compute resource and releases its allocated assets.
	TerminateInstance(ctx context.Context, idOrName string) error
	// Exec runs a command inside the instance and returns the output.
	Exec(ctx context.Context, idOrName string, cmd []string) (string, error)
}
