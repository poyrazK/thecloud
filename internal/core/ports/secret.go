// Package ports defines service and repository interfaces.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

// SecretRepository handles the persistence of sensitive data and encryption-at-rest metadata.
type SecretRepository interface {
	// Create persists a new secret record (encrypted).
	Create(ctx context.Context, secret *domain.Secret) error
	// GetByID retrieves a secret's metadata by its UUID.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Secret, error)
	// GetByName retrieves a secret's metadata by its unique name (within user scope).
	GetByName(ctx context.Context, name string) (*domain.Secret, error)
	// List returns metadata for all secrets owned by the current user.
	List(ctx context.Context) ([]*domain.Secret, error)
	// Update modifies a secret's value or metadata.
	Update(ctx context.Context, secret *domain.Secret) error
	// Delete removes a secret permanently from storage.
	Delete(ctx context.Context, id uuid.UUID) error
}

// SecretService provides business logic for secure credential storage and retrieval (e.g., KMS/Vault-like).
type SecretService interface {
	// CreateSecret encrypts and stores a new sensitive configuration value.
	CreateSecret(ctx context.Context, name, value, description string) (*domain.Secret, error)
	// GetSecret retrieves a secret's metadata and its decrypted value.
	GetSecret(ctx context.Context, id uuid.UUID) (*domain.Secret, error)
	// GetSecretByName retrieves a secret by name, including its decrypted value.
	GetSecretByName(ctx context.Context, name string) (*domain.Secret, error)
	// ListSecrets returns all secrets (metadata only) for the current authorized user.
	ListSecrets(ctx context.Context) ([]*domain.Secret, error)
	// DeleteSecret decommissioning a sensitive configuration value.
	DeleteSecret(ctx context.Context, id uuid.UUID) error
}
