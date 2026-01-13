package docker

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type fakeDockerClient struct {
	pingErr    error
	pullErr    error
	inspectErr error

	inspect types.ContainerJSON

	removeErr error

	statsErr error
	statsRC  io.ReadCloser

	// Exec
	execCreateID   string
	execCreateErr  error
	execAttachErr  error
	execAttachRead io.Reader
	execInspect    container.ExecInspect
	execInspectErr error
}

func (f *fakeDockerClient) Ping(ctx context.Context) (types.Ping, error) {
	return types.Ping{}, f.pingErr
}

func (f *fakeDockerClient) ImagePull(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error) {
	if f.pullErr != nil {
		return nil, f.pullErr
	}
	// Return a harmless stream (adapter discards it)
	return io.NopCloser(strings.NewReader("ok")), nil
}

func (f *fakeDockerClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
	return container.CreateResponse{ID: "cid"}, nil
}

func (f *fakeDockerClient) ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error {
	return nil
}

func (f *fakeDockerClient) ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error {
	return nil
}

func (f *fakeDockerClient) ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error {
	return f.removeErr
}

func (f *fakeDockerClient) ContainerLogs(ctx context.Context, containerID string, options container.LogsOptions) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("")), nil
}

func (f *fakeDockerClient) ContainerStats(ctx context.Context, containerID string, stream bool) (container.StatsResponseReader, error) {
	if f.statsErr != nil {
		return container.StatsResponseReader{}, f.statsErr
	}
	if f.statsRC == nil {
		f.statsRC = io.NopCloser(bytes.NewReader([]byte("{}")))
	}
	return container.StatsResponseReader{Body: f.statsRC}, nil
}

func (f *fakeDockerClient) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	if f.inspectErr != nil {
		return types.ContainerJSON{}, f.inspectErr
	}
	return f.inspect, nil
}

func (f *fakeDockerClient) NetworkCreate(ctx context.Context, name string, options network.CreateOptions) (network.CreateResponse, error) {
	return network.CreateResponse{ID: "nid"}, nil
}

func (f *fakeDockerClient) NetworkRemove(ctx context.Context, networkID string) error {
	return nil
}

func (f *fakeDockerClient) VolumeCreate(ctx context.Context, options volume.CreateOptions) (volume.Volume, error) {
	return volume.Volume{Name: options.Name}, nil
}

func (f *fakeDockerClient) VolumeRemove(ctx context.Context, volumeID string, force bool) error {
	return nil
}

func (f *fakeDockerClient) ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error) {
	statusCh := make(chan container.WaitResponse, 1)
	errCh := make(chan error, 1)
	statusCh <- container.WaitResponse{StatusCode: 0}
	close(statusCh)
	close(errCh)
	return statusCh, errCh
}

func (f *fakeDockerClient) ContainerExecCreate(ctx context.Context, containerID string, config container.ExecOptions) (container.ExecCreateResponse, error) {
	if f.execCreateErr != nil {
		return container.ExecCreateResponse{}, f.execCreateErr
	}
	id := f.execCreateID
	if id == "" {
		id = "execid"
	}
	return container.ExecCreateResponse{ID: id}, nil
}

func (f *fakeDockerClient) ContainerExecStart(ctx context.Context, execID string, config container.ExecStartOptions) error {
	return nil
}

func (f *fakeDockerClient) ContainerExecAttach(ctx context.Context, execID string, config container.ExecStartOptions) (types.HijackedResponse, error) {
	if f.execAttachErr != nil {
		return types.HijackedResponse{}, f.execAttachErr
	}
	r := f.execAttachRead
	if r == nil {
		r = strings.NewReader("")
	}
	// HijackedResponse.Close() closes Conn; provide a real in-memory conn to avoid panics.
	c1, c2 := net.Pipe()
	_ = c2.Close()
	return types.HijackedResponse{Conn: c1, Reader: bufio.NewReader(r)}, nil
}

func (f *fakeDockerClient) ContainerExecInspect(ctx context.Context, execID string) (container.ExecInspect, error) {
	if f.execInspectErr != nil {
		return container.ExecInspect{}, f.execInspectErr
	}
	return f.execInspect, nil
}

var errFakeNotFound = errors.New("not found")
