package services_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	gatewayV1Prefix  = "/v1"
	gatewayTargetURL = "http://target:80"
)

func setupGatewayServiceTest(initialRoutes []*domain.GatewayRoute) (*MockGatewayRepo, *MockAuditService, ports.GatewayService) {
	repo := new(MockGatewayRepo)
	auditSvc := new(MockAuditService)

	// NewGatewayService calls GetAllActiveRoutes during initialization
	repo.On("GetAllActiveRoutes", mock.Anything).Return(initialRoutes, nil).Once()

	svc := services.NewGatewayService(repo, auditSvc)
	return repo, auditSvc, svc
}

func TestGatewayServiceCreateRoute(t *testing.T) {
	repo, auditSvc, svc := setupGatewayServiceTest([]*domain.GatewayRoute{})
	defer repo.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)

	repo.On("CreateRoute", ctx, mock.AnythingOfType("*domain.GatewayRoute")).Return(nil)
	repo.On("GetAllActiveRoutes", ctx).Return([]*domain.GatewayRoute{}, nil)
	auditSvc.On("Log", ctx, userID, "gateway.route_create", "gateway", mock.Anything, mock.MatchedBy(func(details map[string]interface{}) bool {
		return details["prefix"] == gatewayV1Prefix
	})).Return(nil)

	route, err := svc.CreateRoute(ctx, "test-api", gatewayV1Prefix, gatewayTargetURL, true, 100)

	assert.NoError(t, err)
	assert.NotNil(t, route)
	assert.Equal(t, gatewayV1Prefix, route.PathPrefix)
}

func TestGatewayServiceRefreshAndGetProxy(t *testing.T) {
	route := &domain.GatewayRoute{
		PathPrefix: "/api",
		TargetURL:  "http://localhost:8080",
	}

	repo, _, svc := setupGatewayServiceTest([]*domain.GatewayRoute{route})
	defer repo.AssertExpectations(t)

	proxy, ok := svc.GetProxy("/api/users")
	assert.True(t, ok)
	assert.NotNil(t, proxy)

	_, ok = svc.GetProxy("/wrong")
	assert.False(t, ok)
}

func TestGatewayServiceListRoutes(t *testing.T) {
	repo, _, svc := setupGatewayServiceTest(nil)
	defer repo.AssertExpectations(t)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	routes := []*domain.GatewayRoute{{ID: uuid.New(), UserID: userID}}

	repo.On("ListRoutes", ctx, userID).Return(routes, nil)

	res, err := svc.ListRoutes(ctx)

	assert.NoError(t, err)
	assert.Equal(t, routes, res)
}

func TestGatewayServiceDeleteRoute(t *testing.T) {
	repo, audit, svc := setupGatewayServiceTest(nil)
	defer repo.AssertExpectations(t)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	routeID := uuid.New()
	route := &domain.GatewayRoute{ID: routeID, UserID: userID, Name: "test"}

	repo.On("GetRouteByID", ctx, routeID, userID).Return(route, nil)
	repo.On("DeleteRoute", ctx, routeID).Return(nil)
	repo.On("GetAllActiveRoutes", ctx).Return([]*domain.GatewayRoute{}, nil)
	audit.On("Log", ctx, userID, "gateway.route_delete", "gateway", routeID.String(), mock.Anything).Return(nil)

	err := svc.DeleteRoute(ctx, routeID)
	assert.NoError(t, err)
}

func TestGatewayServiceCreateRouteUnauthorized(t *testing.T) {
	_, _, svc := setupGatewayServiceTest(nil)
	ctx := context.Background()

	_, err := svc.CreateRoute(ctx, "test", "/api", "http://target", true, 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized")
}

func TestGatewayServiceErrors(t *testing.T) {
	ctx := appcontext.WithUserID(context.Background(), uuid.New())
	userID := appcontext.UserIDFromContext(ctx)

	t.Run("CreateRoute_RepoError", func(t *testing.T) {
		repo, _, svc := setupGatewayServiceTest(nil)
		repo.On("CreateRoute", ctx, mock.Anything).Return(assert.AnError)
		_, err := svc.CreateRoute(ctx, "test", "/api", "http://target", true, 100)
		assert.Error(t, err)
	})

	t.Run("ListRoutes_RepoError", func(t *testing.T) {
		repo, _, svc := setupGatewayServiceTest(nil)
		repo.On("ListRoutes", ctx, userID).Return(nil, assert.AnError)
		_, err := svc.ListRoutes(ctx)
		assert.Error(t, err)
	})

	t.Run("DeleteRoute_GetError", func(t *testing.T) {
		repo, _, svc := setupGatewayServiceTest(nil)
		repo.On("GetRouteByID", ctx, mock.Anything, userID).Return(nil, assert.AnError)
		err := svc.DeleteRoute(ctx, uuid.New())
		assert.Error(t, err)
	})

	t.Run("RefreshRoutes_RepoError", func(t *testing.T) {
		repo, _, svc := setupGatewayServiceTest(nil)
		repo.On("GetAllActiveRoutes", mock.Anything).Return(nil, assert.AnError)
		err := svc.RefreshRoutes(context.Background())
		assert.Error(t, err)
	})

	t.Run("RefreshRoutes_ParseError", func(t *testing.T) {
		repo, _, svc := setupGatewayServiceTest(nil)
		routes := []*domain.GatewayRoute{{PathPrefix: "/api", TargetURL: "::invalid"}}
		repo.On("GetAllActiveRoutes", mock.Anything).Return(routes, nil)
		err := svc.RefreshRoutes(context.Background())
		assert.NoError(t, err) // Should skip invalid URLs but not return error
	})
}

func TestGatewayServiceListRoutesUnauthorized(t *testing.T) {
	_, _, svc := setupGatewayServiceTest(nil)
	ctx := context.Background()
	_, err := svc.ListRoutes(ctx)
	assert.Error(t, err)
}

func TestGatewayServiceDeleteRouteDeleteError(t *testing.T) {
	repo, _, svc := setupGatewayServiceTest(nil)
	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	routeID := uuid.New()

	repo.On("GetRouteByID", ctx, routeID, userID).Return(&domain.GatewayRoute{ID: routeID, UserID: userID}, nil)
	repo.On("DeleteRoute", ctx, routeID).Return(assert.AnError)

	err := svc.DeleteRoute(ctx, routeID)
	assert.Error(t, err)
}

func TestGatewayServiceProxyDirector(t *testing.T) {
	repo, _, svc := setupGatewayServiceTest(nil)
	ctx := context.Background()

	route := &domain.GatewayRoute{
		PathPrefix:  "/api",
		TargetURL:   "http://localhost:8080",
		StripPrefix: true,
	}
	repo.On("GetAllActiveRoutes", ctx).Return([]*domain.GatewayRoute{route}, nil)
	_ = svc.RefreshRoutes(ctx)

	proxy, ok := svc.GetProxy("/api/users")
	assert.True(t, ok)

	req := httptest.NewRequest("GET", "http://gateway/gw/api/users", nil)
	// We need to trigger the director. ReverseProxy triggers it during ServeHTTP.
	// But we can just call it if we can access it.
	// Since we can't easily call it without a full HTTP roundtrip, let's use a dummy recorder.
	_ = httptest.NewRecorder()

	// This will call the director
	// Note: It might try to actually send the request, so we should be careful.
	// Actually, we can just call the director function directly if we can reach it.
	// But it's assigned to proxy.Director.

	proxy.Director(req)

	assert.Equal(t, "/users", req.URL.Path)
	assert.Equal(t, "localhost:8080", req.Host)
}
