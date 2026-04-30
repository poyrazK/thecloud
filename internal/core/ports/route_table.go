package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

// RouteTableRepository manages the persistent state of route tables.
type RouteTableRepository interface {
	Create(ctx context.Context, rt *domain.RouteTable) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.RouteTable, error)
	GetByVPC(ctx context.Context, vpcID uuid.UUID) ([]*domain.RouteTable, error)
	GetMainByVPC(ctx context.Context, vpcID uuid.UUID) (*domain.RouteTable, error)
	Update(ctx context.Context, rt *domain.RouteTable) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Route operations
	AddRoute(ctx context.Context, rtID uuid.UUID, route *domain.Route) error
	RemoveRoute(ctx context.Context, rtID, routeID uuid.UUID) error
	ListRoutes(ctx context.Context, rtID uuid.UUID) ([]domain.Route, error)

	// Association operations
	AssociateSubnet(ctx context.Context, rtID, subnetID uuid.UUID) error
	DisassociateSubnet(ctx context.Context, rtID, subnetID uuid.UUID) error
	ListAssociatedSubnets(ctx context.Context, rtID uuid.UUID) ([]uuid.UUID, error)
}

// RouteTableService provides business logic for route table management.
type RouteTableService interface {
	// CreateRouteTable creates a new custom route table for a VPC.
	// If isMain is true, this becomes the main route table (replacing existing main).
	CreateRouteTable(ctx context.Context, vpcID uuid.UUID, name string, isMain bool) (*domain.RouteTable, error)

	// GetRouteTable retrieves a route table by ID.
	GetRouteTable(ctx context.Context, id uuid.UUID) (*domain.RouteTable, error)

	// ListRouteTables returns all route tables for a VPC.
	ListRouteTables(ctx context.Context, vpcID uuid.UUID) ([]*domain.RouteTable, error)

	// DeleteRouteTable removes a custom route table (cannot delete main route table).
	DeleteRouteTable(ctx context.Context, id uuid.UUID) error

	// AddRoute adds a route to an existing route table.
	// destinationCIDR: CIDR block (e.g., "0.0.0.0/0" or "10.0.1.0/24")
	// targetType: local, igw, nat, or peering
	// targetID: UUID of the target resource (IGW, NAT, or Peering) - nil for local
	AddRoute(ctx context.Context, rtID uuid.UUID, destinationCIDR string, targetType domain.RouteTargetType, targetID *uuid.UUID) (*domain.Route, error)

	// RemoveRoute removes a route from a route table.
	RemoveRoute(ctx context.Context, rtID, routeID uuid.UUID) error

	// AssociateSubnet links a subnet to a route table.
	// The subnet must belong to the same VPC as the route table.
	AssociateSubnet(ctx context.Context, rtID, subnetID uuid.UUID) error

	// DisassociateSubnet removes a subnet's association with a route table.
	// The subnet will then use the main route table.
	DisassociateSubnet(ctx context.Context, rtID, subnetID uuid.UUID) error

	// ReplaceRoute replaces an existing route with a new target.
	ReplaceRoute(ctx context.Context, rtID, routeID uuid.UUID, newTargetID *uuid.UUID) error
}
