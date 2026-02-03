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

func TestDatabaseRepository_Integration(t *testing.T) {
	dbPool := SetupDB(t)
	defer dbPool.Close()
	repo := NewDatabaseRepository(dbPool)
	ctx := SetupTestUser(t, dbPool)
	userID := appcontext.UserIDFromContext(ctx)

	// Cleanup before test
	_, err := dbPool.Exec(context.Background(), "DELETE FROM databases")
	require.NoError(t, err)

	t.Run("Create and Get", func(t *testing.T) {
		id := uuid.New()
		db := &domain.Database{
			ID:        id,
			UserID:    userID,
			Name:      "test-db",
			Engine:    domain.EnginePostgres,
			Version:   "16",
			Status:    domain.DatabaseStatusCreating,
			Username:  "admin",
			Password:  "pass",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := repo.Create(ctx, db)
		require.NoError(t, err)

		fetched, err := repo.GetByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, db.Name, fetched.Name)
		assert.Equal(t, db.Engine, fetched.Engine)
	})

	t.Run("List", func(t *testing.T) {
		list, err := repo.List(ctx)
		require.NoError(t, err)
		assert.Len(t, list, 1)
	})

	t.Run("Update", func(t *testing.T) {
		list, _ := repo.List(ctx)
		db := list[0]

		db.Status = domain.DatabaseStatusRunning
		db.Port = 54321
		db.ContainerID = "cont-123"

		err := repo.Update(ctx, db)
		require.NoError(t, err)

		fetched, err := repo.GetByID(ctx, db.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.DatabaseStatusRunning, fetched.Status)
		assert.Equal(t, 54321, fetched.Port)
		assert.Equal(t, "cont-123", fetched.ContainerID)
	})

	t.Run("Delete", func(t *testing.T) {
		list, _ := repo.List(ctx)
		db := list[0]

		err := repo.Delete(ctx, db.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, db.ID)
		assert.Error(t, err)
	})
}
