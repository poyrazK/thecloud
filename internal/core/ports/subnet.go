// Package ports defines service and repository interfaces.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

// SubnetRepository manages the persistent state of VPC subdivisions (subnets).
type SubnetRepository interface {
	// Create saves a new subnet record.
	Create(ctx context.Context, subnet *domain.Subnet) error
	// GetByID retrieves a subnet definition by its unique UUID.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Subnet, error)
	// GetByName retrieves a subnet definition by its friendly name (scoped to a VPC).
	GetByName(ctx context.Context, vpcID uuid.UUID, name string) (*domain.Subnet, error)
	// ListByVPC returns all subnets currently defined within a specific virtual network.
	ListByVPC(ctx context.Context, vpcID uuid.UUID) ([]*domain.Subnet, error)
	// Delete removes a subnet definition from persistent storage.
	Delete(ctx context.Context, id uuid.UUID) error
}

// SubnetService provides business logic for managing virtual network subdivisions.
type SubnetService interface {
	// CreateSubnet provisions a new address range subdivision within a VPC.
	CreateSubnet(ctx context.Context, vpcID uuid.UUID, name, cidrBlock, az string) (*domain.Subnet, error)
	// GetSubnet retrieves details for a specific virtual network subdivision.
	GetSubnet(ctx context.Context, idOrName string, vpcID uuid.UUID) (*domain.Subnet, error)
	// ListSubnets returns all subdivisions within a specified authorized VPC.
	ListSubnets(ctx context.Context, vpcID uuid.UUID) ([]*domain.Subnet, error)
	// DeleteSubnet decommissioning a virtual network subdivision.
	DeleteSubnet(ctx context.Context, id uuid.UUID) error
}
