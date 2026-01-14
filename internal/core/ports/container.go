// Package ports defines service and repository interfaces.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

// ContainerRepository manages the persistent state of container deployments and their individual replicas.
type ContainerRepository interface {
	// CreateDeployment saves a new container deployment configuration.
	CreateDeployment(ctx context.Context, d *domain.Deployment) error
	// GetDeploymentByID retrieves a specific deployment for a user.
	GetDeploymentByID(ctx context.Context, id, userID uuid.UUID) (*domain.Deployment, error)
	// ListDeployments returns all deployments owned by a user.
	ListDeployments(ctx context.Context, userID uuid.UUID) ([]*domain.Deployment, error)
	// UpdateDeployment modifies an existing deployment's metadata or desired state.
	UpdateDeployment(ctx context.Context, d *domain.Deployment) error
	// DeleteDeployment removes a deployment configuration from storage.
	DeleteDeployment(ctx context.Context, id uuid.UUID) error

	// Replication management

	// AddContainer links a specific instance ID to a deployment.
	AddContainer(ctx context.Context, deploymentID, instanceID uuid.UUID) error
	// RemoveContainer unlinks a specific instance from a deployment.
	RemoveContainer(ctx context.Context, deploymentID, instanceID uuid.UUID) error
	// GetContainers retrieves the IDs of all instances belonging to a deployment.
	GetContainers(ctx context.Context, deploymentID uuid.UUID) ([]uuid.UUID, error)

	// Worker

	// ListAllDeployments returns every deployment in the system for background reconciliation.
	ListAllDeployments(ctx context.Context) ([]*domain.Deployment, error)
}

// ContainerService provides business logic for managing Container-as-a-Service (CaaS) deployments.
type ContainerService interface {
	// CreateDeployment provisions a new managed container set.
	CreateDeployment(ctx context.Context, name, image string, replicas int, ports string) (*domain.Deployment, error)
	// ListDeployments returns deployments for the current authorized user.
	ListDeployments(ctx context.Context) ([]*domain.Deployment, error)
	// GetDeployment retrieves details for a specific deployment.
	GetDeployment(ctx context.Context, id uuid.UUID) (*domain.Deployment, error)
	// ScaleDeployment adjusts the desired replica count for an existing deployment.
	ScaleDeployment(ctx context.Context, id uuid.UUID, replicas int) error
	// DeleteDeployment decommission an entire deployment and stops all replicas.
	DeleteDeployment(ctx context.Context, id uuid.UUID) error
}
