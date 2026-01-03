//go:build integration

package docker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDockerAdapter_Integration(t *testing.T) {
	adapter, err := NewDockerAdapter()
	require.NoError(t, err)
	ctx := context.Background()

	t.Run("Container Lifecycle", func(t *testing.T) {
		name := "integration-test-container-" + time.Now().Format("20060102150405")
		image := "alpine"

		// 1. Create
		// Using a minimal sleep command so it stays running but exits eventually
		id, err := adapter.CreateContainer(ctx, name, image, []string{}, "", []string{}, nil, nil)
		require.NoError(t, err)
		assert.NotEmpty(t, id)

		// 2. Stop
		err = adapter.StopContainer(ctx, id)
		assert.NoError(t, err)

		// 3. Remove
		err = adapter.RemoveContainer(ctx, id)
		assert.NoError(t, err)
	})

	t.Run("Network Lifecycle", func(t *testing.T) {
		netName := "integration-test-net-" + time.Now().Format("20060102150405")

		// 1. Create
		id, err := adapter.CreateNetwork(ctx, netName)
		require.NoError(t, err)
		assert.NotEmpty(t, id)

		// 2. Remove
		err = adapter.RemoveNetwork(ctx, id)
		assert.NoError(t, err)
	})
}
