//go:build integration

package docker

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/poyrazk/thecloud/internal/core/ports"
)

func TestDockerSSHInjectionIntegration(t *testing.T) {
	adapter, err := NewDockerAdapter(slog.Default())
	require.NoError(t, err)
	ctx := context.Background()

	t.Run("Inject SSH Keys via Cloud-Config", func(t *testing.T) {
		name := "ssh-test-" + time.Now().Format("20060102150405")
		const testKey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC test-key"
		userData := "#cloud-config\nssh_authorized_keys:\n  - " + testKey

		// 1. Launch with UserData
		id, err := adapter.LaunchInstanceWithOptions(ctx, ports.CreateInstanceOptions{
			Name:      name,
			ImageName: "alpine",
			UserData:  userData,
		})
		require.NoError(t, err)
		defer func() { _ = adapter.DeleteInstance(ctx, id) }()

		// 2. Verify file exists and has content
		require.Eventually(t, func() bool {
			out, err := adapter.Exec(ctx, id, []string{"cat", "/root/.ssh/authorized_keys"})
			if err != nil {
				return false
			}
			return assert.Contains(t, out, testKey)
		}, 15*time.Second, 1*time.Second, "SSH key not found in /root/.ssh/authorized_keys")

		// 3. Verify permissions
		out, err := adapter.Exec(ctx, id, []string{"stat", "-c", "%a", "/root/.ssh/authorized_keys"})
		require.NoError(t, err)
		assert.Contains(t, out, "600")

		out, err = adapter.Exec(ctx, id, []string{"stat", "-c", "%a", "/root/.ssh"})
		require.NoError(t, err)
		assert.Contains(t, out, "700")
	})
}
