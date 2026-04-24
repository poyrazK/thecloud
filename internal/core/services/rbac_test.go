//go:build integration
// +build integration

package services_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRBACServiceIntegrationTest(t *testing.T) (ports.RBACService, ports.RoleRepository, ports.UserRepository, ports.TenantRepository, context.Context) {
	t.Helper()
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	userRepo := postgres.NewUserRepo(db)
	roleRepo := postgres.NewRBACRepository(db)
	tenantRepo := postgres.NewTenantRepo(db)
	iamRepo := postgres.NewIAMRepository(db)
	evaluator := services.NewIAMEvaluator()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := services.NewRBACService(services.RBACServiceParams{
		UserRepo:   userRepo,
		RoleRepo:   roleRepo,
		TenantRepo: tenantRepo,
		IAMRepo:    iamRepo,
		Evaluator:  evaluator,
		Logger:     logger,
	})

	return svc, roleRepo, userRepo, tenantRepo, ctx
}

func TestRBACServiceIntegration(t *testing.T) {
	svc, roleRepo, userRepo, tenantRepo, ctx := setupRBACServiceIntegrationTest(t)

	t.Run("Authorize_Global_Success", func(t *testing.T) {
		roleID := uuid.New()
		role := &domain.Role{
			ID:   roleID,
			Name: "test-developer",
			Permissions: []domain.Permission{
				domain.PermissionInstanceLaunch,
			},
		}
		err := roleRepo.CreateRole(ctx, role)
		require.NoError(t, err)

		// Create user with this role
		userID := uuid.New()
		user := &domain.User{
			ID:    userID,
			Email: "dev@test.com",
			Role:  "test-developer",
		}
		err = userRepo.Create(ctx, user)
		require.NoError(t, err)

		err = svc.Authorize(ctx, userID, uuid.Nil, domain.PermissionInstanceLaunch, "*")
		assert.NoError(t, err)
	})

	t.Run("Authorize_Tenant_Success", func(t *testing.T) {
		// 1. Setup Tenant and Member Role
		tenantID := uuid.New()
		userID := uuid.New()
		roleName := "tenant-admin"

		role := &domain.Role{
			ID:   uuid.New(),
			Name: roleName,
			Permissions: []domain.Permission{
				domain.PermissionVpcCreate,
			},
		}
		require.NoError(t, roleRepo.CreateRole(ctx, role))

		user := &domain.User{
			ID:    userID,
			Email: "tenant@test.com",
			Role:  "viewer", // Global role is viewer
		}
		require.NoError(t, userRepo.Create(ctx, user))

		tenant := &domain.Tenant{
			ID:      tenantID,
			Name:    "Test Tenant",
			OwnerID: userID,
		}
		require.NoError(t, tenantRepo.Create(ctx, tenant))

		require.NoError(t, tenantRepo.AddMember(ctx, tenantID, userID, roleName))

		// 2. Authorize in Tenant Context
		err := svc.Authorize(ctx, userID, tenantID, domain.PermissionVpcCreate, "*")
		require.NoError(t, err)

		// 3. Verify Global Context (viewer) still Denied for VpcCreate
		err = svc.Authorize(ctx, userID, uuid.Nil, domain.PermissionVpcCreate, "*")
		require.Error(t, err)
	})

	t.Run("Authorize_Denied", func(t *testing.T) {
		userID := uuid.New()
		user := &domain.User{
			ID:    userID,
			Email: "viewer@test.com",
			Role:  "viewer",
		}
		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Viewer doesn't have launch permission by default even if role not in DB (fallback might allow read but not launch)
		err = svc.Authorize(ctx, userID, uuid.Nil, domain.PermissionInstanceLaunch, "*")
		assert.Error(t, err)
	})

	t.Run("Roles_CRUD", func(t *testing.T) {
		role := &domain.Role{
			ID:   uuid.New(),
			Name: "test-crud-role",
			Permissions: []domain.Permission{
				domain.PermissionVpcRead,
			},
		}

		err := svc.CreateRole(ctx, role)
		require.NoError(t, err)

		fetched, err := svc.GetRoleByID(ctx, role.ID)
		require.NoError(t, err)
		assert.Equal(t, role.Name, fetched.Name)
		assert.Contains(t, fetched.Permissions, domain.PermissionVpcRead)

		// Update
		role.Permissions = append(role.Permissions, domain.PermissionVpcCreate)
		err = svc.UpdateRole(ctx, role)
		require.NoError(t, err)

		fetched, _ = svc.GetRoleByID(ctx, role.ID)
		assert.Len(t, fetched.Permissions, 2)

		// Delete
		err = svc.DeleteRole(ctx, role.ID)
		require.NoError(t, err)

		_, err = svc.GetRoleByID(ctx, role.ID)
		require.Error(t, err)
	})

	t.Run("BindRole", func(t *testing.T) {
		role := &domain.Role{ID: uuid.New(), Name: "manager"}
		_ = roleRepo.CreateRole(ctx, role)

		userID := uuid.New()
		tenantID := appcontext.TenantIDFromContext(ctx); user := &domain.User{ID: userID, TenantID: tenantID, Email: "manager@test.com", Role: "none"}
		_ = userRepo.Create(ctx, user)

		err := svc.BindRole(ctx, "manager@test.com", "manager")
		require.NoError(t, err)

		updated, _ := userRepo.GetByID(ctx, userID)
		assert.Equal(t, "manager", updated.Role)
	})
}
