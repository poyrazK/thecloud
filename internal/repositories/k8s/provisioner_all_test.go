package k8s_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/repositories/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func prepareProvisioner() (*k8s.KubeadmProvisioner, *MockInstanceService, *MockClusterRepo, *MockSecretService, *MockSecurityGroupService, *MockStorageServiceV2, *MockLBService) {
	instSvc := new(MockInstanceService)
	repo := new(MockClusterRepo)
	secretSvc := new(MockSecretService)
	sgSvc := new(MockSecurityGroupService)
	storageSvc := new(MockStorageServiceV2)
	lbSvc := new(MockLBService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	p := k8s.NewKubeadmProvisioner(instSvc, repo, secretSvc, sgSvc, storageSvc, lbSvc, logger)
	return p, instSvc, repo, secretSvc, sgSvc, storageSvc, lbSvc
}

func TestKubeadmProvisionerK8sOps(t *testing.T) {
	ctx := context.Background()
	clusterID := uuid.New()
	vpcID := uuid.New()
	cluster := &domain.Cluster{
		ID:              clusterID,
		Name:            "test-cluster",
		VpcID:           vpcID,
		Version:         "v1.29.0",
		ControlPlaneIPs: []string{testIPMaster},
	}
	masterID := uuid.New()
	masterNode := &domain.ClusterNode{InstanceID: masterID, Role: domain.NodeRoleControlPlane}

	const sgName = "sg-test-cluster"

	t.Run("SecurityGroupOps", func(t *testing.T) {
		p, _, _, _, sgSvc, _, _ := prepareProvisioner()
		sgSvc.On("GetGroup", mock.Anything, sgName, vpcID).Return(nil, fmt.Errorf("not found")).Once()
		sgSvc.On("CreateGroup", mock.Anything, vpcID, sgName, mock.Anything).Return(&domain.SecurityGroup{ID: uuid.New()}, nil).Once()
		sgSvc.On("AddRule", mock.Anything, mock.Anything, mock.Anything).Return(&domain.SecurityRule{}, nil)

		err := p.ExportEnsureClusterSecurityGroup(ctx, cluster)
		assert.NoError(t, err)
		sgSvc.AssertExpectations(t)
	})

	t.Run("CreateNode", func(t *testing.T) {
		p, instSvc, repo, _, sgSvc, _, _ := prepareProvisioner()
		instSvc.On("LaunchInstance", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&domain.Instance{ID: masterID}, nil).Once()
		sgSvc.On("GetGroup", mock.Anything, sgName, vpcID).Return(&domain.SecurityGroup{ID: uuid.New()}, nil).Once()
		sgSvc.On("AttachToInstance", mock.Anything, masterID, mock.Anything).Return(nil).Once()
		repo.On("AddNode", mock.Anything, mock.Anything).Return(nil).Once()

		inst, err := p.ExportCreateNode(ctx, cluster, "master-0", domain.NodeRoleControlPlane)
		assert.NoError(t, err)
		assert.Equal(t, masterID, inst.ID)
		instSvc.AssertExpectations(t)
		sgSvc.AssertExpectations(t)
		repo.AssertExpectations(t)
	})

	t.Run("HealthAndRepair", func(t *testing.T) {
		p, instSvc, repo, _, _, _, _ := prepareProvisioner()
		repo.On("GetNodes", ctx, clusterID).Return([]*domain.ClusterNode{masterNode}, nil)
		instSvc.On("GetInstance", ctx, masterID.String()).Return(&domain.Instance{ID: masterID, PrivateIP: testIPMaster}, nil)

		// GetHealth Success
		healthCheckCmd := "kubectl --kubeconfig /etc/kubernetes/admin.conf get nodes"
		instSvc.On("Exec", ctx, masterID.String(), mock.MatchedBy(func(args []string) bool {
			return len(args) > 0 && args[len(args)-1] == healthCheckCmd
		})).Return("master Ready v1.29.0", nil).Once()

		healthNoHeadersCmd := "kubectl --kubeconfig /etc/kubernetes/admin.conf get nodes --no-headers"
		instSvc.On("Exec", ctx, masterID.String(), mock.MatchedBy(func(args []string) bool {
			return len(args) > 0 && args[len(args)-1] == healthNoHeadersCmd
		})).Return("master Ready v1.29.0", nil).Once()

		h, err := p.GetHealth(ctx, cluster)
		assert.NoError(t, err)
		assert.True(t, h.APIServer)
		assert.Equal(t, 1, h.NodesTotal)

		// Repair
		instSvc.On("Exec", ctx, masterID.String(), mock.Anything).Return("success", nil)
		err = p.Repair(ctx, cluster)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
		instSvc.AssertExpectations(t)
	})

	t.Run("KubeConfig", func(t *testing.T) {
		p, instSvc, repo, _, _, _, _ := prepareProvisioner()
		repo.On("GetNodes", ctx, clusterID).Return([]*domain.ClusterNode{masterNode}, nil)
		instSvc.On("GetInstance", ctx, masterID.String()).Return(&domain.Instance{ID: masterID, PrivateIP: testIPMaster}, nil)

		expectedCmd := "cat /etc/kubernetes/admin.conf"
		instSvc.On("Exec", ctx, masterID.String(), mock.MatchedBy(func(args []string) bool {
			return args[len(args)-1] == expectedCmd
		})).Return("kubeconfig-content", nil).Once()

		conf, err := p.GetKubeconfig(ctx, cluster, "admin")
		assert.NoError(t, err)
		assert.Equal(t, "kubeconfig-content", conf)
		repo.AssertExpectations(t)
		instSvc.AssertExpectations(t)
	})

	t.Run("ScaleDown", func(t *testing.T) {
		p, instSvc, repo, _, _, _, _ := prepareProvisioner()
		workerID := uuid.New()
		workerNode := &domain.ClusterNode{ID: uuid.New(), InstanceID: workerID, Role: domain.NodeRoleWorker}
		repo.On("GetNodes", ctx, clusterID).Return([]*domain.ClusterNode{masterNode, workerNode}, nil)

		// Scale down to 0 workers
		cluster.WorkerCount = 0
		instSvc.On("TerminateInstance", ctx, workerID.String()).Return(nil).Once()
		repo.On("DeleteNode", ctx, workerNode.ID).Return(nil).Once()

		err := p.Scale(ctx, cluster)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
		instSvc.AssertExpectations(t)
	})
}
