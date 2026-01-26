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

func TestPostgresGatewayRepository(t *testing.T) {
	db := setupDB(t)
	defer db.Close()
	repo := NewPostgresGatewayRepository(db)

	cleanDB(t, db)
	ctx := setupTestUser(t, db)
	userID := appcontext.UserIDFromContext(ctx)

	t.Run("CreateAndListRoutes", func(t *testing.T) {
		route := &domain.GatewayRoute{
			ID:          uuid.New(),
			UserID:      userID,
			Name:        "test-route",
			PathPrefix:  "/v1-test",
			TargetURL:   "http://test:80",
			StripPrefix: true,
			RateLimit:   100,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := repo.CreateRoute(ctx, route)
		require.NoError(t, err)

		routes, err := repo.ListRoutes(ctx, userID)
		require.NoError(t, err)
		assert.NotEmpty(t, routes)
	})
}
