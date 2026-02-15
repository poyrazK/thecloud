package services_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRBACService struct {
	mock.Mock
}

const (
	rbacPermPrefix     = "rbac:perm:"
	rbacRoleIDPrefix   = "rbac:role:id:"
	rbacRoleNamePrefix = "rbac:role:name:"
)

func rbacPermKey(userID uuid.UUID, permission domain.Permission, resource string) string {
	return rbacPermPrefix + userID.String() + ":" + string(permission) + ":" + resource
}

func rbacRoleIDKey(roleID uuid.UUID) string {
	return rbacRoleIDPrefix + roleID.String()
}

func rbacRoleNameKey(name string) string {
	return rbacRoleNamePrefix + name
}

func (m *mockRBACService) Authorize(ctx context.Context, userID uuid.UUID, permission domain.Permission, resource string) error {
	args := m.Called(ctx, userID, permission, resource)
	return args.Error(0)
}
func (m *mockRBACService) HasPermission(ctx context.Context, userID uuid.UUID, permission domain.Permission, resource string) (bool, error) {
	args := m.Called(ctx, userID, permission, resource)
	return args.Bool(0), args.Error(1)
}
func (m *mockRBACService) CreateRole(ctx context.Context, role *domain.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}
func (m *mockRBACService) GetRoleByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Role), args.Error(1)
}
func (m *mockRBACService) GetRoleByName(ctx context.Context, name string) (*domain.Role, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Role), args.Error(1)
}
func (m *mockRBACService) ListRoles(ctx context.Context) ([]*domain.Role, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Role), args.Error(1)
}
func (m *mockRBACService) UpdateRole(ctx context.Context, role *domain.Role) error {
	_ = "UpdateRole"
	args := m.Called(ctx, role)
	return args.Error(0)
}
func (m *mockRBACService) DeleteRole(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockRBACService) AddPermissionToRole(ctx context.Context, roleID uuid.UUID, permission domain.Permission) error {
	args := m.Called(ctx, roleID, permission)
	return args.Error(0)
}
func (m *mockRBACService) RemovePermissionFromRole(ctx context.Context, roleID uuid.UUID, permission domain.Permission) error {
	_ = "RemovePermissionFromRole"
	args := m.Called(ctx, roleID, permission)
	return args.Error(0)
}
func (m *mockRBACService) BindRole(ctx context.Context, userIdentifier string, roleName string) error {
	args := m.Called(ctx, userIdentifier, roleName)
	return args.Error(0)
}
func (m *mockRBACService) ListRoleBindings(ctx context.Context) ([]*domain.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.User), args.Error(1)
}

func (m *mockRBACService) EvaluatePolicy(ctx context.Context, userID uuid.UUID, action string, resource string, context map[string]interface{}) (bool, error) {
	args := m.Called(ctx, userID, action, resource, context)
	return args.Bool(0), args.Error(1)
}

func setupCachedRBACTest(t *testing.T) (*mockRBACService, *redis.Client, *miniredis.Miniredis) {
	t.Helper()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return new(mockRBACService), client, mr
}

func TestCachedRBACServiceListRoles(t *testing.T) {
	t.Parallel()
	mockSvc, cache, mr := setupCachedRBACTest(t)
	defer mr.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := services.NewCachedRBACService(mockSvc, cache, logger)

	roles := []*domain.Role{{ID: uuid.New(), Name: "admin"}}
	mockSvc.On("ListRoles", mock.Anything).Return(roles, nil).Once()

	res, err := svc.ListRoles(context.Background())
	assert.NoError(t, err)
	assert.Len(t, res, 1)
	mockSvc.AssertExpectations(t)
}

func TestCachedRBACServiceAuthorizeDelegates(t *testing.T) {
	t.Parallel()
	mockSvc, cache, mr := setupCachedRBACTest(t)
	defer mr.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := services.NewCachedRBACService(mockSvc, cache, logger)

	userID := uuid.New()
	mockSvc.On("Authorize", mock.Anything, userID, domain.PermissionInstanceRead, "*").Return(nil).Once()

	err := svc.Authorize(context.Background(), userID, domain.PermissionInstanceRead, "*")
	assert.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestCachedRBACServiceHasPermissionCacheHit(t *testing.T) {
	t.Parallel()
	mockSvc, cache, mr := setupCachedRBACTest(t)
	defer mr.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := services.NewCachedRBACService(mockSvc, cache, logger)

	ctx := context.Background()
	userID := uuid.New()
	key := rbacPermKey(userID, domain.PermissionInstanceRead, "*")
	cache.Set(ctx, key, "1", time.Minute)

	allowed, err := svc.HasPermission(ctx, userID, domain.PermissionInstanceRead, "*")
	assert.NoError(t, err)
	assert.True(t, allowed)
	mockSvc.AssertNotCalled(t, "HasPermission", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestCachedRBACServiceHasPermissionCacheMiss(t *testing.T) {
	t.Parallel()
	mockSvc, cache, mr := setupCachedRBACTest(t)
	defer mr.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := services.NewCachedRBACService(mockSvc, cache, logger)

	ctx := context.Background()
	userID := uuid.New()
	permission := domain.PermissionInstanceRead

	mockSvc.On("HasPermission", mock.Anything, userID, permission, "*").Return(true, nil).Once()

	allowed, err := svc.HasPermission(ctx, userID, permission, "*")
	assert.NoError(t, err)
	assert.True(t, allowed)

	key := rbacPermKey(userID, permission, "*")
	assert.Equal(t, "1", cache.Get(ctx, key).Val())
}

func TestCachedRBACServiceHasPermissionError(t *testing.T) {
	t.Parallel()
	mockSvc, cache, mr := setupCachedRBACTest(t)
	defer mr.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := services.NewCachedRBACService(mockSvc, cache, logger)

	ctx := context.Background()
	userID := uuid.New()
	permission := domain.PermissionInstanceRead

	mockSvc.On("HasPermission", mock.Anything, userID, permission, "*").Return(false, assert.AnError).Once()

	allowed, err := svc.HasPermission(ctx, userID, permission, "*")
	assert.Error(t, err)
	assert.False(t, allowed)

	key := "rbac:perm:" + userID.String() + ":" + string(permission)
	assert.False(t, cache.Exists(ctx, key).Val() > 0)
}

func TestCachedRBACServiceCreateRoleDelegates(t *testing.T) {
	t.Parallel()
	mockSvc, cache, mr := setupCachedRBACTest(t)
	defer mr.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := services.NewCachedRBACService(mockSvc, cache, logger)

	role := &domain.Role{ID: uuid.New(), Name: "dev"}
	mockSvc.On("CreateRole", mock.Anything, role).Return(nil).Once()

	err := svc.CreateRole(context.Background(), role)
	assert.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestCachedRBACServiceGetRoleByIDCaches(t *testing.T) {
	t.Parallel()
	mockSvc, cache, mr := setupCachedRBACTest(t)
	defer mr.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := services.NewCachedRBACService(mockSvc, cache, logger)

	roleID := uuid.New()
	role := &domain.Role{ID: roleID, Name: "dev"}
	mockSvc.On("GetRoleByID", mock.Anything, roleID).Return(role, nil).Once()

	res, err := svc.GetRoleByID(context.Background(), roleID)
	assert.NoError(t, err)
	assert.Equal(t, roleID, res.ID)

	key := rbacRoleIDKey(roleID)
	assert.True(t, cache.Exists(context.Background(), key).Val() > 0)
}

func TestCachedRBACServiceGetRoleByNameCaches(t *testing.T) {
	t.Parallel()
	mockSvc, cache, mr := setupCachedRBACTest(t)
	defer mr.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := services.NewCachedRBACService(mockSvc, cache, logger)

	role := &domain.Role{ID: uuid.New(), Name: "ops"}
	mockSvc.On("GetRoleByName", mock.Anything, "ops").Return(role, nil).Once()

	res, err := svc.GetRoleByName(context.Background(), "ops")
	assert.NoError(t, err)
	assert.Equal(t, "ops", res.Name)

	key := rbacRoleNameKey("ops")
	assert.True(t, cache.Exists(context.Background(), key).Val() > 0)
}

func TestCachedRBACServiceUpdateRoleInvalidatesCache(t *testing.T) {
	t.Parallel()
	mockSvc, cache, mr := setupCachedRBACTest(t)
	defer mr.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := services.NewCachedRBACService(mockSvc, cache, logger)

	roleID := uuid.New()
	role := &domain.Role{ID: roleID, Name: "dev"}

	ctx := context.Background()
	cache.Set(ctx, rbacRoleIDKey(roleID), "cached", time.Minute)
	cache.Set(ctx, rbacRoleNameKey(role.Name), "cached", time.Minute)

	mockSvc.On("UpdateRole", mock.Anything, role).Return(nil).Once()

	err := svc.UpdateRole(ctx, role)
	assert.NoError(t, err)

	existsID := cache.Exists(ctx, rbacRoleIDKey(roleID)).Val()
	existsName := cache.Exists(ctx, rbacRoleNameKey(role.Name)).Val()
	assert.Equal(t, int64(0), existsID)
	assert.Equal(t, int64(0), existsName)
}

func TestCachedRBACServiceDeleteRoleInvalidatesCache(t *testing.T) {
	t.Parallel()
	mockSvc, cache, mr := setupCachedRBACTest(t)
	defer mr.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := services.NewCachedRBACService(mockSvc, cache, logger)

	roleID := uuid.New()
	role := &domain.Role{ID: roleID, Name: "ops"}

	ctx := context.Background()
	cache.Set(ctx, rbacRoleIDKey(roleID), "cached", time.Minute)
	cache.Set(ctx, rbacRoleNameKey(role.Name), "cached", time.Minute)

	mockSvc.On("GetRoleByID", mock.Anything, roleID).Return(role, nil).Once()
	mockSvc.On("DeleteRole", mock.Anything, roleID).Return(nil).Once()

	err := svc.DeleteRole(ctx, roleID)
	assert.NoError(t, err)

	existsID := cache.Exists(ctx, rbacRoleIDKey(roleID)).Val()
	existsName := cache.Exists(ctx, rbacRoleNameKey(role.Name)).Val()
	assert.Equal(t, int64(0), existsID)
	assert.Equal(t, int64(0), existsName)
}

func TestCachedRBACServiceAddPermissionInvalidatesCache(t *testing.T) {
	t.Parallel()
	mockSvc, cache, mr := setupCachedRBACTest(t)
	defer mr.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := services.NewCachedRBACService(mockSvc, cache, logger)

	roleID := uuid.New()
	role := &domain.Role{ID: roleID, Name: "ops"}

	ctx := context.Background()
	cache.Set(ctx, rbacRoleIDKey(roleID), "cached", time.Minute)
	cache.Set(ctx, rbacRoleNameKey(role.Name), "cached", time.Minute)

	mockSvc.On("AddPermissionToRole", mock.Anything, roleID, domain.PermissionInstanceRead).Return(nil).Once()
	mockSvc.On("GetRoleByID", mock.Anything, roleID).Return(role, nil).Once()

	err := svc.AddPermissionToRole(ctx, roleID, domain.PermissionInstanceRead)
	assert.NoError(t, err)

	existsID := cache.Exists(ctx, rbacRoleIDKey(roleID)).Val()
	existsName := cache.Exists(ctx, rbacRoleNameKey(role.Name)).Val()
	assert.Equal(t, int64(0), existsID)
	assert.Equal(t, int64(0), existsName)
}

func TestCachedRBACServiceRemovePermissionInvalidatesCache(t *testing.T) {
	t.Parallel()
	mockSvc, cache, mr := setupCachedRBACTest(t)
	defer mr.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := services.NewCachedRBACService(mockSvc, cache, logger)

	roleID := uuid.New()
	role := &domain.Role{ID: roleID, Name: "ops"}

	ctx := context.Background()
	cache.Set(ctx, rbacRoleIDKey(roleID), "cached", time.Minute)
	cache.Set(ctx, rbacRoleNameKey(role.Name), "cached", time.Minute)

	mockSvc.On("RemovePermissionFromRole", mock.Anything, roleID, domain.PermissionInstanceRead).Return(nil).Once()
	mockSvc.On("GetRoleByID", mock.Anything, roleID).Return(role, nil).Once()

	err := svc.RemovePermissionFromRole(ctx, roleID, domain.PermissionInstanceRead)
	assert.NoError(t, err)

	existsID := cache.Exists(ctx, rbacRoleIDKey(roleID)).Val()
	existsName := cache.Exists(ctx, rbacRoleNameKey(role.Name)).Val()
	assert.Equal(t, int64(0), existsID)
	assert.Equal(t, int64(0), existsName)
}

func TestCachedRBACServiceBindRoleDelegates(t *testing.T) {
	t.Parallel()
	mockSvc, cache, mr := setupCachedRBACTest(t)
	defer mr.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := services.NewCachedRBACService(mockSvc, cache, logger)

	mockSvc.On("BindRole", mock.Anything, "user@example.com", "admin").Return(nil).Once()

	err := svc.BindRole(context.Background(), "user@example.com", "admin")
	assert.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestCachedRBACServiceListRoleBindings(t *testing.T) {
	t.Parallel()
	mockSvc, cache, mr := setupCachedRBACTest(t)
	defer mr.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := services.NewCachedRBACService(mockSvc, cache, logger)

	mockSvc.On("ListRoleBindings", mock.Anything).Return([]*domain.User{{ID: uuid.New()}}, nil).Once()

	users, err := svc.ListRoleBindings(context.Background())
	assert.NoError(t, err)
	assert.Len(t, users, 1)
	mockSvc.AssertExpectations(t)
}
