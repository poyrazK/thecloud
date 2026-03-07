package ports

import (
	"context"
)

// SecretsManager defines the interface for an external secrets storage backend (e.g., HashiCorp Vault).
type SecretsManager interface {
	// StoreSecret saves a secret at the specified path.
	StoreSecret(ctx context.Context, path string, data map[string]interface{}) error
	// GetSecret retrieves a secret from the specified path.
	GetSecret(ctx context.Context, path string) (map[string]interface{}, error)
	// DeleteSecret removes a secret from the specified path.
	DeleteSecret(ctx context.Context, path string) error
	// Ping checks the health of the secrets manager backend.
	Ping(ctx context.Context) error
}
