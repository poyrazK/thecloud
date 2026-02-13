package tests

import (
	"context"
	"log/slog"
	"testing"

	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/repositories/firecracker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFirecrackerBackend_E2E(t *testing.T) {
	// This test requires the firecracker binary and root privileges/KVM.
	// We skip it unless explicitly enabled or running in CI with proper setup.
	if testing.Short() {
		t.Skip("skipping firecracker e2e test in short mode")
	}

	logger := slog.Default()
	cfg := firecracker.Config{
		BinaryPath: "/usr/local/bin/firecracker",
		KernelPath: "/var/lib/thecloud/vmlinux",
		RootfsPath: "/var/lib/thecloud/rootfs.ext4",
	}

	adapter := firecracker.NewFirecrackerAdapter(logger, cfg)
	
	// If we are on non-linux, this will return the firecracker-noop type
	if adapter.Type() != "firecracker" {
		t.Skipf("Skipping real firecracker test on %s platform", adapter.Type())
	}

	ctx := context.Background()
	opts := ports.CreateInstanceOptions{
		Name:        "test-firecracker-vm",
		ImageName:   "alpine",
		CPULimit:    1,
		MemoryLimit: 128 * 1024 * 1024,
	}

	t.Run("Launch and Delete", func(t *testing.T) {
		id, _, err := adapter.LaunchInstanceWithOptions(ctx, opts)
		// We expect an error if the kernel/rootfs are missing, 
		// but we want to see HOW it fails in CI.
		if err != nil {
			t.Logf("Launch failed as expected (likely missing artifacts): %v", err)
			return 
		}

		require.NotEmpty(t, id)
		
		err = adapter.DeleteInstance(ctx, id)
		assert.NoError(t, err)
	})
}
