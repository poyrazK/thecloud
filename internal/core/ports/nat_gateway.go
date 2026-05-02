package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

// NATGatewayRepository manages persistence of NAT Gateways.
type NATGatewayRepository interface {
	Create(ctx context.Context, nat *domain.NATGateway) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.NATGateway, error)
	ListBySubnet(ctx context.Context, subnetID uuid.UUID) ([]*domain.NATGateway, error)
	ListByVPC(ctx context.Context, vpcID uuid.UUID) ([]*domain.NATGateway, error)
	Update(ctx context.Context, nat *domain.NATGateway) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// NATGatewayService provides business logic for NAT Gateway management.
type NATGatewayService interface {
	// CreateNATGateway creates a NAT Gateway in a subnet with an allocated Elastic IP.
	// The NAT gateway must be placed in a public subnet (a subnet that has a route
	// through an Internet Gateway) to enable instances in private subnets to access
	// the internet.
	CreateNATGateway(ctx context.Context, subnetID uuid.UUID, eipID uuid.UUID) (*domain.NATGateway, error)

	// GetNATGateway retrieves a NAT Gateway by ID.
	GetNATGateway(ctx context.Context, natID uuid.UUID) (*domain.NATGateway, error)

	// ListNATGateways returns all NAT Gateways for a VPC.
	ListNATGateways(ctx context.Context, vpcID uuid.UUID) ([]*domain.NATGateway, error)

	// DeleteNATGateway removes a NAT Gateway and releases the associated EIP.
	DeleteNATGateway(ctx context.Context, natID uuid.UUID) error
}
