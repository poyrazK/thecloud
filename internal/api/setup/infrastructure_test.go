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

	t.Run("firecracker", func(t *testing.T) {
		cfg := &platform.Config{
			ComputeBackend:      "firecracker",
			FirecrackerMockMode: true,
		}
		backend, err := InitComputeBackend(cfg, logger)
		assert.NoError(t, err)
		// type depends on platform if not in mock mode, but with mock mode it should be firecracker-mock
		assert.Contains(t, backend.Type(), "firecracker")
	})

	t.Run("libvirt", func(t *testing.T) {
		cfg := &platform.Config{
			ComputeBackend: "libvirt",
			LibvirtURI:     "test:///default",
		}
		backend, err := InitComputeBackend(cfg, logger)
		if err != nil {
			t.Skipf("skipping libvirt test as Init failed (likely missing libvirt-dev): %v", err)
		}
		assert.Equal(t, "libvirt", backend.Type())
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

	t.Run("lvm", func(t *testing.T) {
		cfg := &platform.Config{
			StorageBackend: "lvm",
			LvmVgName:      "test-vg",
		}
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

func TestInitLBProxy(t *testing.T) {
	cfg := &platform.Config{}
	
	t.Run("default", func(t *testing.T) {
		proxy, err := InitLBProxy(cfg, nil, nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, proxy)
	})

	t.Run("libvirt", func(t *testing.T) {
		cfgLibvirt := &platform.Config{ComputeBackend: "libvirt"}
		proxy, err := InitLBProxy(cfgLibvirt, nil, nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, proxy)
	})
}
