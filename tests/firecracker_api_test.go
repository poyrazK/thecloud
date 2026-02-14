package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFirecrackerAPI_E2E(t *testing.T) {
	// Only run this if we are in an environment that supports firecracker 
	// or explicitly enabled for CI.
	backend := os.Getenv("COMPUTE_BACKEND")
	if backend != "firecracker" && backend != "firecracker-mock" {
		t.Skipf("Skipping firecracker API test: COMPUTE_BACKEND (%s) is not firecracker or firecracker-mock", backend)
	}

	if err := waitForServer(); err != nil {
		t.Fatalf("Server not ready: %v", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	token := registerAndLogin(t, client, "fc-admin@thecloud.local", "FC Admin")

	var instanceID string

	t.Run("Launch Firecracker Instance", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":          "fc-e2e-vm",
			"image":         "alpine",
			"instance_type": "basic-2", // Matches defaults in setup_test.go
		}
		
		body, err := json.Marshal(payload)
		require.NoError(t, err, "failed to marshal payload")

		req, err := http.NewRequest("POST", testutil.TestBaseURL+"/instances", bytes.NewBuffer(body))
		require.NoError(t, err, "failed to create request")

		req.Header.Set("Content-Type", testutil.TestContentTypeAppJSON)
		req.Header.Set(testutil.TestHeaderAPIKey, token)
		applyTenantHeader(t, req, token)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// status 202 is expected for async provisioning
		if resp.StatusCode != http.StatusAccepted {
			t.Fatalf("Launch failed with status %d (expected 202)", resp.StatusCode)
		}

		var res struct {
			Data domain.Instance `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		instanceID = res.Data.ID.String()
		assert.NotEmpty(t, instanceID)

		// Wait for provision worker to finish (especially in mock mode)
		time.Sleep(5 * time.Second)
	})

	if instanceID == "" {
		t.Skip("Skipping subsequent steps because instance launch failed or was skipped")
	}

	t.Run("Get Firecracker Instance", func(t *testing.T) {
		req, err := http.NewRequest("GET", testutil.TestBaseURL+"/instances/"+instanceID, nil)
		require.NoError(t, err, "failed to create request")

		req.Header.Set(testutil.TestHeaderAPIKey, token)
		applyTenantHeader(t, req, token)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res struct {
			Data domain.Instance `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		// In mock mode, the status should transition to RUNNING quickly
		assert.Equal(t, domain.StatusRunning, res.Data.Status)
	})

	t.Run("Terminate Firecracker Instance", func(t *testing.T) {
		req, err := http.NewRequest("DELETE", testutil.TestBaseURL+"/instances/"+instanceID, nil)
		require.NoError(t, err, "failed to create request")

		req.Header.Set(testutil.TestHeaderAPIKey, token)
		applyTenantHeader(t, req, token)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
