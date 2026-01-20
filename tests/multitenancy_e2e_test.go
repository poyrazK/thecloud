// Package tests contains end-to-end and integration tests for the platform.
package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/poyrazk/thecloud/pkg/testutil"
)

const instancesPathFmt = "%s/instances/%s"

// baseURL is now usually defined by constant, but let's assume we use the one from testutil
// const baseURL = "http://localhost:8080"

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	APIKey string `json:"api_key"`
}

type ResponseWrapper struct {
	Data AuthResponse `json:"data"`
}

type Instance struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func TestMultiTenancyE2E(t *testing.T) {
	// Skip if server is not reachable (e.g., in CI without live services)
	if err := waitForServer(); err != nil {
		t.Skipf("Skipping E2E test: %v (server at %s not available)", err, testutil.TestBaseURL)
	}

	client := &http.Client{Timeout: 5 * time.Second}

	// 1. Register and Login User A
	tokenA := registerAndLogin(t, client, "userA@test.com", "User A")

	// 2. Register and Login User B
	tokenB := registerAndLogin(t, client, "userB@test.com", "User B")

	// 3. User A Creates an Instance
	instA := createInstance(t, client, tokenA, "inst-a")
	assert.NotEmpty(t, instA.ID)

	t.Run("User B cannot see User A's instance in List", func(t *testing.T) {
		listB := listInstances(t, client, tokenB)
		for _, inst := range listB {
			assert.NotEqual(t, instA.ID, inst.ID, "User B should not see User A's instance")
		}
	})

	t.Run("User B cannot Get User A's instance", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf(instancesPathFmt, testutil.TestBaseURL, instA.ID), nil)
		req.Header.Set(testutil.TestHeaderAPIKey, tokenB)
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Expect 404 (Not Found) or 403 (Forbidden).
		// Since repo filters by user_id, it likely returns Not Found (like it doesn't exist for them).
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("User A can see their instance", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf(instancesPathFmt, testutil.TestBaseURL, instA.ID), nil)
		req.Header.Set(testutil.TestHeaderAPIKey, tokenA)
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Cleanup
	deleteInstance(t, client, tokenA, instA.ID)
}

func waitForServer() error {
	for i := 0; i < 30; i++ {
		resp, err := http.Get(testutil.TestBaseURL + "/health")
		if err == nil && resp.StatusCode == 200 {
			_ = resp.Body.Close()
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("server not ready")
}

func registerAndLogin(t *testing.T, client *http.Client, email, name string) string {
	// Register
	regReq := RegisterRequest{Email: email, Password: testutil.TestPasswordStrong, Name: name}
	body, _ := json.Marshal(regReq)
	resp, err := client.Post(testutil.TestBaseURL+"/auth/register", testutil.TestContentTypeAppJSON, bytes.NewBuffer(body))
	if err == nil {
		_ = resp.Body.Close()
	}
	// Ignore error if already registered, proceed to login

	// Login
	loginReq := LoginRequest{Email: email, Password: testutil.TestPasswordStrong}
	body, _ = json.Marshal(loginReq)
	resp, err = client.Post(testutil.TestBaseURL+"/auth/login", testutil.TestContentTypeAppJSON, bytes.NewBuffer(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Login failed for %s: status %d", email, resp.StatusCode)
	}

	var authResp ResponseWrapper
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&authResp))
	return authResp.Data.APIKey
}

func createInstance(t *testing.T, client *http.Client, token, name string) Instance {
	reqBody := map[string]string{
		"name":  name,
		"image": "alpine",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", testutil.TestBaseURL+"/instances", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", testutil.TestContentTypeAppJSON)
	req.Header.Set(testutil.TestHeaderAPIKey, token)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Create instance failed: status %d body %s", resp.StatusCode, body)
	}

	// Response is wrapped in data
	type InstanceWrapper struct {
		Data Instance `json:"data"`
	}
	var instWrapper InstanceWrapper
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&instWrapper))
	return instWrapper.Data
}

func listInstances(t *testing.T, client *http.Client, token string) []Instance {
	req, _ := http.NewRequest("GET", testutil.TestBaseURL+"/instances", nil)
	req.Header.Set(testutil.TestHeaderAPIKey, token)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	type ListWrapper struct {
		Data []Instance `json:"data"`
	}
	var listWrapper ListWrapper
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&listWrapper))
	return listWrapper.Data
}

func deleteInstance(_ *testing.T, client *http.Client, token, id string) {
	req, _ := http.NewRequest("DELETE", fmt.Sprintf(instancesPathFmt, testutil.TestBaseURL, id), nil)
	req.Header.Set(testutil.TestHeaderAPIKey, token)
	_, _ = client.Do(req)
}
