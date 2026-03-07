package vault

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/hashicorp/vault/api"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

// VaultAdapter implements ports.SecretsManager using HashiCorp Vault.
type VaultAdapter struct {
	client *api.Client
	logger *slog.Logger
}

// NewVaultAdapter constructs a new VaultAdapter.
func NewVaultAdapter(address, token string, logger *slog.Logger) (*VaultAdapter, error) {
	config := api.DefaultConfig()
	config.Address = address

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	client.SetToken(token)

	return &VaultAdapter{
		client: client,
		logger: logger,
	}, nil
}

func (a *VaultAdapter) StoreSecret(ctx context.Context, path string, data map[string]interface{}) error {
	_, err := a.client.Logical().Write(path, data)
	if err != nil {
		return fmt.Errorf("failed to write secret to vault at %s: %w", path, err)
	}
	return nil
}

func (a *VaultAdapter) GetSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	secret, err := a.client.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret from vault at %s: %w", path, err)
	}

	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("secret not found at path: %s", path)
	}

	return secret.Data, nil
}

func (a *VaultAdapter) DeleteSecret(ctx context.Context, path string) error {
	_, err := a.client.Logical().Delete(path)
	if err != nil {
		return fmt.Errorf("failed to delete secret from vault at %s: %w", path, err)
	}
	return nil
}

func (a *VaultAdapter) Ping(ctx context.Context) error {
	_, err := a.client.Sys().Health()
	if err != nil {
		return fmt.Errorf("vault health check failed: %w", err)
	}
	return nil
}

// Ensure VaultAdapter implements ports.SecretsManager
var _ ports.SecretsManager = (*VaultAdapter)(nil)
