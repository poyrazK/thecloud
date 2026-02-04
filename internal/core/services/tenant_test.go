package services_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTenantServiceIntegrationTest(t *testing.T) (ports.TenantService, ports.TenantRepository, ports.UserRepository, context.Context) {
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	tenantRepo := postgres.NewTenantRepo(db)
	userRepo := postgres.NewUserRepo(db)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := services.NewTenantService(tenantRepo, userRepo, logger)

	return svc, tenantRepo, userRepo, ctx
}

func TestTenantService_Integration(t *testing.T) {
	svc, tenantRepo, userRepo, ctx := setupTenantServiceIntegrationTest(t)

	t.Run("CreateTenant", func(t *testing.T) {
		ownerID := uuid.New()
		err := userRepo.Create(ctx, &domain.User{ID: ownerID, Email: "owner@test.com"})
		require.NoError(t, err)

		name := "Integration Tenant"
		slug := "integration-tenant"

		tenant, err := svc.CreateTenant(ctx, name, slug, ownerID)
		assert.NoError(t, err)
		assert.NotNil(t, tenant)
		assert.Equal(t, name, tenant.Name)
		assert.Equal(t, slug, tenant.Slug)

		// Verify membership
		members, err := tenantRepo.ListMembers(ctx, tenant.ID)
		assert.NoError(t, err)
		assert.Len(t, members, 1)
		assert.Equal(t, ownerID, members[0].UserID)
		assert.Equal(t, "owner", members[0].Role)

		// Verify user default tenant was set
		updatedUser, _ := userRepo.GetByID(ctx, ownerID)
		assert.NotNil(t, updatedUser.DefaultTenantID)
		assert.Equal(t, tenant.ID, *updatedUser.DefaultTenantID)
	})

	t.Run("InviteAndSwitch", func(t *testing.T) {
		ownerID := uuid.New()
		_ = userRepo.Create(ctx, &domain.User{ID: ownerID, Email: "owner2@test.com"})
		tenant, _ := svc.CreateTenant(ctx, "Switch Test", "switch-test", ownerID)

		inviteeID := uuid.New()
		inviteeEmail := "invitee@test.com"
		_ = userRepo.Create(ctx, &domain.User{ID: inviteeID, Email: inviteeEmail})

		// Invite
		err := svc.InviteMember(ctx, tenant.ID, inviteeEmail, "member")
		assert.NoError(t, err)

		// Switch
		err = svc.SwitchTenant(ctx, inviteeID, tenant.ID)
		assert.NoError(t, err)

		updatedInvitee, _ := userRepo.GetByID(ctx, inviteeID)
		assert.NotNil(t, updatedInvitee.DefaultTenantID)
		assert.Equal(t, tenant.ID, *updatedInvitee.DefaultTenantID)
	})

	t.Run("Quota", func(t *testing.T) {
		ownerID := uuid.New()
		_ = userRepo.Create(ctx, &domain.User{ID: ownerID, Email: "quota@test.com"})
		tenant, _ := svc.CreateTenant(ctx, "Quota Test", "quota-test", ownerID)

		// Default quota should be there or we create it
		quota := &domain.TenantQuota{
			TenantID:     tenant.ID,
			MaxInstances: 5,
		}
		err := tenantRepo.UpdateQuota(ctx, quota)
		assert.NoError(t, err)

		// Check within limit
		err = svc.CheckQuota(ctx, tenant.ID, "instances", 1)
		assert.NoError(t, err)

		// Check over limit (used instances is 0 in DB)
		err = svc.CheckQuota(ctx, tenant.ID, "instances", 10)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "quota exceeded")
	})

	t.Run("RemoveMember", func(t *testing.T) {
		ownerID := uuid.New()
		_ = userRepo.Create(ctx, &domain.User{ID: ownerID, Email: "remove@test.com"})
		tenant, _ := svc.CreateTenant(ctx, "Remove Test", "remove-test", ownerID)

		memberID := uuid.New()
		_ = userRepo.Create(ctx, &domain.User{ID: memberID, Email: "mem@test.com"})
		_ = tenantRepo.AddMember(ctx, tenant.ID, memberID, "member")

		// Remove member
		err := svc.RemoveMember(ctx, tenant.ID, memberID)
		assert.NoError(t, err)

		// Verify gone
		mem, _ := tenantRepo.GetMembership(ctx, tenant.ID, memberID)
		assert.Nil(t, mem)

		// Cannot remove owner
		err = svc.RemoveMember(ctx, tenant.ID, ownerID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot remove tenant owner")
	})
}
