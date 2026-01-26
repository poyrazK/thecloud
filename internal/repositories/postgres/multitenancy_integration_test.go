//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMultiTenancy_TenantIsolation(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	// Clear tables in correct order
	ctx := context.Background()
	_, _ = db.Exec(ctx, "DELETE FROM instances")
	_, _ = db.Exec(ctx, "DELETE FROM vpcs")
	_, _ = db.Exec(ctx, "DELETE FROM volumes")
	_, _ = db.Exec(ctx, "DELETE FROM tenant_members")
	_, _ = db.Exec(ctx, "UPDATE users SET default_tenant_id = NULL")
	_, _ = db.Exec(ctx, "DELETE FROM tenants")
	_, _ = db.Exec(ctx, "DELETE FROM users")

	// Create User IDs
	user1 := uuid.New()
	user2 := uuid.New()
	user3 := uuid.New()

	// Repositories
	userRepo := NewUserRepo(db)
	tenantRepo := NewTenantRepo(db)
	vpcRepo := NewVpcRepository(db)
	instRepo := NewInstanceRepository(db)

	// 1. Create Users
	users := []*domain.User{
		{ID: user1, Email: "user1@a.com", Name: "U1", Role: "user", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: user2, Email: "user2@a.com", Name: "U2", Role: "user", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: user3, Email: "user3@b.com", Name: "U3", Role: "user", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	for _, u := range users {
		require.NoError(t, userRepo.Create(ctx, u))
	}

	// 2. Create Tenants
	tenantA := uuid.New()
	tenantB := uuid.New()

	require.NoError(t, tenantRepo.Create(ctx, &domain.Tenant{ID: tenantA, Name: "Tenant A", Slug: "tenant-a", OwnerID: user1, CreatedAt: time.Now(), UpdatedAt: time.Now(), Plan: "free", Status: "active"}))
	require.NoError(t, tenantRepo.Create(ctx, &domain.Tenant{ID: tenantB, Name: "Tenant B", Slug: "tenant-b", OwnerID: user3, CreatedAt: time.Now(), UpdatedAt: time.Now(), Plan: "free", Status: "active"}))

	// 3. Setup Memberships
	require.NoError(t, tenantRepo.AddMember(ctx, tenantA, user1, "owner"))
	require.NoError(t, tenantRepo.AddMember(ctx, tenantA, user2, "member"))
	require.NoError(t, tenantRepo.AddMember(ctx, tenantB, user3, "owner"))

	// Contexts
	ctxA1 := appcontext.WithTenantID(appcontext.WithUserID(ctx, user1), tenantA)
	ctxA2 := appcontext.WithTenantID(appcontext.WithUserID(ctx, user2), tenantA)
	ctxB3 := appcontext.WithTenantID(appcontext.WithUserID(ctx, user3), tenantB)

	t.Run("Resource Visibility Within Tenant", func(t *testing.T) {
		vpc := &domain.VPC{
			ID:        uuid.New(),
			UserID:    user1,
			TenantID:  tenantA,
			Name:      "shared-vpc",
			CreatedAt: time.Now(),
		}
		require.NoError(t, vpcRepo.Create(ctxA1, vpc))

		// User 1 (Owner) can see it
		fetched1, err := vpcRepo.GetByID(ctxA1, vpc.ID)
		require.NoError(t, err)
		assert.Equal(t, vpc.ID, fetched1.ID)

		// User 2 (Member of same tenant) can also see it
		fetched2, err := vpcRepo.GetByID(ctxA2, vpc.ID)
		require.NoError(t, err)
		assert.Equal(t, vpc.ID, fetched2.ID)

		// User 3 (Different tenant) cannot see it
		_, err = vpcRepo.GetByID(ctxB3, vpc.ID)
		assert.Error(t, err)
	})

	t.Run("Instance Isolation Between Tenants", func(t *testing.T) {
		instA := &domain.Instance{
			ID:        uuid.New(),
			UserID:    user1,
			TenantID:  tenantA,
			Name:      "tenant-a-inst",
			Image:     "alpine",
			Status:    domain.StatusRunning,
			CreatedAt: time.Now(),
		}
		require.NoError(t, instRepo.Create(ctxA1, instA))

		instB := &domain.Instance{
			ID:        uuid.New(),
			UserID:    user3,
			TenantID:  tenantB,
			Name:      "tenant-b-inst",
			Image:     "ubuntu",
			Status:    domain.StatusRunning,
			CreatedAt: time.Now(),
		}
		require.NoError(t, instRepo.Create(ctxB3, instB))

		// Tenant A list
		listA, err := instRepo.List(ctxA1)
		require.NoError(t, err)
		assert.Len(t, listA, 1)
		assert.Equal(t, instA.ID, listA[0].ID)

		// Tenant B list
		listB, err := instRepo.List(ctxB3)
		require.NoError(t, err)
		assert.Len(t, listB, 1)
		assert.Equal(t, instB.ID, listB[0].ID)
	})
}
