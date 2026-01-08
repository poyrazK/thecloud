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

type mockVolumeService struct {
	mock.Mock
}

func (m *mockVolumeService) CreateVolume(ctx context.Context, name string, sizeGB int) (*domain.Volume, error) {
	args := m.Called(ctx, name, sizeGB)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Volume), args.Error(1)
}

func (m *mockVolumeService) ListVolumes(ctx context.Context) ([]*domain.Volume, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Volume), args.Error(1)
}

func (m *mockVolumeService) GetVolume(ctx context.Context, idOrName string) (*domain.Volume, error) {
	args := m.Called(ctx, idOrName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Volume), args.Error(1)
}

func (m *mockVolumeService) DeleteVolume(ctx context.Context, idOrName string) error {
	args := m.Called(ctx, idOrName)
	return args.Error(0)
}

func (m *mockVolumeService) ReleaseVolumesForInstance(ctx context.Context, instanceID uuid.UUID) error {
	args := m.Called(ctx, instanceID)
	return args.Error(0)
}

func setupVolumeHandlerTest(t *testing.T) (*mockVolumeService, *VolumeHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := new(mockVolumeService)
	handler := NewVolumeHandler(svc)
	r := gin.New()
	return svc, handler, r
}

func TestVolumeHandler_Create(t *testing.T) {
	svc, handler, r := setupVolumeHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.POST("/volumes", handler.Create)

	vol := &domain.Volume{ID: uuid.New(), Name: "vol-1"}
	svc.On("CreateVolume", mock.Anything, "vol-1", 10).Return(vol, nil)

	body, err := json.Marshal(map[string]interface{}{
		"name":    "vol-1",
		"size_gb": 10,
	})
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/volumes", bytes.NewBuffer(body))
	assert.NoError(t, err)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestVolumeHandler_List(t *testing.T) {
	svc, handler, r := setupVolumeHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.GET("/volumes", handler.List)

	vols := []*domain.Volume{{ID: uuid.New(), Name: "vol-1"}}
	svc.On("ListVolumes", mock.Anything).Return(vols, nil)

	req, err := http.NewRequest(http.MethodGet, "/volumes", nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestVolumeHandler_Get(t *testing.T) {
	svc, handler, r := setupVolumeHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.GET("/volumes/:id", handler.Get)

	id := uuid.New().String()
	vol := &domain.Volume{ID: uuid.New(), Name: "vol-1"}
	svc.On("GetVolume", mock.Anything, id).Return(vol, nil)

	req, err := http.NewRequest(http.MethodGet, "/volumes/"+id, nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestVolumeHandler_Delete(t *testing.T) {
	svc, handler, r := setupVolumeHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.DELETE("/volumes/:id", handler.Delete)

	id := uuid.New().String()
	svc.On("DeleteVolume", mock.Anything, id).Return(nil)

	req, err := http.NewRequest(http.MethodDelete, "/volumes/"+id, nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
