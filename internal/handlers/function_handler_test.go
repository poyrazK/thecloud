package httphandlers

import (
	"bytes"
	"context"
	"mime/multipart"
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
	functionsPath    = "/functions"
	testFunctionName = "fn-1"
	invokeSuffix     = "/invoke"
	logsSuffix       = "/logs"
	hdrContentType   = "Content-Type"
	fnPathInvalid      = "/invalid"
)

type mockFunctionService struct {
	mock.Mock
}

func (m *mockFunctionService) CreateFunction(ctx context.Context, name, runtime, handler string, code []byte) (*domain.Function, error) {
	args := m.Called(ctx, name, runtime, handler, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Function), args.Error(1)
}

func (m *mockFunctionService) ListFunctions(ctx context.Context) ([]*domain.Function, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Function), args.Error(1)
}

func (m *mockFunctionService) GetFunction(ctx context.Context, id uuid.UUID) (*domain.Function, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Function), args.Error(1)
}

func (m *mockFunctionService) DeleteFunction(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockFunctionService) InvokeFunction(ctx context.Context, id uuid.UUID, payload []byte, async bool) (*domain.Invocation, error) {
	args := m.Called(ctx, id, payload, async)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Invocation), args.Error(1)
}

func (m *mockFunctionService) GetFunctionLogs(ctx context.Context, id uuid.UUID, limit int) ([]*domain.Invocation, error) {
	args := m.Called(ctx, id, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Invocation), args.Error(1)
}

func setupFunctionHandlerTest(_ *testing.T) (*mockFunctionService, *FunctionHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := new(mockFunctionService)
	handler := NewFunctionHandler(svc)
	r := gin.New()
	return svc, handler, r
}

func TestFunctionHandlerCreate(t *testing.T) {
	svc, handler, r := setupFunctionHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.POST(functionsPath, handler.Create)

	fn := &domain.Function{ID: uuid.New(), Name: testFunctionName}
	svc.On("CreateFunction", mock.Anything, testFunctionName, "nodejs18", "index.handler", []byte("code")).Return(fn, nil)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	err := writer.WriteField("name", testFunctionName)
	assert.NoError(t, err)
	err = writer.WriteField("runtime", "nodejs18")
	assert.NoError(t, err)
	err = writer.WriteField("handler", "index.handler")
	assert.NoError(t, err)
	part, err := writer.CreateFormFile("code", "index.js")
	assert.NoError(t, err)
	_, err = part.Write([]byte("code"))
	assert.NoError(t, err)
	err = writer.Close()
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, functionsPath, body)
	assert.NoError(t, err)
	req.Header.Set(hdrContentType, writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestFunctionHandlerList(t *testing.T) {
	svc, handler, r := setupFunctionHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.GET(functionsPath, handler.List)

	fns := []*domain.Function{{ID: uuid.New(), Name: testFunctionName}}
	svc.On("ListFunctions", mock.Anything).Return(fns, nil)

	req, err := http.NewRequest(http.MethodGet, functionsPath, nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFunctionHandlerGet(t *testing.T) {
	svc, handler, r := setupFunctionHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.GET(functionsPath+"/:id", handler.Get)

	id := uuid.New()
	fn := &domain.Function{ID: id, Name: testFunctionName}
	svc.On("GetFunction", mock.Anything, id).Return(fn, nil)

	req, err := http.NewRequest(http.MethodGet, functionsPath+"/"+id.String(), nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFunctionHandlerDelete(t *testing.T) {
	svc, handler, r := setupFunctionHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.DELETE(functionsPath+"/:id", handler.Delete)

	id := uuid.New()
	svc.On("DeleteFunction", mock.Anything, id).Return(nil)

	req, err := http.NewRequest(http.MethodDelete, functionsPath+"/"+id.String(), nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFunctionHandlerInvoke(t *testing.T) {
	svc, handler, r := setupFunctionHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.POST(functionsPath+"/:id"+invokeSuffix, handler.Invoke)

	id := uuid.New()
	inv := &domain.Invocation{ID: uuid.New(), Status: "completed"}
	svc.On("InvokeFunction", mock.Anything, id, []byte("{}"), false).Return(inv, nil)

	req, err := http.NewRequest(http.MethodPost, functionsPath+"/"+id.String()+invokeSuffix, strings.NewReader("{}"))
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFunctionHandlerGetLogs(t *testing.T) {
	svc, handler, r := setupFunctionHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.GET(functionsPath+"/:id"+logsSuffix, handler.GetLogs)

	id := uuid.New()
	svc.On("GetFunctionLogs", mock.Anything, id, 100).Return([]*domain.Invocation{}, nil)

	req, err := http.NewRequest(http.MethodGet, functionsPath+"/"+id.String()+logsSuffix, nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFunctionHandlerCreateError(t *testing.T) {
	t.Run("InvalidForm", func(t *testing.T) {
		_, handler, r := setupFunctionHandlerTest(t)
		r.POST(functionsPath, handler.Create)
		req, _ := http.NewRequest("POST", functionsPath, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("MissingFile", func(t *testing.T) {
		_, handler, r := setupFunctionHandlerTest(t)
		r.POST(functionsPath, handler.Create)
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("name", "n")
		_ = writer.WriteField("runtime", "r")
		_ = writer.WriteField("handler", "h")
		_ = writer.Close()
		req, _ := http.NewRequest("POST", functionsPath, body)
		req.Header.Set(hdrContentType, writer.FormDataContentType())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ServiceError", func(t *testing.T) {
		svc, handler, r := setupFunctionHandlerTest(t)
		r.POST(functionsPath, handler.Create)
		svc.On("CreateFunction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil, errors.New(errors.Internal, "error"))

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("name", "n")
		_ = writer.WriteField("runtime", "r")
		_ = writer.WriteField("handler", "h")
		part, _ := writer.CreateFormFile("code", "index.js")
		_, _ = part.Write([]byte("code"))
		_ = writer.Close()

		req, _ := http.NewRequest("POST", functionsPath, body)
		req.Header.Set(hdrContentType, writer.FormDataContentType())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})
}

func TestFunctionHandlerListError(t *testing.T) {
	svc, handler, r := setupFunctionHandlerTest(t)
	r.GET(functionsPath, handler.List)
	svc.On("ListFunctions", mock.Anything).Return(nil, errors.New(errors.Internal, "error"))
	req, _ := http.NewRequest(http.MethodGet, functionsPath, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

func TestFunctionHandlerGetError(t *testing.T) {
	t.Run("InvalidID", func(t *testing.T) {
		_, handler, r := setupFunctionHandlerTest(t)
		r.GET(functionsPath+"/:id", handler.Get)
		req, _ := http.NewRequest(http.MethodGet, functionsPath+fnPathInvalid, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("NotFound", func(t *testing.T) {
		svc, handler, r := setupFunctionHandlerTest(t)
		r.GET(functionsPath+"/:id", handler.Get)
		id := uuid.New()
		svc.On("GetFunction", mock.Anything, id).Return(nil, errors.New(errors.NotFound, errNotFound))
		req, _ := http.NewRequest(http.MethodGet, functionsPath+"/"+id.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
		svc.AssertExpectations(t)
	})
}

func TestFunctionHandlerInvokeError(t *testing.T) {
	t.Run("InvalidID", func(t *testing.T) {
		_, handler, r := setupFunctionHandlerTest(t)
		r.POST(functionsPath+"/:id"+invokeSuffix, handler.Invoke)
		req, _ := http.NewRequest("POST", functionsPath+fnPathInvalid+invokeSuffix, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ServiceError", func(t *testing.T) {
		svc, handler, r := setupFunctionHandlerTest(t)
		r.POST(functionsPath+"/:id"+invokeSuffix, handler.Invoke)
		id := uuid.New()
		svc.On("InvokeFunction", mock.Anything, id, mock.Anything, false).Return(nil, errors.New(errors.Internal, "error"))
		req, _ := http.NewRequest("POST", functionsPath+"/"+id.String()+invokeSuffix, strings.NewReader("{}"))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("AsyncSuccess", func(t *testing.T) {
		svc, handler, r := setupFunctionHandlerTest(t)
		r.POST(functionsPath+"/:id"+invokeSuffix, handler.Invoke)
		id := uuid.New()
		inv := &domain.Invocation{ID: uuid.New(), Status: "accepted"}
		svc.On("InvokeFunction", mock.Anything, id, mock.Anything, true).Return(inv, nil)
		req, _ := http.NewRequest("POST", functionsPath+"/"+id.String()+invokeSuffix+"?async=true", strings.NewReader("{}"))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusAccepted, w.Code)
		svc.AssertExpectations(t)
	})
}

func TestFunctionHandlerGetLogsError(t *testing.T) {
	t.Run("InvalidID", func(t *testing.T) {
		_, handler, r := setupFunctionHandlerTest(t)
		r.GET(functionsPath+"/:id"+logsSuffix, handler.GetLogs)
		req, _ := http.NewRequest(http.MethodGet, functionsPath+fnPathInvalid+logsSuffix, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ServiceError", func(t *testing.T) {
		svc, handler, r := setupFunctionHandlerTest(t)
		r.GET(functionsPath+"/:id"+logsSuffix, handler.GetLogs)
		id := uuid.New()
		svc.On("GetFunctionLogs", mock.Anything, id, mock.Anything).Return(nil, errors.New(errors.Internal, "error"))
		req, _ := http.NewRequest(http.MethodGet, functionsPath+"/"+id.String()+logsSuffix, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})
}

func TestFunctionHandlerDeleteError(t *testing.T) {
	t.Run("InvalidID", func(t *testing.T) {
		_, handler, r := setupFunctionHandlerTest(t)
		r.DELETE(functionsPath+"/:id", handler.Delete)
		req, _ := http.NewRequest(http.MethodDelete, functionsPath+fnPathInvalid, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ServiceError", func(t *testing.T) {
		svc, handler, r := setupFunctionHandlerTest(t)
		r.DELETE(functionsPath+"/:id", handler.Delete)
		id := uuid.New()
		svc.On("DeleteFunction", mock.Anything, id).Return(errors.New(errors.Internal, "error"))
		req, _ := http.NewRequest(http.MethodDelete, functionsPath+"/"+id.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})
}
