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

const (
	gatewayV1Prefix  = "/v1"
	gatewayTargetURL = "http://target:80"
)

func setupGatewayServiceTest(t *testing.T, initialRoutes []*domain.GatewayRoute) (*MockGatewayRepo, *MockAuditService, ports.GatewayService) {
	repo := new(MockGatewayRepo)
	auditSvc := new(MockAuditService)

	// NewGatewayService calls GetAllActiveRoutes during initialization
	repo.On("GetAllActiveRoutes", mock.Anything).Return(initialRoutes, nil).Once()

	svc := services.NewGatewayService(repo, auditSvc)
	return repo, auditSvc, svc
}

func TestGatewayServiceCreateRoute(t *testing.T) {
	repo, auditSvc, svc := setupGatewayServiceTest(t, []*domain.GatewayRoute{})
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

	repo, _, svc := setupGatewayServiceTest(t, []*domain.GatewayRoute{route})
	defer repo.AssertExpectations(t)

	proxy, ok := svc.GetProxy("/api/users")
	assert.True(t, ok)
	assert.NotNil(t, proxy)

	_, ok = svc.GetProxy("/wrong")
	assert.False(t, ok)
}
