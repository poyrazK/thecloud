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

type mockStackService struct {
	mock.Mock
}

func (m *mockStackService) CreateStack(ctx context.Context, name, template string, parameters map[string]string) (*domain.Stack, error) {
	args := m.Called(ctx, name, template, parameters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Stack), args.Error(1)
}

func (m *mockStackService) ListStacks(ctx context.Context) ([]*domain.Stack, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Stack), args.Error(1)
}

func (m *mockStackService) GetStack(ctx context.Context, id uuid.UUID) (*domain.Stack, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Stack), args.Error(1)
}

func (m *mockStackService) DeleteStack(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockStackService) ValidateTemplate(ctx context.Context, template string) (*domain.TemplateValidateResponse, error) {
	args := m.Called(ctx, template)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TemplateValidateResponse), args.Error(1)
}

const (
	testStackName     = "test-stack"
	testStackTmpl     = "version: 1"
	testStackPath     = "/iac/stacks"
	testStackVPath    = "/iac/validate"
	testStackAppJSON  = "application/json"
	stackPathInvalid  = "invalid"
	headerContentType = "Content-Type"
)

func TestStackHandlerCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mockStackService)
	handler := NewStackHandler(svc)

	stack := &domain.Stack{ID: uuid.New(), Name: "test-stack"}

	svc.On("CreateStack", mock.Anything, testStackName, testStackTmpl, mock.Anything).Return(stack, nil)

	reqBody := CreateStackRequest{
		Name:     testStackName,
		Template: testStackTmpl,
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", testStackPath, bytes.NewBuffer(body))
	c.Request.Header.Set(headerContentType, testStackAppJSON)

	handler.Create(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	svc.AssertExpectations(t)
}

func TestStackHandlerList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mockStackService)
	handler := NewStackHandler(svc)

	stacks := []*domain.Stack{
		{ID: uuid.New(), Name: "stack-1"},
	}

	svc.On("ListStacks", mock.Anything).Return(stacks, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", testStackPath, nil)

	handler.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestStackHandlerGet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mockStackService)
	handler := NewStackHandler(svc)

	id := uuid.New()
	stack := &domain.Stack{ID: id, Name: testStackName}

	svc.On("GetStack", mock.Anything, id).Return(stack, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", testStackPath+"/"+id.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: id.String()}}

	handler.Get(c)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestStackHandlerDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mockStackService)
	handler := NewStackHandler(svc)

	id := uuid.New()

	svc.On("DeleteStack", mock.Anything, id).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("DELETE", testStackPath+"/"+id.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: id.String()}}

	handler.Delete(c)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestStackHandlerValidate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mockStackService)
	handler := NewStackHandler(svc)

	resp := &domain.TemplateValidateResponse{Valid: true}

	svc.On("ValidateTemplate", mock.Anything, testStackTmpl).Return(resp, nil)

	reqBody := map[string]string{
		"template": testStackTmpl,
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", testStackVPath, bytes.NewBuffer(body))
	c.Request.Header.Set(headerContentType, testStackAppJSON)

	handler.Validate(c)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestStackHandlerErrorPaths(t *testing.T) {
	setup := func(_ *testing.T) (*mockStackService, *StackHandler, *gin.Engine) {
		svc := new(mockStackService)
		handler := NewStackHandler(svc)
		r := gin.New()
		return svc, handler, r
	}

	t.Run("CreateInvalidJSON", func(t *testing.T) {
		_, handler, r := setup(t)
		r.POST(testStackPath, handler.Create)
		req, _ := http.NewRequest("POST", testStackPath, bytes.NewBufferString("invalid"))
		req.Header.Set(headerContentType, testStackAppJSON)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CreateServiceError", func(t *testing.T) {
		svc, handler, r := setup(t)
		r.POST(testStackPath, handler.Create)
		svc.On("CreateStack", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New(errors.Internal, "error"))
		body, _ := json.Marshal(map[string]interface{}{"name": "n", "template": "t"})
		req, _ := http.NewRequest("POST", testStackPath, bytes.NewBuffer(body))
		req.Header.Set(headerContentType, testStackAppJSON)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("ListServiceError", func(t *testing.T) {
		svc, handler, r := setup(t)
		r.GET(testStackPath, handler.List)
		svc.On("ListStacks", mock.Anything).Return(nil, errors.New(errors.Internal, "error"))
		req, _ := http.NewRequest("GET", testStackPath, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("GetInvalidID", func(t *testing.T) {
		_, handler, r := setup(t)
		r.GET(testStackPath+"/:id", handler.Get)
		req, _ := http.NewRequest("GET", testStackPath+"/"+stackPathInvalid, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("GetServiceError", func(t *testing.T) {
		svc, handler, r := setup(t)
		r.GET(testStackPath+"/:id", handler.Get)
		id := uuid.New()
		svc.On("GetStack", mock.Anything, id).Return(nil, errors.New(errors.Internal, "error"))
		req, _ := http.NewRequest("GET", testStackPath+"/"+id.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("DeleteInvalidID", func(t *testing.T) {
		_, handler, r := setup(t)
		r.DELETE(testStackPath+"/:id", handler.Delete)
		req, _ := http.NewRequest("DELETE", testStackPath+"/"+stackPathInvalid, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("DeleteServiceError", func(t *testing.T) {
		svc, handler, r := setup(t)
		r.DELETE(testStackPath+"/:id", handler.Delete)
		id := uuid.New()
		svc.On("DeleteStack", mock.Anything, id).Return(errors.New(errors.Internal, "error"))
		req, _ := http.NewRequest("DELETE", testStackPath+"/"+id.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("ValidateInvalidJSON", func(t *testing.T) {
		_, handler, r := setup(t)
		r.POST(testStackVPath, handler.Validate)
		req, _ := http.NewRequest("POST", testStackVPath, bytes.NewBufferString("invalid"))
		req.Header.Set(headerContentType, testStackAppJSON)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ValidateServiceError", func(t *testing.T) {
		svc, handler, r := setup(t)
		r.POST(testStackVPath, handler.Validate)
		svc.On("ValidateTemplate", mock.Anything, mock.Anything).Return(nil, errors.New(errors.Internal, "error"))
		body, _ := json.Marshal(map[string]interface{}{"template": "t"})
		req, _ := http.NewRequest("POST", testStackVPath, bytes.NewBuffer(body))
		req.Header.Set(headerContentType, testStackAppJSON)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
