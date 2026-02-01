package setup

import (
	"log/slog"
	"os"
	"testing"

	"github.com/poyrazk/thecloud/internal/platform"
	"github.com/stretchr/testify/assert"
)

func TestInitComputeBackend(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	t.Run("noop", func(t *testing.T) {
		cfg := &platform.Config{ComputeBackend: "noop"}
		backend, err := InitComputeBackend(cfg, logger)
		assert.NoError(t, err)
		assert.Equal(t, "noop", backend.Type())
	})

	t.Run("docker", func(t *testing.T) {
		cfg := &platform.Config{ComputeBackend: "docker"}
		backend, err := InitComputeBackend(cfg, logger)
		assert.NoError(t, err)
		assert.Equal(t, "docker", backend.Type())
	})
}

func TestInitStorageBackend(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	t.Run("noop", func(t *testing.T) {
		cfg := &platform.Config{StorageBackend: "noop"}
		backend, err := InitStorageBackend(cfg, logger)
		assert.NoError(t, err)
		assert.NotNil(t, backend)
	})

	t.Run("default to noop", func(t *testing.T) {
		cfg := &platform.Config{StorageBackend: "invalid"}
		backend, err := InitStorageBackend(cfg, logger)
		assert.NoError(t, err)
		assert.NotNil(t, backend)
	})
}

func TestInitNetworkBackend(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	t.Run("noop", func(t *testing.T) {
		cfg := &platform.Config{NetworkBackend: "noop"}
		backend := InitNetworkBackend(cfg, logger)
		assert.Equal(t, "noop", backend.Type())
	})
}
