package services_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testClusterName     = "test-cluster"
	clusterEncryptedKey = "encrypted-key"
	clusterVersion      = "v1.29.0"
)

func setupClusterServiceTest(t *testing.T) (*MockClusterRepo, *MockClusterProvisioner, *MockVpcService, *MockTaskQueue, *MockSecretService, ports.ClusterService) {
	t.Helper()
	repo := new(MockClusterRepo)
	provisioner := new(MockClusterProvisioner)
	vpcSvc := new(MockVpcService)
	instSvc := new(MockInstanceService)
	secretSvc := new(MockSecretService)
	taskQueue := new(MockTaskQueue)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc, _ := services.NewClusterService(services.ClusterServiceParams{
		Repo:        repo,
		Provisioner: provisioner,
		VpcSvc:      vpcSvc,
		InstanceSvc: instSvc,
		SecretSvc:   secretSvc,
		TaskQueue:   taskQueue,
		Logger:      logger,
	})
	return repo, provisioner, vpcSvc, taskQueue, secretSvc, svc
}

func TestClusterServiceCreate(t *testing.T) {
	t.Parallel()
	repo, _, vpcSvc, taskQueue, secretSvc, svc := setupClusterServiceTest(t)
	ctx := context.Background()
	userID := uuid.New()
	vpcID := uuid.New()

	t.Run("success", func(t *testing.T) {
		vpcSvc.On("GetVPC", mock.Anything, vpcID.String()).Return(&domain.VPC{ID: vpcID}, nil).Once()
		repo.On("Create", mock.Anything, mock.MatchedBy(func(c *domain.Cluster) bool {
			return c.Name == testClusterName && c.UserID == userID
		})).Return(nil).Once()
		repo.On("AddNodeGroup", mock.Anything, mock.MatchedBy(func(ng *domain.NodeGroup) bool {
			return ng.Name == "default-pool"
		})).Return(nil).Once()
		secretSvc.On("Encrypt", mock.Anything, mock.Anything, mock.Anything).Return(clusterEncryptedKey, nil).Once()
		repo.On("Update", mock.Anything, mock.Anything).Return(nil).Twice()

		// Expect task queue enqueue
		taskQueue.On("Enqueue", mock.Anything, "k8s_jobs", mock.MatchedBy(func(job domain.ClusterJob) bool {
			return job.Type == domain.ClusterJobProvision && job.UserID == userID
		})).Return(nil).Once()

		cluster, err := svc.CreateCluster(ctx, ports.CreateClusterParams{
			UserID:  userID,
			Name:    testClusterName,
			VpcID:   vpcID,
			Version: clusterVersion,
			Workers: 2,
		})

		require.NoError(t, err)
		assert.NotNil(t, cluster)
		assert.Equal(t, testClusterName, cluster.Name)
		taskQueue.AssertExpectations(t)
	})

	t.Run("VPC_NotFound", func(t *testing.T) {
		vpcSvc.On("GetVPC", mock.Anything, vpcID.String()).Return(nil, errors.New("not found")).Once()
		_, err := svc.CreateCluster(ctx, ports.CreateClusterParams{VpcID: vpcID})
		assert.Error(t, err)
	})

	t.Run("RepoError", func(t *testing.T) {
		vpcSvc.On("GetVPC", mock.Anything, vpcID.String()).Return(&domain.VPC{ID: vpcID}, nil).Once()
		secretSvc.On("Encrypt", mock.Anything, mock.Anything, mock.Anything).Return(clusterEncryptedKey, nil).Maybe()
		repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db fail")).Once()
		_, err := svc.CreateCluster(ctx, ports.CreateClusterParams{VpcID: vpcID, Name: "test"})
		assert.Error(t, err)
	})
}

func TestClusterServiceDelete(t *testing.T) {
	t.Parallel()
	repo, _, _, taskQueue, _, svc := setupClusterServiceTest(t)
	ctx := context.Background()
	id := uuid.New()
	cluster := &domain.Cluster{ID: id, Status: domain.ClusterStatusRunning}

	t.Run("success", func(t *testing.T) {
		repo.On("GetByID", mock.Anything, id).Return(cluster, nil).Once()
		repo.On("Update", mock.Anything, mock.MatchedBy(func(c *domain.Cluster) bool {
			return c.Status == domain.ClusterStatusDeleting
		})).Return(nil).Once()
		repo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
		repo.On("Delete", mock.Anything, id).Return(nil).Once()

		taskQueue.On("Enqueue", mock.Anything, "k8s_jobs", mock.MatchedBy(func(job domain.ClusterJob) bool {
			return job.Type == domain.ClusterJobDeprovision && job.ClusterID == id
		})).Return(nil).Once()

		err := svc.DeleteCluster(ctx, id)
		require.NoError(t, err)
		taskQueue.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		repo.On("GetByID", mock.Anything, id).Return(nil, errors.New("not found")).Once()
		err := svc.DeleteCluster(ctx, id)
		assert.Error(t, err)
	})
}

func TestClusterServiceNodeGroupOps(t *testing.T) {
	t.Parallel()
	repo, _, _, _, _, svc := setupClusterServiceTest(t)
	ctx := context.Background()
	clusterID := uuid.New()

	t.Run("Add_Success", func(t *testing.T) {
		cluster := &domain.Cluster{ID: clusterID, Status: domain.ClusterStatusRunning, WorkerCount: 2}
		repo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()
		repo.On("AddNodeGroup", mock.Anything, mock.Anything).Return(nil).Once()
		repo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()

		params := ports.NodeGroupParams{
			Name: "test-pool", InstanceType: "standard-1",
			MinSize: 1, MaxSize: 5, DesiredSize: 3,
		}
		ng, err := svc.AddNodeGroup(ctx, clusterID, params)
		require.NoError(t, err)
		assert.Equal(t, "test-pool", ng.Name)
	})

	t.Run("Add_RepoError", func(t *testing.T) {
		cluster := &domain.Cluster{ID: clusterID, Status: domain.ClusterStatusRunning}
		repo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()
		repo.On("AddNodeGroup", mock.Anything, mock.Anything).Return(errors.New("db error")).Once()
		
		_, err := svc.AddNodeGroup(ctx, clusterID, ports.NodeGroupParams{Name: "test"})
		assert.Error(t, err)
	})

	t.Run("Delete_Success", func(t *testing.T) {
		clusterWithNG := &domain.Cluster{
			ID: clusterID, Status: domain.ClusterStatusRunning, WorkerCount: 5,
			NodeGroups: []domain.NodeGroup{
				{Name: "test-pool", CurrentSize: 3},
			},
		}
		repo.On("GetByID", mock.Anything, clusterID).Return(clusterWithNG, nil).Once()
		repo.On("DeleteNodeGroup", mock.Anything, clusterID, "test-pool").Return(nil).Once()
		repo.On("Update", mock.Anything, mock.MatchedBy(func(c *domain.Cluster) bool {
			return c.WorkerCount == 2
		})).Return(nil).Once()

		err := svc.DeleteNodeGroup(ctx, clusterID, "test-pool")
		require.NoError(t, err)
	})
}

func TestClusterServiceSetBackupPolicyExtra(t *testing.T) {
	t.Parallel()
	repo, _, _, _, _, svc := setupClusterServiceTest(t)
	ctx := context.Background()
	clusterID := uuid.New()
	cluster := &domain.Cluster{ID: clusterID, Status: domain.ClusterStatusRunning}

	t.Run("success", func(t *testing.T) {
		repo.On("GetByID", ctx, clusterID).Return(cluster, nil).Once()
		repo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()

		err := svc.SetBackupPolicy(ctx, clusterID, ports.BackupPolicyParams{
			Schedule:      "@weekly",
			RetentionDays: 30,
		})
		require.NoError(t, err)
	})

	t.Run("NotFound", func(t *testing.T) {
		repo.On("GetByID", ctx, clusterID).Return(nil, errors.New("not found")).Once()
		err := svc.SetBackupPolicy(ctx, clusterID, ports.BackupPolicyParams{})
		assert.Error(t, err)
	})
}

func TestClusterServiceScaleCluster(t *testing.T) {
	t.Parallel()
	repo, provisioner, _, _, _, svc := setupClusterServiceTest(t)
	ctx := context.Background()
	clusterID := uuid.New()
	cluster := &domain.Cluster{ID: clusterID, Status: domain.ClusterStatusRunning, WorkerCount: 2}

	t.Run("Scale_Up", func(t *testing.T) {
		repo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()
		repo.On("Update", mock.Anything, mock.MatchedBy(func(c *domain.Cluster) bool {
			return c.WorkerCount == 4
		})).Return(nil).Once()
		provisioner.On("Scale", mock.Anything, mock.Anything).Return(nil).Once()

		err := svc.ScaleCluster(ctx, clusterID, 4)
		require.NoError(t, err)
		
		// Wait for async scale
		time.Sleep(50 * time.Millisecond)
		provisioner.AssertExpectations(t)
	})

	t.Run("InvalidWorkers", func(t *testing.T) {
		repo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()
		err := svc.ScaleCluster(ctx, clusterID, 0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least 1 worker required")
	})

	t.Run("UpdateError", func(t *testing.T) {
		repo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()
		repo.On("Update", mock.Anything, mock.Anything).Return(errors.New("db fail")).Once()
		
		err := svc.ScaleCluster(ctx, clusterID, 5)
		assert.Error(t, err)
	})
}

func TestClusterServiceRepairCluster(t *testing.T) {
	t.Parallel()
	repo, provisioner, _, _, _, svc := setupClusterServiceTest(t)
	ctx := context.Background()
	clusterID := uuid.New()
	cluster := &domain.Cluster{ID: clusterID, Status: domain.ClusterStatusFailed}

	t.Run("Success", func(t *testing.T) {
		repo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()
		provisioner.On("Repair", mock.Anything, mock.Anything).Return(nil).Once()

		err := svc.RepairCluster(ctx, clusterID)
		require.NoError(t, err)
		
		// Wait for async repair
		time.Sleep(50 * time.Millisecond)
		provisioner.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		repo.On("GetByID", mock.Anything, clusterID).Return(nil, errors.New("not found")).Once()
		err := svc.RepairCluster(ctx, clusterID)
		assert.Error(t, err)
	})
}

func TestClusterServiceGetClusterHealth(t *testing.T) {
	t.Parallel()
	repo, provisioner, _, _, _, svc := setupClusterServiceTest(t)
	ctx := context.Background()
	clusterID := uuid.New()
	cluster := &domain.Cluster{ID: clusterID, Status: domain.ClusterStatusRunning}

	repo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()
	provisioner.On("GetHealth", mock.Anything, cluster).Return(&ports.ClusterHealth{Status: domain.ClusterStatusRunning}, nil).Once()

	health, err := svc.GetClusterHealth(ctx, clusterID)
	require.NoError(t, err)
	assert.Equal(t, domain.ClusterStatusRunning, health.Status)
}

func TestClusterServiceGetKubeconfigExtra(t *testing.T) {
	t.Parallel()
	repo, provisioner, _, _, secretSvc, svc := setupClusterServiceTest(t)
	ctx := context.Background()
	clusterID := uuid.New()
	userID := uuid.New()
	cluster := &domain.Cluster{
		ID:                  clusterID,
		UserID:              userID,
		Status:              domain.ClusterStatusRunning,
		KubeconfigEncrypted: "encrypted",
	}

	t.Run("Admin_Success", func(t *testing.T) {
		repo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()
		secretSvc.On("Decrypt", mock.Anything, userID, "encrypted").Return("decrypted-kubeconfig", nil).Once()

		cfg, err := svc.GetKubeconfig(ctx, clusterID, "admin")
		require.NoError(t, err)
		assert.Equal(t, "decrypted-kubeconfig", cfg)
	})

	t.Run("Decrypt_Error", func(t *testing.T) {
		repo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()
		secretSvc.On("Decrypt", mock.Anything, userID, "encrypted").Return("", errors.New("decrypt fail")).Once()

		_, err := svc.GetKubeconfig(ctx, clusterID, "admin")
		assert.Error(t, err)
	})

	t.Run("Guest_Success", func(t *testing.T) {
		repo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()
		provisioner.On("GetKubeconfig", mock.Anything, cluster, "viewer").Return("viewer-kubeconfig", nil).Once()

		cfg, err := svc.GetKubeconfig(ctx, clusterID, "viewer")
		require.NoError(t, err)
		assert.Equal(t, "viewer-kubeconfig", cfg)
	})

	t.Run("NotRunning", func(t *testing.T) {
		clusterPending := &domain.Cluster{ID: clusterID, Status: domain.ClusterStatusPending}
		repo.On("GetByID", mock.Anything, clusterID).Return(clusterPending, nil).Once()

		_, err := svc.GetKubeconfig(ctx, clusterID, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "only available when cluster is running")
	})
}
