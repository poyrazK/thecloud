package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGatewayServiceCreateRoute(t *testing.T) {
	repo := new(MockGatewayRepo)
	auditSvc := new(MockAuditService)

	// RefreshRoutes called during init
	repo.On("GetAllActiveRoutes", mock.Anything).Return([]*domain.GatewayRoute{}, nil)

	svc := services.NewGatewayService(repo, auditSvc)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)

	repo.On("CreateRoute", mock.Anything, mock.MatchedBy(func(r *domain.GatewayRoute) bool {
		return r.UserID == userID && r.Name == "test-route"
	})).Return(nil)

	auditSvc.On("Log", mock.Anything, userID, "gateway.route_create", "gateway", mock.Anything, mock.Anything).Return(nil)

	route, err := svc.CreateRoute(ctx, "test-route", "/test", "http://example.com", false, 100)
	assert.NoError(t, err)
	assert.NotNil(t, route)
	assert.Equal(t, "test-route", route.Name)

	repo.AssertExpectations(t)
	auditSvc.AssertExpectations(t)
}
