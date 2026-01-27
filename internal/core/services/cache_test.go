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
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	cacheContainerID = "cont-123"
	cacheName        = "test-cache"
	cacheDomainType  = "*domain.Cache"
)

// MockCacheRepo
type MockCacheRepo struct{ mock.Mock }

func (m *MockCacheRepo) Create(ctx context.Context, c *domain.Cache) error {
	return m.Called(ctx, c).Error(0)
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Cache), args.Error(1)
}
func (m *MockCacheRepo) Update(ctx context.Context, c *domain.Cache) error {
	return m.Called(ctx, c).Error(0)
}
func (m *MockCacheRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupCacheServiceTest(_ *testing.T) (*MockCacheRepo, *MockComputeBackend, *MockVpcRepo, *MockEventService, *MockAuditService, *services.CacheService) {
	repo := new(MockCacheRepo)
	docker := new(MockComputeBackend)
	vpcRepo := new(MockVpcRepo)
	eventSvc := new(MockEventService)
	auditSvc := new(MockAuditService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := services.NewCacheService(repo, docker, vpcRepo, eventSvc, auditSvc, logger)
	return repo, docker, vpcRepo, eventSvc, auditSvc, svc
}

func TestCreateCacheSuccess(t *testing.T) {
	repo, docker, _, eventSvc, auditSvc, svc := setupCacheServiceTest(t)
	defer repo.AssertExpectations(t)
	defer docker.AssertExpectations(t)
	defer eventSvc.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	ctx := appcontext.WithUserID(context.Background(), uuid.New())
	name := cacheName
	version := "7.2"
	memory := 128

	// Use strict matchers where possible, but Anything for generated IDs in names
	// Use strict matchers where possible, but Anything for generated IDs in names
	docker.On("CreateInstance", ctx, mock.MatchedBy(func(opts ports.CreateInstanceOptions) bool {
		return strings.HasPrefix(opts.Name, "thecloud-cache-") &&
			opts.ImageName == "redis:7.2-alpine" &&
			len(opts.Ports) == 1 && opts.Ports[0] == "0:6379"
	})).Return(cacheContainerID, nil)
	docker.On("GetInstancePort", ctx, cacheContainerID, "6379").Return(30000, nil)
	repo.On("Create", ctx, mock.AnythingOfType(cacheDomainType)).Return(nil)
	repo.On("Update", ctx, mock.AnythingOfType(cacheDomainType)).Return(nil)
	eventSvc.On("RecordEvent", ctx, "CACHE_CREATE", mock.Anything, "CACHE", mock.Anything).Return(nil)
	auditSvc.On("Log", ctx, mock.Anything, "cache.create", "cache", mock.Anything, mock.Anything).Return(nil)

	cache, err := svc.CreateCache(ctx, name, version, memory, nil)

	assert.NoError(t, err)
	assert.NotNil(t, cache)
	assert.Equal(t, name, cache.Name)
	assert.Equal(t, domain.EngineRedis, cache.Engine)
	assert.Equal(t, 30000, cache.Port)
	assert.Equal(t, cacheContainerID, cache.ContainerID)
	assert.NotEmpty(t, cache.Password)
	assert.Equal(t, 128, cache.MemoryMB)
}

func TestCreateCacheWithVpc(t *testing.T) {
	repo, docker, vpcRepo, eventSvc, auditSvc, svc := setupCacheServiceTest(t)
	defer repo.AssertExpectations(t)
	defer docker.AssertExpectations(t)
	defer vpcRepo.AssertExpectations(t)
	defer eventSvc.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	ctx := appcontext.WithUserID(context.Background(), uuid.New())
	name := cacheName
	version := "7.2"
	memory := 128
	vpcID := uuid.New()

	vpcRepo.On("GetByID", ctx, vpcID).Return(&domain.VPC{ID: vpcID, NetworkID: "net-1"}, nil)
	docker.On("CreateInstance", ctx, mock.MatchedBy(func(opts ports.CreateInstanceOptions) bool {
		return opts.NetworkID == "net-1" && opts.ImageName == "redis:7.2-alpine"
	})).Return(cacheContainerID, nil)
	docker.On("GetInstancePort", ctx, cacheContainerID, "6379").Return(30000, nil)
	repo.On("Create", ctx, mock.AnythingOfType(cacheDomainType)).Return(nil)
	repo.On("Update", ctx, mock.AnythingOfType(cacheDomainType)).Return(nil)
	eventSvc.On("RecordEvent", ctx, "CACHE_CREATE", mock.Anything, "CACHE", mock.Anything).Return(nil)
	auditSvc.On("Log", ctx, mock.Anything, "cache.create", "cache", mock.Anything, mock.Anything).Return(nil)

	cache, err := svc.CreateCache(ctx, name, version, memory, &vpcID)
	assert.NoError(t, err)
	assert.Equal(t, cacheContainerID, cache.ContainerID)
}

func TestCreateCacheVpcError(t *testing.T) {
	repo, _, vpcRepo, _, _, svc := setupCacheServiceTest(t)
	ctx := appcontext.WithUserID(context.Background(), uuid.New())
	vpcID := uuid.New()

	repo.On("Create", ctx, mock.AnythingOfType(cacheDomainType)).Return(nil)
	vpcRepo.On("GetByID", ctx, vpcID).Return(nil, assert.AnError)

	_, err := svc.CreateCache(ctx, cacheName, "7.2", 128, &vpcID)
	assert.Error(t, err)
}

func TestDeleteCacheSuccess(t *testing.T) {
	repo, docker, _, eventSvc, auditSvc, svc := setupCacheServiceTest(t)
	defer repo.AssertExpectations(t)
	defer docker.AssertExpectations(t)
	defer eventSvc.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	ctx := context.Background()
	cacheID := uuid.New()
	cache := &domain.Cache{
		ID:          cacheID,
		Name:        cacheName,
		ContainerID: cacheContainerID,
	}

	repo.On("GetByID", ctx, cacheID).Return(cache, nil)
	docker.On("StopInstance", ctx, cacheContainerID).Return(nil)
	docker.On("DeleteInstance", ctx, cacheContainerID).Return(nil)
	repo.On("Delete", ctx, cacheID).Return(nil)
	eventSvc.On("RecordEvent", ctx, "CACHE_DELETE", cacheID.String(), "CACHE", mock.Anything).Return(nil)
	auditSvc.On("Log", ctx, mock.Anything, "cache.delete", "cache", cacheID.String(), mock.Anything).Return(nil)

	err := svc.DeleteCache(ctx, cacheID.String())

	assert.NoError(t, err)
}

func TestFlushCacheSuccess(t *testing.T) {
	repo, docker, _, _, auditSvc, svc := setupCacheServiceTest(t)
	defer repo.AssertExpectations(t)
	defer docker.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	ctx := context.Background()
	cacheID := uuid.New()
	cache := &domain.Cache{
		ID:          cacheID,
		Status:      domain.CacheStatusRunning,
		ContainerID: cacheContainerID,
		Password:    "secret",
	}

	repo.On("GetByID", ctx, cacheID).Return(cache, nil)
	// Expect Exec call: redis-cli -a password FLUSHALL
	docker.On("Exec", ctx, cacheContainerID, []string{"redis-cli", "-a", "secret", "FLUSHALL"}).Return("OK", nil)
	auditSvc.On("Log", ctx, mock.Anything, "cache.flush", "cache", cacheID.String(), mock.Anything).Return(nil)

	err := svc.FlushCache(ctx, cacheID.String())

	assert.NoError(t, err)
}

func TestGetCacheStatsSuccess(t *testing.T) {
	repo, docker, _, _, _, svc := setupCacheServiceTest(t)
	defer repo.AssertExpectations(t)
	defer docker.AssertExpectations(t)

	ctx := context.Background()
	cacheID := uuid.New()
	containerID := cacheContainerID
	cache := &domain.Cache{
		ID:          cacheID,
		Status:      domain.CacheStatusRunning,
		ContainerID: containerID,
		Password:    "secret",
	}

	repo.On("GetByID", ctx, cacheID).Return(cache, nil)

	// Mock Docker Stats
	statsJSON := `{"memory_stats":{"usage":1024,"limit":2048}}`
	docker.On("GetInstanceStats", ctx, containerID).Return(io.NopCloser(strings.NewReader(statsJSON)), nil)

	// Mock Exec INFO
	infoOutput := "connected_clients:5\r\ndb0:keys=10,expires=0,avg_ttl=0\r\n"
	docker.On("Exec", ctx, containerID, []string{"redis-cli", "-a", "secret", "INFO"}).Return(infoOutput, nil)

	stats, err := svc.GetCacheStats(ctx, cacheID.String())

	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, int64(1024), stats.UsedMemoryBytes)
	assert.Equal(t, int64(2048), stats.MaxMemoryBytes)
	assert.Equal(t, 5, stats.ConnectedClients)
	assert.Equal(t, int64(10), stats.TotalKeys)
}

func TestCreateCacheDockerFailure(t *testing.T) {
	repo, docker, _, _, _, svc := setupCacheServiceTest(t)
	defer repo.AssertExpectations(t)
	defer docker.AssertExpectations(t)

	ctx := appcontext.WithUserID(context.Background(), uuid.New())
	name := "fail-cache"

	// Expect Repo Create
	repo.On("Create", ctx, mock.Anything).Return(nil)

	// Expect Docker Create to fail
	docker.On("CreateInstance", ctx, mock.Anything).Return("", assert.AnError)

	// Expect Rollback Delete
	repo.On("Delete", ctx, mock.Anything).Return(nil)

	cache, err := svc.CreateCache(ctx, name, "7.2", 128, nil)

	assert.Error(t, err)
	assert.Nil(t, cache)
}

func TestGetCacheByID(t *testing.T) {
	repo, _, _, _, _, svc := setupCacheServiceTest(t)
	defer repo.AssertExpectations(t)

	ctx := context.Background()
	cacheID := uuid.New()
	cache := &domain.Cache{ID: cacheID, Name: "my-cache"}

	repo.On("GetByID", ctx, cacheID).Return(cache, nil)

	result, err := svc.GetCache(ctx, cacheID.String())

	assert.NoError(t, err)
	assert.Equal(t, cacheID, result.ID)
}

func TestGetCacheByName(t *testing.T) {
	repo, _, _, _, _, svc := setupCacheServiceTest(t)
	defer repo.AssertExpectations(t)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	name := "named-cache"
	cache := &domain.Cache{ID: uuid.New(), Name: name}

	repo.On("GetByName", ctx, userID, name).Return(cache, nil)

	result, err := svc.GetCache(ctx, name)

	assert.NoError(t, err)
	assert.Equal(t, name, result.Name)
}

func TestListCaches(t *testing.T) {
	repo, _, _, _, _, svc := setupCacheServiceTest(t)
	defer repo.AssertExpectations(t)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)

	caches := []*domain.Cache{{Name: "cache1"}, {Name: "cache2"}}
	repo.On("List", ctx, userID).Return(caches, nil)

	result, err := svc.ListCaches(ctx)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestGetCacheConnectionString(t *testing.T) {
	repo, _, _, _, _, svc := setupCacheServiceTest(t)
	defer repo.AssertExpectations(t)

	ctx := context.Background()
	cacheID := uuid.New()
	cache := &domain.Cache{
		ID:       cacheID,
		Name:     "conn-cache",
		Port:     6379,
		Password: "secret",
	}

	repo.On("GetByID", ctx, cacheID).Return(cache, nil)

	connStr, err := svc.GetConnectionString(ctx, cacheID.String())

	assert.NoError(t, err)
	assert.Contains(t, connStr, "redis://")
	assert.Contains(t, connStr, "secret")
	assert.Contains(t, connStr, "6379")
}
