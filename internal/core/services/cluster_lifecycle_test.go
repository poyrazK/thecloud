package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testK8sVersion = "v1.29.0"
	testK8sBase    = "v1.28.0"
)

func TestClusterServiceUpgrade(t *testing.T) {
	t.Parallel()
	repo, provisioner, _, _, taskQueue, _, svc := setupClusterServiceTest()
	ctx := context.Background()
	id := uuid.New()
	cluster := &domain.Cluster{
		ID:      id,
		Status:  domain.ClusterStatusRunning,
		Version: testK8sBase,
		UserID:  uuid.New(),
	}

	t.Run("success initiates upgrade", func(t *testing.T) {
		repo.On("GetByID", mock.Anything, id).Return(cluster, nil).Once()
		repo.On("Update", mock.Anything, mock.MatchedBy(func(c *domain.Cluster) bool {
			return c.Status == domain.ClusterStatusUpgrading
		})).Return(nil).Once()

		// Expect task queue enqueue
		taskQueue.On("Enqueue", mock.Anything, "k8s_jobs", mock.MatchedBy(func(job domain.ClusterJob) bool {
			return job.Type == domain.ClusterJobUpgrade && job.ClusterID == id && job.Version == testK8sVersion
		})).Return(nil).Once()

		err := svc.UpgradeCluster(ctx, id, testK8sVersion)
		assert.NoError(t, err)

		// Wait for async upgrade
		time.Sleep(100 * time.Millisecond)
		provisioner.AssertExpectations(t)
		repo.AssertExpectations(t)
		taskQueue.AssertExpectations(t)
	})

	t.Run("fails if not running", func(t *testing.T) {
		cluster.Status = domain.ClusterStatusPending
		repo.On("GetByID", mock.Anything, id).Return(cluster, nil).Once()

		err := svc.UpgradeCluster(ctx, id, testK8sVersion)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be in running state")
	})

	t.Run("fails if same version", func(t *testing.T) {
		cluster.Status = domain.ClusterStatusRunning
		cluster.Version = testK8sVersion
		repo.On("GetByID", mock.Anything, id).Return(cluster, nil).Once()

		err := svc.UpgradeCluster(ctx, id, testK8sVersion)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already at this version")
	})

	t.Run("fails if skipping minor versions", func(t *testing.T) {
		cluster.Status = domain.ClusterStatusRunning
		cluster.Version = "v1.27.0"
		repo.On("GetByID", mock.Anything, id).Return(cluster, nil).Once()

		err := svc.UpgradeCluster(ctx, id, testK8sVersion)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot skip minor versions")
	})

	t.Run("fails if downgrading", func(t *testing.T) {
		cluster.Status = domain.ClusterStatusRunning
		cluster.Version = testK8sVersion
		repo.On("GetByID", mock.Anything, id).Return(cluster, nil).Once()

		err := svc.UpgradeCluster(ctx, id, testK8sBase)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be higher than current")
	})

	t.Run("fails with invalid target version", func(t *testing.T) {
		cluster.Status = domain.ClusterStatusRunning
		cluster.Version = testK8sBase
		repo.On("GetByID", mock.Anything, id).Return(cluster, nil).Once()

		err := svc.UpgradeCluster(ctx, id, "not-a-version")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid target version")
	})

	t.Run("allows upgrade with invalid current version", func(t *testing.T) {
		cluster.Status = domain.ClusterStatusRunning
		cluster.Version = "bad-current"
		repo.On("GetByID", mock.Anything, id).Return(cluster, nil).Once()
		repo.On("Update", mock.Anything, mock.MatchedBy(func(c *domain.Cluster) bool {
			return c.Status == domain.ClusterStatusUpgrading
		})).Return(nil).Once()
		taskQueue.On("Enqueue", mock.Anything, "k8s_jobs", mock.MatchedBy(func(job domain.ClusterJob) bool {
			return job.Type == domain.ClusterJobUpgrade && job.ClusterID == id && job.Version == testK8sVersion
		})).Return(nil).Once()

		err := svc.UpgradeCluster(ctx, id, testK8sVersion)
		assert.NoError(t, err)
	})

	t.Run("fails when enqueue fails", func(t *testing.T) {
		cluster.Status = domain.ClusterStatusRunning
		cluster.Version = testK8sBase
		repo.On("GetByID", mock.Anything, id).Return(cluster, nil).Once()
		repo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
		taskQueue.On("Enqueue", mock.Anything, "k8s_jobs", mock.Anything).Return(assert.AnError).Once()

		err := svc.UpgradeCluster(ctx, id, testK8sVersion)
		assert.Error(t, err)
	})
}

func TestClusterServiceRotateSecrets(t *testing.T) {
	t.Parallel()
	repo, provisioner, _, _, _, _, svc := setupClusterServiceTest()
	ctx := context.Background()
	id := uuid.New()
	cluster := &domain.Cluster{
		ID:     id,
		Status: domain.ClusterStatusRunning,
	}

	t.Run("success rotates secrets", func(t *testing.T) {
		repo.On("GetByID", mock.Anything, id).Return(cluster, nil).Once()
		repo.On("Update", mock.Anything, mock.MatchedBy(func(c *domain.Cluster) bool {
			return c.Status == domain.ClusterStatusUpdating
		})).Return(nil).Once()

		provisioner.On("RotateSecrets", mock.Anything, mock.Anything).Return(nil).Once()
		repo.On("Update", mock.Anything, mock.MatchedBy(func(c *domain.Cluster) bool {
			return c.Status == domain.ClusterStatusRunning
		})).Return(nil).Once()

		err := svc.RotateSecrets(ctx, id)
		assert.NoError(t, err)

		provisioner.AssertExpectations(t)
		repo.AssertExpectations(t)
	})
}

func TestClusterServiceBackup(t *testing.T) {
	t.Parallel()
	repo, provisioner, _, _, _, _, svc := setupClusterServiceTest()
	ctx := context.Background()
	id := uuid.New()
	cluster := &domain.Cluster{
		ID:     id,
		Status: domain.ClusterStatusRunning,
	}

	t.Run("success creates backup", func(t *testing.T) {
		repo.On("GetByID", mock.Anything, id).Return(cluster, nil).Once()
		provisioner.On("CreateBackup", mock.Anything, mock.Anything).Return(nil).Once()

		err := svc.CreateBackup(ctx, id)
		assert.NoError(t, err)
		provisioner.AssertExpectations(t)
	})

	t.Run("fails when not running", func(t *testing.T) {
		cluster.Status = domain.ClusterStatusPending
		repo.On("GetByID", mock.Anything, id).Return(cluster, nil).Once()

		err := svc.CreateBackup(ctx, id)
		assert.Error(t, err)
	})
}

func TestClusterServiceRestore(t *testing.T) {
	t.Parallel()
	repo, provisioner, _, _, _, _, svc := setupClusterServiceTest()
	ctx := context.Background()
	id := uuid.New()
	cluster := &domain.Cluster{
		ID:     id,
		Status: domain.ClusterStatusRunning,
	}

	t.Run("success restores from backup", func(t *testing.T) {
		backupPath := "/tmp/backup.db"
		repo.On("GetByID", mock.Anything, id).Return(cluster, nil).Once()
		repo.On("Update", mock.Anything, mock.MatchedBy(func(c *domain.Cluster) bool {
			return c.Status == domain.ClusterStatusRepairing
		})).Return(nil).Once()

		provisioner.On("Restore", mock.Anything, mock.Anything, backupPath).Return(nil).Once()
		repo.On("Update", mock.Anything, mock.MatchedBy(func(c *domain.Cluster) bool {
			return c.Status == domain.ClusterStatusRunning
		})).Return(nil).Once()

		err := svc.RestoreBackup(ctx, id, backupPath)
		assert.NoError(t, err)

		provisioner.AssertExpectations(t)
		repo.AssertExpectations(t)
	})
}
