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

func TestFirecrackerAdapter_InterfaceCompliance(t *testing.T) {
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

func TestFirecrackerAdapter_NoopMethods(t *testing.T) {
	logger := slog.Default()
	adapter, err := NewFirecrackerAdapter(logger, Config{})
	require.NoError(t, err)
	ctx := context.Background()

	t.Run("LaunchInstanceWithOptions", func(t *testing.T) {
		_, _, err := adapter.LaunchInstanceWithOptions(ctx, ports.CreateInstanceOptions{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not supported")
	})

	t.Run("StartInstance", func(t *testing.T) {
		err := adapter.StartInstance(ctx, "id")
		assert.Error(t, err)
	})

	t.Run("StopInstance", func(t *testing.T) {
		err := adapter.StopInstance(ctx, "id")
		assert.Error(t, err)
	})

	t.Run("DeleteInstance", func(t *testing.T) {
		err := adapter.DeleteInstance(ctx, "id")
		assert.NoError(t, err) // Delete is safe
	})

	t.Run("GetInstanceLogs", func(t *testing.T) {
		_, err := adapter.GetInstanceLogs(ctx, "id")
		assert.Error(t, err)
	})

	t.Run("GetInstanceIP", func(t *testing.T) {
		ip, err := adapter.GetInstanceIP(ctx, "id")
		assert.NoError(t, err)
		assert.Equal(t, "0.0.0.0", ip)
	})

	t.Run("Ping", func(t *testing.T) {
		err := adapter.Ping(ctx)
		assert.NoError(t, err)
	})

	t.Run("AttachVolume", func(t *testing.T) {
		err := adapter.AttachVolume(ctx, "id", "path")
		assert.Error(t, err)
	})
}
