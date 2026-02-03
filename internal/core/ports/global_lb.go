// Package ports defines the interfaces for the global load balancer service and repository.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

// GlobalLBRepository manages persistent storage for Global Load Balancers.
type GlobalLBRepository interface {
	Create(ctx context.Context, glb *domain.GlobalLoadBalancer) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.GlobalLoadBalancer, error)
	GetByHostname(ctx context.Context, hostname string) (*domain.GlobalLoadBalancer, error)
	List(ctx context.Context, userID uuid.UUID) ([]*domain.GlobalLoadBalancer, error)
	Update(ctx context.Context, glb *domain.GlobalLoadBalancer) error
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	// Endpoint management
	AddEndpoint(ctx context.Context, ep *domain.GlobalEndpoint) error
	RemoveEndpoint(ctx context.Context, endpointID uuid.UUID) error
	GetEndpointByID(ctx context.Context, endpointID uuid.UUID) (*domain.GlobalEndpoint, error)
	ListEndpoints(ctx context.Context, glbID uuid.UUID) ([]*domain.GlobalEndpoint, error)
	UpdateEndpointHealth(ctx context.Context, epID uuid.UUID, healthy bool) error
}

// GlobalLBService provides business logic for multi-region routing.
type GlobalLBService interface {
	Create(ctx context.Context, name, hostname string, policy domain.RoutingPolicy, healthCheck domain.GlobalHealthCheckConfig) (*domain.GlobalLoadBalancer, error)
	Get(ctx context.Context, id uuid.UUID) (*domain.GlobalLoadBalancer, error)
	List(ctx context.Context, userID uuid.UUID) ([]*domain.GlobalLoadBalancer, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	AddEndpoint(ctx context.Context, glbID uuid.UUID, region string, targetType string, targetID *uuid.UUID, targetIP *string, weight, priority int) (*domain.GlobalEndpoint, error)
	RemoveEndpoint(ctx context.Context, glbID, endpointID uuid.UUID) error
	ListEndpoints(ctx context.Context, glbID uuid.UUID) ([]*domain.GlobalEndpoint, error)
}

// GeoDNSBackend abstracts the underlying DNS provider capable of geo-routing.
type GeoDNSBackend interface {
	// CreateGeoRecord creates or updates a record set with regional answers.
	CreateGeoRecord(ctx context.Context, hostname string, endpoints []domain.GlobalEndpoint) error
	// DeleteGeoRecord removes the record set.
	DeleteGeoRecord(ctx context.Context, hostname string) error
}
