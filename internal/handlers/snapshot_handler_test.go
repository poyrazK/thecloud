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
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockSnapshotService struct {
	mock.Mock
}

func (m *mockSnapshotService) CreateSnapshot(ctx context.Context, volumeID uuid.UUID, description string) (*domain.Snapshot, error) {
	args := m.Called(ctx, volumeID, description)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Snapshot), args.Error(1)
}

func (m *mockSnapshotService) ListSnapshots(ctx context.Context) ([]*domain.Snapshot, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Snapshot), args.Error(1)
}

func (m *mockSnapshotService) GetSnapshot(ctx context.Context, id uuid.UUID) (*domain.Snapshot, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Snapshot), args.Error(1)
}

func (m *mockSnapshotService) DeleteSnapshot(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockSnapshotService) RestoreSnapshot(ctx context.Context, id uuid.UUID, newVolumeName string) (*domain.Volume, error) {
	args := m.Called(ctx, id, newVolumeName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Volume), args.Error(1)
}

const (
	testSSnap             = "test snapshot"
	testSpath             = "/snapshots"
	testSpathID           = "/snapshots/"
	testSRestore          = "/restore"
	testVolRest           = "restored-volume"
	snapPathInvalid       = "invalid"
	snapHeaderContentType = "Content-Type"
	snapAppJSON           = "application/json"
)

func TestSnapshotHandlerCreate(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := new(mockSnapshotService)
	handler := NewSnapshotHandler(svc)

	volumeID := uuid.New()
	snapshot := &domain.Snapshot{ID: uuid.New(), VolumeID: volumeID, Description: "test snapshot"}

	svc.On("CreateSnapshot", mock.Anything, volumeID, testSSnap).Return(snapshot, nil)

	reqBody := CreateSnapshotRequest{
		VolumeID:    volumeID,
		Description: testSSnap,
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", testSpath, bytes.NewBuffer(body))
	c.Request.Header.Set(snapHeaderContentType, snapAppJSON)

	handler.Create(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	svc.AssertExpectations(t)
}

func TestSnapshotHandlerList(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := new(mockSnapshotService)
	handler := NewSnapshotHandler(svc)

	snapshots := []*domain.Snapshot{
		{ID: uuid.New()},
		{ID: uuid.New()},
	}

	svc.On("ListSnapshots", mock.Anything).Return(snapshots, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", testSpath, nil)

	handler.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestSnapshotHandlerGet(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := new(mockSnapshotService)
	handler := NewSnapshotHandler(svc)

	id := uuid.New()
	snapshot := &domain.Snapshot{ID: id}

	svc.On("GetSnapshot", mock.Anything, id).Return(snapshot, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", testSpathID+id.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: id.String()}}

	handler.Get(c)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestSnapshotHandlerDelete(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := new(mockSnapshotService)
	handler := NewSnapshotHandler(svc)

	id := uuid.New()

	svc.On("DeleteSnapshot", mock.Anything, id).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("DELETE", testSpathID+id.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: id.String()}}

	handler.Delete(c)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestSnapshotHandlerRestore(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := new(mockSnapshotService)
	handler := NewSnapshotHandler(svc)

	id := uuid.New()
	vol := &domain.Volume{ID: uuid.New(), Name: testVolRest}

	svc.On("RestoreSnapshot", mock.Anything, id, testVolRest).Return(vol, nil)

	reqBody := RestoreSnapshotRequest{
		NewVolumeName: testVolRest,
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", testSpathID+id.String()+testSRestore, bytes.NewBuffer(body))
	c.Request.Header.Set(snapHeaderContentType, snapAppJSON)
	c.Params = gin.Params{{Key: "id", Value: id.String()}}

	handler.Restore(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	svc.AssertExpectations(t)
}

func TestSnapshotHandlerErrorPaths(t *testing.T) {
	t.Parallel()
	setup := func(_ *testing.T) (*mockSnapshotService, *SnapshotHandler, *gin.Engine) {
		svc := new(mockSnapshotService)
		handler := NewSnapshotHandler(svc)
		r := gin.New()
		return svc, handler, r
	}

	t.Run("CreateInvalidJSON", func(t *testing.T) {
		_, handler, r := setup(t)
		r.POST(testSpath, handler.Create)
		req, _ := http.NewRequest("POST", testSpath, bytes.NewBufferString("invalid"))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CreateServiceError", func(t *testing.T) {
		svc, handler, r := setup(t)
		r.POST(testSpath, handler.Create)
		svc.On("CreateSnapshot", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New(errors.Internal, "error"))
		body, _ := json.Marshal(map[string]interface{}{"volume_id": uuid.New(), "description": "d"})
		req, _ := http.NewRequest("POST", testSpath, bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("ListServiceError", func(t *testing.T) {
		svc, handler, r := setup(t)
		r.GET(testSpath, handler.List)
		svc.On("ListSnapshots", mock.Anything).Return(nil, errors.New(errors.Internal, "error"))
		req, _ := http.NewRequest("GET", testSpath, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("GetInvalidID", func(t *testing.T) {
		_, handler, r := setup(t)
		r.GET(testSpath+"/:id", handler.Get)
		req, _ := http.NewRequest("GET", testSpath+"/"+snapPathInvalid, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("GetServiceError", func(t *testing.T) {
		svc, handler, r := setup(t)
		r.GET(testSpath+"/:id", handler.Get)
		id := uuid.New()
		svc.On("GetSnapshot", mock.Anything, id).Return(nil, errors.New(errors.Internal, "error"))
		req, _ := http.NewRequest("GET", testSpath+"/"+id.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("DeleteInvalidID", func(t *testing.T) {
		_, handler, r := setup(t)
		r.DELETE(testSpath+"/:id", handler.Delete)
		req, _ := http.NewRequest("DELETE", testSpath+"/"+snapPathInvalid, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("DeleteServiceError", func(t *testing.T) {
		svc, handler, r := setup(t)
		r.DELETE(testSpath+"/:id", handler.Delete)
		id := uuid.New()
		svc.On("DeleteSnapshot", mock.Anything, id).Return(errors.New(errors.Internal, "error"))
		req, _ := http.NewRequest("DELETE", testSpath+"/"+id.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("RestoreInvalidID", func(t *testing.T) {
		_, handler, r := setup(t)
		r.POST(testSpathID+":id"+testSRestore, handler.Restore)
		req, _ := http.NewRequest("POST", testSpath+"/"+snapPathInvalid+"/restore", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("RestoreInvalidJSON", func(t *testing.T) {
		_, handler, r := setup(t)
		r.POST(testSpathID+":id"+testSRestore, handler.Restore)
		id := uuid.New()
		req, _ := http.NewRequest("POST", testSpath+"/"+id.String()+testSRestore, bytes.NewBufferString("invalid"))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("RestoreServiceError", func(t *testing.T) {
		svc, handler, r := setup(t)
		r.POST(testSpathID+":id"+testSRestore, handler.Restore)
		id := uuid.New()
		svc.On("RestoreSnapshot", mock.Anything, id, mock.Anything).Return(nil, errors.New(errors.Internal, "error"))
		body, _ := json.Marshal(map[string]interface{}{"new_volume_name": "vol"})
		req, _ := http.NewRequest("POST", testSpath+"/"+id.String()+testSRestore, bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
