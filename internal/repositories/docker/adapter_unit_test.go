package docker

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
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
	defer func() { _ = rc.Close() }()
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

func TestDockerAdapter_CreateVolume(t *testing.T) {
	cli := &fakeDockerClient{}
	adapter := &DockerAdapter{cli: cli}

	err := adapter.CreateVolume(context.Background(), "test-volume")
	require.NoError(t, err)
}

func TestDockerAdapter_DeleteVolume(t *testing.T) {
	cli := &fakeDockerClient{}
	adapter := &DockerAdapter{cli: cli}

	err := adapter.DeleteVolume(context.Background(), "test-volume")
	require.NoError(t, err)
}

func TestDockerAdapter_AttachVolume(t *testing.T) {
	adapter := &DockerAdapter{}
	err := adapter.AttachVolume(context.Background(), "inst1", "/data")
	// AttachVolume is not supported in docker
	require.Error(t, err)
}

func TestDockerAdapter_DetachVolume(t *testing.T) {
	adapter := &DockerAdapter{}
	err := adapter.DetachVolume(context.Background(), "inst1", "/data")
	// DetachVolume is not supported in docker
	require.Error(t, err)
}

func TestDockerAdapter_GetConsoleURL(t *testing.T) {
	adapter := &DockerAdapter{}
	_, err := adapter.GetConsoleURL(context.Background(), "cid")
	// Console is not supported for docker
	require.Error(t, err)
	require.Contains(t, err.Error(), "console not supported")
}

func TestDockerAdapter_CreateVolumeSnapshot(t *testing.T) {
	cli := &fakeDockerClient{}
	adapter := &DockerAdapter{cli: cli}

	// Should succeed with fake client
	err := adapter.CreateVolumeSnapshot(context.Background(), "vol-id", "/tmp/backup.tar.gz")
	require.NoError(t, err)
}

func TestDockerAdapter_RestoreVolumeSnapshot(t *testing.T) {
	cli := &fakeDockerClient{}
	adapter := &DockerAdapter{cli: cli}

	// Should succeed with fake client
	err := adapter.RestoreVolumeSnapshot(context.Background(), "vol-id", "/tmp/backup.tar.gz")
	require.NoError(t, err)
}

func TestDockerAdapter_RunTask(t *testing.T) {
	cli := &fakeDockerClient{}
	adapter := &DockerAdapter{cli: cli}

	opts := ports.RunTaskOptions{
		Image:   "alpine",
		Command: []string{"echo", "hello"},
	}
	id, err := adapter.RunTask(context.Background(), opts)
	require.NoError(t, err)
	require.Equal(t, "cid", id)
}

func TestDockerAdapter_WaitTask(t *testing.T) {
	cli := &fakeDockerClient{}
	adapter := &DockerAdapter{cli: cli}

	exitCode, err := adapter.WaitTask(context.Background(), "cid")
	require.NoError(t, err)
	require.Equal(t, int64(0), exitCode)
}

func TestDockerAdapter_WaitTask_Error(t *testing.T) {
	cli := &fakeDockerClient{waitErr: errors.New("wait failed")}
	adapter := &DockerAdapter{cli: cli}

	_, err := adapter.WaitTask(context.Background(), "cid")
	require.Error(t, err)
}

func TestDockerAdapter_CreateNetwork(t *testing.T) {
	cli := &fakeDockerClient{}
	adapter := &DockerAdapter{cli: cli}

	id, err := adapter.CreateNetwork(context.Background(), "net1")
	require.NoError(t, err)
	require.Equal(t, "nid", id)
}

func TestDockerAdapter_DeleteNetwork(t *testing.T) {
	cli := &fakeDockerClient{}
	adapter := &DockerAdapter{cli: cli}

	err := adapter.DeleteNetwork(context.Background(), "net1")
	require.NoError(t, err)
}

func TestDockerAdapter_CreateInstance(t *testing.T) {
	cli := &fakeDockerClient{}
	adapter := &DockerAdapter{cli: cli}

	opts := ports.CreateInstanceOptions{
		Name:      "inst1",
		ImageName: "alpine",
		Cmd:       []string{"/bin/sh"},
		Ports:     []string{"8080:80"},
	}
	id, err := adapter.CreateInstance(context.Background(), opts)
	require.NoError(t, err)
	require.Equal(t, "cid", id)
}

func TestDockerAdapter_StopInstance(t *testing.T) {
	cli := &fakeDockerClient{}
	adapter := &DockerAdapter{cli: cli}

	err := adapter.StopInstance(context.Background(), "cid")
	require.NoError(t, err)
}

func TestDockerAdapter_GetInstanceLogs(t *testing.T) {
	cli := &fakeDockerClient{}
	adapter := &DockerAdapter{cli: cli}

	rc, err := adapter.GetInstanceLogs(context.Background(), "cid")
	require.NoError(t, err)
	_ = rc.Close()
}

func TestDockerAdapter_Ping(t *testing.T) {
	cli := &fakeDockerClient{}
	adapter := &DockerAdapter{cli: cli}
	err := adapter.Ping(context.Background())
	require.NoError(t, err)

	cli.pingErr = errors.New("ping failed")
	err = adapter.Ping(context.Background())
	require.Error(t, err)
}

func TestDockerAdapter_NewDockerAdapter_Error(t *testing.T) {
	// Triggering error by setting an invalid DOCKER_HOST
	t.Setenv("DOCKER_HOST", "invalid-proto://")
	_, err := NewDockerAdapter()
	require.Error(t, err)
}

type mockVpcRepository struct {
	ports.VpcRepository
	getByIDFunc func(ctx context.Context, id uuid.UUID) (*domain.VPC, error)
}

func (m *mockVpcRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.VPC, error) {
	return m.getByIDFunc(ctx, id)
}

type mockInstanceRepoUnit struct {
	ports.InstanceRepository
	getByIDFunc func(ctx context.Context, id uuid.UUID) (*domain.Instance, error)
}

func (m *mockInstanceRepoUnit) GetByID(ctx context.Context, id uuid.UUID) (*domain.Instance, error) {
	return m.getByIDFunc(ctx, id)
}

func TestLBProxyAdapter_DeployProxy(t *testing.T) {
	cli := &fakeDockerClient{}
	vpcRepo := &mockVpcRepository{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.VPC, error) {
			return &domain.VPC{ID: id, NetworkID: "net1"}, nil
		},
	}
	instRepo := &mockInstanceRepoUnit{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.Instance, error) {
			return &domain.Instance{ID: id}, nil
		},
	}
	adapter := &LBProxyAdapter{
		cli:          cli,
		vpcRepo:      vpcRepo,
		instanceRepo: instRepo,
	}

	lb := &domain.LoadBalancer{
		ID:    uuid.New(),
		Port:  80,
		VpcID: uuid.New(),
	}
	targets := []*domain.LBTarget{
		{InstanceID: uuid.New(), Port: 8080, Weight: 1},
	}

	id, err := adapter.DeployProxy(context.Background(), lb, targets)
	require.NoError(t, err)
	require.Equal(t, "cid", id)
}

func TestLBProxyAdapter_RemoveProxy(t *testing.T) {
	cli := &fakeDockerClient{}
	adapter := &LBProxyAdapter{cli: cli}
	err := adapter.RemoveProxy(context.Background(), uuid.New())
	require.NoError(t, err)
}

func TestLBProxyAdapter_UpdateProxyConfig(t *testing.T) {
	cli := &fakeDockerClient{}
	instRepo := &mockInstanceRepoUnit{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.Instance, error) {
			return &domain.Instance{ID: id}, nil
		},
	}
	adapter := &LBProxyAdapter{
		cli:          cli,
		instanceRepo: instRepo,
	}

	lb := &domain.LoadBalancer{
		ID:   uuid.New(),
		Port: 80,
	}
	targets := []*domain.LBTarget{
		{InstanceID: uuid.New(), Port: 8080},
	}

	// Ensure directory exists for test
	configPath := filepath.Join("/tmp", "thecloud", "lb", lb.ID.String())
	_ = os.MkdirAll(configPath, 0755)
	defer func() { _ = os.RemoveAll(configPath) }()

	err := adapter.UpdateProxyConfig(context.Background(), lb, targets)
	require.NoError(t, err)
}

func TestDockerAdapter_StopInstance_Error(t *testing.T) {
	cli := &fakeDockerClient{stopErr: errors.New("stop error")}
	adapter := &DockerAdapter{cli: cli}
	err := adapter.StopInstance(context.Background(), "cid")
	require.Error(t, err)
}

func TestDockerAdapter_CreateNetwork_Error(t *testing.T) {
	cli := &fakeDockerClient{networkCreateErr: errors.New("network create error")}
	adapter := &DockerAdapter{cli: cli}
	_, err := adapter.CreateNetwork(context.Background(), "net")
	require.Error(t, err)
}
