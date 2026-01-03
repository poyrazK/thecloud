package services_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCacheRepo
type MockCacheRepo struct{ mock.Mock }

func (m *MockCacheRepo) Create(ctx context.Context, c *domain.Cache) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}
func (m *MockCacheRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Cache, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Cache), args.Error(1)
}
func (m *MockCacheRepo) GetByName(ctx context.Context, userID uuid.UUID, name string) (*domain.Cache, error) {
	args := m.Called(ctx, userID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Cache), args.Error(1)
}
func (m *MockCacheRepo) List(ctx context.Context, userID uuid.UUID) ([]*domain.Cache, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*domain.Cache), args.Error(1)
}
func (m *MockCacheRepo) Update(ctx context.Context, c *domain.Cache) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}
func (m *MockCacheRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestCreateCache_Success(t *testing.T) {
	repo := new(MockCacheRepo)
	docker := new(MockDockerClient)
	vpcRepo := new(MockVpcRepo)
	eventSvc := new(MockEventService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := services.NewCacheService(repo, docker, vpcRepo, eventSvc, logger)

	ctx := appcontext.WithUserID(context.Background(), uuid.New())
	name := "test-cache"
	version := "7.2"
	memory := 128

	docker.On("CreateContainer", ctx, mock.Anything, "redis:7.2-alpine", []string{"0:6379"}, "", []string(nil), mock.Anything, mock.Anything).Return("cont-123", nil)
	docker.On("GetContainerPort", ctx, "cont-123", "6379").Return(30000, nil)
	repo.On("Create", ctx, mock.AnythingOfType("*domain.Cache")).Return(nil)
	repo.On("Update", ctx, mock.AnythingOfType("*domain.Cache")).Return(nil)
	eventSvc.On("RecordEvent", ctx, "CACHE_CREATE", mock.Anything, "CACHE", mock.Anything).Return(nil)

	cache, err := svc.CreateCache(ctx, name, version, memory, nil)

	assert.NoError(t, err)
	assert.NotNil(t, cache)
	assert.Equal(t, name, cache.Name)
	assert.Equal(t, domain.EngineRedis, cache.Engine)
	assert.Equal(t, 30000, cache.Port)
	assert.Equal(t, "cont-123", cache.ContainerID)
	assert.NotEmpty(t, cache.Password)
	assert.Equal(t, 128, cache.MemoryMB)

	repo.AssertExpectations(t)
	docker.AssertExpectations(t)
}

func TestDeleteCache_Success(t *testing.T) {
	repo := new(MockCacheRepo)
	docker := new(MockDockerClient)
	vpcRepo := new(MockVpcRepo)
	eventSvc := new(MockEventService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := services.NewCacheService(repo, docker, vpcRepo, eventSvc, logger)

	ctx := context.Background()
	cacheID := uuid.New()
	cache := &domain.Cache{
		ID:          cacheID,
		ContainerID: "cont-123",
	}

	repo.On("GetByID", ctx, cacheID).Return(cache, nil)
	docker.On("StopContainer", ctx, "cont-123").Return(nil)
	docker.On("RemoveContainer", ctx, "cont-123").Return(nil)
	repo.On("Delete", ctx, cacheID).Return(nil)
	eventSvc.On("RecordEvent", ctx, "CACHE_DELETE", cacheID.String(), "CACHE", mock.Anything).Return(nil)

	err := svc.DeleteCache(ctx, cacheID.String())

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	docker.AssertExpectations(t)
}

func TestFlushCache_Success(t *testing.T) {
	repo := new(MockCacheRepo)
	docker := new(MockDockerClient)
	vpcRepo := new(MockVpcRepo)
	eventSvc := new(MockEventService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := services.NewCacheService(repo, docker, vpcRepo, eventSvc, logger)

	ctx := context.Background()
	cacheID := uuid.New()
	cache := &domain.Cache{
		ID:          cacheID,
		Status:      domain.CacheStatusRunning,
		ContainerID: "cont-123",
		Password:    "secret",
	}

	repo.On("GetByID", ctx, cacheID).Return(cache, nil)
	// Expect Exec call: redis-cli -a password FLUSHALL
	docker.On("Exec", ctx, "cont-123", []string{"redis-cli", "-a", "secret", "FLUSHALL"}).Return("OK", nil)

	err := svc.FlushCache(ctx, cacheID.String())

	assert.NoError(t, err)
	docker.AssertExpectations(t)
}
