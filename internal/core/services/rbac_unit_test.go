package services_test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) CreateRole(ctx context.Context, role *domain.Role) error {
	return m.Called(ctx, role).Error(0)
}
func (m *MockRoleRepository) GetRoleByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	args := m.Called(ctx, id)
	r0, _ := args.Get(0).(*domain.Role)
	return r0, args.Error(1)
}
func (m *MockRoleRepository) GetRoleByName(ctx context.Context, name string) (*domain.Role, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).(*domain.Role)
	return r0, args.Error(1)
}
func (m *MockRoleRepository) ListRoles(ctx context.Context) ([]*domain.Role, error) {
	args := m.Called(ctx)
	r0, _ := args.Get(0).([]*domain.Role)
	return r0, args.Error(1)
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
		require.NoError(t, err)
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
		require.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("Authorize_Denied", func(t *testing.T) {
		user := &domain.User{ID: userID, Role: domain.RoleViewer, TenantID: tenantID}
		mockUserRepo.On("GetByID", mock.Anything, userID).Return(user, nil).Once()
		mockIAMRepo.On("GetPoliciesForUser", mock.Anything, tenantID, userID).Return([]*domain.Policy{}, nil).Once()
		mockRoleRepo.On("GetRoleByName", mock.Anything, domain.RoleViewer).Return(&domain.Role{
			Name: domain.RoleViewer,
			Permissions: []domain.Permission{domain.PermissionInstanceRead},
		}, nil).Once()

		err := svc.Authorize(ctx, userID, domain.PermissionInstanceLaunch, "*")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "permission denied")
	})

	t.Run("HasPermission_UserError", func(t *testing.T) {
		mockUserRepo.On("GetByID", mock.Anything, userID).Return(nil, fmt.Errorf("db fail")).Once()
		_, err := svc.HasPermission(ctx, userID, domain.PermissionInstanceRead, "*")
		require.Error(t, err)
	})

	t.Run("BindRole_Success", func(t *testing.T) {
		user := &domain.User{ID: userID, Email: "test@test.com", Role: domain.RoleViewer}
		mockRoleRepo.On("GetRoleByName", mock.Anything, domain.RoleAdmin).Return(&domain.Role{Name: domain.RoleAdmin}, nil).Once()
		mockUserRepo.On("GetByEmail", mock.Anything, "test@test.com").Return(user, nil).Once()
		mockUserRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
			return u.Role == domain.RoleAdmin
		})).Return(nil).Once()

		err := svc.BindRole(ctx, "test@test.com", domain.RoleAdmin)
		require.NoError(t, err)
	})
}
