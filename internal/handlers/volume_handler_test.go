package httphandlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
)

const (
	volumesPath    = "/volumes"
	testVolumeName = "vol-1"
)

type mockVolumeService struct {
	mock.Mock
}

func (m *mockVolumeService) CreateVolume(ctx context.Context, name string, sizeGB int) (*domain.Volume, error) {
	args := m.Called(ctx, name, sizeGB)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).(*domain.Volume)
	return r0, args.Error(1)
}

func (m *mockVolumeService) ListVolumes(ctx context.Context) ([]*domain.Volume, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).([]*domain.Volume)
	return r0, args.Error(1)
}

func (m *mockVolumeService) GetVolume(ctx context.Context, idOrName string) (*domain.Volume, error) {
	args := m.Called(ctx, idOrName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).(*domain.Volume)
	return r0, args.Error(1)
}

func (m *mockVolumeService) DeleteVolume(ctx context.Context, idOrName string) error {
	args := m.Called(ctx, idOrName)
	return args.Error(0)
}

func (m *mockVolumeService) ReleaseVolumesForInstance(ctx context.Context, instanceID uuid.UUID) error {
	return m.Called(ctx, instanceID).Error(0)
}

func (m *mockVolumeService) AttachVolume(ctx context.Context, volumeID string, instanceID string, mountPath string) (string, error) {
	args := m.Called(ctx, volumeID, instanceID, mountPath)
	return args.String(0), args.Error(1)
}

func (m *mockVolumeService) DetachVolume(ctx context.Context, volumeID string) error {
	return m.Called(ctx, volumeID).Error(0)
}

func setupVolumeHandlerTest(_ *testing.T) (*mockVolumeService, *VolumeHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := new(mockVolumeService)
	handler := NewVolumeHandler(svc)
	r := gin.New()
	return svc, handler, r
}

func TestVolumeHandlerCreate(t *testing.T) {
	t.Parallel()
	svc, handler, r := setupVolumeHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.POST(volumesPath, handler.Create)

	vol := &domain.Volume{ID: uuid.New(), Name: testVolumeName}
	svc.On("CreateVolume", mock.Anything, testVolumeName, 10).Return(vol, nil)

	body, err := json.Marshal(map[string]interface{}{
		"name":    testVolumeName,
		"size_gb": 10,
	})
	require.NoError(t, err)
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", volumesPath, bytes.NewBuffer(body))
	require.NoError(t, err)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestVolumeHandlerCreateInvalidJSON(t *testing.T) {
	t.Parallel()
	svc, handler, r := setupVolumeHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.POST(volumesPath, handler.Create)

	req := httptest.NewRequest(http.MethodPost, volumesPath, bytes.NewBufferString("{bad"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "CreateVolume", mock.Anything, mock.Anything, mock.Anything)
}

func TestVolumeHandlerCreateInvalidInputFromService(t *testing.T) {
	t.Parallel()
	svc, handler, r := setupVolumeHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.POST(volumesPath, handler.Create)

	body, err := json.Marshal(map[string]interface{}{
		"name":    testVolumeName,
		"size_gb": 10,
	})
	require.NoError(t, err)

	svc.On("CreateVolume", mock.Anything, testVolumeName, 10).Return(nil, errors.New(errors.InvalidInput, "duplicate"))

	req := httptest.NewRequest(http.MethodPost, volumesPath, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestVolumeHandlerList(t *testing.T) {
	t.Parallel()
	svc, handler, r := setupVolumeHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.GET(volumesPath, handler.List)

	vols := []*domain.Volume{{ID: uuid.New(), Name: testVolumeName}}
	svc.On("ListVolumes", mock.Anything).Return(vols, nil)

	req, err := http.NewRequest(http.MethodGet, volumesPath, nil)
	require.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestVolumeHandlerListError(t *testing.T) {
	t.Parallel()
	svc, handler, r := setupVolumeHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.GET(volumesPath, handler.List)

	svc.On("ListVolumes", mock.Anything).Return(nil, errors.New(errors.Internal, "error"))

	req, err := http.NewRequest(http.MethodGet, volumesPath, nil)
	require.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestVolumeHandlerGet(t *testing.T) {
	t.Parallel()
	svc, handler, r := setupVolumeHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.GET(volumesPath+"/:id", handler.Get)

	id := uuid.New().String()
	vol := &domain.Volume{ID: uuid.New(), Name: testVolumeName}
	svc.On("GetVolume", mock.Anything, id).Return(vol, nil)

	req, err := http.NewRequest(http.MethodGet, volumesPath+"/"+id, nil)
	require.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestVolumeHandlerGetNotFound(t *testing.T) {
	t.Parallel()
	svc, handler, r := setupVolumeHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.GET(volumesPath+"/:id", handler.Get)

	id := uuid.New().String()
	svc.On("GetVolume", mock.Anything, id).Return(nil, errors.New(errors.NotFound, "not found"))

	req := httptest.NewRequest(http.MethodGet, volumesPath+"/"+id, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestVolumeHandlerDelete(t *testing.T) {
	t.Parallel()
	svc, handler, r := setupVolumeHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.DELETE(volumesPath+"/:id", handler.Delete)

	id := uuid.New().String()
	svc.On("DeleteVolume", mock.Anything, id).Return(nil)

	req, err := http.NewRequest(http.MethodDelete, volumesPath+"/"+id, nil)
	require.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestVolumeHandlerDeleteNotFound(t *testing.T) {
	t.Parallel()
	svc, handler, r := setupVolumeHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.DELETE(volumesPath+"/:id", handler.Delete)

	id := uuid.New().String()
	svc.On("DeleteVolume", mock.Anything, id).Return(errors.New(errors.NotFound, "not found"))

	req := httptest.NewRequest(http.MethodDelete, volumesPath+"/"+id, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestVolumeHandlerAttachDetach(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		method         string
		path           string
		body           interface{}
		setupMock      func(m *mockVolumeService)
		expectedStatus int
	}

	volID := uuid.New().String()
	instID := uuid.New().String()
	mountPath := "/mnt/data"
	devicePath := "/dev/vdb"

	testCases := []testCase{
		{
			name:   "Attach Success",
			method: http.MethodPost,
			path:   volumesPath + "/" + volID + "/attach",
			body:   AttachRequest{InstanceID: instID, MountPath: mountPath},
			setupMock: func(m *mockVolumeService) {
				m.On("AttachVolume", mock.Anything, volID, instID, mountPath).Return(devicePath, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Attach Invalid JSON",
			method: http.MethodPost,
			path:   volumesPath + "/" + volID + "/attach",
			body:   "{invalid",
			setupMock: func(m *mockVolumeService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Attach Service Error",
			method: http.MethodPost,
			path:   volumesPath + "/" + volID + "/attach",
			body:   AttachRequest{InstanceID: instID, MountPath: mountPath},
			setupMock: func(m *mockVolumeService) {
				m.On("AttachVolume", mock.Anything, volID, instID, mountPath).Return("", errors.New(errors.Internal, "fail"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "Detach Success",
			method: http.MethodPost,
			path:   volumesPath + "/" + volID + "/detach",
			setupMock: func(m *mockVolumeService) {
				m.On("DetachVolume", mock.Anything, volID).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Detach Service Error",
			method: http.MethodPost,
			path:   volumesPath + "/" + volID + "/detach",
			setupMock: func(m *mockVolumeService) {
				m.On("DetachVolume", mock.Anything, volID).Return(errors.New(errors.Internal, "fail"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, handler, r := setupVolumeHandlerTest(t)
			tc.setupMock(svc)

			r.POST(volumesPath+"/:id/attach", handler.Attach)
			r.POST(volumesPath+"/:id/detach", handler.Detach)

			var reqBody io.Reader
			if tc.body != nil {
				if s, ok := tc.body.(string); ok {
					reqBody = strings.NewReader(s)
				} else {
					b, err := json.Marshal(tc.body)
					require.NoError(t, err)
					reqBody = bytes.NewBuffer(b)
				}
			}

			req := httptest.NewRequest(tc.method, tc.path, reqBody)
			if tc.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			svc.AssertExpectations(t)
		})
	}
}
