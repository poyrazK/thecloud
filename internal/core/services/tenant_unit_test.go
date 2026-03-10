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

func TestTenantService_Unit(t *testing.T) {
	mockRepo := new(MockTenantRepo)
	mockUserRepo := new(MockUserRepo)
	svc := services.NewTenantService(mockRepo, mockUserRepo, slog.Default())

	ctx := context.Background()

	t.Run("CreateTenant_Success", func(t *testing.T) {
		ownerID := uuid.New()
		slug := "new-tenant"
		mockRepo.On("GetBySlug", mock.Anything, slug).Return(nil, nil).Once()
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		mockRepo.On("AddMember", mock.Anything, mock.Anything, ownerID, "owner").Return(nil).Once()
		mockRepo.On("UpdateQuota", mock.Anything, mock.Anything).Return(nil).Once()
		mockUserRepo.On("GetByID", mock.Anything, ownerID).Return(&domain.User{ID: ownerID}, nil).Once()
		mockUserRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
			return u.DefaultTenantID != nil
		})).Return(nil).Once()

		tenant, err := svc.CreateTenant(ctx, "New Tenant", slug, ownerID)
		require.NoError(t, err)
		assert.NotNil(t, tenant)
		assert.Equal(t, slug, tenant.Slug)
	})

	t.Run("CreateTenant_Conflict", func(t *testing.T) {
		slug := "existing"
		mockRepo.On("GetBySlug", mock.Anything, slug).Return(&domain.Tenant{Slug: slug}, nil).Once()

		_, err := svc.CreateTenant(ctx, "Name", slug, uuid.New())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("InviteMember_UserNotFound", func(t *testing.T) {
		email := "unknown@test.com"
		mockUserRepo.On("GetByEmail", mock.Anything, email).Return(nil, fmt.Errorf("not found")).Once()

		err := svc.InviteMember(ctx, uuid.New(), email, "member")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})

	t.Run("CheckQuota_Exceeded", func(t *testing.T) {
		tenantID := uuid.New()
		mockRepo.On("GetQuota", mock.Anything, tenantID).Return(&domain.TenantQuota{
			MaxInstances: 5,
			UsedInstances: 5,
		}, nil).Once()

		err := svc.CheckQuota(ctx, tenantID, "instances", 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "quota exceeded")
	})

	t.Run("CheckQuota_InvalidResource", func(t *testing.T) {
		tenantID := uuid.New()
		mockRepo.On("GetQuota", mock.Anything, tenantID).Return(&domain.TenantQuota{}, nil).Once()

		err := svc.CheckQuota(ctx, tenantID, "invalid", 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown resource type")
	})
}
