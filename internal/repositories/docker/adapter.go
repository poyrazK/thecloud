package docker

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

type DockerAdapter struct {
	cli *client.Client
}

func NewDockerAdapter() (*DockerAdapter, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}
	return &DockerAdapter{cli: cli}, nil
}

func (a *DockerAdapter) CreateContainer(ctx context.Context, name, imageName string) (string, error) {
	// 1. Ensure image exists (pull if not)
	reader, err := a.cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to pull image: %w", err)
	}
	defer reader.Close()
	io.Copy(io.Discard, reader)

	// 2. Configure container
	config := &container.Config{
		Image: imageName,
	}
	hostConfig := &container.HostConfig{}

	// 3. Create container
	resp, err := a.cli.ContainerCreate(ctx, config, hostConfig, nil, nil, name)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	// 4. Start container
	if err := a.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	return resp.ID, nil
}

func (a *DockerAdapter) StopContainer(ctx context.Context, name string) error {
	err := a.cli.ContainerStop(ctx, name, container.StopOptions{})
	if err != nil {
		return fmt.Errorf("failed to stop container %s: %w", name, err)
	}
	return nil
}

func (a *DockerAdapter) RemoveContainer(ctx context.Context, containerID string) error {
	err := a.cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})
	if err != nil {
		return fmt.Errorf("failed to remove container %s: %w", containerID, err)
	}
	return nil
}
