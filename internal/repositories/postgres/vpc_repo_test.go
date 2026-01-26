//go:build integration

package postgres

import (
	"testing"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVpcRepository_Integration(t *testing.T) {
	db := setupDB(t)
	defer db.Close()
	repo := NewVpcRepository(db)
	// Cleanup
	cleanDB(t, db)

	ctx := setupTestUser(t, db)
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	vpcID := uuid.New()
	vpc := &domain.VPC{
		ID:        vpcID,
		UserID:    userID,
		TenantID:  tenantID,
		Name:      "test-vpc",
		NetworkID: "net-123",
		CreatedAt: time.Now(),
	}

	t.Run("Create and Get", func(t *testing.T) {
		err := repo.Create(ctx, vpc)
		require.NoError(t, err)

		fetched, err := repo.GetByID(ctx, vpcID)
		require.NoError(t, err)
		assert.Equal(t, vpc.Name, fetched.Name)
		assert.Equal(t, vpc.NetworkID, fetched.NetworkID)
	})

	t.Run("GetByName", func(t *testing.T) {
		fetched, err := repo.GetByName(ctx, "test-vpc")
		require.NoError(t, err)
		assert.Equal(t, vpcID, fetched.ID)
	})

	t.Run("List", func(t *testing.T) {
		list, err := repo.List(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, list)
	})

	t.Run("Delete", func(t *testing.T) {
		err := repo.Delete(ctx, vpcID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, vpcID)
		assert.Error(t, err)
	})
}
