// Package vault provides a HashiCorp Vault adapter for secret management.
package vault

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/hashicorp/vault/api"
)

// TransitKMSAdapter implements ports.KMSClient using Vault Transit Secrets Engine.
type TransitKMSAdapter struct {
	client *api.Client
}

// NewTransitKMSAdapter creates a new Vault Transit KMS adapter.
func NewTransitKMSAdapter(client *api.Client) *TransitKMSAdapter {
	return &TransitKMSAdapter{client: client}
}

// Encrypt encrypts plaintext using Vault Transit.
func (a *TransitKMSAdapter) Encrypt(ctx context.Context, keyID string, plaintext []byte) ([]byte, error) {
	// Vault Transit expects base64-encoded plaintext
	plaintextB64 := base64.StdEncoding.EncodeToString(plaintext)

	// Call Transit encrypt endpoint
	path := fmt.Sprintf("transit/encrypt/%s", keyID)
	secret, err := a.client.Logical().WriteWithContext(ctx, path, map[string]interface{}{
		"plaintext": plaintextB64,
	})
	if err != nil {
		return nil, fmt.Errorf("vault transit encrypt failed for key %s: %w", keyID, err)
	}
	if secret == nil {
		return nil, fmt.Errorf("vault transit encrypt returned nil secret for key %s", keyID)
	}

	// Response is base64-encoded ciphertext
	ciphertextB64, ok := secret.Data["ciphertext"].(string)
	if !ok {
		return nil, fmt.Errorf("vault transit encrypt response missing ciphertext for key %s", keyID)
	}

	return []byte(ciphertextB64), nil
}

// Decrypt decrypts ciphertext using Vault Transit.
func (a *TransitKMSAdapter) Decrypt(ctx context.Context, keyID string, ciphertext []byte) ([]byte, error) {
	// ciphertext is base64-encoded from Encrypt
	path := fmt.Sprintf("transit/decrypt/%s", keyID)
	secret, err := a.client.Logical().WriteWithContext(ctx, path, map[string]interface{}{
		"ciphertext": string(ciphertext),
	})
	if err != nil {
		return nil, fmt.Errorf("vault transit decrypt failed for key %s: %w", keyID, err)
	}
	if secret == nil {
		return nil, fmt.Errorf("vault transit decrypt returned nil secret for key %s", keyID)
	}

	plaintextB64, ok := secret.Data["plaintext"].(string)
	if !ok {
		return nil, fmt.Errorf("vault transit decrypt response missing plaintext for key %s", keyID)
	}

	plaintext, err := base64.StdEncoding.DecodeString(plaintextB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode plaintext from vault transit: %w", err)
	}

	return plaintext, nil
}

// GenerateKey generates a new random 256-bit key in Vault Transit.
func (a *TransitKMSAdapter) GenerateKey(ctx context.Context, keyID string) ([]byte, error) {
	// Call Transit generate-data-key endpoint (wraps a generated DEK with the transit key)
	path := fmt.Sprintf("transit/gen_wrapping_key/%s", keyID)
	secret, err := a.client.Logical().WriteWithContext(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("vault transit key generation failed for key %s: %w", keyID, err)
	}
	if secret == nil {
		return nil, fmt.Errorf("vault transit key generation returned nil secret for key %s", keyID)
	}

	// plaintext is base64-encoded wrapped key material
	plaintextB64, ok := secret.Data["plaintext"].(string)
	if !ok {
		return nil, fmt.Errorf("vault transit key generation response missing plaintext for key %s", keyID)
	}

	return []byte(plaintextB64), nil
}

// Ensure TransitKMSAdapter implements ports.KMSClient
var _ interface {
	Encrypt(ctx context.Context, keyID string, plaintext []byte) ([]byte, error)
	Decrypt(ctx context.Context, keyID string, ciphertext []byte) ([]byte, error)
	GenerateKey(ctx context.Context, keyID string) ([]byte, error)
} = (*TransitKMSAdapter)(nil)