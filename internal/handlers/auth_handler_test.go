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
	testEmail    = "test@example.com"
	testPassword = "password123"
	testName     = "Test User"
	registerPath = "/auth/register"
	loginPath    = "/auth/login"
)

type mockAuthService struct {
	mock.Mock
}

func (m *mockAuthService) Register(ctx context.Context, email, password, name string) (*domain.User, error) {
	args := m.Called(ctx, email, password, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockAuthService) Login(ctx context.Context, email, password string) (*domain.User, string, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).(*domain.User), args.String(1), args.Error(2)
}

func (m *mockAuthService) ValidateUser(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

type mockPasswordResetService struct {
	mock.Mock
}

func (m *mockPasswordResetService) RequestReset(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *mockPasswordResetService) ResetPassword(ctx context.Context, token, newPassword string) error {
	args := m.Called(ctx, token, newPassword)
	return args.Error(0)
}

func setupAuthHandlerTest(_ *testing.T) (*mockAuthService, *mockPasswordResetService, *AuthHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := new(mockAuthService)
	pwdSvc := new(mockPasswordResetService)
	handler := NewAuthHandler(svc, pwdSvc)
	r := gin.New()
	return svc, pwdSvc, handler, r
}

func TestAuthHandlerRegister(t *testing.T) {
	t.Parallel()
	svc, _, handler, r := setupAuthHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.POST(registerPath, handler.Register)

	user := &domain.User{ID: uuid.New(), Email: testEmail}
	svc.On("Register", mock.Anything, testEmail, testPassword, testName).Return(user, nil)

	body, err := json.Marshal(map[string]string{
		"email":    testEmail,
		"password": testPassword,
		"name":     testName,
	})
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", registerPath, bytes.NewBuffer(body))
	assert.NoError(t, err)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestAuthHandlerRegisterInvalidJSON(t *testing.T) {
	t.Parallel()
	svc, _, handler, r := setupAuthHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.POST(registerPath, handler.Register)

	req := httptest.NewRequest(http.MethodPost, registerPath, bytes.NewBufferString("{bad"))
	req.Header.Set(contentType, applicationJSON)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "Register", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestAuthHandlerRegisterInvalidInputFromService(t *testing.T) {
	t.Parallel()
	svc, _, handler, r := setupAuthHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.POST(registerPath, handler.Register)

	body, err := json.Marshal(map[string]string{
		"email":    testEmail,
		"password": testPassword,
		"name":     "Test User",
	})
	assert.NoError(t, err)

	svc.On("Register", mock.Anything, testEmail, testPassword, testName).Return(nil, errors.New(errors.InvalidInput, "duplicate"))

	req := httptest.NewRequest(http.MethodPost, registerPath, bytes.NewBuffer(body))
	req.Header.Set(contentType, applicationJSON)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandlerLogin(t *testing.T) {
	t.Parallel()
	svc, _, handler, r := setupAuthHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.POST(loginPath, handler.Login)

	user := &domain.User{ID: uuid.New(), Email: testEmail}
	svc.On("Login", mock.Anything, testEmail, testPassword).Return(user, "key123", nil)

	body, err := json.Marshal(map[string]string{
		"email":    testEmail,
		"password": testPassword,
	})
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", loginPath, bytes.NewBuffer(body))
	assert.NoError(t, err)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "key123")
}

func TestAuthHandlerLoginInvalidJSON(t *testing.T) {
	t.Parallel()
	svc, _, handler, r := setupAuthHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.POST(loginPath, handler.Login)

	req := httptest.NewRequest(http.MethodPost, loginPath, bytes.NewBufferString("{oops"))
	req.Header.Set(contentType, applicationJSON)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "Login", mock.Anything, mock.Anything, mock.Anything)
}

func TestAuthHandlerLoginInvalidCredentials(t *testing.T) {
	t.Parallel()
	svc, _, handler, r := setupAuthHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.POST(loginPath, handler.Login)

	body, err := json.Marshal(map[string]string{
		"email":    testEmail,
		"password": testPassword,
	})
	assert.NoError(t, err)

	svc.On("Login", mock.Anything, testEmail, testPassword).Return(nil, "", errors.New(errors.Unauthorized, "invalid credentials"))

	req := httptest.NewRequest(http.MethodPost, loginPath, bytes.NewBuffer(body))
	req.Header.Set(contentType, applicationJSON)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandlerForgotPassword(t *testing.T) {
	t.Parallel()
	_, pwdSvc, handler, r := setupAuthHandlerTest(t)
	defer pwdSvc.AssertExpectations(t)

	r.POST("/auth/forgot-password", handler.ForgotPassword)

	pwdSvc.On("RequestReset", mock.Anything, testEmail).Return(nil)

	body, _ := json.Marshal(map[string]string{
		"email": testEmail,
	})

	req := httptest.NewRequest("POST", "/auth/forgot-password", bytes.NewBuffer(body))
	req.Header.Set(contentType, applicationJSON)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthHandlerResetPassword(t *testing.T) {
	t.Parallel()
	_, pwdSvc, handler, r := setupAuthHandlerTest(t)
	defer pwdSvc.AssertExpectations(t)

	r.POST("/auth/reset-password", handler.ResetPassword)

	token := "reset-token"
	newPwd := "newpass123"

	pwdSvc.On("ResetPassword", mock.Anything, token, newPwd).Return(nil)

	body, _ := json.Marshal(map[string]string{
		"token":        token,
		"new_password": newPwd,
	})

	req := httptest.NewRequest("POST", "/auth/reset-password", bytes.NewBuffer(body))
	req.Header.Set(contentType, applicationJSON)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
