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

// TestGlobalLBRepositoryIntegration executes a comprehensive suite of integration tests
// against a live PostgreSQL backend to verify the GlobalLB repository implementation.
func TestGlobalLBRepositoryIntegration(t *testing.T) {
	db := SetupDB(t)
	defer db.Close()
	repo := NewGlobalLBRepository(db)
	ctx := SetupTestUser(t, db)
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	// Ensure a clean state by removing existing records.
	_, err := db.Exec(context.Background(), "DELETE FROM global_load_balancers")
	require.NoError(t, err)

	glbID := uuid.New()
	hostname := "integration.global.com"

	t.Run("Create and Get", func(t *testing.T) {
		glb := &domain.GlobalLoadBalancer{
			ID:        glbID,
			UserID:    userID,
			TenantID:  tenantID,
			Name:      "integration-glb",
			Hostname:  hostname,
			Policy:    domain.RoutingLatency,
			Status:    "ACTIVE",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			HealthCheck: domain.GlobalHealthCheckConfig{
				Protocol:       "HTTP",
				Port:           80,
				Path:           "/health",
				IntervalSec:    30,
				TimeoutSec:     5,
				HealthyCount:   3,
				UnhealthyCount: 3,
			},
		}

		err := repo.Create(ctx, glb)
		require.NoError(t, err)

		fetched, err := repo.GetByID(ctx, glbID)
		require.NoError(t, err)
		assert.Equal(t, glb.Name, fetched.Name)
		assert.Equal(t, glb.Hostname, fetched.Hostname)
		assert.Equal(t, 80, fetched.HealthCheck.Port)
	})

	t.Run("GetByHostname", func(t *testing.T) {
		fetched, err := repo.GetByHostname(ctx, hostname)
		require.NoError(t, err)
		assert.Equal(t, glbID, fetched.ID)
	})

	t.Run("List", func(t *testing.T) {
		list, err := repo.List(ctx, userID)
		require.NoError(t, err)
		assert.NotEmpty(t, list)
		found := false
		for _, g := range list {
			if g.ID == glbID {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("AddEndpoint", func(t *testing.T) {
		epID := uuid.New()
		ip := "1.2.3.4"
		ep := &domain.GlobalEndpoint{
			ID:         epID,
			GlobalLBID: glbID,
			Region:     "us-east-1",
			TargetType: "IP",
			TargetIP:   &ip,
			Weight:     100,
			Healthy:    true,
			CreatedAt:  time.Now(),
		}

		err := repo.AddEndpoint(ctx, ep)
		require.NoError(t, err)

		endpoints, err := repo.ListEndpoints(ctx, glbID)
		require.NoError(t, err)
		assert.Len(t, endpoints, 1)
		assert.Equal(t, ip, *endpoints[0].TargetIP)
	})

	t.Run("UpdateEndpointHealth", func(t *testing.T) {
		eps, err := repo.ListEndpoints(ctx, glbID)
		require.NoError(t, err)
		require.NotEmpty(t, eps)
		epID := eps[0].ID

		err = repo.UpdateEndpointHealth(ctx, epID, false)
		require.NoError(t, err)

		epsUpdated, err := repo.ListEndpoints(ctx, glbID)
		require.NoError(t, err)
		assert.False(t, epsUpdated[0].Healthy)
	})

	t.Run("Delete", func(t *testing.T) {
		err := repo.Delete(ctx, glbID, userID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, glbID)
		assert.Error(t, err)
	})
}
