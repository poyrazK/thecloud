// Package ports defines service and repository interfaces.
package ports

import "context"

// SecretsManager defines a backend for database credential storage.
type SecretsManager interface {
	StoreSecret(ctx context.Context, path string, data map[string]interface{}) error
	GetSecret(ctx context.Context, path string) (map[string]interface{}, error)
	DeleteSecret(ctx context.Context, path string) error
	Ping(ctx context.Context) error
}
