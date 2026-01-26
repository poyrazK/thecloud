package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/pkg/testutil"
)

func TestRBACE2E(t *testing.T) {
	if err := waitForServer(); err != nil {
		t.Skipf("Skipping RBAC E2E test: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	mainUserEmail := fmt.Sprintf("rbac-tester-%d@thecloud.local", time.Now().UnixNano()%10000)
	token := registerAndLogin(t, client, mainUserEmail, "RBAC Tester")

	var roleID string
	roleName := fmt.Sprintf("e2e-role-%d", time.Now().UnixNano())

	// 1. Create Role
	t.Run("CreateRole", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":        roleName,
			"description": "E2E test role",
			"permissions": []domain.Permission{domain.PermissionInstanceRead, domain.PermissionVolumeRead},
		}
		resp := postRequest(t, client, testutil.TestBaseURL+"/rbac/roles", token, payload)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var res struct {
			Data domain.Role `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		roleID = res.Data.ID.String()
		assert.NotEmpty(t, roleID)
	})

	// 2. Get Role
	t.Run("GetRole", func(t *testing.T) {
		resp := getRequest(t, client, fmt.Sprintf("%s/rbac/roles/%s", testutil.TestBaseURL, roleID), token)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res struct {
			Data domain.Role `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		assert.Equal(t, roleName, res.Data.Name)
	})

	// 3. Add Permission
	t.Run("AddPermission", func(t *testing.T) {
		payload := map[string]interface{}{
			"permission": domain.PermissionVpcRead,
		}
		resp := postRequest(t, client, fmt.Sprintf("%s/rbac/roles/%s/permissions", testutil.TestBaseURL, roleID), token, payload)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	// 4. Bind Role to a second User
	t.Run("BindRole", func(t *testing.T) {
		// Register a second user to bind the role to, so we don't lose permissions for the main user
		secondUserEmail := fmt.Sprintf("rbac-secondary-%d@thecloud.local", time.Now().UnixNano()%10000)
		_ = registerAndLogin(t, client, secondUserEmail, "RBAC Secondary")

		payload := map[string]string{
			"user_identifier": secondUserEmail,
			"role_name":       roleName,
		}
		// Still use the primary token which has management permissions
		resp := postRequest(t, client, testutil.TestBaseURL+"/rbac/bindings", token, payload)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	// 5. List Bindings
	t.Run("ListBindings", func(t *testing.T) {
		resp := getRequest(t, client, testutil.TestBaseURL+"/rbac/bindings", token)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// 6. Cleanup
	t.Run("Cleanup", func(t *testing.T) {
		resp := deleteRequest(t, client, fmt.Sprintf("%s/rbac/roles/%s", testutil.TestBaseURL, roleID), token)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})
}
