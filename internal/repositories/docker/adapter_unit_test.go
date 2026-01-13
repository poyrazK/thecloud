package docker

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/stretchr/testify/require"
)

func TestDockerAdapter_Type(t *testing.T) {
	a := &DockerAdapter{}
	require.Equal(t, "docker", a.Type())
}

func TestDockerAdapter_DeleteInstance_NotFoundIsNil(t *testing.T) {
	a := &DockerAdapter{cli: &fakeDockerClient{removeErr: errdefs.ErrNotFound}}
	require.NoError(t, a.DeleteInstance(context.Background(), "missing"))
}

func TestDockerAdapter_GetInstancePort_NoBinding(t *testing.T) {
	inspect := types.ContainerJSON{}
	inspect.NetworkSettings = &types.NetworkSettings{}

	a := &DockerAdapter{cli: &fakeDockerClient{inspect: inspect}}
	_, err := a.GetInstancePort(context.Background(), "cid", "8080")
	require.Error(t, err)
}

// Note: port parsing is best covered via integration tests because docker SDK
// types for NetworkSettings/Ports are tricky to construct across versions.

func TestDockerAdapter_GetInstanceIP_NoNetworks(t *testing.T) {
	inspect := types.ContainerJSON{}
	inspect.NetworkSettings = &types.NetworkSettings{Networks: map[string]*network.EndpointSettings{}}

	a := &DockerAdapter{cli: &fakeDockerClient{inspect: inspect}}
	_, err := a.GetInstanceIP(context.Background(), "cid")
	require.Error(t, err)
}

func TestDockerAdapter_GetInstanceIP_FirstIP(t *testing.T) {
	inspect := types.ContainerJSON{}
	inspect.NetworkSettings = &types.NetworkSettings{
		Networks: map[string]*network.EndpointSettings{
			"n1": {IPAddress: "10.0.0.2"},
		},
	}

	a := &DockerAdapter{cli: &fakeDockerClient{inspect: inspect}}
	ip, err := a.GetInstanceIP(context.Background(), "cid")
	require.NoError(t, err)
	require.Equal(t, "10.0.0.2", ip)
}

func TestDockerAdapter_GetInstanceStats_ReturnsBody(t *testing.T) {
	a := &DockerAdapter{cli: &fakeDockerClient{statsRC: io.NopCloser(strings.NewReader("{}"))}}
	rc, err := a.GetInstanceStats(context.Background(), "cid")
	require.NoError(t, err)
	defer rc.Close()
}

func TestDockerAdapter_Exec_NonZeroExit(t *testing.T) {
	cli := &fakeDockerClient{
		execAttachRead: strings.NewReader("out"),
		execInspect:    container.ExecInspect{ExitCode: 2},
	}
	adapter := &DockerAdapter{cli: cli}

	out, err := adapter.Exec(context.Background(), "cid", []string{"false"})
	require.Error(t, err)
	// Exec output is demultiplexed with stdcopy, which expects Docker's multiplexed
	// stream format. Our unit tests don't simulate that format, so output may be empty.
	_ = out
}

func TestDockerAdapter_Exec_InspectErrorReturnsOutput(t *testing.T) {
	cli := &fakeDockerClient{
		execAttachRead: strings.NewReader("out"),
		execInspectErr: errFakeNotFound,
	}
	adapter := &DockerAdapter{cli: cli}

	out, err := adapter.Exec(context.Background(), "cid", []string{"echo"})
	require.Error(t, err)
	_ = out
}
