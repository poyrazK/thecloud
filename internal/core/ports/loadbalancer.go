// Package ports defines service and repository interfaces.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

// LBRepository manages the persistence of load balancer configurations and target memberships.
type LBRepository interface {
	// Create persists a new load balancer record.
	Create(ctx context.Context, lb *domain.LoadBalancer) error
	// GetByID retrieves a load balancer by its unique UUID.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.LoadBalancer, error)
	// GetByIdempotencyKey retrieves a load balancer using its idempotency key to prevent duplicate creation.
	GetByIdempotencyKey(ctx context.Context, key string) (*domain.LoadBalancer, error)
	// List returns load balancers authorized for the current user.
	List(ctx context.Context) ([]*domain.LoadBalancer, error)
	// ListAll returns every load balancer in the system for background management tasks.
	ListAll(ctx context.Context) ([]*domain.LoadBalancer, error)
	// Update modifies an existing load balancer's configuration or status.
	Update(ctx context.Context, lb *domain.LoadBalancer) error
	// Delete removes a load balancer record from storage.
	Delete(ctx context.Context, id uuid.UUID) error

	// AddTarget links a compute instance to a load balancer's target pool.
	AddTarget(ctx context.Context, target *domain.LBTarget) error
	// RemoveTarget unlinks an instance from a load balancer's target pool.
	RemoveTarget(ctx context.Context, lbID, instanceID uuid.UUID) error
	// ListTargets retrieves all backend members associated with a load balancer.
	ListTargets(ctx context.Context, lbID uuid.UUID) ([]*domain.LBTarget, error)
	// UpdateTargetHealth updates the operational state of a specific target.
	UpdateTargetHealth(ctx context.Context, lbID, instanceID uuid.UUID, health string) error
	// GetTargetsForInstance retrieves all load balancers that a specific instance is a member of.
	GetTargetsForInstance(ctx context.Context, instanceID uuid.UUID) ([]*domain.LBTarget, error)
}

// LBService provides business logic for distributing traffic across multiple instances.
type LBService interface {
	// Create establishes a new managed load balancer.
	Create(ctx context.Context, name string, vpcID uuid.UUID, port int, algo string, idempotencyKey string) (*domain.LoadBalancer, error)
	// Get retrieves details and current status for a specific load balancer.
	Get(ctx context.Context, id uuid.UUID) (*domain.LoadBalancer, error)
	// List returns all load balancers belonging to the current authorized context.
	List(ctx context.Context) ([]*domain.LoadBalancer, error)
	// Delete decommission a load balancer and stops its proxy traffic distribution.
	Delete(ctx context.Context, id uuid.UUID) error

	// AddTarget registers a new backend instance into the load balancer's rotation.
	AddTarget(ctx context.Context, lbID, instanceID uuid.UUID, port int, weight int) error
	// RemoveTarget unregisters an instance from the load balancer.
	RemoveTarget(ctx context.Context, lbID, instanceID uuid.UUID) error
	// ListTargets returns all current members of the load balancer's pool.
	ListTargets(ctx context.Context, lbID uuid.UUID) ([]*domain.LBTarget, error)
}

// LBProxyAdapter abstracts the platform-specific implementation of the traffic proxy (e.g., Nginx, HAProxy).
type LBProxyAdapter interface {
	// DeployProxy configures and starts the physical or virtual proxy process.
	DeployProxy(ctx context.Context, lb *domain.LoadBalancer, targets []*domain.LBTarget) (string, error)
	// RemoveProxy decommission the traffic proxy process.
	RemoveProxy(ctx context.Context, lbID uuid.UUID) error
	// UpdateProxyConfig reloads the proxy configuration with an updated target set.
	UpdateProxyConfig(ctx context.Context, lb *domain.LoadBalancer, targets []*domain.LBTarget) error
}
