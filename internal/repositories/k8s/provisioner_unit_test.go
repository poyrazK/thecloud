package k8s

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupProvisionerTest() (*KubeadmProvisioner, *mockInstanceService, *mockClusterRepo, *MockSecretService, *MockSecurityGroupService, *MockLBService) {
	instSvc := new(mockInstanceService)
	repo := new(mockClusterRepo)
	secretSvc := new(MockSecretService)
	sgSvc := new(MockSecurityGroupService)
	lbSvc := new(MockLBService)
	storageSvc := new(MockStorageService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	p := NewKubeadmProvisioner(
		instSvc,
		repo,
		secretSvc,
		sgSvc,
		storageSvc,
		lbSvc,
		logger,
	)

	return p, instSvc, repo, secretSvc, sgSvc, lbSvc
}

func TestProvision_SecurityGroupFailure(t *testing.T) {
	p, _, repo, _, sgSvc, _ := setupProvisionerTest()
	ctx := context.Background()
	cluster := &domain.Cluster{ID: uuid.New(), Name: "test-cluster", VpcID: uuid.New()}

	repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	sgSvc.On("GetGroup", mock.Anything, mock.Anything, cluster.VpcID).Return(nil, fmt.Errorf("not found"))
	sgSvc.On("CreateGroup", mock.Anything, cluster.VpcID, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("sg failure"))

	err := p.Provision(ctx, cluster)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sg failure")
}

func TestProvision_LBFailure(t *testing.T) {
	p, _, repo, _, sgSvc, lbSvc := setupProvisionerTest()
	ctx := context.Background()
	cluster := &domain.Cluster{ID: uuid.New(), Name: "test-cluster", VpcID: uuid.New(), HAEnabled: true}

	repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	sgSvc.On("GetGroup", mock.Anything, mock.Anything, cluster.VpcID).Return(nil, fmt.Errorf("not found"))
	sgSvc.On("CreateGroup", mock.Anything, cluster.VpcID, mock.Anything, mock.Anything).Return(&domain.SecurityGroup{ID: uuid.New()}, nil)
	sgSvc.On("AddRule", mock.Anything, mock.Anything, mock.Anything).Return(&domain.SecurityRule{}, nil)
	lbSvc.On("Create", mock.Anything, mock.Anything, cluster.VpcID, 6443, "round-robin", mock.Anything).Return(nil, fmt.Errorf("lb failure"))

	err := p.Provision(ctx, cluster)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "lb failure")
}

func TestProvision_ControlPlaneFailure(t *testing.T) {
	p, instSvc, repo, _, sgSvc, _ := setupProvisionerTest()
	ctx := context.Background()
	cluster := &domain.Cluster{ID: uuid.New(), Name: "test-cluster", VpcID: uuid.New()}

	repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	sgSvc.On("GetGroup", mock.Anything, mock.Anything, cluster.VpcID).Return(nil, fmt.Errorf("not found"))
	sgSvc.On("CreateGroup", mock.Anything, cluster.VpcID, mock.Anything, mock.Anything).Return(&domain.SecurityGroup{ID: uuid.New()}, nil)
	sgSvc.On("AddRule", mock.Anything, mock.Anything, mock.Anything).Return(&domain.SecurityRule{}, nil)
	instSvc.On("LaunchInstanceWithOptions", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("launch failure"))

	err := p.Provision(ctx, cluster)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "launch failure")
}

func TestProvision_WorkerLaunchFailure(t *testing.T) {
	p, instSvc, repo, _, sgSvc, _ := setupProvisionerTest()
	ctx := context.Background()
	cluster := &domain.Cluster{
		ID:          uuid.New(),
		Name:        "test-cluster",
		VpcID:       uuid.New(),
		WorkerCount: 1,
	}

	repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	repo.On("AddNode", mock.Anything, mock.Anything).Return(nil)

	sgSvc.On("GetGroup", mock.Anything, mock.Anything, cluster.VpcID).Return(&domain.SecurityGroup{ID: uuid.New()}, nil)

	// Phase 3 CP Success
	cpInst := &domain.Instance{ID: uuid.New(), PrivateIP: "10.0.0.1"}
	instSvc.On("LaunchInstanceWithOptions", mock.Anything, mock.MatchedBy(func(opts ports.CreateInstanceOptions) bool {
		return opts.Name == "test-cluster-cp-0"
	})).Return(cpInst, nil)
	instSvc.On("GetInstance", mock.Anything, cpInst.ID.String()).Return(cpInst, nil)
	instSvc.On("ListInstances", mock.Anything).Return([]*domain.Instance{cpInst}, nil)
	instSvc.On("Exec", mock.Anything, cpInst.ID.String(), mock.Anything).Return("kubeconfig content", nil)

	// Phase 4 Worker Failure
	instSvc.On("LaunchInstanceWithOptions", mock.Anything, mock.MatchedBy(func(opts ports.CreateInstanceOptions) bool {
		return opts.Name == "test-cluster-worker-0"
	})).Return(nil, fmt.Errorf("worker launch failure"))

	// Provisioning should continue despite worker failure (it just logs and continues)
	err := p.Provision(ctx, cluster)
	assert.NoError(t, err)
}

func TestKubeadmProvisioner_Scale(t *testing.T) {
	p, instSvc, repo, _, _, _ := setupProvisionerTest()
	ctx := context.Background()
	cluster := &domain.Cluster{ID: uuid.New(), Name: "test", ControlPlaneIPs: []string{"10.0.0.1"}, WorkerCount: 2}

	instSvc.On("LaunchInstanceWithOptions", mock.Anything, mock.Anything).Return(&domain.Instance{ID: uuid.New()}, nil)
	repo.On("AddNode", mock.Anything, mock.Anything).Return(nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)

	err := p.Scale(ctx, cluster)
	assert.NoError(t, err)
}

func TestKubeadmProvisioner_Upgrade(t *testing.T) {
	p, instSvc, repo, _, _, _ := setupProvisionerTest()
	ctx := context.Background()
	cluster := &domain.Cluster{ID: uuid.New(), Name: "test", ControlPlaneIPs: []string{"10.0.0.1"}}
	node := &domain.ClusterNode{ID: uuid.New(), Role: domain.NodeRoleControlPlane, InstanceID: uuid.New()}

	repo.On("GetNodes", mock.Anything, cluster.ID).Return([]*domain.ClusterNode{node}, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)

	inst := &domain.Instance{ID: node.InstanceID, PrivateIP: "10.0.0.1"}
	instSvc.On("GetInstance", mock.Anything, node.InstanceID.String()).Return(inst, nil)
	instSvc.On("ListInstances", mock.Anything).Return([]*domain.Instance{inst}, nil)
	instSvc.On("Exec", mock.Anything, inst.ID.String(), mock.Anything).Return("done", nil)

	err := p.Upgrade(ctx, cluster, "v1.28.2")
	assert.NoError(t, err)
}

func TestKubeadmProvisioner_RotateSecrets(t *testing.T) {
	p, instSvc, repo, _, _, _ := setupProvisionerTest()
	ctx := context.Background()
	cluster := &domain.Cluster{ID: uuid.New(), Name: "test", ControlPlaneIPs: []string{"10.0.0.1"}}

	inst := &domain.Instance{ID: uuid.New(), PrivateIP: "10.0.0.1"}
	instSvc.On("ListInstances", mock.Anything).Return([]*domain.Instance{inst}, nil)
	instSvc.On("Exec", mock.Anything, inst.ID.String(), mock.Anything).Return("new-kubeconfig", nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)

	err := p.RotateSecrets(ctx, cluster)
	assert.NoError(t, err)
}

func TestKubeadmProvisioner_CreateBackup(t *testing.T) {
	p, instSvc, _, _, _, _ := setupProvisionerTest()
	ctx := context.Background()
	cluster := &domain.Cluster{ID: uuid.New(), Name: "test", ControlPlaneIPs: []string{"10.0.0.1"}}

	inst := &domain.Instance{ID: uuid.New(), PrivateIP: "10.0.0.1"}
	instSvc.On("ListInstances", mock.Anything).Return([]*domain.Instance{inst}, nil)
	instSvc.On("Exec", mock.Anything, inst.ID.String(), mock.Anything).Return("encoded-data", nil)

	err := p.CreateBackup(ctx, cluster)
	assert.NoError(t, err)
}

func TestKubeadmProvisioner_GetStatus(t *testing.T) {
	p, _, _, _, _, _ := setupProvisionerTest()
	status, err := p.GetStatus(context.Background(), &domain.Cluster{Status: domain.ClusterStatusRunning})
	assert.NoError(t, err)
	assert.Equal(t, domain.ClusterStatusRunning, status)
}

func TestKubeadmProvisioner_Repair(t *testing.T) {
	p, _, _, _, _, _ := setupProvisionerTest()
	err := p.Repair(context.Background(), &domain.Cluster{})
	assert.NoError(t, err)
}

func TestKubeadmProvisioner_Restore(t *testing.T) {
	p, _, _, _, _, _ := setupProvisionerTest()
	err := p.Restore(context.Background(), &domain.Cluster{}, "some-path")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not yet fully implemented")
}
