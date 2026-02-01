package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

const (
	testRouteName    = "test-route"
	testRoutePattern = "/api/*"
	testTargetURL    = "http://localhost"
)

func TestGatewayCreateRoute(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewPostgresGatewayRepository(mock)
	route := &domain.GatewayRoute{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		Name:        testRouteName,
		PathPrefix:  "/api",
		PathPattern: testRoutePattern,
		PatternType: "wildcard",
		ParamNames:  []string{},
		TargetURL:   testTargetURL,
		Methods:     []string{"GET"},
		StripPrefix: true,
		RateLimit:   100,
		Priority:    1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mock.ExpectExec("INSERT INTO gateway_routes").
		WithArgs(route.ID, route.UserID, route.Name, route.PathPrefix, route.PathPattern, route.PatternType, route.ParamNames, route.TargetURL, route.Methods, route.StripPrefix, route.RateLimit, route.Priority, route.CreatedAt, route.UpdatedAt).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.CreateRoute(context.Background(), route)
	assert.NoError(t, err)
}

func TestGatewayGetRouteByID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewPostgresGatewayRepository(mock)
	id := uuid.New()
	userID := uuid.New()
	now := time.Now()

	columns := []string{"id", "user_id", "name", "path_prefix", "path_pattern", "pattern_type", "param_names", "target_url", "methods", "strip_prefix", "rate_limit", "priority", "created_at", "updated_at"}
	mock.ExpectQuery("SELECT id, user_id, name, path_prefix, path_pattern, pattern_type, param_names, target_url, methods, strip_prefix, rate_limit, priority, created_at, updated_at FROM gateway_routes").
		WithArgs(id, userID).
		WillReturnRows(pgxmock.NewRows(columns).
			AddRow(id, userID, testRouteName, "/api", testRoutePattern, "wildcard", []string{}, testTargetURL, []string{"GET"}, true, 100, 1, now, now))

	route, err := repo.GetRouteByID(context.Background(), id, userID)
	assert.NoError(t, err)
	assert.NotNil(t, route)
	assert.Equal(t, id, route.ID)
}

func TestGatewayListRoutes(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewPostgresGatewayRepository(mock)
	userID := uuid.New()
	now := time.Now()

	columns := []string{"id", "user_id", "name", "path_prefix", "path_pattern", "pattern_type", "param_names", "target_url", "methods", "strip_prefix", "rate_limit", "priority", "created_at", "updated_at"}
	mock.ExpectQuery("SELECT id, user_id, name, path_prefix, path_pattern, pattern_type, param_names, target_url, methods, strip_prefix, rate_limit, priority, created_at, updated_at FROM gateway_routes").
		WithArgs(userID).
		WillReturnRows(pgxmock.NewRows(columns).
			AddRow(uuid.New(), userID, testRouteName, "/api", testRoutePattern, "wildcard", []string{}, testTargetURL, []string{"GET"}, true, 100, 1, now, now))

	routes, err := repo.ListRoutes(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, routes, 1)
}
