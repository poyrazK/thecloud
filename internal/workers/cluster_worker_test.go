package workers

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/stretchr/testify/mock"
)

type MockTaskQueue struct {
	mock.Mock
}

func (m *MockTaskQueue) Enqueue(ctx context.Context, queue string, task any) error {
	args := m.Called(ctx, queue, task)
	return args.Error(0)
}

func (m *MockTaskQueue) Dequeue(ctx context.Context, queue string) (string, error) {
	args := m.Called(ctx, queue)
	return args.String(0), args.Error(1)
}

type MockClusterRepo struct{ mock.Mock }

func (m *MockClusterRepo) Create(ctx context.Context, c *domain.Cluster) error { return nil }
func (m *MockClusterRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Cluster, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).(*domain.Cluster)
	return r0, args.Error(1)
}
func (m *MockClusterRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Cluster, error) {
	return nil, nil
}
func (m *MockClusterRepo) ListAll(ctx context.Context) ([]*domain.Cluster, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).([]*domain.Cluster)
	return r0, args.Error(1)
}
func (m *MockClusterRepo) Update(ctx context.Context, c *domain.Cluster) error {
	return m.Called(ctx, c).Error(0)
}
func (m *MockClusterRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockClusterRepo) AddNode(ctx context.Context, n *domain.ClusterNode) error { return nil }
func (m *MockClusterRepo) GetNodes(ctx context.Context, clusterID uuid.UUID) ([]*domain.ClusterNode, error) {
	return nil, nil
}
func (m *MockClusterRepo) DeleteNode(ctx context.Context, nodeID uuid.UUID) error      { return nil }
func (m *MockClusterRepo) UpdateNode(ctx context.Context, n *domain.ClusterNode) error { return nil }

type MockProvisioner struct{ mock.Mock }

func (m *MockProvisioner) Provision(ctx context.Context, c *domain.Cluster) error {
	return m.Called(ctx, c).Error(0)
}
func (m *MockProvisioner) Deprovision(ctx context.Context, c *domain.Cluster) error {
	return m.Called(ctx, c).Error(0)
}
func (m *MockProvisioner) Upgrade(ctx context.Context, c *domain.Cluster, version string) error {
	return m.Called(ctx, c, version).Error(0)
}
func (m *MockProvisioner) GetStatus(ctx context.Context, c *domain.Cluster) (domain.ClusterStatus, error) {
	args := m.Called(ctx, c)
	r0, _ := args.Get(0).(domain.ClusterStatus)
	return r0, args.Error(1)
}
func (m *MockProvisioner) Repair(ctx context.Context, c *domain.Cluster) error {
	return m.Called(ctx, c).Error(0)
}
func (m *MockProvisioner) Scale(ctx context.Context, c *domain.Cluster) error {
	return m.Called(ctx, c).Error(0)
}
func (m *MockProvisioner) GetKubeconfig(ctx context.Context, c *domain.Cluster, role string) (string, error) {
	args := m.Called(ctx, c, role)
	return args.String(0), args.Error(1)
}
func (m *MockProvisioner) GetHealth(ctx context.Context, c *domain.Cluster) (*ports.ClusterHealth, error) {
	args := m.Called(ctx, c)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).(*ports.ClusterHealth)
	return r0, args.Error(1)
}
func (m *MockProvisioner) RotateSecrets(ctx context.Context, c *domain.Cluster) error {
	return m.Called(ctx, c).Error(0)
}
func (m *MockProvisioner) CreateBackup(ctx context.Context, c *domain.Cluster) error {
	return m.Called(ctx, c).Error(0)
}
func (m *MockProvisioner) Restore(ctx context.Context, c *domain.Cluster, path string) error {
	return m.Called(ctx, c, path).Error(0)
}

const workerClusterName = "worker-test"

func TestClusterWorkerProcessProvisionJob(t *testing.T) {
	tq := new(MockTaskQueue)
	repo := new(MockClusterRepo)
	prov := new(MockProvisioner)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	worker := NewClusterWorker(repo, prov, tq, logger)

	clusterID := uuid.New()
	userID := uuid.New()
	cluster := &domain.Cluster{ID: clusterID, UserID: userID, Name: workerClusterName}

	job := domain.ClusterJob{
		Type:      domain.ClusterJobProvision,
		ClusterID: clusterID,
		UserID:    userID,
	}

	repo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(c *domain.Cluster) bool {
		return c.Status == domain.ClusterStatusProvisioning || c.Status == domain.ClusterStatusRunning
	})).Return(nil)
	prov.On("Provision", mock.Anything, cluster).Return(nil)

	worker.processJob(job)

	repo.AssertExpectations(t)
	prov.AssertExpectations(t)
}

func TestClusterWorkerProcessDeprovisionJobSuccess(t *testing.T) {
	tq := new(MockTaskQueue)
	repo := new(MockClusterRepo)
	prov := new(MockProvisioner)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	worker := NewClusterWorker(repo, prov, tq, logger)

	clusterID := uuid.New()
	userID := uuid.New()
	cluster := &domain.Cluster{ID: clusterID, UserID: userID, Name: workerClusterName}

	job := domain.ClusterJob{
		Type:      domain.ClusterJobDeprovision,
		ClusterID: clusterID,
		UserID:    userID,
	}

	repo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Cluster")).Return(nil)
	prov.On("Deprovision", mock.Anything, cluster).Return(nil)
	repo.On("Delete", mock.Anything, clusterID).Return(nil)

	worker.processJob(job)

	repo.AssertExpectations(t)
	prov.AssertExpectations(t)
}

func TestClusterWorkerProcessDeprovisionJobFailure(t *testing.T) {
	tq := new(MockTaskQueue)
	repo := new(MockClusterRepo)
	prov := new(MockProvisioner)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	worker := NewClusterWorker(repo, prov, tq, logger)

	clusterID := uuid.New()
	userID := uuid.New()
	cluster := &domain.Cluster{ID: clusterID, UserID: userID, Name: workerClusterName}

	job := domain.ClusterJob{
		Type:      domain.ClusterJobDeprovision,
		ClusterID: clusterID,
		UserID:    userID,
	}

	repo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Cluster")).Return(nil).Twice()
	prov.On("Deprovision", mock.Anything, cluster).Return(io.EOF)

	worker.processJob(job)

	repo.AssertExpectations(t)
	prov.AssertExpectations(t)
}

func TestClusterWorkerProcessUpgradeJob(t *testing.T) {
	tq := new(MockTaskQueue)
	repo := new(MockClusterRepo)
	prov := new(MockProvisioner)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	worker := NewClusterWorker(repo, prov, tq, logger)

	clusterID := uuid.New()
	userID := uuid.New()
	version := "1.30.0"
	cluster := &domain.Cluster{ID: clusterID, UserID: userID, Name: workerClusterName}

	job := domain.ClusterJob{
		Type:      domain.ClusterJobUpgrade,
		ClusterID: clusterID,
		UserID:    userID,
		Version:   version,
	}

	repo.On("GetByID", mock.Anything, clusterID).Return(cluster, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(c *domain.Cluster) bool {
		return c.Status == domain.ClusterStatusUpgrading || c.Status == domain.ClusterStatusRunning
	})).Return(nil).Twice()
	prov.On("Upgrade", mock.Anything, cluster, version).Return(nil)

	worker.processJob(job)

	repo.AssertExpectations(t)
	prov.AssertExpectations(t)
}

func TestClusterWorkerProcessJobClusterNotFound(t *testing.T) {
	tq := new(MockTaskQueue)
	repo := new(MockClusterRepo)
	prov := new(MockProvisioner)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	worker := NewClusterWorker(repo, prov, tq, logger)

	clusterID := uuid.New()
	userID := uuid.New()
	job := domain.ClusterJob{
		Type:      domain.ClusterJobProvision,
		ClusterID: clusterID,
		UserID:    userID,
	}

	repo.On("GetByID", mock.Anything, clusterID).Return(nil, nil)

	worker.processJob(job)

	prov.AssertNotCalled(t, "Provision", mock.Anything, mock.Anything)
}
