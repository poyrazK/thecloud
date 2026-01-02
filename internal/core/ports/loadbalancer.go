package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

type LBRepository interface {
	Create(ctx context.Context, lb *domain.LoadBalancer) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.LoadBalancer, error)
	GetByIdempotencyKey(ctx context.Context, key string) (*domain.LoadBalancer, error)
	List(ctx context.Context) ([]*domain.LoadBalancer, error)
	ListAll(ctx context.Context) ([]*domain.LoadBalancer, error)
	Update(ctx context.Context, lb *domain.LoadBalancer) error
	Delete(ctx context.Context, id uuid.UUID) error

	AddTarget(ctx context.Context, target *domain.LBTarget) error
	RemoveTarget(ctx context.Context, lbID, instanceID uuid.UUID) error
	ListTargets(ctx context.Context, lbID uuid.UUID) ([]*domain.LBTarget, error)
	UpdateTargetHealth(ctx context.Context, lbID, instanceID uuid.UUID, health string) error
	GetTargetsForInstance(ctx context.Context, instanceID uuid.UUID) ([]*domain.LBTarget, error)
}

type LBService interface {
	Create(ctx context.Context, name string, vpcID uuid.UUID, port int, algo string, idempotencyKey string) (*domain.LoadBalancer, error)
	Get(ctx context.Context, id uuid.UUID) (*domain.LoadBalancer, error)
	List(ctx context.Context) ([]*domain.LoadBalancer, error)
	Delete(ctx context.Context, id uuid.UUID) error

	AddTarget(ctx context.Context, lbID, instanceID uuid.UUID, port int, weight int) error
	RemoveTarget(ctx context.Context, lbID, instanceID uuid.UUID) error
	ListTargets(ctx context.Context, lbID uuid.UUID) ([]*domain.LBTarget, error)
}

type LBProxyAdapter interface {
	DeployProxy(ctx context.Context, lb *domain.LoadBalancer, targets []*domain.LBTarget) (string, error)
	RemoveProxy(ctx context.Context, lbID uuid.UUID) error
	UpdateProxyConfig(ctx context.Context, lb *domain.LoadBalancer, targets []*domain.LBTarget) error
}
