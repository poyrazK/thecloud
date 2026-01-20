package services_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const testClusterName = "test-cluster"

func setupClusterServiceTest() (*MockClusterRepo, *MockClusterProvisioner, *MockVpcService, *MockInstanceService, *MockTaskQueue, *MockSecretService, ports.ClusterService) {
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
	return repo, provisioner, vpcSvc, instSvc, taskQueue, secretSvc, svc
}

func TestClusterServiceCreate(t *testing.T) {
	repo, _, vpcSvc, _, taskQueue, secretSvc, svc := setupClusterServiceTest()
	ctx := context.Background()
	userID := uuid.New()
	vpcID := uuid.New()

	vpcSvc.On("GetVPC", mock.Anything, vpcID.String()).Return(&domain.VPC{ID: vpcID}, nil)
	repo.On("Create", mock.Anything, mock.MatchedBy(func(c *domain.Cluster) bool {
		return c.Name == testClusterName && c.UserID == userID
	})).Return(nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	secretSvc.On("Encrypt", mock.Anything, userID, mock.Anything).Return("encrypted-key", nil)

	// Expect task queue enqueue
	taskQueue.On("Enqueue", mock.Anything, "k8s_jobs", mock.MatchedBy(func(job domain.ClusterJob) bool {
		return job.Type == domain.ClusterJobProvision && job.ClusterID != uuid.Nil && job.UserID == userID
	})).Return(nil).Once()

	cluster, err := svc.CreateCluster(ctx, ports.CreateClusterParams{
		UserID:  userID,
		Name:    testClusterName,
		VpcID:   vpcID,
		Version: "v1.29.0",
		Workers: 2,
	})

	assert.NoError(t, err)
	assert.NotNil(t, cluster)
	assert.Equal(t, testClusterName, cluster.Name)
	assert.Equal(t, domain.ClusterStatusPending, cluster.Status)

	// Wait for background provisioning - Wait time removed as provision is not async in test mock unless explicitly delayed
	taskQueue.AssertExpectations(t)
}

func TestClusterServiceDelete(t *testing.T) {
	repo, _, _, _, taskQueue, _, svc := setupClusterServiceTest()
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

	assert.NoError(t, err)

	taskQueue.AssertExpectations(t)
	repo.AssertCalled(t, "GetByID", mock.Anything, id)
	repo.AssertCalled(t, "Update", mock.Anything, mock.Anything)
}
