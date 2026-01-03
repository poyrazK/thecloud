package ports

import (
	"context"
	"io"
)

type RunTaskOptions struct {
	Image           string
	Command         []string
	Env             []string
	MemoryMB        int64
	CPUs            float64
	NetworkDisabled bool
	ReadOnlyRootfs  bool
	PidsLimit       *int64
	WorkingDir      string
	Binds           []string
}

// DockerClient defines the interface for interacting with the container engine.
type DockerClient interface {
	CreateContainer(ctx context.Context, name, image string, ports []string, networkID string, volumeBinds []string, env []string, cmd []string) (string, error)
	StopContainer(ctx context.Context, containerID string) error
	RemoveContainer(ctx context.Context, containerID string) error
	GetLogs(ctx context.Context, containerID string) (io.ReadCloser, error)
	GetContainerStats(ctx context.Context, containerID string) (io.ReadCloser, error)
	GetContainerPort(ctx context.Context, containerID string, containerPort string) (int, error)
	CreateNetwork(ctx context.Context, name string) (string, error)
	RemoveNetwork(ctx context.Context, networkID string) error
	CreateVolume(ctx context.Context, name string) error
	DeleteVolume(ctx context.Context, name string) error
	RunTask(ctx context.Context, opts RunTaskOptions) (string, error)
	WaitContainer(ctx context.Context, containerID string) (int64, error)
	Exec(ctx context.Context, containerID string, cmd []string) (string, error)
}
