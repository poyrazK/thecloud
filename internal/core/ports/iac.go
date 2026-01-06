package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

type StackRepository interface {
	Create(ctx context.Context, stack *domain.Stack) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Stack, error)
	GetByName(ctx context.Context, userID uuid.UUID, name string) (*domain.Stack, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Stack, error)
	Update(ctx context.Context, stack *domain.Stack) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Resource management
	AddResource(ctx context.Context, resource *domain.StackResource) error
	ListResources(ctx context.Context, stackID uuid.UUID) ([]domain.StackResource, error)
	DeleteResources(ctx context.Context, stackID uuid.UUID) error
}

type StackService interface {
	CreateStack(ctx context.Context, name, template string, parameters map[string]string) (*domain.Stack, error)
	GetStack(ctx context.Context, id uuid.UUID) (*domain.Stack, error)
	ListStacks(ctx context.Context) ([]*domain.Stack, error)
	DeleteStack(ctx context.Context, id uuid.UUID) error
	ValidateTemplate(ctx context.Context, template string) (*domain.TemplateValidateResponse, error)
}
