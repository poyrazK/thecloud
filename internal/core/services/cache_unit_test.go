package services_test

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
)

type MockCacheRepository struct {
	mock.Mock
}

func (m *MockCacheRepository) Create(ctx context.Context, cache *domain.Cache) error {
	return m.Called(ctx, cache).Error(0)
}
func (m *MockCacheRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Cache, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).(*domain.Cache)
	return r0, args.Error(1)
}
func (m *MockCacheRepository) GetByName(ctx context.Context, userID uuid.UUID, name string) (*domain.Cache, error) {
	args := m.Called(ctx, userID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).(*domain.Cache)
	return r0, args.Error(1)
}
func (m *MockCacheRepository) List(ctx context.Context, userID uuid.UUID) ([]*domain.Cache, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).([]*domain.Cache)
	return r0, args.Error(1)
}
func (m *MockCacheRepository) Update(ctx context.Context, cache *domain.Cache) error {
	return m.Called(ctx, cache).Error(0)
}
func (m *MockCacheRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func TestCacheService_Unit_Extended(t *testing.T) {
	repo := new(MockCacheRepository)
	compute := new(MockComputeBackend)
	eventSvc := new(MockEventService)
	auditSvc := new(MockAuditService)
	svc := services.NewCacheService(repo, compute, nil, eventSvc, auditSvc, slog.Default())
	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)

	t.Run("CreateCache_Success", func(t *testing.T) {
		repo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		compute.On("LaunchInstanceWithOptions", mock.Anything, mock.Anything).Return("cid", []string{"30001:6379"}, nil).Once()
		repo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
		eventSvc.On("RecordEvent", mock.Anything, "CACHE_CREATE", mock.Anything, "CACHE", mock.Anything).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "cache.create", "cache", mock.Anything, mock.Anything).Return(nil).Once()

		cache, err := svc.CreateCache(ctx, "my-cache", "7.0", 128, nil)
		require.NoError(t, err)
		assert.NotNil(t, cache)
		assert.Equal(t, 30001, cache.Port)
	})

	t.Run("GetCache", func(t *testing.T) {
		cacheID := uuid.New()
		repo.On("GetByID", mock.Anything, cacheID).Return(&domain.Cache{ID: cacheID}, nil).Once()
		res, err := svc.GetCache(ctx, cacheID.String())
		require.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("ListCaches", func(t *testing.T) {
		repo.On("List", mock.Anything, userID).Return([]*domain.Cache{}, nil).Once()
		res, err := svc.ListCaches(ctx)
		require.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("DeleteCache", func(t *testing.T) {
		cacheID := uuid.New()
		cache := &domain.Cache{ID: cacheID, UserID: userID, Name: "my-cache", ContainerID: "cid"}
		repo.On("GetByID", mock.Anything, cacheID).Return(cache, nil).Once()
		compute.On("StopInstance", mock.Anything, "cid").Return(nil).Once()
		compute.On("DeleteInstance", mock.Anything, "cid").Return(nil).Once()
		repo.On("Delete", mock.Anything, cacheID).Return(nil).Once()
		eventSvc.On("RecordEvent", mock.Anything, "CACHE_DELETE", cacheID.String(), "CACHE", mock.Anything).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "cache.delete", "cache", cacheID.String(), mock.Anything).Return(nil).Once()

		err := svc.DeleteCache(ctx, cacheID.String())
		require.NoError(t, err)
	})

	t.Run("GetConnectionString", func(t *testing.T) {
		cacheID := uuid.New()
		cache := &domain.Cache{
			ID:       cacheID,
			Password: "pass",
			Port:     6379,
		}
		repo.On("GetByID", mock.Anything, cacheID).Return(cache, nil).Once()
		
		conn, err := svc.GetConnectionString(ctx, cacheID.String())
		require.NoError(t, err)
		assert.Contains(t, conn, "redis://:pass@localhost:6379")
	})

	t.Run("FlushCache", func(t *testing.T) {
		cacheID := uuid.New()
		cache := &domain.Cache{ID: cacheID, ContainerID: "cid", UserID: userID, Status: domain.CacheStatusRunning}
		repo.On("GetByID", mock.Anything, cacheID).Return(cache, nil).Once()
		compute.On("Exec", mock.Anything, "cid", []string{"redis-cli", "FLUSHALL"}).Return("", nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "cache.flush", "cache", cacheID.String(), mock.Anything).Return(nil).Once()

		err := svc.FlushCache(ctx, cacheID.String())
		require.NoError(t, err)
	})

	t.Run("GetCacheStats", func(t *testing.T) {
		cacheID := uuid.New()
		cache := &domain.Cache{ID: cacheID, ContainerID: "cid", Status: domain.CacheStatusRunning}
		repo.On("GetByID", mock.Anything, cacheID).Return(cache, nil).Once()
		
		statsJSON := `{"memory_stats": {"usage": 1024, "limit": 2048}}`
		compute.On("GetInstanceStats", mock.Anything, "cid").Return(io.NopCloser(strings.NewReader(statsJSON)), nil).Once()
		compute.On("Exec", mock.Anything, "cid", mock.Anything).Return("connected_clients:1\r\ndb0:keys=5,expires=0,avg_ttl=0", nil).Once()

		stats, err := svc.GetCacheStats(ctx, cacheID.String())
		require.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, int64(1024), stats.UsedMemoryBytes)
	})
}
