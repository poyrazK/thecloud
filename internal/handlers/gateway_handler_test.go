package httphandlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockGatewayService struct {
	mock.Mock
}

func (m *mockGatewayService) CreateRoute(ctx context.Context, name, prefix, target string, strip bool, rateLimit int) (*domain.GatewayRoute, error) {
	args := m.Called(ctx, name, prefix, target, strip, rateLimit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.GatewayRoute), args.Error(1)
}

func (m *mockGatewayService) ListRoutes(ctx context.Context) ([]*domain.GatewayRoute, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.GatewayRoute), args.Error(1)
}

func (m *mockGatewayService) DeleteRoute(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockGatewayService) RefreshRoutes(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockGatewayService) GetProxy(path string) (*httputil.ReverseProxy, bool) {
	args := m.Called(path)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(*httputil.ReverseProxy), args.Bool(1)
}

func setupGatewayHandlerTest(t *testing.T) (*mockGatewayService, *GatewayHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := new(mockGatewayService)
	handler := NewGatewayHandler(svc)
	r := gin.New()
	return svc, handler, r
}

func TestGatewayHandler_CreateRoute(t *testing.T) {
	svc, handler, r := setupGatewayHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.POST("/gateway/routes", handler.CreateRoute)

	route := &domain.GatewayRoute{ID: uuid.New(), Name: "route-1"}
	svc.On("CreateRoute", mock.Anything, "route-1", "/api/v1", "http://example.com", false, 100).Return(route, nil)

	body, err := json.Marshal(map[string]interface{}{
		"name":        "route-1",
		"path_prefix": "/api/v1",
		"target_url":  "http://example.com",
		"rate_limit":  100,
	})
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/gateway/routes", bytes.NewBuffer(body))
	assert.NoError(t, err)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestGatewayHandler_ListRoutes(t *testing.T) {
	svc, handler, r := setupGatewayHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.GET("/gateway/routes", handler.ListRoutes)

	routes := []*domain.GatewayRoute{{ID: uuid.New(), Name: "route-1"}}
	svc.On("ListRoutes", mock.Anything).Return(routes, nil)

	req, err := http.NewRequest(http.MethodGet, "/gateway/routes", nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGatewayHandler_DeleteRoute(t *testing.T) {
	svc, handler, r := setupGatewayHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.DELETE("/gateway/routes/:id", handler.DeleteRoute)

	id := uuid.New()
	svc.On("DeleteRoute", mock.Anything, id).Return(nil)

	req, err := http.NewRequest(http.MethodDelete, "/gateway/routes/"+id.String(), nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGatewayHandler_Proxy_NotFound(t *testing.T) {
	svc, handler, r := setupGatewayHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.Any("/gw/*proxy", handler.Proxy)

	svc.On("GetProxy", "/unknown").Return(nil, false)

	req, err := http.NewRequest(http.MethodGet, "/gw/unknown", nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
