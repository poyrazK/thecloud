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

type mockContainerService struct {
	mock.Mock
}

func (m *mockContainerService) CreateDeployment(ctx context.Context, name, image string, replicas int, ports string) (*domain.Deployment, error) {
	args := m.Called(ctx, name, image, replicas, ports)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Deployment), args.Error(1)
}

func (m *mockContainerService) ListDeployments(ctx context.Context) ([]*domain.Deployment, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Deployment), args.Error(1)
}

func (m *mockContainerService) GetDeployment(ctx context.Context, id uuid.UUID) (*domain.Deployment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Deployment), args.Error(1)
}

func (m *mockContainerService) ScaleDeployment(ctx context.Context, id uuid.UUID, replicas int) error {
	args := m.Called(ctx, id, replicas)
	return args.Error(0)
}

func (m *mockContainerService) DeleteDeployment(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupContainerHandlerTest(t *testing.T) (*mockContainerService, *ContainerHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := new(mockContainerService)
	handler := NewContainerHandler(svc)
	r := gin.New()
	return svc, handler, r
}

func TestContainerHandler_CreateDeployment(t *testing.T) {
	svc, handler, r := setupContainerHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.POST("/containers/deployments", handler.CreateDeployment)

	dep := &domain.Deployment{ID: uuid.New(), Name: "dep-1"}
	svc.On("CreateDeployment", mock.Anything, "dep-1", "nginx", 3, "80:80").Return(dep, nil)

	body, err := json.Marshal(map[string]interface{}{
		"name":     "dep-1",
		"image":    "nginx",
		"replicas": 3,
		"ports":    "80:80",
	})
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/containers/deployments", bytes.NewBuffer(body))
	assert.NoError(t, err)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestContainerHandler_ListDeployments(t *testing.T) {
	svc, handler, r := setupContainerHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.GET("/containers/deployments", handler.ListDeployments)

	deps := []*domain.Deployment{{ID: uuid.New(), Name: "dep-1"}}
	svc.On("ListDeployments", mock.Anything).Return(deps, nil)

	req, err := http.NewRequest(http.MethodGet, "/containers/deployments", nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestContainerHandler_GetDeployment(t *testing.T) {
	svc, handler, r := setupContainerHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.GET("/containers/deployments/:id", handler.GetDeployment)

	id := uuid.New()
	dep := &domain.Deployment{ID: id, Name: "dep-1"}
	svc.On("GetDeployment", mock.Anything, id).Return(dep, nil)

	req, err := http.NewRequest(http.MethodGet, "/containers/deployments/"+id.String(), nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestContainerHandler_ScaleDeployment(t *testing.T) {
	svc, handler, r := setupContainerHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.POST("/containers/deployments/:id/scale", handler.ScaleDeployment)

	id := uuid.New()
	svc.On("ScaleDeployment", mock.Anything, id, 5).Return(nil)

	body, err := json.Marshal(map[string]interface{}{"replicas": 5})
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/containers/deployments/"+id.String()+"/scale", bytes.NewBuffer(body))
	assert.NoError(t, err)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestContainerHandler_DeleteDeployment(t *testing.T) {
	svc, handler, r := setupContainerHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.DELETE("/containers/deployments/:id", handler.DeleteDeployment)

	id := uuid.New()
	svc.On("DeleteDeployment", mock.Anything, id).Return(nil)

	req, err := http.NewRequest(http.MethodDelete, "/containers/deployments/"+id.String(), nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
