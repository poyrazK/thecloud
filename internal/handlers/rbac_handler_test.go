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
	rolesPath    = "/rbac/roles"
	bindPath     = "/rbac/bind"
	testRoleName = "test-role"
	adminRole    = "admin"
)

type mockRBACService struct {
	mock.Mock
}

func (m *mockRBACService) Authorize(ctx context.Context, userID uuid.UUID, permission domain.Permission) error {
	args := m.Called(ctx, userID, permission)
	return args.Error(0)
}
func (m *mockRBACService) HasPermission(ctx context.Context, userID uuid.UUID, permission domain.Permission) (bool, error) {
	args := m.Called(ctx, userID, permission)
	return args.Bool(0), args.Error(1)
}
func (m *mockRBACService) CreateRole(ctx context.Context, role *domain.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}
func (m *mockRBACService) GetRoleByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Role), args.Error(1)
}
func (m *mockRBACService) GetRoleByName(ctx context.Context, name string) (*domain.Role, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Role), args.Error(1)
}
func (m *mockRBACService) ListRoles(ctx context.Context) ([]*domain.Role, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Role), args.Error(1)
}
func (m *mockRBACService) UpdateRole(ctx context.Context, role *domain.Role) error {
	return m.Called(ctx, role).Error(0)
}
func (m *mockRBACService) DeleteRole(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockRBACService) AddPermissionToRole(ctx context.Context, roleID uuid.UUID, permission domain.Permission) error {
	args := m.Called(ctx, roleID, permission)
	return args.Error(0)
}
func (m *mockRBACService) RemovePermissionFromRole(ctx context.Context, roleID uuid.UUID, permission domain.Permission) error {
	return m.Called(ctx, roleID, permission).Error(0)
}
func (m *mockRBACService) BindRole(ctx context.Context, userIdentifier string, roleName string) error {
	args := m.Called(ctx, userIdentifier, roleName)
	return args.Error(0)
}
func (m *mockRBACService) ListRoleBindings(ctx context.Context) ([]*domain.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.User), args.Error(1)
}

func setupRBACHandlerTest(t *testing.T) (*mockRBACService, *RBACHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := new(mockRBACService)
	handler := NewRBACHandler(svc)
	r := gin.New()
	return svc, handler, r
}

func TestRBACHandlerCreateRole(t *testing.T) {
	svc, handler, r := setupRBACHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.POST(rolesPath, handler.CreateRole)

	req := CreateRoleRequest{
		Name:        testRoleName,
		Description: "A test role",
		Permissions: []domain.Permission{domain.PermissionInstanceRead},
	}
	body, err := json.Marshal(req)
	assert.NoError(t, err)

	svc.On("CreateRole", mock.Anything, mock.MatchedBy(func(r *domain.Role) bool {
		return r.Name == req.Name && len(r.Permissions) == 1
	})).Return(nil)

	w := httptest.NewRecorder()
	httpReq, err := http.NewRequest("POST", rolesPath, bytes.NewBuffer(body))
	assert.NoError(t, err)
	r.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestRBACHandlerListRoles(t *testing.T) {
	svc, handler, r := setupRBACHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.GET(rolesPath, handler.ListRoles)

	roles := []*domain.Role{
		{ID: uuid.New(), Name: adminRole},
		{ID: uuid.New(), Name: "viewer"},
	}

	svc.On("ListRoles", mock.Anything).Return(roles, nil)

	w := httptest.NewRecorder()
	httpReq, err := http.NewRequest("GET", rolesPath, nil)
	assert.NoError(t, err)
	r.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), adminRole)
	assert.Contains(t, w.Body.String(), "viewer")
}

func TestRBACHandlerGetRoleByID(t *testing.T) {
	svc, handler, r := setupRBACHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.GET(rolesPath+"/:id", handler.GetRole)

	roleID := uuid.New()
	role := &domain.Role{ID: roleID, Name: adminRole}

	svc.On("GetRoleByID", mock.Anything, roleID).Return(role, nil)

	w := httptest.NewRecorder()
	httpReq, err := http.NewRequest("GET", rolesPath+"/"+roleID.String(), nil)
	assert.NoError(t, err)
	r.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), adminRole)
}

func TestRBACHandlerGetRoleByName(t *testing.T) {
	svc, handler, r := setupRBACHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.GET(rolesPath+"/:id", handler.GetRole)

	roleName := "developer"
	role := &domain.Role{ID: uuid.New(), Name: roleName}

	svc.On("GetRoleByName", mock.Anything, roleName).Return(role, nil)

	w := httptest.NewRecorder()
	httpReq, err := http.NewRequest("GET", rolesPath+"/"+roleName, nil)
	assert.NoError(t, err)
	r.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), roleName)
}

func TestRBACHandlerDeleteRole(t *testing.T) {
	svc, handler, r := setupRBACHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.DELETE(rolesPath+"/:id", handler.DeleteRole)

	roleID := uuid.New()

	svc.On("DeleteRole", mock.Anything, roleID).Return(nil)

	w := httptest.NewRecorder()
	httpReq, err := http.NewRequest("DELETE", rolesPath+"/"+roleID.String(), nil)
	assert.NoError(t, err)
	r.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestRBACHandlerAddPermission(t *testing.T) {
	svc, handler, r := setupRBACHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.POST(rolesPath+"/:id/permissions", handler.AddPermission)

	roleID := uuid.New()

	// Create a minimal request body with just the Permission field
	body, err := json.Marshal(map[string]interface{}{
		"permission": domain.PermissionInstanceLaunch,
	})
	assert.NoError(t, err)

	svc.On("AddPermissionToRole", mock.Anything, roleID, domain.PermissionInstanceLaunch).Return(nil)

	w := httptest.NewRecorder()
	httpReq, err := http.NewRequest("POST", rolesPath+"/"+roleID.String()+"/permissions", bytes.NewBuffer(body))
	assert.NoError(t, err)
	r.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestRBACHandlerBindRole(t *testing.T) {
	svc, handler, r := setupRBACHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.POST(bindPath, handler.BindRole)

	userEmail := "user@example.com"
	roleName := adminRole

	body, err := json.Marshal(map[string]string{
		"user_identifier": userEmail,
		"role_name":       roleName,
	})
	assert.NoError(t, err)

	svc.On("BindRole", mock.Anything, userEmail, roleName).Return(nil)

	w := httptest.NewRecorder()
	httpReq, err := http.NewRequest("POST", bindPath, bytes.NewBuffer(body))
	assert.NoError(t, err)
	r.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusNoContent, w.Code)
}
