package services_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"testing"

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

	vpcSvc.On("GetVPC", mock.Anything, vpcID.String()).Return(&domain.VPC{ID: vpcID}, nil)
	repo.On("Create", mock.Anything, mock.MatchedBy(func(c *domain.Cluster) bool {
		return c.Name == testClusterName && c.UserID == userID
	})).Return(nil)
	repo.On("AddNodeGroup", mock.Anything, mock.MatchedBy(func(ng *domain.NodeGroup) bool {
		return ng.Name == "default-pool"
	})).Return(nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	secretSvc.On("Encrypt", mock.Anything, userID, mock.Anything).Return(clusterEncryptedKey, nil)

	// Expect task queue enqueue
	taskQueue.On("Enqueue", mock.Anything, "k8s_jobs", mock.MatchedBy(func(job domain.ClusterJob) bool {
		return job.Type == domain.ClusterJobProvision && job.ClusterID != uuid.Nil && job.UserID == userID
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
	assert.Equal(t, domain.ClusterStatusPending, cluster.Status)

	taskQueue.AssertExpectations(t)
}

func TestClusterServiceDelete(t *testing.T) {
	t.Parallel()
	repo, _, _, taskQueue, _, svc := setupClusterServiceTest(t)
	ctx := context.Background()
	id := uuid.New()
	cluster := &domain.Cluster{ID: id, Status: domain.ClusterStatusRunning}

	repo.On("GetByID", mock.Anything, id).Return(cluster, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	repo.On("Delete", mock.Anything, id).Return(nil)

	// Expect task queue enqueue
	taskQueue.On("Enqueue", mock.Anything, "k8s_jobs", mock.MatchedBy(func(job domain.ClusterJob) bool {
		return job.Type == domain.ClusterJobDeprovision && job.ClusterID == id
	})).Return(nil).Once()

	err := svc.DeleteCluster(ctx, id)

	require.NoError(t, err)

	taskQueue.AssertExpectations(t)
}

func TestClusterServiceUpgradeCluster(t *testing.T) {
	t.Parallel()
	repo, _, _, taskQueue, _, svc := setupClusterServiceTest(t)
	ctx := context.Background()
	clusterID := uuid.New()
	
	t.Run("Success", func(t *testing.T) {
		cluster := &domain.Cluster{ID: clusterID, Status: domain.ClusterStatusRunning, Version: "v1.28.0"}
		repo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()
		repo.On("Update", mock.Anything, mock.MatchedBy(func(c *domain.Cluster) bool {
			return c.Status == domain.ClusterStatusUpgrading
		})).Return(nil).Once()
		taskQueue.On("Enqueue", mock.Anything, "k8s_jobs", mock.MatchedBy(func(j domain.ClusterJob) bool {
			return j.Type == domain.ClusterJobUpgrade && j.Version == "v1.29.0"
		})).Return(nil).Once()

		err := svc.UpgradeCluster(ctx, clusterID, "v1.29.0")
		require.NoError(t, err)
	})

	t.Run("InvalidVersion_Same", func(t *testing.T) {
		cluster := &domain.Cluster{ID: clusterID, Status: domain.ClusterStatusRunning, Version: "v1.29.0"}
		repo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()
		err := svc.UpgradeCluster(ctx, clusterID, "v1.29.0")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already at this version")
	})

	t.Run("InvalidVersion_SkipMinor", func(t *testing.T) {
		cluster := &domain.Cluster{ID: clusterID, Status: domain.ClusterStatusRunning, Version: "v1.27.0"}
		repo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()
		err := svc.UpgradeCluster(ctx, clusterID, "v1.29.0")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot skip minor versions")
	})
}

func TestClusterServiceNodeGroups(t *testing.T) {
	t.Parallel()
	repo, _, _, _, _, svc := setupClusterServiceTest(t)
	ctx := context.Background()
	clusterID := uuid.New()
	cluster := &domain.Cluster{ID: clusterID, Status: domain.ClusterStatusRunning, WorkerCount: 2}

	t.Run("Add_Success", func(t *testing.T) {
		repo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()
		repo.On("AddNodeGroup", mock.Anything, mock.Anything).Return(nil).Once()
		repo.On("Update", mock.Anything, mock.MatchedBy(func(c *domain.Cluster) bool {
			return c.WorkerCount == 5 // 2 existing + 3 new
		})).Return(nil).Once()

		params := ports.NodeGroupParams{
			Name: "test-pool", InstanceType: "standard-1",
			MinSize: 1, MaxSize: 5, DesiredSize: 3,
		}
		ng, err := svc.AddNodeGroup(ctx, clusterID, params)
		require.NoError(t, err)
		assert.Equal(t, "test-pool", ng.Name)
	})

	t.Run("Update_Success", func(t *testing.T) {
		clusterWithNG := &domain.Cluster{
			ID: clusterID, WorkerCount: 5,
			NodeGroups: []domain.NodeGroup{
				{Name: "test-pool", MinSize: 1, MaxSize: 5, CurrentSize: 3},
			},
		}
		repo.On("GetByID", mock.Anything, clusterID).Return(clusterWithNG, nil).Once()
		repo.On("UpdateNodeGroup", mock.Anything, mock.Anything).Return(nil).Once()
		repo.On("Update", mock.Anything, mock.MatchedBy(func(c *domain.Cluster) bool {
			return c.WorkerCount == 6
		})).Return(nil).Once()

		desired := 4
		params := ports.UpdateNodeGroupParams{DesiredSize: &desired}
		ng, err := svc.UpdateNodeGroup(ctx, clusterID, "test-pool", params)
		require.NoError(t, err)
		assert.Equal(t, 4, ng.CurrentSize)
	})

	t.Run("Update_NotFound", func(t *testing.T) {
		repo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()
		desired := 4
		_, err := svc.UpdateNodeGroup(ctx, clusterID, "non-existent", ports.UpdateNodeGroupParams{DesiredSize: &desired})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestClusterServiceRestoreBackup(t *testing.T) {
	t.Parallel()
	repo, provisioner, _, _, _, svc := setupClusterServiceTest(t)
	ctx := context.Background()
	clusterID := uuid.New()
	cluster := &domain.Cluster{ID: clusterID, Status: domain.ClusterStatusRunning}
	backupPath := "s3://bucket/backup"

	repo.On("GetByID", ctx, clusterID).Return(cluster, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(c *domain.Cluster) bool {
		return c.Status == domain.ClusterStatusRepairing
	})).Return(nil).Once()
	provisioner.On("Restore", ctx, cluster, backupPath).Return(nil).Once()
	repo.On("Update", mock.Anything, mock.MatchedBy(func(c *domain.Cluster) bool {
		return c.Status == domain.ClusterStatusRunning
	})).Return(nil).Once()

	err := svc.RestoreBackup(ctx, clusterID, backupPath)
	require.NoError(t, err)
}

func TestClusterServiceRestoreBackupFailure(t *testing.T) {
	t.Parallel()
	repo, provisioner, _, _, _, svc := setupClusterServiceTest(t)
	ctx := context.Background()
	clusterID := uuid.New()
	cluster := &domain.Cluster{ID: clusterID, Status: domain.ClusterStatusRunning}
	backupPath := "path/to/backup"

	repo.On("GetByID", ctx, clusterID).Return(cluster, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(c *domain.Cluster) bool {
		return c.Status == domain.ClusterStatusRepairing
	})).Return(nil).Once()
	
	provisioner.On("Restore", mock.Anything, cluster, backupPath).Return(fmt.Errorf("restore failed")).Once()
	
	repo.On("Update", mock.Anything, mock.MatchedBy(func(c *domain.Cluster) bool {
		return c.Status == domain.ClusterStatusFailed
	})).Return(nil).Once()

	err := svc.RestoreBackup(ctx, clusterID, backupPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "restore failed")
}

func TestClusterServiceSetBackupPolicy(t *testing.T) {
	t.Parallel()
	repo, _, _, _, _, svc := setupClusterServiceTest(t)
	ctx := context.Background()
	clusterID := uuid.New()
	cluster := &domain.Cluster{ID: clusterID, Status: domain.ClusterStatusRunning}

	repo.On("GetByID", ctx, clusterID).Return(cluster, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(c *domain.Cluster) bool {
		return c.BackupSchedule == "@weekly" && c.BackupRetentionDays == 30
	})).Return(nil).Once()

	err := svc.SetBackupPolicy(ctx, clusterID, ports.BackupPolicyParams{
		Schedule:      "@weekly",
		RetentionDays: 30,
	})
	require.NoError(t, err)
}
