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

func TestPostgresContainerRepository(t *testing.T) {
	db, _ := SetupDB(t)
	defer db.Close()
	repo := NewPostgresContainerRepository(db)
	ctx := SetupTestUser(t, db)
	userID := appcontext.UserIDFromContext(ctx)

	t.Run("CreateAndGetDeployment", func(t *testing.T) {
		dep := &domain.Deployment{
			ID:           uuid.New(),
			UserID:       userID,
			Name:         "test-dep",
			Image:        "nginx",
			Replicas:     3,
			CurrentCount: 0,
			Status:       domain.DeploymentStatusScaling,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err := repo.CreateDeployment(ctx, dep)
		require.NoError(t, err)

		found, err := repo.GetDeploymentByID(ctx, dep.ID, userID)
		require.NoError(t, err)
		assert.Equal(t, dep.Name, found.Name)
	})

	t.Run("Containers", func(t *testing.T) {
		// Placeholder for container relationship tests
	})
}
