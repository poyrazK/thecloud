package services_test

import (
	"context"
	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTenantService_Unit(t *testing.T) {
	mockRepo := new(MockTenantRepo)
	mockUserRepo := new(MockUserRepo)
	rbacSvc := new(MockRBACService)
	svc := services.NewTenantService(services.TenantServiceParams{
		Repo: mockRepo, UserRepo: mockUserRepo, RBACSvc: rbacSvc,
	})

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	t.Run("CreateTenant_Success", func(t *testing.T) {
		rbacSvc.On("Authorize", mock.Anything, userID, tenantID, domain.PermissionTenantCreate, "*").Return(nil).Once()
		mockRepo.On("GetBySlug", mock.Anything, "new-tenant").Return(nil, nil).Once()
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		mockRepo.On("AddMember", mock.Anything, mock.Anything, mock.Anything, "owner").Return(nil).Once()
		mockRepo.On("UpdateQuota", mock.Anything, mock.Anything).Return(nil).Once()
		mockUserRepo.On("GetByID", mock.Anything, userID).Return(&domain.User{ID: userID}, nil).Once()
		mockUserRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()

		tenant, err := svc.CreateTenant(ctx, "New Tenant", "new-tenant", userID)
		require.NoError(t, err)
		assert.NotNil(t, tenant)
	})

	t.Run("InviteMember_UserNotFound", func(t *testing.T) {
		rbacSvc.On("Authorize", mock.Anything, userID, tenantID, domain.PermissionTenantUpdate, tenantID.String()).Return(nil).Once()
		mockUserRepo.On("GetByEmail", mock.Anything, "unknown@test.com").Return(nil, assert.AnError).Once()

		err := svc.InviteMember(ctx, tenantID, "unknown@test.com", "member")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})

	t.Run("CheckQuota_Exceeded", func(t *testing.T) {
		quota := &domain.TenantQuota{UsedInstances: 10, MaxInstances: 10}
		mockRepo.On("GetQuota", mock.Anything, tenantID).Return(quota, nil).Once()

		err := svc.CheckQuota(ctx, tenantID, "instances", 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "quota exceeded")
	})

	t.Run("CheckQuota_InvalidResource", func(t *testing.T) {
		mockRepo.On("GetQuota", mock.Anything, tenantID).Return(&domain.TenantQuota{}, nil).Once()

		err := svc.CheckQuota(ctx, tenantID, "invalid", 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown resource type")
	})

	t.Run("GetTenant", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, tenantID).Return(&domain.Tenant{ID: tenantID}, nil).Once()
		res, err := svc.GetTenant(ctx, tenantID)
		require.NoError(t, err)
		assert.Equal(t, tenantID, res.ID)
	})

	t.Run("ListUserTenants", func(t *testing.T) {
		mockRepo.On("ListUserTenants", mock.Anything, userID).Return([]domain.Tenant{}, nil).Once()
		res, err := svc.ListUserTenants(ctx, userID)
		require.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("RemoveMember_Success", func(t *testing.T) {
		targetUserID := uuid.New()
		rbacSvc.On("Authorize", mock.Anything, userID, tenantID, domain.PermissionTenantUpdate, tenantID.String()).Return(nil).Once()
		mockRepo.On("GetByID", mock.Anything, tenantID).Return(&domain.Tenant{ID: tenantID, OwnerID: userID}, nil).Once()
		mockRepo.On("RemoveMember", mock.Anything, tenantID, targetUserID).Return(nil).Once()

		err := svc.RemoveMember(ctx, tenantID, targetUserID)
		require.NoError(t, err)
	})

	t.Run("SwitchTenant_Success", func(t *testing.T) {
		mockRepo.On("GetMembership", mock.Anything, tenantID, userID).Return(&domain.TenantMember{}, nil).Once()
		mockUserRepo.On("GetByID", mock.Anything, userID).Return(&domain.User{ID: userID}, nil).Once()
		mockUserRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
			return *u.DefaultTenantID == tenantID
		})).Return(nil).Once()

		err := svc.SwitchTenant(ctx, userID, tenantID)
		require.NoError(t, err)
	})

	t.Run("Usage_Tracking", func(t *testing.T) {
		mockRepo.On("IncrementUsage", mock.Anything, tenantID, "vpcs", 1).Return(nil).Once()
		mockRepo.On("DecrementUsage", mock.Anything, tenantID, "vpcs", 1).Return(nil).Once()

		err := svc.IncrementUsage(ctx, tenantID, "vpcs", 1)
		require.NoError(t, err)

		err = svc.DecrementUsage(ctx, tenantID, "vpcs", 1)
		require.NoError(t, err)
	})
}
