// Package ports defines service and repository interfaces.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

// StackRepository manages the persistent state of Infrastructure-as-Code (IaC) stacks and their resources.
type StackRepository interface {
	// Create saves a new stack configuration.
	Create(ctx context.Context, stack *domain.Stack) error
	// GetByID retrieves a stack by its unique UUID.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Stack, error)
	// GetByName retrieves a stack by name for a specific user.
	GetByName(ctx context.Context, userID uuid.UUID, name string) (*domain.Stack, error)
	// ListByUserID returns all stacks defined by a user.
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Stack, error)
	// Update modifies an existing stack's state or configuration.
	Update(ctx context.Context, stack *domain.Stack) error
	// Delete removes a stack definition.
	Delete(ctx context.Context, id uuid.UUID) error

	// Resource management

	// AddResource links a physical infrastructure resource to a logical stack component.
	AddResource(ctx context.Context, resource *domain.StackResource) error
	// ListResources retrieves all physical resources provisioned as part of a stack.
	ListResources(ctx context.Context, stackID uuid.UUID) ([]domain.StackResource, error)
	// DeleteResources removes resource mappings for a stack.
	DeleteResources(ctx context.Context, stackID uuid.UUID) error
}

// StackService provides business logic for orchestrating complex infrastructure via templates.
type StackService interface {
	// CreateStack parses a template and provisions the defined resources as an atomic unit.
	CreateStack(ctx context.Context, name, template string, parameters map[string]string) (*domain.Stack, error)
	// GetStack retrieves details to track the provisioning status and resources of a stack.
	GetStack(ctx context.Context, id uuid.UUID) (*domain.Stack, error)
	// ListStacks returns stacks registered to the current authorized user.
	ListStacks(ctx context.Context) ([]*domain.Stack, error)
	// DeleteStack decommission all resources managed by the stack in reverse order.
	DeleteStack(ctx context.Context, id uuid.UUID) error
	// ValidateTemplate performs a syntax and capability check on a raw IaC template string.
	ValidateTemplate(ctx context.Context, template string) (*domain.TemplateValidateResponse, error)
}
