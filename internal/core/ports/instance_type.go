package ports

import (
	"context"

	"github.com/poyrazk/thecloud/internal/core/domain"
)

// InstanceTypeRepository handles the persistence and retrieval of instance types.
type InstanceTypeRepository interface {
	// List returns all available instance types.
	List(ctx context.Context) ([]*domain.InstanceType, error)
	// GetByID retrieves a specific instance type by its unique identifier.
	GetByID(ctx context.Context, id string) (*domain.InstanceType, error)
	// Create persists a new instance type.
	Create(ctx context.Context, it *domain.InstanceType) (*domain.InstanceType, error)
	// Update modifies an existing instance type.
	Update(ctx context.Context, it *domain.InstanceType) (*domain.InstanceType, error)
	// Delete removes an instance type by its ID.
	Delete(ctx context.Context, id string) error
}

// InstanceTypeService defines the business logic for instance types.
type InstanceTypeService interface {
	// List returns all available instance types.
	List(ctx context.Context) ([]*domain.InstanceType, error)
}
