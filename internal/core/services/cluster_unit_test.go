package services_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestClusterService_Unit(t *testing.T) {
	mockRepo := new(MockClusterRepo)
	mockProv := new(MockClusterProvisioner)
	mockVpcSvc := new(MockVpcService)
	mockInstSvc := new(MockInstanceService)
	mockSecretSvc := new(MockSecretService)
	mockTaskQueue := new(MockTaskQueue)
	rbacSvc := new(MockRBACService)

	rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	params := services.ClusterServiceParams{
		Repo:        mockRepo,
		Provisioner: mockProv,
		VpcSvc:      mockVpcSvc,
		InstanceSvc: mockInstSvc,
		SecretSvc:   mockSecretSvc,
		TaskQueue:   mockTaskQueue,
		RBAC:        rbacSvc,
		Logger:      slog.Default(),
	}

	svc, err := services.NewClusterService(params)
	require.NoError(t, err)

	userID := uuid.New()
	tenantID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)
	vpcID := uuid.New()

	t.Run("CreateCluster", func(t *testing.T) {
		mockVpcSvc.On("GetVPC", mock.Anything, vpcID.String()).Return(&domain.VPC{ID: vpcID}, nil).Once()
		mockSecretSvc.On("Encrypt", mock.Anything, userID, mock.Anything).Return("encrypted-key", nil).Once()
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		mockRepo.On("AddNodeGroup", mock.Anything, mock.MatchedBy(func(ng *domain.NodeGroup) bool {
			return ng.Name == "default-pool"
		})).Return(nil).Once()
		mockTaskQueue.On("Enqueue", mock.Anything, "k8s_jobs", mock.Anything).Return(nil).Once()

		params := ports.CreateClusterParams{
			UserID:  userID,
			Name:    "test-cluster",
			VpcID:   vpcID,
			Workers: 3,
		}

		cluster, err := svc.CreateCluster(ctx, params)
		require.NoError(t, err)
		assert.NotNil(t, cluster)
		assert.Equal(t, "test-cluster", cluster.Name)
		assert.Equal(t, 3, cluster.WorkerCount)
		mockRepo.AssertExpectations(t)
	})

	t.Run("UpgradeCluster_Success", func(t *testing.T) {
		clusterID := uuid.New()
		cluster := &domain.Cluster{
			ID:      clusterID,
			UserID:  userID,
			Version: "v1.28.0",
			Status:  domain.ClusterStatusRunning,
		}
		mockRepo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
		mockTaskQueue.On("Enqueue", mock.Anything, "k8s_jobs", mock.Anything).Return(nil).Once()

		err := svc.UpgradeCluster(ctx, clusterID, "v1.29.0")
		require.NoError(t, err)
	})

	t.Run("UpgradeCluster_InvalidVersion", func(t *testing.T) {
		clusterID := uuid.New()
		cluster := &domain.Cluster{
			ID:      clusterID,
			UserID:  userID,
			Version: "v1.28.0",
			Status:  domain.ClusterStatusRunning,
		}
		mockRepo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()

		err := svc.UpgradeCluster(ctx, clusterID, "v1.27.0")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "higher than current")
	})

	t.Run("GetCluster_Success", func(t *testing.T) {
		clusterID := uuid.New()
		cluster := &domain.Cluster{ID: clusterID, UserID: userID, TenantID: tenantID}
		mockRepo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()

		res, err := svc.GetCluster(ctx, clusterID)
		require.NoError(t, err)
		assert.Equal(t, clusterID, res.ID)
	})

	t.Run("GetCluster_NotFound", func(t *testing.T) {
		clusterID := uuid.New()
		mockRepo.On("GetByID", mock.Anything, clusterID).Return(nil, nil).Once()

		_, err := svc.GetCluster(ctx, clusterID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("ListClusters_Success", func(t *testing.T) {
		clusters := []*domain.Cluster{
			{ID: uuid.New(), UserID: userID},
			{ID: uuid.New(), UserID: userID},
		}
		mockRepo.On("ListByUserID", mock.Anything, userID).Return(clusters, nil).Once()

		res, err := svc.ListClusters(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, res, 2)
	})

	t.Run("DeleteCluster_Success", func(t *testing.T) {
		clusterID := uuid.New()
		cluster := &domain.Cluster{ID: clusterID, UserID: userID, TenantID: tenantID, Status: domain.ClusterStatusRunning}
		mockRepo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
		mockTaskQueue.On("Enqueue", mock.Anything, "k8s_jobs", mock.Anything).Return(nil).Once()

		err := svc.DeleteCluster(ctx, clusterID)
		require.NoError(t, err)
	})

	t.Run("GetKubeconfig_Success", func(t *testing.T) {
		clusterID := uuid.New()
		cluster := &domain.Cluster{
			ID:                   clusterID,
			UserID:               userID,
			TenantID:             tenantID,
			Status:               domain.ClusterStatusRunning,
			KubeconfigEncrypted:  "encrypted-kubeconfig",
		}
		mockRepo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()
		mockSecretSvc.On("Decrypt", mock.Anything, userID, "encrypted-kubeconfig").Return("decrypted-kubeconfig", nil).Once()

		res, err := svc.GetKubeconfig(ctx, clusterID, "admin")
		require.NoError(t, err)
		assert.Equal(t, "decrypted-kubeconfig", res)
	})

	t.Run("GetKubeconfig_NotRunning", func(t *testing.T) {
		clusterID := uuid.New()
		cluster := &domain.Cluster{
			ID:      clusterID,
			UserID:  userID,
			TenantID: tenantID,
			Status:  domain.ClusterStatusPending,
		}
		mockRepo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()

		_, err := svc.GetKubeconfig(ctx, clusterID, "admin")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "only available when cluster is running")
	})

	t.Run("ScaleCluster_Success", func(t *testing.T) {
		clusterID := uuid.New()
		cluster := &domain.Cluster{
			ID:           clusterID,
			UserID:       userID,
			TenantID:     tenantID,
			Status:       domain.ClusterStatusRunning,
			WorkerCount:  2,
		}
		mockRepo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
		mockProv.On("Scale", mock.Anything, mock.Anything).Return(nil).Maybe()

		err := svc.ScaleCluster(ctx, clusterID, 5)
		require.NoError(t, err)
		assert.Equal(t, 5, cluster.WorkerCount)
	})

	t.Run("ScaleCluster_InvalidWorkers", func(t *testing.T) {
		clusterID := uuid.New()
		cluster := &domain.Cluster{
			ID:      clusterID,
			UserID:  userID,
			TenantID: tenantID,
			Status:  domain.ClusterStatusRunning,
		}
		mockRepo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()

		err := svc.ScaleCluster(ctx, clusterID, 0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least 1 worker required")
	})

	t.Run("GetClusterHealth_Success", func(t *testing.T) {
		clusterID := uuid.New()
		cluster := &domain.Cluster{ID: clusterID, UserID: userID, TenantID: tenantID}
		health := &ports.ClusterHealth{Status: "healthy"}
		mockRepo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()
		mockProv.On("GetHealth", mock.Anything, cluster).Return(health, nil).Once()

		res, err := svc.GetClusterHealth(ctx, clusterID)
		require.NoError(t, err)
		assert.Equal(t, domain.ClusterStatus("healthy"), res.Status)
	})

	t.Run("RotateSecrets_Success", func(t *testing.T) {
		clusterID := uuid.New()
		cluster := &domain.Cluster{
			ID:      clusterID,
			UserID:  userID,
			TenantID: tenantID,
			Status:  domain.ClusterStatusRunning,
		}
		mockRepo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Twice()
		mockProv.On("RotateSecrets", mock.Anything, cluster).Return(nil).Once()

		err := svc.RotateSecrets(ctx, clusterID)
		require.NoError(t, err)
	})

	t.Run("CreateBackup_Success", func(t *testing.T) {
		clusterID := uuid.New()
		cluster := &domain.Cluster{
			ID:      clusterID,
			UserID:  userID,
			TenantID: tenantID,
			Status:  domain.ClusterStatusRunning,
		}
		mockRepo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()
		mockProv.On("CreateBackup", mock.Anything, cluster).Return(nil).Once()

		err := svc.CreateBackup(ctx, clusterID)
		require.NoError(t, err)
	})

	t.Run("CreateBackup_NotRunning", func(t *testing.T) {
		clusterID := uuid.New()
		cluster := &domain.Cluster{
			ID:      clusterID,
			UserID:  userID,
			TenantID: tenantID,
			Status:  domain.ClusterStatusPending,
		}
		mockRepo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()

		err := svc.CreateBackup(ctx, clusterID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "running state")
	})

	t.Run("RestoreBackup_Success", func(t *testing.T) {
		clusterID := uuid.New()
		cluster := &domain.Cluster{
			ID:      clusterID,
			UserID:  userID,
			TenantID: tenantID,
			Status:  domain.ClusterStatusRunning,
		}
		mockRepo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Twice()
		mockProv.On("Restore", mock.Anything, cluster, "/path/to/backup").Return(nil).Once()

		err := svc.RestoreBackup(ctx, clusterID, "/path/to/backup")
		require.NoError(t, err)
	})

	t.Run("SetBackupPolicy_Success", func(t *testing.T) {
		clusterID := uuid.New()
		cluster := &domain.Cluster{
			ID:      clusterID,
			UserID:  userID,
			TenantID: tenantID,
			Status:  domain.ClusterStatusRunning,
		}
		mockRepo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()

		retention := 14
		err := svc.SetBackupPolicy(ctx, clusterID, ports.BackupPolicyParams{RetentionDays: &retention})
		require.NoError(t, err)
		assert.Equal(t, 14, cluster.BackupRetentionDays)
	})

	t.Run("SetBackupPolicy_InvalidRetention", func(t *testing.T) {
		clusterID := uuid.New()
		cluster := &domain.Cluster{
			ID:      clusterID,
			UserID:  userID,
			TenantID: tenantID,
			Status:  domain.ClusterStatusRunning,
		}
		mockRepo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()

		retention := 0
		err := svc.SetBackupPolicy(ctx, clusterID, ports.BackupPolicyParams{RetentionDays: &retention})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid retention")
	})

	t.Run("AddNodeGroup_Success", func(t *testing.T) {
		clusterID := uuid.New()
		cluster := &domain.Cluster{
			ID:           clusterID,
			UserID:       userID,
			TenantID:     tenantID,
			WorkerCount:  2,
			NodeGroups:   []domain.NodeGroup{{Name: "default-pool"}},
		}
		mockRepo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()
		mockRepo.On("AddNodeGroup", mock.Anything, mock.Anything).Return(nil).Once()
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()

		params := ports.NodeGroupParams{
			Name:        "new-pool",
			InstanceType: "standard-2",
			MinSize:     1,
			MaxSize:     5,
			DesiredSize: 3,
		}
		ng, err := svc.AddNodeGroup(ctx, clusterID, params)
		require.NoError(t, err)
		assert.Equal(t, "new-pool", ng.Name)
	})

	t.Run("UpdateNodeGroup_Success", func(t *testing.T) {
		clusterID := uuid.New()
		cluster := &domain.Cluster{
			ID:       clusterID,
			UserID:   userID,
			TenantID: tenantID,
			NodeGroups: []domain.NodeGroup{
				{Name: "existing-pool", MinSize: 1, MaxSize: 3, CurrentSize: 2},
			},
		}
		newMin := 2
		mockRepo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()
		mockRepo.On("UpdateNodeGroup", mock.Anything, mock.Anything).Return(nil).Once()
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()

		params := ports.UpdateNodeGroupParams{MinSize: &newMin}
		ng, err := svc.UpdateNodeGroup(ctx, clusterID, "existing-pool", params)
		require.NoError(t, err)
		assert.Equal(t, 2, ng.MinSize)
	})

	t.Run("UpdateNodeGroup_NotFound", func(t *testing.T) {
		clusterID := uuid.New()
		cluster := &domain.Cluster{
			ID:         clusterID,
			UserID:     userID,
			TenantID:   tenantID,
			NodeGroups: []domain.NodeGroup{},
		}
		mockRepo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()

		params := ports.UpdateNodeGroupParams{}
		_, err := svc.UpdateNodeGroup(ctx, clusterID, "nonexistent", params)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("DeleteNodeGroup_Success", func(t *testing.T) {
		clusterID := uuid.New()
		cluster := &domain.Cluster{
			ID:           clusterID,
			UserID:       userID,
			TenantID:     tenantID,
			WorkerCount:  5,
			NodeGroups: []domain.NodeGroup{
				{Name: "default-pool"},
				{Name: "extra-pool", CurrentSize: 3},
			},
		}
		mockRepo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()
		mockRepo.On("DeleteNodeGroup", mock.Anything, clusterID, "extra-pool").Return(nil).Once()
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()

		err := svc.DeleteNodeGroup(ctx, clusterID, "extra-pool")
		require.NoError(t, err)
	})

	t.Run("DeleteNodeGroup_DefaultPool", func(t *testing.T) {
		clusterID := uuid.New()
		cluster := &domain.Cluster{
			ID:         clusterID,
			UserID:     userID,
			TenantID:   tenantID,
			NodeGroups: []domain.NodeGroup{{Name: "default-pool"}},
		}
		mockRepo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()

		err := svc.DeleteNodeGroup(ctx, clusterID, "default-pool")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot delete default node group")
	})

	t.Run("RepairCluster_Success", func(t *testing.T) {
		clusterID := uuid.New()
		cluster := &domain.Cluster{
			ID:      clusterID,
			UserID:  userID,
			TenantID: tenantID,
			Status:  domain.ClusterStatusRunning,
		}
		mockRepo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil).Once()
		mockProv.On("Repair", mock.Anything, cluster).Return(nil).Maybe()

		err := svc.RepairCluster(ctx, clusterID)
		require.NoError(t, err)
	})
}
