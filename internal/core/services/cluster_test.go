package services_test

import (
	"context"
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
)

const testClusterName = "test-cluster"

func setupClusterServiceTest() (*MockClusterRepo, *MockClusterProvisioner, *MockVpcService, *MockInstanceService, ports.ClusterService) {
	repo := new(MockClusterRepo)
	provisioner := new(MockClusterProvisioner)
	vpcSvc := new(MockVpcService)
	instSvc := new(MockInstanceService)
	secretSvc := new(MockSecretService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := services.NewClusterService(services.ClusterServiceParams{
		Repo:        repo,
		Provisioner: provisioner,
		VpcSvc:      vpcSvc,
		InstanceSvc: instSvc,
		SecretSvc:   secretSvc,
		Logger:      logger,
	})
	return repo, provisioner, vpcSvc, instSvc, svc
}

func TestClusterServiceCreate(t *testing.T) {
	repo, provisioner, vpcSvc, _, svc := setupClusterServiceTest()
	ctx := context.Background()
	userID := uuid.New()
	vpcID := uuid.New()

	vpcSvc.On("GetVPC", mock.Anything, vpcID.String()).Return(&domain.VPC{ID: vpcID}, nil)
	repo.On("Create", mock.Anything, mock.MatchedBy(func(c *domain.Cluster) bool {
		return c.Name == testClusterName && c.UserID == userID
	})).Return(nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	provisioner.On("Provision", mock.Anything, mock.Anything).Return(nil)

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

	// Wait for background provisioning
	time.Sleep(100 * time.Millisecond)
	provisioner.AssertCalled(t, "Provision", mock.Anything, mock.Anything)
}

func TestClusterServiceDelete(t *testing.T) {
	repo, provisioner, _, _, svc := setupClusterServiceTest()
	ctx := context.Background()
	id := uuid.New()
	cluster := &domain.Cluster{ID: id, Status: domain.ClusterStatusRunning}

	repo.On("GetByID", mock.Anything, id).Return(cluster, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	provisioner.On("Deprovision", mock.Anything, mock.Anything).Return(nil)
	repo.On("Delete", mock.Anything, id).Return(nil)

	err := svc.DeleteCluster(ctx, id)

	assert.NoError(t, err)

	// Wait for background deprovisioning
	time.Sleep(100 * time.Millisecond)
	provisioner.AssertCalled(t, "Deprovision", mock.Anything, mock.Anything)
	repo.AssertCalled(t, "Delete", mock.Anything, id)
}
