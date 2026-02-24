//go:build !linux

package firecracker

import (
	"context"
	"log/slog"
	"testing"

	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFirecrackerAdapterInterfaceCompliance(t *testing.T) {
	var _ ports.ComputeBackend = (*FirecrackerAdapter)(nil)
}

func TestNewFirecrackerAdapter(t *testing.T) {
	logger := slog.Default()
	cfg := Config{}

	adapter, err := NewFirecrackerAdapter(logger, cfg)
	require.NoError(t, err)

	assert.NotNil(t, adapter)
	assert.Equal(t, "firecracker-noop", adapter.Type())
}

func TestFirecrackerAdapterNoopMethods(t *testing.T) {
	logger := slog.Default()
	adapter, err := NewFirecrackerAdapter(logger, Config{})
	require.NoError(t, err)
	ctx := context.Background()

	t.Run("LaunchInstanceWithOptions", func(t *testing.T) {
		_, _, err := adapter.LaunchInstanceWithOptions(ctx, ports.CreateInstanceOptions{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not supported")
	})

	t.Run("StartInstance", func(t *testing.T) {
		err := adapter.StartInstance(ctx, "id")
		require.Error(t, err)
	})

	t.Run("StopInstance", func(t *testing.T) {
		err := adapter.StopInstance(ctx, "id")
		require.Error(t, err)
	})

	t.Run("DeleteInstance", func(t *testing.T) {
		err := adapter.DeleteInstance(ctx, "id")
		require.NoError(t, err) // Delete is safe
	})

	t.Run("GetInstanceLogs", func(t *testing.T) {
		_, err := adapter.GetInstanceLogs(ctx, "id")
		require.Error(t, err)
	})

	t.Run("GetInstanceIP", func(t *testing.T) {
		ip, err := adapter.GetInstanceIP(ctx, "id")
		require.NoError(t, err)
		assert.Equal(t, "0.0.0.0", ip)
	})

	t.Run("Ping", func(t *testing.T) {
		err := adapter.Ping(ctx)
		require.NoError(t, err)
	})

	t.Run("AttachVolume", func(t *testing.T) {
		err := adapter.AttachVolume(ctx, "id", "path")
		require.Error(t, err)
	})

	t.Run("GetInstanceStats", func(t *testing.T) {
		res, err := adapter.GetInstanceStats(ctx, "id")
		require.Error(t, err)
		assert.Nil(t, res)
	})

	t.Run("GetInstancePort", func(t *testing.T) {
		res, err := adapter.GetInstancePort(ctx, "id", "80")
		require.Error(t, err)
		assert.Equal(t, 0, res)
	})

	t.Run("GetConsoleURL", func(t *testing.T) {
		res, err := adapter.GetConsoleURL(ctx, "id")
		require.Error(t, err)
		assert.Empty(t, res)
	})

	t.Run("Exec", func(t *testing.T) {
		res, err := adapter.Exec(ctx, "id", []string{"ls"})
		require.Error(t, err)
		assert.Empty(t, res)
	})

	t.Run("RunTask", func(t *testing.T) {
		id, ips, err := adapter.RunTask(ctx, ports.RunTaskOptions{})
		require.Error(t, err)
		assert.Empty(t, id)
		assert.Empty(t, ips)
	})

	t.Run("WaitTask", func(t *testing.T) {
		code, err := adapter.WaitTask(ctx, "id")
		require.Error(t, err)
		assert.Equal(t, int64(-1), code)
	})

	t.Run("CreateNetwork", func(t *testing.T) {
		_, err := adapter.CreateNetwork(ctx, "10.0.0.0/24")
		require.Error(t, err)
	})

	t.Run("DeleteNetwork", func(t *testing.T) {
		err := adapter.DeleteNetwork(ctx, "id")
		require.NoError(t, err)
	})

	t.Run("DetachVolume", func(t *testing.T) {
		err := adapter.DetachVolume(ctx, "id", "path")
		require.Error(t, err)
	})
}
