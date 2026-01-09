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

const (
	deploymentsPath   = "/containers/deployments"
	testDepName       = "dep-1"
	imageNginx        = "nginx"
	containerPort8080 = "80:80"
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

func TestContainerHandlerCreateDeployment(t *testing.T) {
	svc, handler, r := setupContainerHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.POST(deploymentsPath, handler.CreateDeployment)

	dep := &domain.Deployment{ID: uuid.New(), Name: testDepName}
	svc.On("CreateDeployment", mock.Anything, testDepName, imageNginx, 3, containerPort8080).Return(dep, nil)

	body, err := json.Marshal(map[string]interface{}{
		"name":     testDepName,
		"image":    imageNginx,
		"replicas": 3,
		"ports":    containerPort8080,
	})
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", deploymentsPath, bytes.NewBuffer(body))
	assert.NoError(t, err)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestContainerHandlerListDeployments(t *testing.T) {
	svc, handler, r := setupContainerHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.GET(deploymentsPath, handler.ListDeployments)

	deps := []*domain.Deployment{{ID: uuid.New(), Name: testDepName}}
	svc.On("ListDeployments", mock.Anything).Return(deps, nil)

	req, err := http.NewRequest(http.MethodGet, deploymentsPath, nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestContainerHandlerGetDeployment(t *testing.T) {
	svc, handler, r := setupContainerHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.GET(deploymentsPath+"/:id", handler.GetDeployment)

	id := uuid.New()
	dep := &domain.Deployment{ID: id, Name: testDepName}
	svc.On("GetDeployment", mock.Anything, id).Return(dep, nil)

	req, err := http.NewRequest(http.MethodGet, deploymentsPath+"/"+id.String(), nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestContainerHandlerScaleDeployment(t *testing.T) {
	svc, handler, r := setupContainerHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.POST(deploymentsPath+"/:id/scale", handler.ScaleDeployment)

	id := uuid.New()
	svc.On("ScaleDeployment", mock.Anything, id, 5).Return(nil)

	body, err := json.Marshal(map[string]interface{}{"replicas": 5})
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", deploymentsPath+"/"+id.String()+"/scale", bytes.NewBuffer(body))
	assert.NoError(t, err)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestContainerHandlerDeleteDeployment(t *testing.T) {
	svc, handler, r := setupContainerHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.DELETE(deploymentsPath+"/:id", handler.DeleteDeployment)

	id := uuid.New()
	svc.On("DeleteDeployment", mock.Anything, id).Return(nil)

	req, err := http.NewRequest(http.MethodDelete, deploymentsPath+"/"+id.String(), nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
