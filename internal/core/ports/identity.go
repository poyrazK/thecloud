// Package ports defines service and repository interfaces.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

// IdentityRepository manages the persistent storage of programmatic credentials (API keys).
type IdentityRepository interface {
	// CreateAPIKey saves a new API key record.
	CreateAPIKey(ctx context.Context, apiKey *domain.APIKey) error
	// GetAPIKeyByKey retrieves a key's metadata based on its raw string value.
	GetAPIKeyByKey(ctx context.Context, key string) (*domain.APIKey, error)
	// GetAPIKeyByID retrieves key metadata by its unique UUID.
	GetAPIKeyByID(ctx context.Context, id uuid.UUID) (*domain.APIKey, error)
	// ListAPIKeysByUserID returns all active and revoked keys for a specific user.
	ListAPIKeysByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.APIKey, error)
	// DeleteAPIKey removes an API key record from storage.
	DeleteAPIKey(ctx context.Context, id uuid.UUID) error
}

// IdentityService provides business logic for managing programmatic access and multi-user authentication.
type IdentityService interface {
	// CreateKey generates a new unique API key for a user.
	CreateKey(ctx context.Context, userID uuid.UUID, name string) (*domain.APIKey, error)
	// ValidateAPIKey authenticates a request using a raw API key string.
	ValidateAPIKey(ctx context.Context, key string) (*domain.APIKey, error)
	// ListKeys returns all keys associated with an authorized user.
	ListKeys(ctx context.Context, userID uuid.UUID) ([]*domain.APIKey, error)
	// RevokeKey immediately invalidates an API key.
	RevokeKey(ctx context.Context, userID uuid.UUID, id uuid.UUID) error
	// RotateKey invalidates an existing key and generates a replacement.
	RotateKey(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*domain.APIKey, error)
}
