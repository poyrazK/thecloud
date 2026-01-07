package ports

import (
	"context"
	"io"
)

// ComputeBackend abstracts the underlying infrastructure provider (Docker, Libvirt, etc.)
type ComputeBackend interface {
	// Instance Lifecycle
	CreateInstance(ctx context.Context, name, imageName string, ports []string, networkID string, volumeBinds []string, env []string, cmd []string) (string, error)
	StopInstance(ctx context.Context, id string) error
	DeleteInstance(ctx context.Context, id string) error
	GetInstanceLogs(ctx context.Context, id string) (io.ReadCloser, error)
	GetInstanceStats(ctx context.Context, id string) (io.ReadCloser, error)
	GetInstancePort(ctx context.Context, id string, internalPort string) (int, error)
	GetInstanceIP(ctx context.Context, id string) (string, error)

	// Execution
	Exec(ctx context.Context, id string, cmd []string) (string, error)
	RunTask(ctx context.Context, opts RunTaskOptions) (string, error)
	WaitTask(ctx context.Context, id string) (int64, error)

	// Network Management
	CreateNetwork(ctx context.Context, name string) (string, error)
	DeleteNetwork(ctx context.Context, id string) error

	// Volume Management
	CreateVolume(ctx context.Context, name string) error
	DeleteVolume(ctx context.Context, name string) error
	CreateVolumeSnapshot(ctx context.Context, volumeID string, destinationPath string) error
	RestoreVolumeSnapshot(ctx context.Context, volumeID string, sourcePath string) error

	// Health
	Ping(ctx context.Context) error
	Type() string
}
