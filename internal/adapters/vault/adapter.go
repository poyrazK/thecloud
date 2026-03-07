// Package vault provides a HashiCorp Vault adapter for secret management.
package vault

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/hashicorp/vault/api"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

// Adapter implements ports.SecretsManager using HashiCorp Vault.
type Adapter struct {
	client *api.Client
	logger *slog.Logger
}

// NewVaultAdapter creates a new Vault adapter.
func NewVaultAdapter(address, token string, logger *slog.Logger) (*Adapter, error) {
	config := api.DefaultConfig()
	config.Address = address

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	client.SetToken(token)

	return &Adapter{
		client: client,
		logger: logger,
	}, nil
}

// StoreSecret saves a secret at the specified path.
func (a *Adapter) StoreSecret(_ context.Context, path string, data map[string]interface{}) error {
	_, err := a.client.Logical().Write(path, data)
	if err != nil {
		return fmt.Errorf("failed to store secret in vault: %w", err)
	}
	return nil
}

// GetSecret retrieves a secret from the specified path.
func (a *Adapter) GetSecret(_ context.Context, path string) (map[string]interface{}, error) {
	secret, err := a.client.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret from vault: %w", err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("secret not found at path: %s", path)
	}
	return secret.Data, nil
}

// DeleteSecret removes a secret from the specified path.
func (a *Adapter) DeleteSecret(_ context.Context, path string) error {
	_, err := a.client.Logical().Delete(path)
	if err != nil {
		return fmt.Errorf("failed to delete secret from vault: %w", err)
	}
	return nil
}

// Ping checks the connectivity and health of the Vault server.
func (a *Adapter) Ping(_ context.Context) error {
	_, err := a.client.Sys().Health()
	if err != nil {
		return fmt.Errorf("vault health check failed: %w", err)
	}
	return nil
}

// Ensure Adapter implements ports.SecretsManager
var _ ports.SecretsManager = (*Adapter)(nil)
