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

type mockSecretService struct {
	mock.Mock
}

func (m *mockSecretService) CreateSecret(ctx context.Context, name, value, description string) (*domain.Secret, error) {
	args := m.Called(ctx, name, value, description)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Secret), args.Error(1)
}

func (m *mockSecretService) ListSecrets(ctx context.Context) ([]*domain.Secret, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Secret), args.Error(1)
}

func (m *mockSecretService) GetSecret(ctx context.Context, id uuid.UUID) (*domain.Secret, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Secret), args.Error(1)
}

func (m *mockSecretService) GetSecretByName(ctx context.Context, name string) (*domain.Secret, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Secret), args.Error(1)
}

func (m *mockSecretService) DeleteSecret(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupSecretHandlerTest(t *testing.T) (*mockSecretService, *SecretHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := new(mockSecretService)
	handler := NewSecretHandler(svc)
	r := gin.New()
	return svc, handler, r
}

func TestSecretHandler_Create(t *testing.T) {
	svc, handler, r := setupSecretHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.POST("/secrets", handler.Create)

	secret := &domain.Secret{ID: uuid.New(), Name: "sec-1"}
	svc.On("CreateSecret", mock.Anything, "sec-1", "value", "desc").Return(secret, nil)

	body, err := json.Marshal(map[string]interface{}{
		"name":        "sec-1",
		"value":       "value",
		"description": "desc",
	})
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/secrets", bytes.NewBuffer(body))
	assert.NoError(t, err)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestSecretHandler_List(t *testing.T) {
	svc, handler, r := setupSecretHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.GET("/secrets", handler.List)

	secrets := []*domain.Secret{{ID: uuid.New(), Name: "sec-1"}}
	svc.On("ListSecrets", mock.Anything).Return(secrets, nil)

	req, err := http.NewRequest(http.MethodGet, "/secrets", nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSecretHandler_Get_ByID(t *testing.T) {
	svc, handler, r := setupSecretHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.GET("/secrets/:id", handler.Get)

	id := uuid.New()
	secret := &domain.Secret{ID: id, Name: "sec-1"}
	svc.On("GetSecret", mock.Anything, id).Return(secret, nil)

	req, err := http.NewRequest(http.MethodGet, "/secrets/"+id.String(), nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSecretHandler_Get_ByName(t *testing.T) {
	svc, handler, r := setupSecretHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.GET("/secrets/:id", handler.Get)

	secret := &domain.Secret{ID: uuid.New(), Name: "sec-1"}
	svc.On("GetSecretByName", mock.Anything, "sec-1").Return(secret, nil)

	req, err := http.NewRequest(http.MethodGet, "/secrets/sec-1", nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSecretHandler_Delete(t *testing.T) {
	svc, handler, r := setupSecretHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.DELETE("/secrets/:id", handler.Delete)

	id := uuid.New()
	svc.On("DeleteSecret", mock.Anything, id).Return(nil)

	req, err := http.NewRequest(http.MethodDelete, "/secrets/"+id.String(), nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
