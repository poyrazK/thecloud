package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

type SSHKeyRepository interface {
	Create(ctx context.Context, key *domain.SSHKey) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.SSHKey, error)
	GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*domain.SSHKey, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]*domain.SSHKey, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type SSHKeyService interface {
	CreateKey(ctx context.Context, name, publicKey string) (*domain.SSHKey, error)
	GetKey(ctx context.Context, id uuid.UUID) (*domain.SSHKey, error)
	ListKeys(ctx context.Context) ([]*domain.SSHKey, error)
	DeleteKey(ctx context.Context, id uuid.UUID) error
}
