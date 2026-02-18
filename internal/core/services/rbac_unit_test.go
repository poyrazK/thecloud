package services_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) CreateRole(ctx context.Context, role *domain.Role) error {
	return m.Called(ctx, role).Error(0)
}
func (m *MockRoleRepository) GetRoleByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Role), args.Error(1)
}
func (m *MockRoleRepository) GetRoleByName(ctx context.Context, name string) (*domain.Role, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Role), args.Error(1)
}
func (m *MockRoleRepository) ListRoles(ctx context.Context) ([]*domain.Role, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Role), args.Error(1)
}
func (m *MockRoleRepository) UpdateRole(ctx context.Context, role *domain.Role) error {
	return m.Called(ctx, role).Error(0)
}
func (m *MockRoleRepository) DeleteRole(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockRoleRepository) AddPermissionToRole(ctx context.Context, roleID uuid.UUID, p domain.Permission) error {
	return m.Called(ctx, roleID, p).Error(0)
}
func (m *MockRoleRepository) RemovePermissionFromRole(ctx context.Context, roleID uuid.UUID, p domain.Permission) error {
	return m.Called(ctx, roleID, p).Error(0)
}
func (m *MockRoleRepository) GetPermissionsForRole(ctx context.Context, id uuid.UUID) ([]domain.Permission, error) {
	return nil, nil
}

type MockPolicyEvaluator struct {
	mock.Mock
}

func (m *MockPolicyEvaluator) Evaluate(ctx context.Context, policies []*domain.Policy, action, resource string, evalCtx map[string]interface{}) (bool, error) {
	args := m.Called(ctx, policies, action, resource, evalCtx)
	return args.Bool(0), args.Error(1)
}

func TestRBACService_Unit(t *testing.T) {
	mockUserRepo := new(MockUserRepo)
	mockRoleRepo := new(MockRoleRepository)
	mockIAMRepo := new(MockIAMRepository)
	mockEval := new(MockPolicyEvaluator)
	svc := services.NewRBACService(mockUserRepo, mockRoleRepo, mockIAMRepo, mockEval, slog.Default())

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()

	t.Run("HasPermission_AdminRole", func(t *testing.T) {
		user := &domain.User{ID: userID, Role: domain.RoleAdmin, TenantID: tenantID}
		mockUserRepo.On("GetByID", mock.Anything, userID).Return(user, nil).Once()
		mockIAMRepo.On("GetPoliciesForUser", mock.Anything, tenantID, userID).Return([]*domain.Policy{}, nil).Once()
		mockRoleRepo.On("GetRoleByName", mock.Anything, domain.RoleAdmin).Return(nil, assert.AnError).Once() // Fallback to hardcoded

		allowed, err := svc.HasPermission(ctx, userID, domain.PermissionInstanceRead, "*")
		assert.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("HasPermission_IAMPolicy", func(t *testing.T) {
		user := &domain.User{ID: userID, Role: domain.RoleViewer, TenantID: tenantID}
		policy := &domain.Policy{Name: "custom"}
		mockUserRepo.On("GetByID", mock.Anything, userID).Return(user, nil).Once()
		mockIAMRepo.On("GetPoliciesForUser", mock.Anything, tenantID, userID).Return([]*domain.Policy{policy}, nil).Once()
		mockEval.On("Evaluate", mock.Anything, mock.Anything, string(domain.PermissionInstanceLaunch), "*", mock.Anything).
			Return(true, nil).Once()

		allowed, err := svc.HasPermission(ctx, userID, domain.PermissionInstanceLaunch, "*")
		assert.NoError(t, err)
		assert.True(t, allowed)
	})
}
