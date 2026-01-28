package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockCheckable struct {
	mock.Mock
}

func (m *mockCheckable) Ping(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

type fullMockComputeBackend struct {
	ports.ComputeBackend
	mock.Mock
}

func (m *fullMockComputeBackend) Ping(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

type mockClusterService struct {
	ports.ClusterService
	mock.Mock
}

func (m *mockClusterService) ListClusters(ctx context.Context, userID uuid.UUID) ([]*domain.Cluster, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Cluster), args.Error(1)
}

func TestHealthServiceCheck(t *testing.T) {
	t.Run("Healthy", func(t *testing.T) {
		db := new(mockCheckable)
		compute := new(fullMockComputeBackend)
		cluster := new(mockClusterService)
		svc := NewHealthServiceImpl(db, compute, cluster)

		db.On("Ping", mock.Anything).Return(nil)
		compute.On("Ping", mock.Anything).Return(nil)
		cluster.On("ListClusters", mock.Anything, mock.Anything).Return([]*domain.Cluster{}, nil)

		res := svc.Check(context.Background())

		assert.Equal(t, "UP", res.Status)
		assert.Equal(t, "CONNECTED", res.Checks["database_primary"])
		assert.Equal(t, "CONNECTED", res.Checks["docker"])
		assert.Equal(t, "OK", res.Checks["kubernetes_service"])
		db.AssertExpectations(t)
		compute.AssertExpectations(t)
	})

	t.Run("Degraded_DB", func(t *testing.T) {
		db := new(mockCheckable)
		compute := new(fullMockComputeBackend)
		cluster := new(mockClusterService)
		svc := NewHealthServiceImpl(db, compute, cluster)

		db.On("Ping", mock.Anything).Return(errors.New("db error"))
		compute.On("Ping", mock.Anything).Return(nil)
		cluster.On("ListClusters", mock.Anything, mock.Anything).Return([]*domain.Cluster{}, nil)

		res := svc.Check(context.Background())

		assert.Equal(t, "DEGRADED", res.Status)
		assert.Contains(t, res.Checks["database_primary"], "DISCONNECTED")
		assert.Equal(t, "CONNECTED", res.Checks["docker"])
	})

	t.Run("Degraded_Docker", func(t *testing.T) {
		db := new(mockCheckable)
		compute := new(fullMockComputeBackend)
		cluster := new(mockClusterService)
		svc := NewHealthServiceImpl(db, compute, cluster)

		db.On("Ping", mock.Anything).Return(nil)
		compute.On("Ping", mock.Anything).Return(errors.New("docker error"))
		cluster.On("ListClusters", mock.Anything, mock.Anything).Return([]*domain.Cluster{}, nil)

		res := svc.Check(context.Background())

		assert.Equal(t, "DEGRADED", res.Status)
		assert.Equal(t, "CONNECTED", res.Checks["database_primary"])
		assert.Contains(t, res.Checks["docker"], "DISCONNECTED")
	})

	t.Run("Degraded_K8s", func(t *testing.T) {
		db := new(mockCheckable)
		compute := new(fullMockComputeBackend)
		cluster := new(mockClusterService)
		svc := NewHealthServiceImpl(db, compute, cluster)

		db.On("Ping", mock.Anything).Return(nil)
		compute.On("Ping", mock.Anything).Return(nil)
		cluster.On("ListClusters", mock.Anything, mock.Anything).Return(nil, errors.New("k8s error"))

		res := svc.Check(context.Background())

		assert.Equal(t, "DEGRADED", res.Status)
		assert.Equal(t, "DEGRADED", res.Checks["kubernetes_service"])
	})
}
