package vault

import (
	"context"
	"fmt"
	"sync"
)

// NoOpSecretsManager is a mock/no-op implementation of SecretsManager for testing/dev.
// It now stores secrets in memory to support basic E2E workflows without a real Vault.
type NoOpSecretsManager struct {
	secrets map[string]map[string]interface{}
	mu      sync.RWMutex
}

func NewNoOpSecretsManager() *NoOpSecretsManager {
	return &NoOpSecretsManager{
		secrets: make(map[string]map[string]interface{}),
	}
}

func (m *NoOpSecretsManager) StoreSecret(ctx context.Context, path string, data map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.secrets[path] = data
	return nil
}

func (m *NoOpSecretsManager) GetSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, ok := m.secrets[path]
	if !ok {
		return nil, fmt.Errorf("secret not found at path: %s", path)
	}
	return data, nil
}

func (m *NoOpSecretsManager) DeleteSecret(ctx context.Context, path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.secrets, path)
	return nil
}

func (m *NoOpSecretsManager) Ping(ctx context.Context) error {
	return nil
}
