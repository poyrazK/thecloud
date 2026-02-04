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
	tenantsPath       = "/tenants"
	tenantsPrefix     = "/tenants/"
	tenantMembersPath = "/tenants/:id/members"
	tenantSwitchPath  = "/tenants/:id/switch"
)

type mockTenantService struct {
	mock.Mock
}

func (m *mockTenantService) CreateTenant(ctx context.Context, name, slug string, ownerID uuid.UUID) (*domain.Tenant, error) {
	args := m.Called(ctx, name, slug, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Tenant), args.Error(1)
}
func (m *mockTenantService) GetTenant(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Tenant), args.Error(1)
}
func (m *mockTenantService) InviteMember(ctx context.Context, tenantID uuid.UUID, email, role string) error {
	return m.Called(ctx, tenantID, email, role).Error(0)
}
func (m *mockTenantService) RemoveMember(ctx context.Context, tenantID, userID uuid.UUID) error {
	return m.Called(ctx, tenantID, userID).Error(0)
}
func (m *mockTenantService) SwitchTenant(ctx context.Context, userID, tenantID uuid.UUID) error {
	return m.Called(ctx, userID, tenantID).Error(0)
}
func (m *mockTenantService) CheckQuota(ctx context.Context, tenantID uuid.UUID, resource string, requested int) error {
	return m.Called(ctx, tenantID, resource, requested).Error(0)
}
func (m *mockTenantService) GetMembership(ctx context.Context, tenantID, userID uuid.UUID) (*domain.TenantMember, error) {
	args := m.Called(ctx, tenantID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TenantMember), args.Error(1)
}

func setupTenantHandlerTest(userID uuid.UUID) (*mockTenantService, *TenantHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := new(mockTenantService)
	handler := NewTenantHandler(svc)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	})
	return svc, handler, r
}

func TestTenantHandlerCreate(t *testing.T) {
	t.Parallel()
	userID := uuid.New()
	svc, handler, r := setupTenantHandlerTest(userID)

	r.POST(tenantsPath, handler.Create)

	reqBody := CreateTenantRequest{
		Name: "My Tenant",
		Slug: "my-tenant",
	}
	body, _ := json.Marshal(reqBody)

	expectedTenant := &domain.Tenant{
		ID:      uuid.New(),
		Name:    reqBody.Name,
		Slug:    reqBody.Slug,
		OwnerID: userID,
	}

	svc.On("CreateTenant", mock.Anything, reqBody.Name, reqBody.Slug, userID).Return(expectedTenant, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", tenantsPath, bytes.NewBuffer(body))
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "My Tenant")
}

func TestTenantHandlerInviteMember(t *testing.T) {
	t.Parallel()
	userID := uuid.New()
	svc, handler, r := setupTenantHandlerTest(userID)

	r.POST(tenantMembersPath, handler.InviteMember)

	tenantID := uuid.New()
	reqBody := InviteMemberRequest{
		Email: "test@example.com",
		Role:  "member",
	}
	body, _ := json.Marshal(reqBody)

	svc.On("InviteMember", mock.Anything, tenantID, reqBody.Email, reqBody.Role).Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", tenantsPrefix+tenantID.String()+"/members", bytes.NewBuffer(body))
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "member invited")
}

func TestTenantHandlerSwitchTenant(t *testing.T) {
	t.Parallel()
	userID := uuid.New()
	svc, handler, r := setupTenantHandlerTest(userID)

	r.POST(tenantSwitchPath, handler.SwitchTenant)

	tenantID := uuid.New()

	svc.On("SwitchTenant", mock.Anything, userID, tenantID).Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", tenantsPrefix+tenantID.String()+"/switch", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "tenant switched")
}

func TestTenantHandlerCreateInvalidInput(t *testing.T) {
	t.Parallel()
	userID := uuid.New()
	_, handler, r := setupTenantHandlerTest(userID)

	r.POST(tenantsPath, handler.Create)

	body := []byte(`{"invalid": "json"}`) // Missing required fields

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", tenantsPath, bytes.NewBuffer(body))
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTenantHandlerCreateServiceError(t *testing.T) {
	t.Parallel()
	userID := uuid.New()
	svc, handler, r := setupTenantHandlerTest(userID)

	r.POST(tenantsPath, handler.Create)

	reqBody := CreateTenantRequest{Name: "T", Slug: "t"}
	body, _ := json.Marshal(reqBody)

	svc.On("CreateTenant", mock.Anything, reqBody.Name, reqBody.Slug, userID).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", tenantsPath, bytes.NewBuffer(body))
	r.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusCreated, w.Code)
	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestTenantHandlerInviteMemberInvalidID(t *testing.T) {
	t.Parallel()
	userID := uuid.New()
	_, handler, r := setupTenantHandlerTest(userID)

	r.POST(tenantMembersPath, handler.InviteMember)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", tenantsPrefix+"invalid-uuid/members", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTenantHandlerInviteMemberServiceError(t *testing.T) {
	t.Parallel()
	userID := uuid.New()
	svc, handler, r := setupTenantHandlerTest(userID)

	r.POST(tenantMembersPath, handler.InviteMember)

	tenantID := uuid.New()
	reqBody := InviteMemberRequest{Email: "e", Role: "r"}
	body, _ := json.Marshal(reqBody)

	svc.On("InviteMember", mock.Anything, tenantID, reqBody.Email, reqBody.Role).Return(assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", tenantsPrefix+tenantID.String()+"/members", bytes.NewBuffer(body))
	r.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestTenantHandlerSwitchTenantInvalidID(t *testing.T) {
	t.Parallel()
	userID := uuid.New()
	_, handler, r := setupTenantHandlerTest(userID)

	r.POST(tenantSwitchPath, handler.SwitchTenant)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", tenantsPrefix+"invalid-uuid/switch", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTenantHandlerSwitchTenantServiceError(t *testing.T) {
	t.Parallel()
	userID := uuid.New()
	svc, handler, r := setupTenantHandlerTest(userID)

	r.POST(tenantSwitchPath, handler.SwitchTenant)

	tenantID := uuid.New()
	svc.On("SwitchTenant", mock.Anything, userID, tenantID).Return(assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", tenantsPrefix+tenantID.String()+"/switch", nil)
	r.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusOK, w.Code)
}
