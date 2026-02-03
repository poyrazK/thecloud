package testutil

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
)

// FaultInjector provides a reusable mechanism to inject errors into repositories.
type FaultInjector struct {
	ShouldFail bool
	ErrorMsg   string
}

// CheckFail returns an error if ShouldFail is true.
func (f *FaultInjector) CheckFail() error {
	if f.ShouldFail {
		msg := f.ErrorMsg
		if msg == "" {
			msg = "simulated failure"
		}
		return fmt.Errorf("%s", msg)
	}
	return nil
}

// KillContainer forcefully kills a Docker container by ID.
// This is useful for simulating sudden failures (e.g., OOM kill, power loss).
func KillContainer(ctx context.Context, containerID string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer func() { _ = cli.Close() }()

	return cli.ContainerKill(ctx, containerID, "SIGKILL")
}
