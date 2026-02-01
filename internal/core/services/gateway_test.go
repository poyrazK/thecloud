package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const testRouteName = "test-route"

func TestGatewayServiceCreateRoute(t *testing.T) {
	repo := new(MockGatewayRepo)
	auditSvc := new(MockAuditService)

	// RefreshRoutes called during init
	repo.On("GetAllActiveRoutes", mock.Anything).Return([]*domain.GatewayRoute{}, nil)

	svc := services.NewGatewayService(repo, auditSvc)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)

	repo.On("CreateRoute", mock.Anything, mock.MatchedBy(func(r *domain.GatewayRoute) bool {
		return r.UserID == userID && r.Name == testRouteName
	})).Return(nil)

	auditSvc.On("Log", mock.Anything, userID, "gateway.route_create", "gateway", mock.Anything, mock.Anything).Return(nil)

	params := ports.CreateRouteParams{
		Name:      testRouteName,
		Pattern:   "/test",
		Target:    "http://example.com",
		RateLimit: 100,
	}
	route, err := svc.CreateRoute(ctx, params)
	assert.NoError(t, err)
	assert.NotNil(t, route)
	assert.Equal(t, testRouteName, route.Name)

	repo.AssertExpectations(t)
	auditSvc.AssertExpectations(t)
}

func TestGatewayServiceListRoutes(t *testing.T) {
	repo := new(MockGatewayRepo)
	auditSvc := new(MockAuditService)

	repo.On("GetAllActiveRoutes", mock.Anything).Return([]*domain.GatewayRoute{}, nil)
	svc := services.NewGatewayService(repo, auditSvc)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)

	routes := []*domain.GatewayRoute{{ID: uuid.New(), Name: "r1"}}
	repo.On("ListRoutes", mock.Anything, userID).Return(routes, nil)

	res, err := svc.ListRoutes(ctx)
	assert.NoError(t, err)
	assert.Equal(t, routes, res)
}

func TestGatewayServiceDeleteRoute(t *testing.T) {
	repo := new(MockGatewayRepo)
	auditSvc := new(MockAuditService)

	repo.On("GetAllActiveRoutes", mock.Anything).Return([]*domain.GatewayRoute{}, nil)
	svc := services.NewGatewayService(repo, auditSvc)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	routeID := uuid.New()
	route := &domain.GatewayRoute{ID: routeID, UserID: userID, Name: "r1"}

	repo.On("GetRouteByID", mock.Anything, routeID, userID).Return(route, nil)
	repo.On("DeleteRoute", mock.Anything, routeID).Return(nil)
	repo.On("GetAllActiveRoutes", mock.Anything).Return([]*domain.GatewayRoute{}, nil)
	auditSvc.On("Log", mock.Anything, userID, "gateway.route_delete", "gateway", routeID.String(), mock.Anything).Return(nil)

	err := svc.DeleteRoute(ctx, routeID)
	assert.NoError(t, err)
}

func TestGatewayServiceGetProxy(t *testing.T) {
	repo := new(MockGatewayRepo)
	auditSvc := new(MockAuditService)

	route := &domain.GatewayRoute{
		PathPrefix: "/api",
		TargetURL:  "http://localhost:8080",
	}
	repo.On("GetAllActiveRoutes", mock.Anything).Return([]*domain.GatewayRoute{route}, nil)

	svc := services.NewGatewayService(repo, auditSvc)

	proxy, params, ok := svc.GetProxy("GET", "/api/users")
	assert.True(t, ok)
	assert.NotNil(t, proxy)
	assert.Nil(t, params)

	_, _, ok = svc.GetProxy("GET", "/other")
	assert.False(t, ok)
}

func TestGatewayServiceGetProxyPattern(t *testing.T) {
	repo := new(MockGatewayRepo)
	auditSvc := new(MockAuditService)

	route := &domain.GatewayRoute{
		ID:          uuid.New(),
		PathPattern: "/users/{id}",
		PatternType: "pattern",
		TargetURL:   "http://localhost:8080",
	}
	repo.On("GetAllActiveRoutes", mock.Anything).Return([]*domain.GatewayRoute{route}, nil)

	svc := services.NewGatewayService(repo, auditSvc)

	proxy, params, ok := svc.GetProxy("GET", "/users/123")
	assert.True(t, ok)
	assert.NotNil(t, proxy)
	assert.Equal(t, "123", params["id"])

	_, _, ok = svc.GetProxy("GET", "/users/123/posts")
	assert.False(t, ok)
}
