package services_test

import (
	"context"
	"testing"

	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
)

func setupGatewayServiceTest(t *testing.T) (*services.GatewayService, *postgres.PostgresGatewayRepository, context.Context) {
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewPostgresGatewayRepository(db)

	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(auditRepo)

	svc := services.NewGatewayService(repo, auditSvc)
	return svc, repo.(*postgres.PostgresGatewayRepository), ctx
}

func TestGatewayServiceCreateRoute(t *testing.T) {
	svc, repo, ctx := setupGatewayServiceTest(t)
	userID := appcontext.UserIDFromContext(ctx)

	params := ports.CreateRouteParams{
		Name:      "test-route",
		Pattern:   "/test",
		Target:    "http://example.com",
		RateLimit: 100,
	}
	route, err := svc.CreateRoute(ctx, params)
	assert.NoError(t, err)
	assert.NotNil(t, route)
	assert.Equal(t, "test-route", route.Name)

	// Verify in DB
	fetched, err := repo.GetRouteByID(ctx, route.ID, userID)
	assert.NoError(t, err)
	assert.Equal(t, route.ID, fetched.ID)
}

func TestGatewayServiceListRoutes(t *testing.T) {
	svc, _, ctx := setupGatewayServiceTest(t)

	_, _ = svc.CreateRoute(ctx, ports.CreateRouteParams{Name: "r1", Pattern: "/r1", Target: "http://e.com"})
	_, _ = svc.CreateRoute(ctx, ports.CreateRouteParams{Name: "r2", Pattern: "/r2", Target: "http://e.com"})

	res, err := svc.ListRoutes(ctx)
	assert.NoError(t, err)
	assert.Len(t, res, 2)
}

func TestGatewayServiceDeleteRoute(t *testing.T) {
	svc, repo, ctx := setupGatewayServiceTest(t)
	userID := appcontext.UserIDFromContext(ctx)

	route, _ := svc.CreateRoute(ctx, ports.CreateRouteParams{Name: "r1", Pattern: "/r1", Target: "http://e.com"})

	err := svc.DeleteRoute(ctx, route.ID)
	assert.NoError(t, err)

	// Verify deleted
	_, err = repo.GetRouteByID(ctx, route.ID, userID)
	assert.Error(t, err)
}

func TestGatewayServiceGetProxy(t *testing.T) {
	svc, _, ctx := setupGatewayServiceTest(t)

	_, _ = svc.CreateRoute(ctx, ports.CreateRouteParams{Name: "api", Pattern: "/api", Target: "http://localhost:8080"})

	proxy, paramsMap, ok := svc.GetProxy("GET", "/api/users")
	assert.True(t, ok)
	assert.NotNil(t, proxy)
	assert.Nil(t, paramsMap)
}

func TestGatewayServiceGetProxyPattern(t *testing.T) {
	svc, _, ctx := setupGatewayServiceTest(t)

	_, _ = svc.CreateRoute(ctx, ports.CreateRouteParams{Name: "users", Pattern: "/users/{id}", Target: "http://localhost:8080"})

	proxy, paramsMap, ok := svc.GetProxy("GET", "/users/123")
	assert.True(t, ok)
	assert.NotNil(t, proxy)
	assert.Equal(t, "123", paramsMap["id"])
}
