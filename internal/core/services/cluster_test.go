package services_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
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

func setupClusterServiceTest(t *testing.T) (*MockClusterRepo, *MockClusterProvisioner, *MockVpcService, *MockTaskQueue, *MockSecretService, *MockRBACService, ports.ClusterService) {
	t.Helper()
	repo := new(MockClusterRepo)
	provisioner := new(MockClusterProvisioner)
	vpcSvc := new(MockVpcService)
	instSvc := new(MockInstanceService)
	secretSvc := new(MockSecretService)
	taskQueue := new(MockTaskQueue)
	rbacSvc := new(MockRBACService)
	rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc, _ := services.NewClusterService(services.ClusterServiceParams{
		Repo:        repo,
		RBAC:        rbacSvc,
		Provisioner: provisioner,
		VpcSvc:      vpcSvc,
		InstanceSvc: instSvc,
		SecretSvc:   secretSvc,
		TaskQueue:   taskQueue,
		Logger:      logger,
	})
	return repo, provisioner, vpcSvc, taskQueue, secretSvc, rbacSvc, svc
}


func TestClusterServiceCreate(t *testing.T) {
	t.Parallel()
	repo, _, vpcSvc, taskQueue, secretSvc, _, svc := setupClusterServiceTest(t)
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), userID), tenantID)
	vpcID := uuid.New()

	vpcSvc.On("GetVPC", mock.Anything, vpcID.String()).Return(&domain.VPC{ID: vpcID}, nil)
	repo.On("Create", mock.Anything, mock.MatchedBy(func(c *domain.Cluster) bool {
		return c.Name == testClusterName && c.UserID == userID
	})).Return(nil)
	repo.On("AddNodeGroup", mock.Anything, mock.MatchedBy(func(ng *domain.NodeGroup) bool {
		return ng.Name == "default-pool"
	})).Return(nil)
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

func TestClusterServiceCreateVpcNotFound(t *testing.T) {
	t.Parallel()
	_, _, vpcSvc, _, _, _, svc := setupClusterServiceTest(t)
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), userID), tenantID)
	vpcID := uuid.New()

	vpcSvc.On("GetVPC", mock.Anything, vpcID.String()).Return(nil, assert.AnError)

	cluster, err := svc.CreateCluster(ctx, ports.CreateClusterParams{
		UserID:  userID,
		Name:    testClusterName,
		VpcID:   vpcID,
		Version: clusterVersion,
		Workers: 2,
	})

	require.Error(t, err)
	assert.Nil(t, cluster)
}

func TestClusterServiceCreateEncryptError(t *testing.T) {
	t.Parallel()
	repo, _, vpcSvc, _, secretSvc, _, svc := setupClusterServiceTest(t)
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), userID), tenantID)
	vpcID := uuid.New()

	vpcSvc.On("GetVPC", mock.Anything, vpcID.String()).Return(&domain.VPC{ID: vpcID}, nil)
	secretSvc.On("Encrypt", mock.Anything, userID, mock.Anything).Return("", assert.AnError)

	cluster, err := svc.CreateCluster(ctx, ports.CreateClusterParams{
		UserID:  userID,
		Name:    testClusterName,
		VpcID:   vpcID,
		Version: clusterVersion,
		Workers: 2,
	})

	require.Error(t, err)
	assert.Nil(t, cluster)
	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestClusterServiceCreateRepoError(t *testing.T) {
	t.Parallel()
	repo, _, vpcSvc, _, secretSvc, _, svc := setupClusterServiceTest(t)
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), userID), tenantID)
	vpcID := uuid.New()

	vpcSvc.On("GetVPC", mock.Anything, vpcID.String()).Return(&domain.VPC{ID: vpcID}, nil)
	secretSvc.On("Encrypt", mock.Anything, userID, mock.Anything).Return(clusterEncryptedKey, nil)
	repo.On("Create", mock.Anything, mock.Anything).Return(assert.AnError)

	cluster, err := svc.CreateCluster(ctx, ports.CreateClusterParams{
		UserID:  userID,
		Name:    testClusterName,
		VpcID:   vpcID,
		Version: clusterVersion,
		Workers: 2,
	})

	require.Error(t, err)
	assert.Nil(t, cluster)
}

func TestClusterServiceCreateEnqueueError(t *testing.T) {
	t.Parallel()
	repo, _, vpcSvc, taskQueue, secretSvc, _, svc := setupClusterServiceTest(t)
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), userID), tenantID)
	vpcID := uuid.New()

	vpcSvc.On("GetVPC", mock.Anything, vpcID.String()).Return(&domain.VPC{ID: vpcID}, nil)
	secretSvc.On("Encrypt", mock.Anything, userID, mock.Anything).Return(clusterEncryptedKey, nil)
	repo.On("Create", mock.Anything, mock.Anything).Return(nil)
	repo.On("AddNodeGroup", mock.Anything, mock.MatchedBy(func(ng *domain.NodeGroup) bool {
		return ng.Name == "default-pool"
	})).Return(nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	taskQueue.On("Enqueue", mock.Anything, "k8s_jobs", mock.Anything).Return(assert.AnError).Once()

	cluster, err := svc.CreateCluster(ctx, ports.CreateClusterParams{
		UserID:  userID,
		Name:    testClusterName,
		VpcID:   vpcID,
		Version: clusterVersion,
		Workers: 2,
	})

	require.Error(t, err)
	assert.Nil(t, cluster)
}

func TestClusterServiceDelete(t *testing.T) {
	t.Parallel()
	repo, _, _, taskQueue, _, _, svc := setupClusterServiceTest(t)
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), userID), tenantID)
	id := uuid.New()
	cluster := &domain.Cluster{ID: id, Status: domain.ClusterStatusRunning}

	repo.On("GetByID", mock.Anything, id).Return(cluster, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	repo.On("Delete", mock.Anything, id).Return(nil)

	taskQueue.On("Enqueue", mock.Anything, "k8s_jobs", mock.MatchedBy(func(job domain.ClusterJob) bool {
		return job.Type == domain.ClusterJobDeprovision && job.ClusterID == id
	})).Return(nil).Once()

	err := svc.DeleteCluster(ctx, id)

	require.NoError(t, err)
	taskQueue.AssertExpectations(t)
}

func TestClusterServiceListClusters(t *testing.T) {
	t.Parallel()
	repo, _, _, _, _, _, svc := setupClusterServiceTest(t)
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), userID), tenantID)

	repo.On("ListByUserID", mock.Anything, userID).Return([]*domain.Cluster{{ID: uuid.New()}}, nil)

	clusters, err := svc.ListClusters(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, clusters, 1)
}

func TestClusterServiceGetKubeconfigAdmin(t *testing.T) {
	t.Parallel()
	repo, _, _, _, secretSvc, _, svc := setupClusterServiceTest(t)
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), userID), tenantID)
	clusterID := uuid.New()
	cluster := &domain.Cluster{
		ID:                  clusterID,
		UserID:              userID,
		Status:              domain.ClusterStatusRunning,
		KubeconfigEncrypted: "encrypted",
	}

	repo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil)
	secretSvc.On("Decrypt", mock.Anything, userID, "encrypted").Return("decrypted", nil)

	config, err := svc.GetKubeconfig(ctx, clusterID, "admin")
	require.NoError(t, err)
	assert.Equal(t, "decrypted", config)
}

func TestClusterServiceGetKubeconfigNonAdmin(t *testing.T) {
	t.Parallel()
	repo, provisioner, _, _, _, _, svc := setupClusterServiceTest(t)
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), userID), tenantID)
	clusterID := uuid.New()
	cluster := &domain.Cluster{ID: clusterID, Status: domain.ClusterStatusRunning}

	repo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil)
	provisioner.On("GetKubeconfig", mock.Anything, cluster, "viewer").Return("generated", nil)

	config, err := svc.GetKubeconfig(ctx, clusterID, "viewer")
	require.NoError(t, err)
	assert.Equal(t, "generated", config)
}

func TestClusterServiceGetKubeconfigNotRunning(t *testing.T) {
	t.Parallel()
	repo, _, _, _, _, _, svc := setupClusterServiceTest(t)
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), userID), tenantID)
	clusterID := uuid.New()
	cluster := &domain.Cluster{ID: clusterID, Status: domain.ClusterStatusPending}

	repo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil)

	_, err := svc.GetKubeconfig(ctx, clusterID, "admin")
	require.Error(t, err)
}

func TestClusterServiceRepairCluster(t *testing.T) {
	t.Parallel()
	repo, provisioner, _, _, _, _, svc := setupClusterServiceTest(t)
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), userID), tenantID)
	clusterID := uuid.New()
	cluster := &domain.Cluster{ID: clusterID, UserID: userID}

	repo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil)

	done := make(chan struct{})
	provisioner.On("Repair", mock.Anything, cluster).Return(nil).Run(func(_ mock.Arguments) {
		close(done)
	})

	err := svc.RepairCluster(ctx, clusterID)
	require.NoError(t, err)

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected repair to be invoked")
	}
}

func TestClusterServiceScaleCluster(t *testing.T) {
	t.Parallel()
	repo, provisioner, _, _, _, _, svc := setupClusterServiceTest(t)
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), userID), tenantID)
	clusterID := uuid.New()
	cluster := &domain.Cluster{ID: clusterID, UserID: userID, WorkerCount: 1}

	repo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(c *domain.Cluster) bool {
		return c.ID == clusterID && c.WorkerCount == 3
	})).Return(nil)

	done := make(chan struct{})
	provisioner.On("Scale", mock.Anything, mock.MatchedBy(func(c *domain.Cluster) bool {
		return c.ID == clusterID && c.WorkerCount == 3
	})).Return(nil).Run(func(_ mock.Arguments) {
		close(done)
	})

	err := svc.ScaleCluster(ctx, clusterID, 3)
	require.NoError(t, err)

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected scale to be invoked")
	}
}

func TestClusterServiceScaleClusterInvalidWorkers(t *testing.T) {
	t.Parallel()
	repo, _, _, _, _, _, svc := setupClusterServiceTest(t)
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), userID), tenantID)
	clusterID := uuid.New()
	cluster := &domain.Cluster{ID: clusterID, WorkerCount: 1}

	repo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil)

	err := svc.ScaleCluster(ctx, clusterID, 0)
	require.Error(t, err)
}

func TestClusterServiceGetClusterHealth(t *testing.T) {
	t.Parallel()
	repo, provisioner, _, _, _, _, svc := setupClusterServiceTest(t)
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), userID), tenantID)
	clusterID := uuid.New()
	cluster := &domain.Cluster{ID: clusterID}
	health := &ports.ClusterHealth{Status: domain.ClusterStatusRunning}

	repo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil)
	provisioner.On("GetHealth", mock.Anything, cluster).Return(health, nil)

	resp, err := svc.GetClusterHealth(ctx, clusterID)
	require.NoError(t, err)
	assert.Equal(t, domain.ClusterStatusRunning, resp.Status)
}

func TestClusterServiceRotateSecretsSuccess(t *testing.T) {
	t.Parallel()
	repo, provisioner, _, _, _, _, svc := setupClusterServiceTest(t)
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), userID), tenantID)
	clusterID := uuid.New()
	cluster := &domain.Cluster{ID: clusterID, Status: domain.ClusterStatusRunning}

	repo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(c *domain.Cluster) bool {
		return c.Status == domain.ClusterStatusUpdating
	})).Return(nil).Once()
	provisioner.On("RotateSecrets", mock.Anything, cluster).Return(nil).Once()
	repo.On("Update", mock.Anything, mock.MatchedBy(func(c *domain.Cluster) bool {
		return c.Status == domain.ClusterStatusRunning
	})).Return(nil).Once()

	err := svc.RotateSecrets(ctx, clusterID)
	require.NoError(t, err)
}

func TestClusterServiceRotateSecretsNotRunning(t *testing.T) {
	t.Parallel()
	repo, _, _, _, _, _, svc := setupClusterServiceTest(t)
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), userID), tenantID)
	clusterID := uuid.New()
	cluster := &domain.Cluster{ID: clusterID, Status: domain.ClusterStatusPending}

	repo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil)

	err := svc.RotateSecrets(ctx, clusterID)
	require.Error(t, err)
}

func TestClusterServiceRestoreBackup(t *testing.T) {
	t.Parallel()
	repo, provisioner, _, _, _, _, svc := setupClusterServiceTest(t)
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), userID), tenantID)
	clusterID := uuid.New()
	cluster := &domain.Cluster{ID: clusterID, Status: domain.ClusterStatusRunning}
	backupPath := "s3://bucket/backup"

	repo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(c *domain.Cluster) bool {
		return c.Status == domain.ClusterStatusRepairing
	})).Return(nil).Once()
	provisioner.On("Restore", mock.Anything, cluster, backupPath).Return(nil).Once()
	repo.On("Update", mock.Anything, mock.MatchedBy(func(s *domain.Cluster) bool {
		return s.Status == domain.ClusterStatusRunning
	})).Return(nil).Once()

	err := svc.RestoreBackup(ctx, clusterID, backupPath)
	require.NoError(t, err)
}

func TestClusterServiceRestoreBackupNotRunning(t *testing.T) {
	t.Parallel()
	repo, _, _, _, _, _, svc := setupClusterServiceTest(t)
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), userID), tenantID)
	clusterID := uuid.New()
	cluster := &domain.Cluster{ID: clusterID, Status: domain.ClusterStatusPending}

	repo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil)

	err := svc.RestoreBackup(ctx, clusterID, "s3://bucket/backup")
	require.Error(t, err)
}
