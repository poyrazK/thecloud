package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCheckable struct {
	mock.Mock
}

func (m *MockCheckable) Ping(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

type MockClusterService struct {
	mock.Mock
}

func (m *MockClusterService) CreateCluster(ctx context.Context, params ports.CreateClusterParams) (*domain.Cluster, error) {
	return nil, nil
}
func (m *MockClusterService) GetCluster(ctx context.Context, id uuid.UUID) (*domain.Cluster, error) {
	return nil, nil
}
func (m *MockClusterService) ListClusters(ctx context.Context, userID uuid.UUID) ([]*domain.Cluster, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Cluster), args.Error(1)
}
func (m *MockClusterService) DeleteCluster(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *MockClusterService) GetKubeconfig(ctx context.Context, id uuid.UUID, role string) (string, error) {
	return "", nil
}
func (m *MockClusterService) RepairCluster(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *MockClusterService) ScaleCluster(ctx context.Context, id uuid.UUID, workers int) error {
	return nil
}
func (m *MockClusterService) GetClusterHealth(ctx context.Context, id uuid.UUID) (*ports.ClusterHealth, error) {
	return nil, nil
}
func (m *MockClusterService) UpgradeCluster(ctx context.Context, id uuid.UUID, version string) error {
	return nil
}
func (m *MockClusterService) RotateSecrets(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *MockClusterService) CreateBackup(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *MockClusterService) RestoreBackup(ctx context.Context, id uuid.UUID, backupPath string) error {
	return nil
}

func TestHealthService_Check_Unit(t *testing.T) {
	mockDB := new(MockCheckable)
	mockCompute := new(MockComputeBackend)
	mockCluster := new(MockClusterService)
	svc := services.NewHealthServiceImpl(mockDB, mockCompute, mockCluster)

	ctx := context.Background()

	t.Run("AllUP", func(t *testing.T) {
		mockDB.On("Ping", mock.Anything).Return(nil).Once()
		mockCompute.On("Ping", mock.Anything).Return(nil).Once()
		mockCluster.On("ListClusters", mock.Anything, uuid.Nil).Return([]*domain.Cluster{}, nil).Once()

		res := svc.Check(ctx)
		assert.Equal(t, "UP", res.Status)
		assert.Equal(t, "CONNECTED", res.Checks["database_primary"])
		assert.Equal(t, "CONNECTED", res.Checks["docker"])
		assert.Equal(t, "OK", res.Checks["kubernetes_service"])
	})

	t.Run("DBDown", func(t *testing.T) {
		mockDB.On("Ping", mock.Anything).Return(assert.AnError).Once()
		mockCompute.On("Ping", mock.Anything).Return(nil).Once()
		mockCluster.On("ListClusters", mock.Anything, uuid.Nil).Return([]*domain.Cluster{}, nil).Once()

		res := svc.Check(ctx)
		assert.Equal(t, "DEGRADED", res.Status)
		assert.Contains(t, res.Checks["database_primary"], "DISCONNECTED")
	})
}
