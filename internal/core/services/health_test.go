package services

import (
	"context"
	"errors"
	"testing"

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

func TestHealthServiceCheck(t *testing.T) {
	t.Run("Healthy", func(t *testing.T) {
		db := new(mockCheckable)
		compute := new(fullMockComputeBackend)
		svc := NewHealthServiceImpl(db, compute)

		db.On("Ping", mock.Anything).Return(nil)
		compute.On("Ping", mock.Anything).Return(nil)

		res := svc.Check(context.Background())

		assert.Equal(t, "UP", res.Status)
		assert.Equal(t, "CONNECTED", res.Checks["database"])
		assert.Equal(t, "CONNECTED", res.Checks["docker"])
		db.AssertExpectations(t)
		compute.AssertExpectations(t)
	})

	t.Run("Degraded_DB", func(t *testing.T) {
		db := new(mockCheckable)
		compute := new(fullMockComputeBackend)
		svc := NewHealthServiceImpl(db, compute)

		db.On("Ping", mock.Anything).Return(errors.New("db error"))
		compute.On("Ping", mock.Anything).Return(nil)

		res := svc.Check(context.Background())

		assert.Equal(t, "DEGRADED", res.Status)
		assert.Contains(t, res.Checks["database"], "DISCONNECTED")
		assert.Equal(t, "CONNECTED", res.Checks["docker"])
	})

	t.Run("Degraded_Docker", func(t *testing.T) {
		db := new(mockCheckable)
		compute := new(fullMockComputeBackend)
		svc := NewHealthServiceImpl(db, compute)

		db.On("Ping", mock.Anything).Return(nil)
		compute.On("Ping", mock.Anything).Return(errors.New("docker error"))

		res := svc.Check(context.Background())

		assert.Equal(t, "DEGRADED", res.Status)
		assert.Equal(t, "CONNECTED", res.Checks["database"])
		assert.Contains(t, res.Checks["docker"], "DISCONNECTED")
	})
}
