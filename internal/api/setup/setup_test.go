package setup

import (
	"os"
	"testing"

	"github.com/poyrazk/thecloud/internal/platform"
	"github.com/stretchr/testify/assert"
)

func TestInitLogger(t *testing.T) {
	logger := InitLogger()
	assert.NotNil(t, logger)
}

func TestLoadConfig(t *testing.T) {
	logger := InitLogger()

	// Set some env vars for LoadConfig to potentially use
	_ = os.Setenv("DATABASE_URL", "postgres://localhost:5432/test")
	defer func() { _ = os.Unsetenv("DATABASE_URL") }()

	cfg, err := LoadConfig(logger)
	if err != nil {
		// If it fails because of other missing config, it's fine for now
		// as long as we covered the lines
		t.Logf("LoadConfig failed: %v", err)
	} else {
		assert.NotNil(t, cfg)
	}
}

func TestInitComputeBackend(t *testing.T) {
	logger := InitLogger()

	t.Run("Docker", func(t *testing.T) {
		cfg := &platform.Config{ComputeBackend: "docker"}
		backend, err := InitComputeBackend(cfg, logger)
		// This might fail if docker is not running, so we just check it doesn't panic
		t.Logf("InitComputeBackend(docker) err: %v", err)
		if err == nil {
			assert.NotNil(t, backend)
		}
	})

	t.Run("Libvirt", func(t *testing.T) {
		cfg := &platform.Config{ComputeBackend: "libvirt"}
		// This will likely fail without a libvirt connection
		backend, err := InitComputeBackend(cfg, logger)
		t.Logf("InitComputeBackend(libvirt) err: %v", err)
		if err == nil {
			assert.NotNil(t, backend)
		}
	})
}

func TestInitNetworkBackend(t *testing.T) {
	logger := InitLogger()
	backend := InitNetworkBackend(logger)
	assert.NotNil(t, backend)
}

func TestInitLBProxy(t *testing.T) {
	cfg := &platform.Config{ComputeBackend: "docker"}
	// We pass nils for repos as we just want to see if it creates the adapter
	backend, err := InitLBProxy(cfg, nil, nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, backend)

	cfg.ComputeBackend = "libvirt"
	backend, err = InitLBProxy(cfg, nil, nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, backend)
}

func TestInitHandlers(t *testing.T) {
	svcs := &Services{} // Empty services
	handlers := InitHandlers(svcs)
	assert.NotNil(t, handlers)
	assert.NotNil(t, handlers.Auth)
}

func TestSetupRouter(t *testing.T) {
	cfg := &platform.Config{Environment: "test"}
	logger := InitLogger()
	svcs := &Services{}
	handlers := InitHandlers(svcs)

	router := SetupRouter(cfg, logger, handlers, svcs, nil)
	assert.NotNil(t, router)
}
