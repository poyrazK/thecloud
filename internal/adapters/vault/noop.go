package vault

import (
	"context"
)

// NoOpSecretsManager is a mock/no-op implementation of SecretsManager for testing/dev.
type NoOpSecretsManager struct{}

func NewNoOpSecretsManager() *NoOpSecretsManager {
	return &NoOpSecretsManager{}
}

func (m *NoOpSecretsManager) StoreSecret(ctx context.Context, path string, data map[string]interface{}) error {
	return nil
}

func (m *NoOpSecretsManager) GetSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	return nil, nil
}

func (m *NoOpSecretsManager) DeleteSecret(ctx context.Context, path string) error {
	return nil
}

func (m *NoOpSecretsManager) Ping(ctx context.Context) error {
	return nil
}
