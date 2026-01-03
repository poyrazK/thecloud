package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

type CacheStats struct {
	UsedMemoryBytes  int64
	MaxMemoryBytes   int64
	ConnectedClients int
	TotalKeys        int64
}

type CacheService interface {
	CreateCache(ctx context.Context, name, version string, memoryMB int, vpcID *uuid.UUID) (*domain.Cache, error)
	GetCache(ctx context.Context, id uuid.UUID) (*domain.Cache, error)
	ListCaches(ctx context.Context) ([]*domain.Cache, error)
	DeleteCache(ctx context.Context, id uuid.UUID) error
	GetConnectionString(ctx context.Context, id uuid.UUID) (string, error)
	FlushCache(ctx context.Context, id uuid.UUID) error
	GetCacheStats(ctx context.Context, id uuid.UUID) (*CacheStats, error)
}

type CacheRepository interface {
	Create(ctx context.Context, cache *domain.Cache) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Cache, error)
	GetByName(ctx context.Context, userID uuid.UUID, name string) (*domain.Cache, error)
	List(ctx context.Context, userID uuid.UUID) ([]*domain.Cache, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
