//go:build linux

package firecracker

import (
	"context"
	"log/slog"
	"os"
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

	t.Run("InvalidSocketDir", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "fc-not-a-dir")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = NewFirecrackerAdapter(logger, Config{SocketDir: tmpFile.Name()})
		assert.Error(t, err)
	})
}

func TestFirecrackerAdapter_DeleteInstance(t *testing.T) {
	logger := slog.Default()
	cfg := Config{
		SocketDir: t.TempDir(),
		MockMode:  true,
	}
	adapter, err := NewFirecrackerAdapter(logger, cfg)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("InvalidID", func(t *testing.T) {
		err := adapter.DeleteInstance(ctx, "../invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid instance ID format")
	})

	t.Run("NonExistentInstance", func(t *testing.T) {
		err := adapter.DeleteInstance(ctx, "nonexistent")
		assert.NoError(t, err) // Should return nil if not found
	})
}

func TestFirecrackerAdapter_WaitTask_Mock(t *testing.T) {
	logger := slog.Default()
	cfg := Config{
		MockMode: true,
	}
	adapter, err := NewFirecrackerAdapter(logger, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	exitCode, err := adapter.WaitTask(ctx, "any")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), exitCode)
}

func TestFirecrackerAdapter_UnimplementedMethods(t *testing.T) {
	logger := slog.Default()
	adapter, err := NewFirecrackerAdapter(logger, Config{MockMode: true})
	require.NoError(t, err)
	ctx := context.Background()

	t.Run("GetInstanceLogs", func(t *testing.T) {
		_, err := adapter.GetInstanceLogs(ctx, "id")
		assert.Error(t, err)
	})

	t.Run("GetInstanceStats", func(t *testing.T) {
		_, err := adapter.GetInstanceStats(ctx, "id")
		assert.Error(t, err)
	})

	t.Run("GetInstancePort", func(t *testing.T) {
		_, err := adapter.GetInstancePort(ctx, "id", "80")
		assert.Error(t, err)
	})

	t.Run("GetInstanceIP", func(t *testing.T) {
		ip, err := adapter.GetInstanceIP(ctx, "id")
		assert.NoError(t, err)
		assert.Equal(t, "0.0.0.0", ip)
	})

	t.Run("GetConsoleURL", func(t *testing.T) {
		_, err := adapter.GetConsoleURL(ctx, "id")
		assert.Error(t, err)
	})

	t.Run("Exec", func(t *testing.T) {
		_, err := adapter.Exec(ctx, "id", []string{"ls"})
		assert.Error(t, err)
	})

	t.Run("AttachDetachVolume", func(t *testing.T) {
		err := adapter.AttachVolume(ctx, "id", "path")
		assert.Error(t, err)
		err = adapter.DetachVolume(ctx, "id", "path")
		assert.Error(t, err)
	})

	t.Run("Ping", func(t *testing.T) {
		err := adapter.Ping(ctx)
		assert.NoError(t, err)
	})

	t.Run("Type", func(t *testing.T) {
		assert.Equal(t, "firecracker-mock", adapter.Type())
	})
}
