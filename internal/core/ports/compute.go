// Package ports defines service and repository interfaces.
package ports

import (
	"context"
	"io"
)

// ComputeBackend abstracts the underlying infrastructure provider (Docker, Libvirt, Cloud Hypervisor, etc.)
// It provides a unified set of operations to manage compute instances and their environment.
type ComputeBackend interface {
	// Instance Lifecycle

	// LaunchInstanceWithOptions provisions a new compute entity based on the provided options.
	// Returns the unique identifier of the instance and any allocated ports.
	LaunchInstanceWithOptions(ctx context.Context, opts CreateInstanceOptions) (string, []string, error)
	// StartInstance boots up a stopped instance.
	StartInstance(ctx context.Context, id string) error
	// StopInstance gracefully shuts down or forcibly terminates a running instance.
	StopInstance(ctx context.Context, id string) error
	// DeleteInstance removes an instance and its ephemeral resources.
	DeleteInstance(ctx context.Context, id string) error
	// GetInstanceLogs returns a stream of stdout/stderr from the instance.
	GetInstanceLogs(ctx context.Context, id string) (io.ReadCloser, error)
	// GetInstanceStats retrieves real-time resource utilization (CPU, RAM, Network).
	GetInstanceStats(ctx context.Context, id string) (io.ReadCloser, error)
	// GetInstancePort retrieves the mapped host port for a specific internal port.
	GetInstancePort(ctx context.Context, id string, internalPort string) (int, error)
	// GetInstanceIP retrieves the primary internal IP address of the instance.
	GetInstanceIP(ctx context.Context, id string) (string, error)
	// GetConsoleURL returns a secure URL for interacting with the instance's serial/VNC console.
	GetConsoleURL(ctx context.Context, id string) (string, error)

	// Execution

	// Exec runs a command within the context of a running instance and returns its output.
	Exec(ctx context.Context, id string, cmd []string) (string, error)
	// RunTask launches a short-lived execution task.
	RunTask(ctx context.Context, opts RunTaskOptions) (string, []string, error)
	// WaitTask blocks until a specific task completes and returns its exit code.
	WaitTask(ctx context.Context, id string) (int64, error)

	// Network Management

	// CreateNetwork establishes a new L2/L3 isolation boundary.
	CreateNetwork(ctx context.Context, name string) (string, error)
	// DeleteNetwork removes a network boundary.
	DeleteNetwork(ctx context.Context, id string) error

	// Volume/Disk Attachment (Physical/Block)

	// AttachVolume connects a storage resource to the instance.
	AttachVolume(ctx context.Context, id string, volumePath string) error
	// DetachVolume disconnects a storage resource from the instance.
	DetachVolume(ctx context.Context, id string, volumePath string) error

	// Health

	// Ping checks the connectivity and health of the backend provider.
	Ping(ctx context.Context) error
	// Type returns a string identifier of the backend (e.g., "docker", "kvm").
	Type() string
}
