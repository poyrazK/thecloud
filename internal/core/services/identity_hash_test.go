package services

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeKeyHash(t *testing.T) {
	// Test that the hash function produces consistent output
	key := "test-api-key-123"
	hash1 := computeKeyHash(key)
	hash2 := computeKeyHash(key)

	// Should be deterministic
	require.NotEmpty(t, hash1)
	assert.Equal(t, hash1, hash2)

	// Different keys should produce different hashes
	differentHash := computeKeyHash("different-key")
	assert.NotEqual(t, hash1, differentHash)
}

func TestGetServerSecret(t *testing.T) {
	// Save original value
	origVal := os.Getenv("SECRETS_ENCRYPTION_KEY")
	defer func() {
		if origVal != "" {
			os.Setenv("SECRETS_ENCRYPTION_KEY", origVal)
		} else {
			os.Unsetenv("SECRETS_ENCRYPTION_KEY")
		}
	}()

	t.Run("WithEnvVar", func(t *testing.T) {
		os.Setenv("SECRETS_ENCRYPTION_KEY", "test-secret-key")
		// getServerSecret reads env directly, no need to modify global
		secret := getServerSecret()
		assert.Equal(t, "test-secret-key", secret)
	})

	t.Run("WithoutEnvVar", func(t *testing.T) {
		os.Unsetenv("SECRETS_ENCRYPTION_KEY")
		// getServerSecret will return fallback
		secret := getServerSecret()
		assert.Equal(t, "thecloud-development-secret-do-not-use-in-production", secret)
	})
}