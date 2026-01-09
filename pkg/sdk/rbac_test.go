package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

func TestClient_CreateRole(t *testing.T) {
	expectedRole := domain.Role{
		ID:          uuid.New(),
		Name:        "test-role",
		Description: "test description",
		Permissions: []domain.Permission{domain.Permission("perm-1")},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rbac/roles", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "test-api-key", r.Header.Get("X-API-Key"))

		var req map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, expectedRole.Name, req["name"])
		assert.Equal(t, expectedRole.Description, req["description"])

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedRole)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	role, err := client.CreateRole(expectedRole.Name, expectedRole.Description, expectedRole.Permissions)

	assert.NoError(t, err)
	assert.NotNil(t, role)
	assert.Equal(t, expectedRole.ID, role.ID)
	assert.Equal(t, expectedRole.Name, role.Name)
}

func TestClient_ListRoles(t *testing.T) {
	expectedRoles := []domain.Role{
		{ID: uuid.New(), Name: "role-1"},
		{ID: uuid.New(), Name: "role-2"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rbac/roles", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedRoles)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	roles, err := client.ListRoles()

	assert.NoError(t, err)
	assert.Len(t, roles, 2)
	assert.Equal(t, expectedRoles[0].Name, roles[0].Name)
}

func TestClient_GetRole(t *testing.T) {
	id := uuid.New()
	expectedRole := domain.Role{ID: id, Name: "test-role"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rbac/roles/"+id.String(), r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedRole)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	role, err := client.GetRole(id.String())

	assert.NoError(t, err)
	assert.NotNil(t, role)
	assert.Equal(t, expectedRole.ID, role.ID)
}

func TestClient_DeleteRole(t *testing.T) {
	id := uuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rbac/roles/"+id.String(), r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	err := client.DeleteRole(id)

	assert.NoError(t, err)
}

func TestClient_UpdateRole(t *testing.T) {
	id := uuid.New()
	expectedRole := domain.Role{
		ID:          id,
		Name:        "updated-role",
		Description: "updated description",
		Permissions: []domain.Permission{domain.Permission("perm-2")},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rbac/roles/"+id.String(), r.URL.Path)
		assert.Equal(t, http.MethodPut, r.Method)

		var req map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, expectedRole.Name, req["name"])

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedRole)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	role, err := client.UpdateRole(id, expectedRole.Name, expectedRole.Description, expectedRole.Permissions)

	assert.NoError(t, err)
	assert.NotNil(t, role)
	assert.Equal(t, expectedRole.Name, role.Name)
}

func TestClient_BindRole(t *testing.T) {
	userIdentifier := "user@example.com"
	roleName := "admin"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rbac/bindings", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var req map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, userIdentifier, req["user_identifier"])
		assert.Equal(t, roleName, req["role_name"])

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	err := client.BindRole(userIdentifier, roleName)

	assert.NoError(t, err)
}

func TestClient_ListRoleBindings(t *testing.T) {
	expectedUsers := []domain.User{
		{ID: uuid.New(), Email: "user1@example.com"},
		{ID: uuid.New(), Email: "user2@example.com"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rbac/bindings", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedUsers)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	users, err := client.ListRoleBindings()

	assert.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, expectedUsers[0].Email, users[0].Email)
}
