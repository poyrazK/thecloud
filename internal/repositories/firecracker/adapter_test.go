//go:build linux

package firecracker

import (
	"log/slog"
	"testing"

	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFirecrackerAdapter_InterfaceCompliance(t *testing.T) {
	var _ ports.ComputeBackend = (*FirecrackerAdapter)(nil)
}

func TestNewFirecrackerAdapter(t *testing.T) {
	logger := slog.Default()
	cfg := Config{
		BinaryPath: "/usr/local/bin/firecracker",
		KernelPath: "/var/lib/thecloud/vmlinux",
		RootfsPath: "/var/lib/thecloud/rootfs.ext4",
		SocketDir:  "/tmp/firecracker-test",
	}

	adapter, err := NewFirecrackerAdapter(logger, cfg)
	require.NoError(t, err)

	assert.NotNil(t, adapter)
	assert.Equal(t, "firecracker", adapter.Type())
	assert.Equal(t, cfg.SocketDir, adapter.cfg.SocketDir)
}
