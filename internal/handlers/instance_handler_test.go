package httphandlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testInstanceName = "test-inst"
	instancesPath    = "/instances"
	contentType      = "Content-Type"
	applicationJSON  = "application/json"
	complexTestName  = "test-complex"
)

type instanceServiceMock struct {
	mock.Mock
}

func (m *instanceServiceMock) LaunchInstance(ctx context.Context, name, image, ports, instanceType string, vpcID, subnetID *uuid.UUID, volumes []domain.VolumeAttachment) (*domain.Instance, error) {
	args := m.Called(ctx, name, image, ports, instanceType, vpcID, subnetID, volumes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Instance), args.Error(1)
}

func (m *instanceServiceMock) StopInstance(ctx context.Context, idOrName string) error {
	return m.Called(ctx, idOrName).Error(0)
}

func (m *instanceServiceMock) ListInstances(ctx context.Context) ([]*domain.Instance, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*domain.Instance), args.Error(1)
}

func (m *instanceServiceMock) GetInstance(ctx context.Context, idOrName string) (*domain.Instance, error) {
	args := m.Called(ctx, idOrName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Instance), args.Error(1)
}

func (m *instanceServiceMock) GetInstanceLogs(ctx context.Context, idOrName string) (string, error) {
	args := m.Called(ctx, idOrName)
	logs := args.String(0)
	return logs, args.Error(1)
}

func (m *instanceServiceMock) GetInstanceStats(ctx context.Context, idOrName string) (*domain.InstanceStats, error) {
	args := m.Called(ctx, idOrName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.InstanceStats), args.Error(1)
}

func (m *instanceServiceMock) GetConsoleURL(ctx context.Context, idOrName string) (string, error) {
	args := m.Called(ctx, idOrName)
	return args.String(0), args.Error(1)
}

func (m *instanceServiceMock) TerminateInstance(ctx context.Context, idOrName string) error {
	// TerminateInstance permanently deletes an instance
	args := m.Called(ctx, idOrName)
	return args.Error(0)
}

func (m *instanceServiceMock) Exec(ctx context.Context, idOrName string, cmd []string) (string, error) {
	args := m.Called(ctx, idOrName, cmd)
	return args.String(0), args.Error(1)
}

func setupInstanceHandlerTest(_ *testing.T) (*instanceServiceMock, *InstanceHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(instanceServiceMock)
	handler := NewInstanceHandler(mockSvc)
	r := gin.New()
	return mockSvc, handler, r
}

func TestInstanceHandlerLaunchRejectsEmptyImage(t *testing.T) {
	mockSvc, handler, r := setupInstanceHandlerTest(t)
	defer mockSvc.AssertExpectations(t)
	r.POST(instancesPath, handler.Launch)

	body := `{"name":"` + testInstanceName + `","image":"   "}`
	req := httptest.NewRequest(http.MethodPost, instancesPath, strings.NewReader(body))
	req.Header.Set(contentType, applicationJSON)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var wrapper struct {
		Error struct {
			Type string `json:"type"`
		} `json:"error"`
	}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &wrapper))
	assert.Equal(t, "INVALID_INPUT", wrapper.Error.Type)
	mockSvc.AssertNotCalled(t, "LaunchInstance", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestInstanceHandlerLaunchRejectsInvalidJSON(t *testing.T) {
	mockSvc, handler, r := setupInstanceHandlerTest(t)
	defer mockSvc.AssertExpectations(t)
	r.POST(instancesPath, handler.Launch)

	req := httptest.NewRequest(http.MethodPost, instancesPath, strings.NewReader("{invalid"))
	req.Header.Set(contentType, applicationJSON)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockSvc.AssertNotCalled(t, "LaunchInstance", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestInstanceHandlerLaunchRejectsInvalidNameCharacters(t *testing.T) {
	mockSvc, handler, r := setupInstanceHandlerTest(t)
	defer mockSvc.AssertExpectations(t)
	r.POST(instancesPath, handler.Launch)

	body := `{"name":"bad$name","image":"alpine"}`
	req := httptest.NewRequest(http.MethodPost, instancesPath, strings.NewReader(body))
	req.Header.Set(contentType, applicationJSON)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockSvc.AssertNotCalled(t, "LaunchInstance", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestInstanceHandlerLaunch(t *testing.T) {
	mockSvc, handler, r := setupInstanceHandlerTest(t)
	defer mockSvc.AssertExpectations(t)
	r.POST(instancesPath, handler.Launch)

	inst := &domain.Instance{ID: uuid.New(), Name: testInstanceName}
	mockSvc.On("LaunchInstance", mock.Anything, testInstanceName, "alpine", "", "", (*uuid.UUID)(nil), (*uuid.UUID)(nil), []domain.VolumeAttachment(nil)).Return(inst, nil)

	body := `{"name":"` + testInstanceName + `","image":"alpine"}`
	req := httptest.NewRequest(http.MethodPost, instancesPath, strings.NewReader(body))
	req.Header.Set(contentType, applicationJSON)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)
}

func TestInstanceHandlerLaunchRejectsInvalidMountPath(t *testing.T) {
	mockSvc, handler, r := setupInstanceHandlerTest(t)
	defer mockSvc.AssertExpectations(t)
	r.POST(instancesPath, handler.Launch)

	body := `{"name":"` + testInstanceName + `","image":"alpine","volumes":[{"volume_id":"vol-1","mount_path":"mnt/data"}]}`
	req := httptest.NewRequest(http.MethodPost, instancesPath, strings.NewReader(body))
	req.Header.Set(contentType, applicationJSON)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockSvc.AssertNotCalled(t, "LaunchInstance", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestInstanceHandlerList(t *testing.T) {
	mockSvc, handler, r := setupInstanceHandlerTest(t)
	defer mockSvc.AssertExpectations(t)
	r.GET(instancesPath, handler.List)

	instances := []*domain.Instance{{ID: uuid.New(), Name: testInstanceName}}
	mockSvc.On("ListInstances", mock.Anything).Return(instances, nil)

	req := httptest.NewRequest(http.MethodGet, instancesPath, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInstanceHandlerGet(t *testing.T) {
	mockSvc, handler, r := setupInstanceHandlerTest(t)
	defer mockSvc.AssertExpectations(t)
	r.GET(instancesPath+"/:id", handler.Get)

	id := uuid.New().String()
	inst := &domain.Instance{ID: uuid.MustParse(id), Name: testInstanceName}
	mockSvc.On("GetInstance", mock.Anything, id).Return(inst, nil)

	req := httptest.NewRequest(http.MethodGet, instancesPath+"/"+id, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInstanceHandlerStop(t *testing.T) {
	mockSvc, handler, r := setupInstanceHandlerTest(t)
	defer mockSvc.AssertExpectations(t)
	r.POST(instancesPath+"/:id/stop", handler.Stop)

	id := uuid.New().String()
	mockSvc.On("StopInstance", mock.Anything, id).Return(nil)

	req := httptest.NewRequest(http.MethodPost, instancesPath+"/"+id+"/stop", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInstanceHandlerTerminate(t *testing.T) {
	mockSvc, handler, r := setupInstanceHandlerTest(t)
	defer mockSvc.AssertExpectations(t)
	r.DELETE(instancesPath+"/:id", handler.Terminate)

	id := uuid.New().String()
	mockSvc.On("TerminateInstance", mock.Anything, id).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, instancesPath+"/"+id, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInstanceHandlerStopNotFound(t *testing.T) {
	mockSvc, handler, r := setupInstanceHandlerTest(t)
	defer mockSvc.AssertExpectations(t)
	r.POST(instancesPath+"/:id/stop", handler.Stop)

	id := uuid.New().String()
	mockSvc.On("StopInstance", mock.Anything, id).Return(errors.New(errors.NotFound, "not found"))

	req := httptest.NewRequest(http.MethodPost, instancesPath+"/"+id+"/stop", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestInstanceHandlerTerminateNotFound(t *testing.T) {
	mockSvc, handler, r := setupInstanceHandlerTest(t)
	defer mockSvc.AssertExpectations(t)
	r.DELETE(instancesPath+"/:id", handler.Terminate)

	id := uuid.New().String()
	mockSvc.On("TerminateInstance", mock.Anything, id).Return(errors.New(errors.NotFound, "not found"))

	req := httptest.NewRequest(http.MethodDelete, instancesPath+"/"+id, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestInstanceHandlerGetLogs(t *testing.T) {
	mockSvc, handler, r := setupInstanceHandlerTest(t)
	defer mockSvc.AssertExpectations(t)
	r.GET(instancesPath+"/:id/logs", handler.GetLogs)

	id := uuid.New().String()
	mockSvc.On("GetInstanceLogs", mock.Anything, id).Return("logs content", nil)

	req := httptest.NewRequest(http.MethodGet, instancesPath+"/"+id+"/logs", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "logs content", w.Body.String())
}

func TestInstanceHandlerGetStats(t *testing.T) {
	mockSvc, handler, r := setupInstanceHandlerTest(t)
	defer mockSvc.AssertExpectations(t)
	r.GET(instancesPath+"/:id/stats", handler.GetStats)

	id := uuid.New().String()
	stats := &domain.InstanceStats{CPUPercentage: 10.5, MemoryUsageBytes: 128}
	mockSvc.On("GetInstanceStats", mock.Anything, id).Return(stats, nil)

	req := httptest.NewRequest(http.MethodGet, instancesPath+"/"+id+"/stats", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInstanceHandlerLaunchWithVolumesAndVPC(t *testing.T) {
	mockSvc, handler, r := setupInstanceHandlerTest(t)
	defer mockSvc.AssertExpectations(t)
	r.POST(instancesPath, handler.Launch)

	vpcID := uuid.New()
	volID := "vol-123"

	inst := &domain.Instance{ID: uuid.New(), Name: complexTestName, VpcID: &vpcID}

	expectedVolumes := []domain.VolumeAttachment{
		{VolumeIDOrName: volID, MountPath: "/mnt/data"},
	}

	mockSvc.On("LaunchInstance", mock.Anything, complexTestName, "ubuntu", "80:80", "", &vpcID, (*uuid.UUID)(nil), expectedVolumes).Return(inst, nil)

	body := map[string]interface{}{
		"name":   complexTestName,
		"image":  "ubuntu",
		"ports":  "80:80",
		"vpc_id": vpcID.String(),
		"volumes": []map[string]string{
			{"volume_id": volID, "mount_path": "/mnt/data"},
		},
	}
	jsonBody, err := json.Marshal(body)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, instancesPath, bytes.NewBuffer(jsonBody))
	req.Header.Set(contentType, applicationJSON)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)
}

func TestInstanceHandlerLaunchRejectsInvalidVPCID(t *testing.T) {
	mockSvc, handler, r := setupInstanceHandlerTest(t)
	defer mockSvc.AssertExpectations(t)
	r.POST(instancesPath, handler.Launch)

	body := `{"name":"` + testInstanceName + `","image":"ubuntu","vpc_id":"not-a-uuid"}`
	req := httptest.NewRequest(http.MethodPost, instancesPath, strings.NewReader(body))
	req.Header.Set(contentType, applicationJSON)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockSvc.AssertNotCalled(t, "LaunchInstance", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestInstanceHandlerGetConsole(t *testing.T) {
	mockSvc, handler, r := setupInstanceHandlerTest(t)
	defer mockSvc.AssertExpectations(t)
	r.GET(instancesPath+"/:id/console", handler.GetConsole)

	id := uuid.New().String()
	url := "vnc://localhost:5900"
	mockSvc.On("GetConsoleURL", mock.Anything, id).Return(url, nil)

	req := httptest.NewRequest(http.MethodGet, instancesPath+"/"+id+"/console", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Data map[string]string `json:"data"`
	}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, url, resp.Data["url"])
}

func TestInstanceHandlerGetConsoleError(t *testing.T) {
	mockSvc, handler, r := setupInstanceHandlerTest(t)
	defer mockSvc.AssertExpectations(t)
	r.GET(instancesPath+"/:id/console", handler.GetConsole)

	id := uuid.New().String()
	mockSvc.On("GetConsoleURL", mock.Anything, id).Return("", errors.New(errors.Internal, "console error"))

	req := httptest.NewRequest(http.MethodGet, instancesPath+"/"+id+"/console", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
