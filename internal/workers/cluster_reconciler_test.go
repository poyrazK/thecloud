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

type mockProvisioner struct {
	mock.Mock
}

func (m *mockProvisioner) Provision(ctx context.Context, c *domain.Cluster) error {
	return m.Called(ctx, c).Error(0)
}
func (m *mockProvisioner) Deprovision(ctx context.Context, c *domain.Cluster) error {
	return m.Called(ctx, c).Error(0)
}
func (m *mockProvisioner) Upgrade(ctx context.Context, c *domain.Cluster, version string) error {
	return m.Called(ctx, c, version).Error(0)
}
func (m *mockProvisioner) GetStatus(ctx context.Context, c *domain.Cluster) (domain.ClusterStatus, error) {
	args := m.Called(ctx, c)
	r0, _ := args.Get(0).(domain.ClusterStatus)
	return r0, args.Error(1)
}
func (m *mockProvisioner) Repair(ctx context.Context, c *domain.Cluster) error {
	return m.Called(ctx, c).Error(0)
}
func (m *mockProvisioner) Scale(ctx context.Context, c *domain.Cluster) error {
	return m.Called(ctx, c).Error(0)
}
func (m *mockProvisioner) GetKubeconfig(ctx context.Context, c *domain.Cluster, role string) (string, error) {
	args := m.Called(ctx, c, role)
	return args.String(0), args.Error(1)
}
func (m *mockProvisioner) GetHealth(ctx context.Context, c *domain.Cluster) (*ports.ClusterHealth, error) {
	args := m.Called(ctx, c)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).(*ports.ClusterHealth)
	return r0, args.Error(1)
}
func (m *mockProvisioner) RotateSecrets(ctx context.Context, c *domain.Cluster) error {
	return m.Called(ctx, c).Error(0)
}
func (m *mockProvisioner) CreateBackup(ctx context.Context, c *domain.Cluster) error {
	return m.Called(ctx, c).Error(0)
}
func (m *mockProvisioner) Restore(ctx context.Context, c *domain.Cluster, path string) error {
	return m.Called(ctx, c, path).Error(0)
}

func TestClusterReconcilerReconcile(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	clusterID := uuid.New()
	cluster := &domain.Cluster{
		ID:     clusterID,
		Name:   "test-cluster",
		Status: domain.ClusterStatusRunning,
	}

	t.Run("Healthy Cluster", func(t *testing.T) {
		repo := new(MockClusterRepo)
		prov := new(mockProvisioner)
		reconciler := NewClusterReconciler(repo, prov, logger)

		repo.On("ListAll", mock.Anything).Return([]*domain.Cluster{cluster}, nil).Once()
		prov.On("GetHealth", mock.Anything, cluster).Return(&ports.ClusterHealth{
			APIServer:  true,
			NodesReady: 3,
			NodesTotal: 3,
		}, nil).Once()

		reconciler.reconcile(ctx)

		prov.AssertNotCalled(t, "Repair", mock.Anything, mock.Anything)
		repo.AssertExpectations(t)
		prov.AssertExpectations(t)
	})

	t.Run("Unhealthy Cluster - API Down", func(t *testing.T) {
		repo := new(MockClusterRepo)
		prov := new(mockProvisioner)
		reconciler := NewClusterReconciler(repo, prov, logger)

		repo.On("ListAll", mock.Anything).Return([]*domain.Cluster{cluster}, nil).Once()
		prov.On("GetHealth", mock.Anything, cluster).Return(&ports.ClusterHealth{
			APIServer:  false,
			NodesReady: 3,
			NodesTotal: 3,
		}, nil).Once()
		prov.On("Repair", mock.Anything, cluster).Return(nil).Once()

		reconciler.reconcile(ctx)

		repo.AssertExpectations(t)
		prov.AssertExpectations(t)
	})

	t.Run("Unhealthy Cluster - Missing Nodes", func(t *testing.T) {
		repo := new(MockClusterRepo)
		prov := new(mockProvisioner)
		reconciler := NewClusterReconciler(repo, prov, logger)

		repo.On("ListAll", mock.Anything).Return([]*domain.Cluster{cluster}, nil).Once()
		prov.On("GetHealth", mock.Anything, cluster).Return(&ports.ClusterHealth{
			APIServer:  true,
			NodesReady: 1,
			NodesTotal: 3,
		}, nil).Once()
		prov.On("Repair", mock.Anything, cluster).Return(nil).Once()

		reconciler.reconcile(ctx)

		repo.AssertExpectations(t)
		prov.AssertExpectations(t)
	})

	t.Run("Skipping Non-Running Cluster", func(t *testing.T) {
		repo := new(MockClusterRepo)
		prov := new(mockProvisioner)
		reconciler := NewClusterReconciler(repo, prov, logger)

		provisioningCluster := &domain.Cluster{
			ID:     uuid.New(),
			Status: domain.ClusterStatusProvisioning,
		}
		repo.On("ListAll", mock.Anything).Return([]*domain.Cluster{provisioningCluster}, nil).Once()

		reconciler.reconcile(ctx)

		prov.AssertNotCalled(t, "GetHealth", mock.Anything, mock.Anything)
		repo.AssertExpectations(t)
	})
}
