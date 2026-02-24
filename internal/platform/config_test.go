package platform

import (
	"os"
	"testing"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	// Save original env and restore after test
	originalPort := os.Getenv("PORT")
	err := os.Setenv("PORT", originalPort)
	require.NoError(t, err)

	t.Run("Default values", func(t *testing.T) {
		err := os.Unsetenv("PORT")
		require.NoError(t, err)
		cfg, err := NewConfig()
		require.NoError(t, err)
		assert.Equal(t, "8080", cfg.Port)
	})

	t.Run("Env override", func(t *testing.T) {
		err := os.Setenv("PORT", "9090")
		require.NoError(t, err)
		cfg, err := NewConfig()
		require.NoError(t, err)
		assert.Equal(t, "9090", cfg.Port)
	})
}

func TestGetEnv(t *testing.T) {
	t.Run("Existing env", func(t *testing.T) {
		err := os.Setenv("TEST_KEY", "test_value")
		require.NoError(t, err)
		defer func() { _ = os.Unsetenv("TEST_KEY") }()
		assert.Equal(t, "test_value", getEnv("TEST_KEY", "fallback"))
	})

	t.Run("Fallback value", func(t *testing.T) {
		err := os.Unsetenv("NON_EXISTENT_KEY")
		require.NoError(t, err)
		assert.Equal(t, "fallback", getEnv("NON_EXISTENT_KEY", "fallback"))
	})
}
