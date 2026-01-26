// Package docker implements the Docker infrastructure adapters.
package docker

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"strings"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

const (
	// ImagePullTimeout is the maximum time allowed for pulling a Docker image
	ImagePullTimeout = 5 * time.Minute
	// DefaultOperationTimeout is the default timeout for Docker operations
	DefaultOperationTimeout = 30 * time.Second
	// tracerName is the name of the tracer for this package
	tracerName = "docker-adapter"
)

// DockerAdapter implements the compute backend using Docker.
type DockerAdapter struct {
	cli dockerClient
}

// dockerClient is a narrow interface over the Docker SDK client.
// It exists to make DockerAdapter unit-testable without requiring a real Docker daemon.
//
// We keep it in the docker package (not exported) to avoid leaking docker SDK types
// into other packages.
type dockerClient interface {
	Ping(ctx context.Context) (types.Ping, error)
	ImagePull(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error)
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error)
	ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error
	ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error
	ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error
	ContainerLogs(ctx context.Context, containerID string, options container.LogsOptions) (io.ReadCloser, error)
	ContainerStats(ctx context.Context, containerID string, stream bool) (container.StatsResponseReader, error)
	ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error)
	NetworkCreate(ctx context.Context, name string, options network.CreateOptions) (network.CreateResponse, error)
	NetworkRemove(ctx context.Context, networkID string) error
	VolumeCreate(ctx context.Context, options volume.CreateOptions) (volume.Volume, error)
	VolumeRemove(ctx context.Context, volumeID string, force bool) error
	ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error)
	ContainerExecCreate(ctx context.Context, container string, config container.ExecOptions) (container.ExecCreateResponse, error)
	ContainerExecStart(ctx context.Context, execID string, config container.ExecStartOptions) error
	ContainerExecAttach(ctx context.Context, execID string, config container.ExecStartOptions) (types.HijackedResponse, error)
	ContainerExecInspect(ctx context.Context, execID string) (container.ExecInspect, error)
}

// NewDockerAdapter constructs a DockerAdapter with a Docker client.
func NewDockerAdapter() (*DockerAdapter, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}
	return &DockerAdapter{cli: cli}, nil
}

// Ping checks if Docker daemon is reachable
func (a *DockerAdapter) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := a.cli.Ping(ctx)
	return err
}

func (a *DockerAdapter) Type() string {
	return "docker"
}

func (a *DockerAdapter) CreateInstance(ctx context.Context, opts ports.CreateInstanceOptions) (string, error) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "CreateInstance")
	defer span.End()

	span.SetAttributes(
		attribute.String("docker.image", opts.ImageName),
		attribute.String("docker.name", opts.Name),
		attribute.Bool("docker.network_disabled", opts.NetworkID == ""),
	)

	// 1. Ensure image exists (pull if not) - with timeout
	pullCtx, pullCancel := context.WithTimeout(ctx, ImagePullTimeout)
	defer pullCancel()

	// Create a sub-span for pulling
	pullCtx, pullSpan := otel.Tracer(tracerName).Start(pullCtx, "ImagePull")
	reader, err := a.cli.ImagePull(pullCtx, opts.ImageName, image.PullOptions{})
	pullSpan.End()

	if err != nil {
		return "", fmt.Errorf("failed to pull image: %w", err)
	}
	defer func() { _ = reader.Close() }()
	_, _ = io.Copy(io.Discard, reader)

	// 2. Configure container
	config := &container.Config{
		Image:        opts.ImageName,
		Env:          opts.Env,
		Cmd:          opts.Cmd,
		ExposedPorts: make(nat.PortSet),
	}

	// If no command is specified, use a long-running command to keep container alive
	// This is essential for K8s nodes which need to stay running while we bootstrap them.
	// HACK: For kindest/node, we MUST NOT override the entrypoint/command as it starts systemd.
	isKIND := strings.Contains(opts.ImageName, "kindest/node")
	if len(config.Cmd) == 0 && !isKIND {
		config.Cmd = []string{"/bin/bash", "-c", "tail -f /dev/null"}
	}

	hostConfig := &container.HostConfig{
		PortBindings: make(nat.PortMap),
		Binds:        opts.VolumeBinds,
	}

	// For KIND images, we need privileged mode to support systemd and cgroups.
	// We also need to mount /lib/modules and use anonymous volumes for containerd/kubelet
	// to avoid nested overlayfs issues on Mac.
	if isKIND {
		hostConfig.Privileged = true
		if config.Volumes == nil {
			config.Volumes = make(map[string]struct{})
		}
		config.Volumes["/var/lib/containerd"] = struct{}{}
		config.Volumes["/var/lib/kubelet"] = struct{}{}

		hostConfig.Binds = append(hostConfig.Binds,
			"/lib/modules:/lib/modules:ro",
			"/sys/fs/cgroup:/sys/fs/cgroup:ro",
		)
	}
	networkingConfig := &network.NetworkingConfig{}

	if opts.NetworkID != "" {
		networkingConfig.EndpointsConfig = map[string]*network.EndpointSettings{
			opts.NetworkID: {},
		}
	}

	for _, p := range opts.Ports {
		parts := strings.Split(p, ":")
		if len(parts) == 2 {
			hostPort := parts[0]
			containerPort := parts[1]

			// We assume TCP for now as per plan
			cPort := nat.Port(containerPort + "/tcp")
			config.ExposedPorts[cPort] = struct{}{}
			hostConfig.PortBindings[cPort] = []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: hostPort,
				},
			}
		}
	}

	// 3. Create container
	_, createSpan := otel.Tracer(tracerName).Start(ctx, "ContainerCreate")
	resp, err := a.cli.ContainerCreate(ctx, config, hostConfig, networkingConfig, nil, opts.Name)
	createSpan.End()
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	// 4. Start container
	_, startSpan := otel.Tracer(tracerName).Start(ctx, "ContainerStart")
	if err := a.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		startSpan.End()
		return "", fmt.Errorf("failed to start container: %w", err)
	}
	startSpan.End()

	return resp.ID, nil
}

func (a *DockerAdapter) StopInstance(ctx context.Context, name string) error {
	err := a.cli.ContainerStop(ctx, name, container.StopOptions{})
	if err != nil {
		return fmt.Errorf("failed to stop container %s: %w", name, err)
	}
	return nil
}

func (a *DockerAdapter) DeleteInstance(ctx context.Context, containerID string) error {
	err := a.cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})
	if err != nil {
		if errdefs.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to remove container %s: %w", containerID, err)
	}
	return nil
}

func (a *DockerAdapter) GetInstanceLogs(ctx context.Context, containerID string) (io.ReadCloser, error) {
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       "2000",
	}

	src, err := a.cli.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get container logs: %w", err)
	}

	// Use a pipe to clean the stream asynchronously
	r, w := io.Pipe()
	go func() {
		defer func() { _ = w.Close() }()
		defer func() { _ = src.Close() }()
		// stdcopy demultiplexes docker stream into plain text
		_, _ = stdcopy.StdCopy(w, w, src)
	}()

	return r, nil
}

func (a *DockerAdapter) GetInstanceStats(ctx context.Context, containerID string) (io.ReadCloser, error) {
	// Stream: false = get one snapshot
	stats, err := a.cli.ContainerStats(ctx, containerID, false)
	if err != nil {
		return nil, err
	}
	return stats.Body, nil
}

func (a *DockerAdapter) GetInstancePort(ctx context.Context, containerID string, containerPort string) (int, error) {
	inspect, err := a.cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return 0, fmt.Errorf("failed to inspect container: %w", err)
	}

	cPort := nat.Port(containerPort + "/tcp")
	bindings, ok := inspect.NetworkSettings.Ports[cPort]
	if !ok || len(bindings) == 0 {
		return 0, fmt.Errorf("no port binding found for %s", containerPort)
	}

	var hostPort int
	_, err = fmt.Sscanf(bindings[0].HostPort, "%d", &hostPort)
	if err != nil {
		return 0, fmt.Errorf("failed to parse host port: %w", err)
	}

	return hostPort, nil
}

func (a *DockerAdapter) CreateNetwork(ctx context.Context, name string) (string, error) {
	resp, err := a.cli.NetworkCreate(ctx, name, network.CreateOptions{
		Driver: "bridge",
	})
	if err != nil {
		return "", fmt.Errorf("failed to create network %s: %w", name, err)
	}
	return resp.ID, nil
}

func (a *DockerAdapter) DeleteNetwork(ctx context.Context, networkID string) error {
	err := a.cli.NetworkRemove(ctx, networkID)
	if err != nil {
		return fmt.Errorf("failed to remove network %s: %w", networkID, err)
	}
	return nil
}

func (a *DockerAdapter) CreateVolume(ctx context.Context, name string) error {
	_, err := a.cli.VolumeCreate(ctx, volume.CreateOptions{
		Name: name,
	})
	if err != nil {
		return fmt.Errorf("failed to create volume %s: %w", name, err)
	}
	return nil
}

func (a *DockerAdapter) DeleteVolume(ctx context.Context, name string) error {
	if err := a.cli.VolumeRemove(ctx, name, true); err != nil {
		return fmt.Errorf("failed to delete volume %s: %w", name, err)
	}
	return nil
}

func (a *DockerAdapter) CreateVolumeSnapshot(ctx context.Context, volumeID string, destinationPath string) error {
	// volumeID is the docker volume name
	// destinationPath is on the host

	// Check for path traversal in destinationPath
	if strings.Contains(destinationPath, "..") {
		return fmt.Errorf("invalid destination path: traversal detected")
	}

	// Ensure parent dir of destinationPath exists
	// We assume destinationPath is accessible to the docker daemon (bind mount)

	// Create a temp container to do the work
	imageName := "alpine"
	// Ensure image exists
	if _, err := a.cli.ImagePull(ctx, imageName, image.PullOptions{}); err != nil {
		return fmt.Errorf("failed to pull alpine: %w", err)
	}

	cmd := []string{"tar", "czf", "/snapshot/snapshot.tar.gz", "-C", "/data", "."}

	config := &container.Config{
		Image: imageName,
		Cmd:   cmd,
	}

	hostConfig := &container.HostConfig{
		Binds: []string{
			volumeID + ":/data:ro",
			filepath.Dir(destinationPath) + ":/snapshot",
		},
	}

	// We need to name the file correctly inside.
	// We bind the PARENT directory of destinationPath to /snapshot
	// And write to /snapshot/<filename>
	filename := filepath.Base(destinationPath)
	config.Cmd = []string{"tar", "czf", "/snapshot/" + filename, "-C", "/data", "."}

	resp, err := a.cli.ContainerCreate(ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		return fmt.Errorf("failed to create snapshot task: %w", err)
	}
	defer func() { _ = a.cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true}) }()

	if err := a.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start snapshot task: %w", err)
	}

	statusCh, errCh := a.cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("error waiting for snapshot task: %w", err)
		}
	case status := <-statusCh:
		if status.StatusCode != 0 {
			return fmt.Errorf("snapshot task failed with exit code %d", status.StatusCode)
		}
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func (a *DockerAdapter) RestoreVolumeSnapshot(ctx context.Context, volumeID string, sourcePath string) error {
	// volumeID is the docker volume name
	// sourcePath is the .tar.gz on host

	// Check for path traversal in sourcePath
	if strings.Contains(sourcePath, "..") {
		return fmt.Errorf("invalid source path: traversal detected")
	}

	imageName := "alpine"
	if _, err := a.cli.ImagePull(ctx, imageName, image.PullOptions{}); err != nil {
		return fmt.Errorf("failed to pull alpine: %w", err)
	}

	filename := filepath.Base(sourcePath)
	config := &container.Config{
		Image: imageName,
		Cmd:   []string{"tar", "xzf", "/snapshot/" + filename, "-C", "/data"},
	}

	hostConfig := &container.HostConfig{
		Binds: []string{
			volumeID + ":/data",
			filepath.Dir(sourcePath) + ":/snapshot:ro",
		},
	}

	resp, err := a.cli.ContainerCreate(ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		return fmt.Errorf("failed to create restore task: %w", err)
	}
	defer func() { _ = a.cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true}) }()

	if err := a.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start restore task: %w", err)
	}

	statusCh, errCh := a.cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("error waiting for restore task: %w", err)
		}
	case status := <-statusCh:
		if status.StatusCode != 0 {
			return fmt.Errorf("restore task failed with exit code %d", status.StatusCode)
		}
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func (a *DockerAdapter) AttachVolume(ctx context.Context, id string, volumePath string) error {
	return errors.New(errors.NotImplemented, "attaching volumes to running containers is not supported in docker adapter")
}

func (a *DockerAdapter) DetachVolume(ctx context.Context, id string, volumePath string) error {
	return errors.New(errors.NotImplemented, "detaching volumes from running containers is not supported in docker adapter")
}

func (a *DockerAdapter) GetConsoleURL(ctx context.Context, id string) (string, error) {
	return "", fmt.Errorf("console not supported for docker instances")
}

func (a *DockerAdapter) GetInstanceIP(ctx context.Context, id string) (string, error) {
	// Inspect container
	json, err := a.cli.ContainerInspect(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to inspect container: %w", err)
	}

	// Try to get IP from first network
	for _, net := range json.NetworkSettings.Networks {
		if net.IPAddress != "" {
			return net.IPAddress, nil
		}
	}
	return "", fmt.Errorf("no IP address found for container %s", id)
}

func (a *DockerAdapter) RunTask(ctx context.Context, opts ports.RunTaskOptions) (string, error) {
	// 1. Ensure image exists
	pullCtx, pullCancel := context.WithTimeout(ctx, ImagePullTimeout)
	defer pullCancel()

	reader, err := a.cli.ImagePull(pullCtx, opts.Image, image.PullOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to pull image: %w", err)
	}
	defer func() { _ = reader.Close() }()
	_, _ = io.Copy(io.Discard, reader)

	// 2. Configure container with security defaults
	config := &container.Config{
		Image:           opts.Image,
		Cmd:             opts.Command,
		Env:             opts.Env,
		WorkingDir:      opts.WorkingDir,
		NetworkDisabled: opts.NetworkDisabled,
	}

	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			Memory:   opts.MemoryMB * 1024 * 1024,
			NanoCPUs: int64(opts.CPUs * 1e9),
		},
		Binds:          opts.Binds,
		ReadonlyRootfs: opts.ReadOnlyRootfs,
		// SecurityOpt:    []string{"no-new-privileges:true"}, // Removed to allow privileged operations (e.g. kube-proxy)
	}

	if opts.PidsLimit != nil {
		hostConfig.Resources.PidsLimit = opts.PidsLimit
	}

	// 3. Create container
	resp, err := a.cli.ContainerCreate(ctx, config, hostConfig, nil, nil, opts.Name)
	if err != nil {
		return "", fmt.Errorf("failed to create task container: %w", err)
	}

	// 4. Start container
	if err := a.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start task container: %w", err)
	}

	return resp.ID, nil
}

func (a *DockerAdapter) WaitTask(ctx context.Context, containerID string) (int64, error) {
	statusCh, errCh := a.cli.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return 0, fmt.Errorf("error waiting for container: %w", err)
		}
	case status := <-statusCh:
		return status.StatusCode, nil
	case <-ctx.Done():
		return 0, ctx.Err()
	}
	return 0, nil
}

func (a *DockerAdapter) Exec(ctx context.Context, containerID string, cmd []string) (string, error) {
	config := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	}

	execResp, err := a.cli.ContainerExecCreate(ctx, containerID, config)
	if err != nil {
		return "", fmt.Errorf("failed to create exec: %w", err)
	}

	resp, err := a.cli.ContainerExecAttach(ctx, execResp.ID, container.ExecStartOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to attach exec: %w", err)
	}
	defer resp.Close()

	// Capture output
	var outBuf strings.Builder
	if _, err := stdcopy.StdCopy(&outBuf, &outBuf, resp.Reader); err != nil {
		return "", fmt.Errorf("failed to read exec output: %w", err)
	}

	// Wait for completion to get exit code?
	// The attach waits for the stream to close, which happens when the process exits.
	// But to be sure, we can inspect.
	execInspect, err := a.cli.ContainerExecInspect(ctx, execResp.ID)
	if err != nil {
		return outBuf.String(), fmt.Errorf("failed to inspect exec result: %w", err)
	}

	if execInspect.ExitCode != 0 {
		return outBuf.String(), fmt.Errorf("command execution failed with exit code %d: %s", execInspect.ExitCode, outBuf.String())
	}

	return outBuf.String(), nil
}
