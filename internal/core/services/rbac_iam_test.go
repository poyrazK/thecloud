package services

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRBACService_IAMIntegration(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	roleRepo := new(mocks.RoleRepository)
	tenantRepo := new(mocks.TenantRepository)
	iamRepo := new(mocks.IAMRepository)
	evaluator := NewIAMEvaluator()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := NewRBACService(RBACServiceParams{
		UserRepo:   userRepo,
		RoleRepo:   roleRepo,
		TenantRepo: tenantRepo,
		IAMRepo:    iamRepo,
		Evaluator:  evaluator,
		Logger:     logger,
	})
	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()

	t.Run("AllowByPolicy", func(t *testing.T) {
		tenantRepo.On("GetMembership", ctx, tenantID, userID).Return(&domain.TenantMember{UserID: userID, TenantID: tenantID, Role: "viewer"}, nil).Once()
		roleRepo.On("GetRoleByName", mock.Anything, "viewer").Return(&domain.Role{Name: "viewer"}, nil).Once()
		policies := []*domain.Policy{
			{
				Statements: []domain.Statement{
					{Effect: domain.EffectAllow, Action: []string{"instance:launch"}, Resource: []string{"*"}},
				},
			},
		}
		iamRepo.On("GetPoliciesForUser", ctx, tenantID, userID).Return(policies, nil).Once()

		allowed, err := svc.HasPermission(ctx, userID, tenantID, domain.PermissionInstanceLaunch, "*")
		require.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("DenyByPolicyOverridesRole", func(t *testing.T) {
		tenantRepo.On("GetMembership", ctx, tenantID, userID).Return(&domain.TenantMember{UserID: userID, TenantID: tenantID, Role: "admin"}, nil).Once()
		roleRepo.On("GetRoleByName", mock.Anything, "admin").Return(&domain.Role{Name: "admin"}, nil).Once()
		policies := []*domain.Policy{
			{
				Statements: []domain.Statement{
					{Effect: domain.EffectDeny, Action: []string{"instance:terminate"}, Resource: []string{"*"}},
				},
			},
		}
		iamRepo.On("GetPoliciesForUser", ctx, tenantID, userID).Return(policies, nil).Once()

		// Admin would normally have this, but policy Deny should stop it
		allowed, err := svc.HasPermission(ctx, userID, tenantID, domain.PermissionInstanceTerminate, "*")
		require.NoError(t, err)
		assert.False(t, allowed)
	})

	t.Run("FallbackToRole", func(t *testing.T) {
		// Use a custom role name like "custom-dev" so it doesn't match defaultRoleAdmin/Viewer fallbacks
		tenantRepo.On("GetMembership", ctx, tenantID, userID).Return(&domain.TenantMember{UserID: userID, TenantID: tenantID, Role: "custom-dev"}, nil).Once()
		iamRepo.On("GetPoliciesForUser", ctx, tenantID, userID).Return([]*domain.Policy{}, nil).Once()
		roleRepo.On("GetRoleByName", ctx, "custom-dev").Return(&domain.Role{ID: uuid.New(), Name: "custom-dev", Permissions: []domain.Permission{domain.PermissionInstanceLaunch}}, nil).Once()

		// Fallback logic for custom role
		allowed, err := svc.HasPermission(ctx, userID, tenantID, domain.PermissionInstanceLaunch, "*")
		require.NoError(t, err)
		assert.True(t, allowed)
	})
}
