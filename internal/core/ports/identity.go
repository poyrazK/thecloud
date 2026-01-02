package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

type IdentityRepository interface {
	CreateApiKey(ctx context.Context, apiKey *domain.ApiKey) error
	GetApiKeyByKey(ctx context.Context, key string) (*domain.ApiKey, error)
	// list, delete etc can be added later
}

type IdentityService interface {
	CreateKey(ctx context.Context, userID uuid.UUID, name string) (*domain.ApiKey, error)
	ValidateApiKey(ctx context.Context, key string) (bool, error)
}
