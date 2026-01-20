// Package ports defines service and repository interfaces.
package ports

import (
	"context"
	"io"
)

// RunTaskOptions encapsulates configuration for executing a one-off containerized task.
type RunTaskOptions struct {
	Name            string   // Optional name for the task container
	Image           string   // Container image to use
	Command         []string // Command to execute within the container
	Env             []string // Environment variables (e.g., "KEY=VALUE")
	MemoryMB        int64    // RAM limit in MegaBytes
	CPUs            float64  // CPU unit limit (e.g., 0.5 for half a core)
	NetworkDisabled bool     // If true, the container will have no network access
	ReadOnlyRootfs  bool     // If true, the container's root filesystem is mounted as read-only
	PidsLimit       *int64   // Maximum number of processes allowed within the container
	WorkingDir      string   // Initial directory for the command
	Binds           []string // Host-to-container path mappings
}

// DockerClient defines high-level operations for interacting with a containerization engine (e.g., Docker Daemon).
type DockerClient interface {
	// CreateContainer provisions a new container but does not start it until configured.
	CreateContainer(ctx context.Context, name, image string, ports []string, networkID string, volumeBinds []string, env []string, cmd []string) (string, error)
	// StopContainer gracefully shuts down a running container.
	StopContainer(ctx context.Context, containerID string) error
	// RemoveContainer deletes an existing container and its non-persistent resources.
	RemoveContainer(ctx context.Context, containerID string) error
	// GetLogs opens a stream to the container's stdout and stderr.
	GetLogs(ctx context.Context, containerID string) (io.ReadCloser, error)
	// GetContainerStats returns a stream of real-time resource usage data.
	GetContainerStats(ctx context.Context, containerID string) (io.ReadCloser, error)
	// GetContainerPort looks up the dynamically assigned host port mapping for a container port.
	GetContainerPort(ctx context.Context, containerID string, containerPort string) (int, error)
	// CreateNetwork establishes a new virtual network for container communication.
	CreateNetwork(ctx context.Context, name string) (string, error)
	// RemoveNetwork deletes an existing virtual network.
	RemoveNetwork(ctx context.Context, networkID string) error
	// CreateVolume provisions a persistent storage volume managed by the engine.
	CreateVolume(ctx context.Context, name string) error
	// DeleteVolume removes a storage volume (only if not in use).
	DeleteVolume(ctx context.Context, name string) error
	// RunTask launches a detached executing container for batch work.
	RunTask(ctx context.Context, opts RunTaskOptions) (string, error)
	// WaitContainer blocks until a container exits and returns the status code.
	WaitContainer(ctx context.Context, containerID string) (int64, error)
	// Exec runs a supplemental command in an already running container.
	Exec(ctx context.Context, containerID string, cmd []string) (string, error)
	// Ping verifies the connectivity and responsiveness of the Docker engine.
	Ping(ctx context.Context) error
}
