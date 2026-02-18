package services_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClusterService_Unit(t *testing.T) {
	mockRepo := new(MockClusterRepo)
	mockProv := new(MockClusterProvisioner)
	mockVpcSvc := new(MockVpcService)
	mockInstSvc := new(MockInstanceService)
	mockSecretSvc := new(MockSecretService)
	mockTaskQueue := new(MockTaskQueue)
	
	params := services.ClusterServiceParams{
		Repo:        mockRepo,
		Provisioner: mockProv,
		VpcSvc:      mockVpcSvc,
		InstanceSvc: mockInstSvc,
		SecretSvc:   mockSecretSvc,
		TaskQueue:   mockTaskQueue,
		Logger:      slog.Default(),
	}
	
	svc, err := services.NewClusterService(params)
	assert.NoError(t, err)

	ctx := context.Background()
	userID := uuid.New()
	vpcID := uuid.New()

	t.Run("CreateCluster", func(t *testing.T) {
		mockVpcSvc.On("GetVPC", mock.Anything, vpcID.String()).Return(&domain.VPC{ID: vpcID}, nil).Once()
		mockSecretSvc.On("Encrypt", mock.Anything, userID, mock.Anything).Return("encrypted-key", nil).Once()
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		mockTaskQueue.On("Enqueue", mock.Anything, "k8s_jobs", mock.Anything).Return(nil).Once()

		params := ports.CreateClusterParams{
			UserID:  userID,
			Name:    "test-cluster",
			VpcID:   vpcID,
			Workers: 3,
		}

		cluster, err := svc.CreateCluster(ctx, params)
		assert.NoError(t, err)
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
		assert.NoError(t, err)
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
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "higher than current")
	})
}
