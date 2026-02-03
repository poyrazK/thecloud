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
	db := SetupDB(t)
	defer db.Close()
	repo := NewPostgresGatewayRepository(db)

	CleanDB(t, db)
	ctx := SetupTestUser(t, db)
	userID := appcontext.UserIDFromContext(ctx)

	t.Run("CreateAndListRoutes", func(t *testing.T) {
		route := &domain.GatewayRoute{
			ID:          uuid.New(),
			UserID:      userID,
			Name:        "test-route",
			PathPrefix:  "/v1-test",
			PathPattern: "/v1-test/*",
			PatternType: "pattern",
			ParamNames:  []string{},
			TargetURL:   "http://test:80",
			Methods:     []string{"GET", "POST"},
			StripPrefix: true,
			RateLimit:   100,
			Priority:    5,
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
