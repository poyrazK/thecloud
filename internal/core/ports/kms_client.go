// Package ports defines service and repository interfaces.
package ports

import (
	"context"
)

// KMSClient handles key encryption and decryption operations via a KMS backend.
type KMSClient interface {
	// Encrypt encrypts plaintext using the specified key.
	Encrypt(ctx context.Context, keyID string, plaintext []byte) ([]byte, error)
	// Decrypt decrypts ciphertext using the specified key.
	Decrypt(ctx context.Context, keyID string, ciphertext []byte) ([]byte, error)
	// GenerateKey generates a new wrapped DEK under the specified key ID.
	// Returns the wrapped/encrypted DEK bytes; callers should use Decrypt to unwrap.
	GenerateKey(ctx context.Context, keyID string) ([]byte, error)
}
