// Package ports defines service and repository interfaces.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

// VpcRepository manages the persistent state of Virtual Private Clouds (VPCs).
type VpcRepository interface {
	// Create saves a new VPC configuration record.
	Create(ctx context.Context, vpc *domain.VPC) error
	// GetByID retrieves a VPC definition by its unique UUID.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.VPC, error)
	// GetByName retrieves a VPC definition by its friendly name (scoped to authorized user).
	GetByName(ctx context.Context, name string) (*domain.VPC, error)
	// List returns all VPCs accessible to the current user context.
	List(ctx context.Context) ([]*domain.VPC, error)
	// Delete removes a VPC definition from storage.
	Delete(ctx context.Context, id uuid.UUID) error
}

// VpcService provides business logic for managing isolated virtual networks.
type VpcService interface {
	// CreateVPC establishes a new isolated virtual network with a specific address space.
	CreateVPC(ctx context.Context, name, cidrBlock string) (*domain.VPC, error)
	// GetVPC retrieves detailed information for a specific virtual network.
	GetVPC(ctx context.Context, idOrName string) (*domain.VPC, error)
	// ListVPCs returns all virtual networks registered to the current authorized user.
	ListVPCs(ctx context.Context) ([]*domain.VPC, error)
	// DeleteVPC decommissioning an existing virtual network.
	DeleteVPC(ctx context.Context, idOrName string) error
}
