package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

// IGWRepository manages persistence of Internet Gateways.
type IGWRepository interface {
	Create(ctx context.Context, igw *domain.InternetGateway) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.InternetGateway, error)
	GetByVPC(ctx context.Context, vpcID uuid.UUID) (*domain.InternetGateway, error)
	ListAll(ctx context.Context) ([]*domain.InternetGateway, error)
	Update(ctx context.Context, igw *domain.InternetGateway) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// InternetGatewayService provides business logic for IGW management.
type InternetGatewayService interface {
	// CreateIGW creates a new Internet Gateway (starts in detached state).
	CreateIGW(ctx context.Context) (*domain.InternetGateway, error)

	// AttachIGW attaches an IGW to a VPC.
	// This also adds a default route (0.0.0.0/0) to the VPC's main route table
	// pointing to this IGW for internet-bound traffic.
	AttachIGW(ctx context.Context, igwID, vpcID uuid.UUID) error

	// DetachIGW detaches an IGW from its VPC.
	// This removes the default route from the main route table.
	DetachIGW(ctx context.Context, igwID uuid.UUID) error

	// GetIGW retrieves an IGW by ID.
	GetIGW(ctx context.Context, igwID uuid.UUID) (*domain.InternetGateway, error)

	// ListIGWs returns all IGWs for the current tenant.
	ListIGWs(ctx context.Context) ([]*domain.InternetGateway, error)

	// DeleteIGW permanently removes an IGW (must be detached first).
	DeleteIGW(ctx context.Context, igwID uuid.UUID) error
}