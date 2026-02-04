package httphandlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockGlobalLBService provides a mocked implementation of the ports.GlobalLBService interface
// for use in transport-level unit testing.
type mockGlobalLBService struct {
	mock.Mock
}

func (m *mockGlobalLBService) Create(ctx context.Context, name, hostname string, policy domain.RoutingPolicy, healthCheck domain.GlobalHealthCheckConfig) (*domain.GlobalLoadBalancer, error) {
	args := m.Called(ctx, name, hostname, policy, healthCheck)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.GlobalLoadBalancer), args.Error(1)
}

func (m *mockGlobalLBService) Get(ctx context.Context, id uuid.UUID) (*domain.GlobalLoadBalancer, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.GlobalLoadBalancer), args.Error(1)
}

func (m *mockGlobalLBService) List(ctx context.Context, userID uuid.UUID) ([]*domain.GlobalLoadBalancer, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*domain.GlobalLoadBalancer), args.Error(1)
}

func (m *mockGlobalLBService) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *mockGlobalLBService) AddEndpoint(ctx context.Context, glbID uuid.UUID, region string, targetType string, targetID *uuid.UUID, targetIP *string, weight, priority int) (*domain.GlobalEndpoint, error) {
	args := m.Called(ctx, glbID, region, targetType, targetID, targetIP, weight, priority)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.GlobalEndpoint), args.Error(1)
}

func (m *mockGlobalLBService) RemoveEndpoint(ctx context.Context, glbID, endpointID uuid.UUID) error {
	args := m.Called(ctx, glbID, endpointID)
	return args.Error(0)
}

func (m *mockGlobalLBService) ListEndpoints(ctx context.Context, glbID uuid.UUID) ([]*domain.GlobalEndpoint, error) {
	args := m.Called(ctx, glbID)
	return args.Get(0).([]*domain.GlobalEndpoint), args.Error(1)
}

// TestGlobalLBHandlerCreate verifies the behavior of the Create endpoint for Global Load Balancers.
func TestGlobalLBHandlerCreate(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := new(mockGlobalLBService)
		handler := NewGlobalLBHandler(svc)

		req := CreateGlobalLBRequest{
			Name:     "test-glb",
			Hostname: "test.global.com",
			Policy:   domain.RoutingLatency,
			HealthCheck: domain.GlobalHealthCheckConfig{
				Protocol: "HTTP",
				Port:     80,
			},
		}

		glb := &domain.GlobalLoadBalancer{
			ID:       uuid.New(),
			Name:     req.Name,
			Hostname: req.Hostname,
			Status:   "ACTIVE",
		}

		svc.On("Create", mock.Anything, req.Name, req.Hostname, req.Policy, req.HealthCheck).Return(glb, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body, _ := json.Marshal(req)
		c.Request = httptest.NewRequest("POST", "/global-lb", bytes.NewBuffer(body))

		handler.Create(c)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp struct {
			Data domain.GlobalLoadBalancer `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, glb.ID, resp.Data.ID)
	})
}
