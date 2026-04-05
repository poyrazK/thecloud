package services_test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockPolicyEvaluator struct {
	mock.Mock
}

func (m *MockPolicyEvaluator) Evaluate(ctx context.Context, policies []*domain.Policy, action string, resource string, evalCtx map[string]interface{}) (domain.PolicyEffect, error) {
	args := m.Called(ctx, policies, action, resource, evalCtx)
	return args.Get(0).(domain.PolicyEffect), args.Error(1)
}

func TestRBACService_Unit(t *testing.T) {
	mockUserRepo := new(MockUserRepo)
	mockRoleRepo := new(MockRoleRepository)
	mockTenantRepo := new(MockTenantRepo)
	mockIAMRepo := new(MockIAMRepository)
	mockEval := new(MockPolicyEvaluator)
	svc := services.NewRBACService(services.RBACServiceParams{
		UserRepo:   mockUserRepo,
		RoleRepo:   mockRoleRepo,
		TenantRepo: mockTenantRepo,
		IAMRepo:    mockIAMRepo,
		Evaluator:  mockEval,
		Logger:     slog.Default(),
	})

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()

	t.Run("HasPermission_AdminRole", func(t *testing.T) {
		mockTenantRepo.On("GetMembership", mock.Anything, tenantID, userID).Return(&domain.TenantMember{Role: domain.RoleAdmin}, nil).Once()
		mockIAMRepo.On("GetPoliciesForUser", mock.Anything, tenantID, userID).Return([]*domain.Policy{}, nil).Once()
		mockRoleRepo.On("GetRoleByName", mock.Anything, domain.RoleAdmin).Return(nil, errors.New(errors.NotFound, "not found")).Once() // Fallback to hardcoded

		allowed, err := svc.HasPermission(ctx, userID, tenantID, domain.PermissionInstanceRead, "*")
		require.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("HasPermission_IAMPolicy", func(t *testing.T) {
		policy := &domain.Policy{Name: "custom"}
		mockTenantRepo.On("GetMembership", mock.Anything, tenantID, userID).Return(&domain.TenantMember{Role: domain.RoleViewer}, nil).Once()
		mockIAMRepo.On("GetPoliciesForUser", mock.Anything, tenantID, userID).Return([]*domain.Policy{policy}, nil).Once()
		mockEval.On("Evaluate", mock.Anything, mock.Anything, string(domain.PermissionInstanceLaunch), "*", mock.Anything).
			Return(domain.EffectAllow, nil).Once()

		allowed, err := svc.HasPermission(ctx, userID, tenantID, domain.PermissionInstanceLaunch, "*")
		require.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("Authorize_Denied", func(t *testing.T) {
		mockTenantRepo.On("GetMembership", mock.Anything, tenantID, userID).Return(&domain.TenantMember{Role: domain.RoleViewer}, nil).Once()
		mockIAMRepo.On("GetPoliciesForUser", mock.Anything, tenantID, userID).Return([]*domain.Policy{}, nil).Once()
		mockRoleRepo.On("GetRoleByName", mock.Anything, domain.RoleViewer).Return(&domain.Role{
			Name:        domain.RoleViewer,
			Permissions: []domain.Permission{domain.PermissionInstanceRead},
		}, nil).Once()

		err := svc.Authorize(ctx, userID, tenantID, domain.PermissionInstanceLaunch, "*")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "permission denied")
	})

	t.Run("HasPermission_UserError", func(t *testing.T) {
		mockTenantRepo.On("GetMembership", mock.Anything, tenantID, userID).Return(nil, fmt.Errorf("db fail")).Once()
		_, err := svc.HasPermission(ctx, userID, tenantID, domain.PermissionInstanceRead, "*")
		require.Error(t, err)
	})

	t.Run("BindRole_Success", func(t *testing.T) {
		bindUserRepo := new(MockUserRepo)
		bindRoleRepo := new(MockRoleRepository)
		bindTenantRepo := new(MockTenantRepo)
		bindIAMRepo := new(MockIAMRepository)
		bindEval := new(MockPolicyEvaluator)
		bindSvc := services.NewRBACService(services.RBACServiceParams{
			UserRepo: bindUserRepo, RoleRepo: bindRoleRepo, TenantRepo: bindTenantRepo,
			IAMRepo: bindIAMRepo, Evaluator: bindEval, Logger: slog.Default(),
		})

		user := &domain.User{ID: userID, Email: "test@test.com", Role: domain.RoleViewer}
		bindRoleRepo.On("GetRoleByName", mock.Anything, domain.RoleAdmin).Return(&domain.Role{Name: domain.RoleAdmin}, nil).Once()
		bindUserRepo.On("GetByEmail", mock.Anything, "test@test.com").Return(user, nil).Once()
		bindUserRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
			return u.Role == domain.RoleAdmin
		})).Return(nil).Once()

		err := bindSvc.BindRole(ctx, "test@test.com", domain.RoleAdmin)
		require.NoError(t, err)
	})
}
