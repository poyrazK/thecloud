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

func setupRBACServiceIntegrationTest(t *testing.T) (ports.RBACService, ports.RoleRepository, ports.UserRepository, context.Context) {
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	userRepo := postgres.NewUserRepo(db)
	roleRepo := postgres.NewRBACRepository(db)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := services.NewRBACService(userRepo, roleRepo, logger)

	return svc, roleRepo, userRepo, ctx
}

func TestRBACService_Integration(t *testing.T) {
	svc, roleRepo, userRepo, ctx := setupRBACServiceIntegrationTest(t)

	t.Run("Authorize_Success", func(t *testing.T) {
		roleID := uuid.New()
		role := &domain.Role{
			ID:   roleID,
			Name: "developer",
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
			Role:  "developer",
		}
		err = userRepo.Create(ctx, user)
		require.NoError(t, err)

		err = svc.Authorize(ctx, userID, domain.PermissionInstanceLaunch)
		assert.NoError(t, err)
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
		err = svc.Authorize(ctx, userID, domain.PermissionInstanceLaunch)
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
		assert.NoError(t, err)

		fetched, err := svc.GetRoleByID(ctx, role.ID)
		assert.NoError(t, err)
		assert.Equal(t, role.Name, fetched.Name)
		assert.Contains(t, fetched.Permissions, domain.PermissionVpcRead)

		// Update
		role.Permissions = append(role.Permissions, domain.PermissionVpcCreate)
		err = svc.UpdateRole(ctx, role)
		assert.NoError(t, err)

		fetched, _ = svc.GetRoleByID(ctx, role.ID)
		assert.Len(t, fetched.Permissions, 2)

		// Delete
		err = svc.DeleteRole(ctx, role.ID)
		assert.NoError(t, err)

		_, err = svc.GetRoleByID(ctx, role.ID)
		assert.Error(t, err)
	})

	t.Run("BindRole", func(t *testing.T) {
		role := &domain.Role{ID: uuid.New(), Name: "manager"}
		_ = roleRepo.CreateRole(ctx, role)

		userID := uuid.New()
		user := &domain.User{ID: userID, Email: "manager@test.com", Role: "none"}
		_ = userRepo.Create(ctx, user)

		err := svc.BindRole(ctx, "manager@test.com", "manager")
		assert.NoError(t, err)

		updated, _ := userRepo.GetByID(ctx, userID)
		assert.Equal(t, "manager", updated.Role)
	})
}
