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
	"github.com/stretchr/testify/require"
)

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
		mockRoleRepo.On("GetRoleByName", mock.Anything, domain.RoleAdmin).Return(nil, assert.AnError).Once() // Fallback to hardcoded

		allowed, err := svc.HasPermission(ctx, userID, tenantID, domain.PermissionInstanceRead, "*")
		require.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("HasPermission_IAMPolicy", func(t *testing.T) {
		policy := &domain.Policy{Name: "custom"}
		mockTenantRepo.On("GetMembership", mock.Anything, tenantID, userID).Return(&domain.TenantMember{Role: domain.RoleViewer}, nil).Once()
		mockIAMRepo.On("GetPoliciesForUser", mock.Anything, tenantID, userID).Return([]*domain.Policy{policy}, nil).Once()
		mockEval.On("Evaluate", mock.Anything, mock.Anything, string(domain.PermissionInstanceLaunch), "*", mock.Anything).
			Return(true, nil).Once()

		allowed, err := svc.HasPermission(ctx, userID, tenantID, domain.PermissionInstanceLaunch, "*")
		require.NoError(t, err)
		assert.True(t, allowed)
	})
}
