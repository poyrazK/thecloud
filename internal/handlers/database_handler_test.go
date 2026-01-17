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

const (
	// testDBConnStr is a dummy connection string for testing - safe for hardcoding
	testDBConnStr = "postgres://user:pass@host:5432/db"
)

const (
	databasesPath = "/databases"
	testDBName    = "db-1"
	connPath      = "/:id/connection"
	connSuffix    = "/connection"
	dbPathInvalid   = "/invalid"
)

type mockDatabaseService struct {
	mock.Mock
}

func (m *mockDatabaseService) CreateDatabase(ctx context.Context, name, engine, version string, vpcID *uuid.UUID) (*domain.Database, error) {
	args := m.Called(ctx, name, engine, version, vpcID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Database), args.Error(1)
}

func (m *mockDatabaseService) ListDatabases(ctx context.Context) ([]*domain.Database, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Database), args.Error(1)
}

func (m *mockDatabaseService) GetDatabase(ctx context.Context, id uuid.UUID) (*domain.Database, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Database), args.Error(1)
}

func (m *mockDatabaseService) DeleteDatabase(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockDatabaseService) GetConnectionString(ctx context.Context, id uuid.UUID) (string, error) {
	args := m.Called(ctx, id)
	return args.String(0), args.Error(1)
}

func setupDatabaseHandlerTest(_ *testing.T) (*mockDatabaseService, *DatabaseHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := new(mockDatabaseService)
	handler := NewDatabaseHandler(svc)
	r := gin.New()
	return svc, handler, r
}

func TestDatabaseHandlerCreate(t *testing.T) {
	svc, handler, r := setupDatabaseHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.POST(databasesPath, handler.Create)

	db := &domain.Database{ID: uuid.New(), Name: testDBName}
	svc.On("CreateDatabase", mock.Anything, testDBName, "postgres", "15", (*uuid.UUID)(nil)).Return(db, nil)

	body, err := json.Marshal(map[string]interface{}{
		"name":    testDBName,
		"engine":  "postgres",
		"version": "15",
	})
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", databasesPath, bytes.NewBuffer(body))
	assert.NoError(t, err)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestDatabaseHandlerList(t *testing.T) {
	svc, handler, r := setupDatabaseHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.GET(databasesPath, handler.List)

	dbs := []*domain.Database{{ID: uuid.New(), Name: testDBName}}
	svc.On("ListDatabases", mock.Anything).Return(dbs, nil)

	req, err := http.NewRequest(http.MethodGet, databasesPath, nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDatabaseHandlerGet(t *testing.T) {
	svc, handler, r := setupDatabaseHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.GET(databasesPath+"/:id", handler.Get)

	id := uuid.New()
	db := &domain.Database{ID: id, Name: testDBName}
	svc.On("GetDatabase", mock.Anything, id).Return(db, nil)

	req, err := http.NewRequest(http.MethodGet, databasesPath+"/"+id.String(), nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDatabaseHandlerDelete(t *testing.T) {
	svc, handler, r := setupDatabaseHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.DELETE(databasesPath+"/:id", handler.Delete)

	id := uuid.New()
	svc.On("DeleteDatabase", mock.Anything, id).Return(nil)

	req, err := http.NewRequest(http.MethodDelete, databasesPath+"/"+id.String(), nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDatabaseHandlerGetConnectionString(t *testing.T) {
	svc, handler, r := setupDatabaseHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.GET(databasesPath+connPath, handler.GetConnectionString)

	id := uuid.New()
	svc.On("GetConnectionString", mock.Anything, id).Return(testDBConnStr, nil)

	req, err := http.NewRequest(http.MethodGet, databasesPath+"/"+id.String()+connSuffix, nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), testDBConnStr)
}

func TestDatabaseHandlerCreateError(t *testing.T) {
	t.Run("InvalidJSON", func(t *testing.T) {
		_, handler, r := setupDatabaseHandlerTest(t)
		r.POST(databasesPath, handler.Create)
		req, _ := http.NewRequest("POST", databasesPath, bytes.NewBufferString("invalid"))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ServiceError", func(t *testing.T) {
		svc, handler, r := setupDatabaseHandlerTest(t)
		r.POST(databasesPath, handler.Create)
		svc.On("CreateDatabase", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil, errors.New(errors.Internal, "error"))
		body, _ := json.Marshal(map[string]interface{}{"name": "n", "engine": "e", "version": "v"})
		req, _ := http.NewRequest("POST", databasesPath, bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})
}

func TestDatabaseHandlerListError(t *testing.T) {
	svc, handler, r := setupDatabaseHandlerTest(t)
	r.GET(databasesPath, handler.List)
	svc.On("ListDatabases", mock.Anything).Return(nil, errors.New(errors.Internal, "error"))
	req, _ := http.NewRequest(http.MethodGet, databasesPath, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

func TestDatabaseHandlerGetError(t *testing.T) {
	t.Run("InvalidID", func(t *testing.T) {
		_, handler, r := setupDatabaseHandlerTest(t)
		r.GET(databasesPath+"/:id", handler.Get)
		req, _ := http.NewRequest(http.MethodGet, databasesPath+dbPathInvalid, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("NotFound", func(t *testing.T) {
		svc, handler, r := setupDatabaseHandlerTest(t)
		r.GET(databasesPath+"/:id", handler.Get)
		id := uuid.New()
		svc.On("GetDatabase", mock.Anything, id).Return(nil, errors.New(errors.NotFound, errNotFound))
		req, _ := http.NewRequest(http.MethodGet, databasesPath+"/"+id.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
		svc.AssertExpectations(t)
	})
}

func TestDatabaseHandlerDeleteError(t *testing.T) {
	t.Run("InvalidID", func(t *testing.T) {
		_, handler, r := setupDatabaseHandlerTest(t)
		r.DELETE(databasesPath+"/:id", handler.Delete)
		req, _ := http.NewRequest(http.MethodDelete, databasesPath+dbPathInvalid, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ServiceError", func(t *testing.T) {
		svc, handler, r := setupDatabaseHandlerTest(t)
		r.DELETE(databasesPath+"/:id", handler.Delete)
		id := uuid.New()
		svc.On("DeleteDatabase", mock.Anything, id).Return(errors.New(errors.Internal, "error"))
		req, _ := http.NewRequest(http.MethodDelete, databasesPath+"/"+id.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})
}

func TestDatabaseHandlerGetConnectionStringError(t *testing.T) {
	t.Run("InvalidID", func(t *testing.T) {
		_, handler, r := setupDatabaseHandlerTest(t)
		r.GET(databasesPath+connPath, handler.GetConnectionString)
		req, _ := http.NewRequest(http.MethodGet, databasesPath+dbPathInvalid+connSuffix, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ServiceError", func(t *testing.T) {
		svc, handler, r := setupDatabaseHandlerTest(t)
		r.GET(databasesPath+connPath, handler.GetConnectionString)
		id := uuid.New()
		svc.On("GetConnectionString", mock.Anything, id).Return("", errors.New(errors.Internal, "error"))
		req, _ := http.NewRequest(http.MethodGet, databasesPath+"/"+id.String()+connSuffix, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})
}
