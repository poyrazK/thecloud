package k8s

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testMasterIP = "10.0.0.10"
	binSh        = "/bin/sh"
)

type mockInstanceService struct{ mock.Mock }

type mockClusterRepo struct{ mock.Mock }

func (m *mockClusterRepo) Create(ctx context.Context, c *domain.Cluster) error { return nil }
func (m *mockClusterRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Cluster, error) {
	return nil, nil
}
func (m *mockClusterRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Cluster, error) {
	return nil, nil
}
func (m *mockClusterRepo) Update(ctx context.Context, c *domain.Cluster) error      { return nil }
func (m *mockClusterRepo) Delete(ctx context.Context, id uuid.UUID) error           { return nil }
func (m *mockClusterRepo) AddNode(ctx context.Context, n *domain.ClusterNode) error { return nil }
func (m *mockClusterRepo) GetNodes(ctx context.Context, clusterID uuid.UUID) ([]*domain.ClusterNode, error) {
	args := m.Called(ctx, clusterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ClusterNode), args.Error(1)
}
func (m *mockClusterRepo) DeleteNode(ctx context.Context, nodeID uuid.UUID) error      { return nil }
func (m *mockClusterRepo) UpdateNode(ctx context.Context, n *domain.ClusterNode) error { return nil }

func (m *mockInstanceService) LaunchInstance(ctx context.Context, name, image, ports string, vpcID, subnetID *uuid.UUID, volumes []domain.VolumeAttachment) (*domain.Instance, error) {
	args := m.Called(ctx, name, image, ports, vpcID, subnetID, volumes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Instance), args.Error(1)
}
func (m *mockInstanceService) StopInstance(ctx context.Context, idOrName string) error {
	return nil
}
func (m *mockInstanceService) ListInstances(ctx context.Context) ([]*domain.Instance, error) {
	return nil, nil
}
func (m *mockInstanceService) GetInstance(ctx context.Context, idOrName string) (*domain.Instance, error) {
	args := m.Called(ctx, idOrName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Instance), args.Error(1)
}
func (m *mockInstanceService) GetInstanceLogs(ctx context.Context, idOrName string) (string, error) {
	return "", nil
}
func (m *mockInstanceService) GetInstanceStats(ctx context.Context, idOrName string) (*domain.InstanceStats, error) {
	return nil, nil
}
func (m *mockInstanceService) GetConsoleURL(ctx context.Context, idOrName string) (string, error) {
	return "", nil
}
func (m *mockInstanceService) TerminateInstance(ctx context.Context, idOrName string) error {
	return nil
}
func (m *mockInstanceService) Exec(ctx context.Context, idOrName string, cmd []string) (string, error) {
	args := m.Called(ctx, idOrName, cmd)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.String(0), args.Error(1)
}

func newProvisioner(instSvc *mockInstanceService, repo *mockClusterRepo) *KubeadmProvisioner {
	return &KubeadmProvisioner{
		instSvc: instSvc,
		repo:    repo,
		logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

func TestKubeadmProvisionerGetHealthNoControlPlaneIPs(t *testing.T) {
	instSvc := new(mockInstanceService)
	repo := new(mockClusterRepo)
	p := newProvisioner(instSvc, repo)

	cluster := &domain.Cluster{ID: uuid.New()}

	health, err := p.GetHealth(context.Background(), cluster)

	assert.Error(t, err)
	assert.Nil(t, health)
}

func TestKubeadmProvisionerGetHealthServiceExecutor(t *testing.T) {
	instSvc := new(mockInstanceService)
	repo := new(mockClusterRepo)
	p := newProvisioner(instSvc, repo)

	clusterID := uuid.New()
	instanceID := uuid.New()
	masterIP := testMasterIP
	cluster := &domain.Cluster{ID: clusterID, ControlPlaneIPs: []string{masterIP}, Status: domain.ClusterStatusRunning}

	repo.On("GetNodes", mock.Anything, clusterID).Return([]*domain.ClusterNode{{ID: uuid.New(), InstanceID: instanceID}}, nil)
	instSvc.On("GetInstance", mock.Anything, instanceID.String()).Return(&domain.Instance{ID: instanceID, PrivateIP: masterIP}, nil)
	instSvc.On("Exec", mock.Anything, instanceID.String(), []string{binSh, "-c", kubectlBase + " get nodes"}).Return("", nil)
	instSvc.On("Exec", mock.Anything, instanceID.String(), []string{binSh, "-c", kubectlBase + " get nodes --no-headers"}).Return("node-a Ready \nnode-b NotReady", nil)

	health, err := p.GetHealth(context.Background(), cluster)

	assert.NoError(t, err)
	assert.True(t, health.APIServer)
	assert.Equal(t, 2, health.NodesTotal)
	assert.Equal(t, 1, health.NodesReady)
	assert.Equal(t, "1/2 nodes are ready", health.Message)
}

func TestKubeadmProvisionerGetHealthAPIServerDown(t *testing.T) {
	instSvc := new(mockInstanceService)
	repo := new(mockClusterRepo)
	p := newProvisioner(instSvc, repo)

	clusterID := uuid.New()
	instanceID := uuid.New()
	masterIP := testMasterIP
	cluster := &domain.Cluster{ID: clusterID, ControlPlaneIPs: []string{masterIP}, Status: domain.ClusterStatusRunning}

	repo.On("GetNodes", mock.Anything, clusterID).Return([]*domain.ClusterNode{{ID: uuid.New(), InstanceID: instanceID}}, nil)
	instSvc.On("GetInstance", mock.Anything, instanceID.String()).Return(&domain.Instance{ID: instanceID, PrivateIP: masterIP}, nil)
	instSvc.On("Exec", mock.Anything, instanceID.String(), []string{binSh, "-c", kubectlBase + " get nodes"}).Return("", errors.New("down"))
	instSvc.On("Exec", mock.Anything, instanceID.String(), []string{binSh, "-c", kubectlBase + " get nodes --no-headers"}).Return("", io.EOF)

	health, err := p.GetHealth(context.Background(), cluster)

	assert.NoError(t, err)
	assert.False(t, health.APIServer)
	assert.Equal(t, "API server is unreachable", health.Message)
}

func TestKubeadmProvisionerGetKubeconfigNoControlPlaneIPs(t *testing.T) {
	instSvc := new(mockInstanceService)
	repo := new(mockClusterRepo)
	p := newProvisioner(instSvc, repo)

	cluster := &domain.Cluster{ID: uuid.New()}

	kubeconfig, err := p.GetKubeconfig(context.Background(), cluster, "admin")

	assert.Error(t, err)
	assert.Empty(t, kubeconfig)
}

func TestKubeadmProvisionerGetKubeconfigViewerNotImplemented(t *testing.T) {
	instSvc := new(mockInstanceService)
	repo := new(mockClusterRepo)
	p := newProvisioner(instSvc, repo)

	clusterID := uuid.New()
	cluster := &domain.Cluster{ID: clusterID, ControlPlaneIPs: []string{testMasterIP}}

	repo.On("GetNodes", mock.Anything, clusterID).Return(nil, errors.New("no nodes"))

	kubeconfig, err := p.GetKubeconfig(context.Background(), cluster, "viewer")

	assert.Error(t, err)
	assert.Empty(t, kubeconfig)
}

func TestKubeadmProvisionerGetKubeconfigAdmin(t *testing.T) {
	instSvc := new(mockInstanceService)
	repo := new(mockClusterRepo)
	p := newProvisioner(instSvc, repo)

	clusterID := uuid.New()
	instanceID := uuid.New()
	masterIP := testMasterIP
	cluster := &domain.Cluster{ID: clusterID, ControlPlaneIPs: []string{masterIP}}

	repo.On("GetNodes", mock.Anything, clusterID).Return([]*domain.ClusterNode{{ID: uuid.New(), InstanceID: instanceID}}, nil)
	instSvc.On("GetInstance", mock.Anything, instanceID.String()).Return(&domain.Instance{ID: instanceID, PrivateIP: masterIP}, nil)
	instSvc.On("Exec", mock.Anything, instanceID.String(), []string{binSh, "-c", "cat " + adminKubeconfig}).Return("kubeconfig", nil)

	kubeconfig, err := p.GetKubeconfig(context.Background(), cluster, "admin")

	assert.NoError(t, err)
	assert.Equal(t, "kubeconfig", kubeconfig)
}
