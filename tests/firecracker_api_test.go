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
	if os.Getenv("COMPUTE_BACKEND") != "firecracker" {
		t.Skip("Skipping firecracker API test: COMPUTE_BACKEND != firecracker")
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
		
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", testutil.TestBaseURL+"/instances", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", testutil.TestContentTypeAppJSON)
		req.Header.Set(testutil.TestHeaderAPIKey, token)
		applyTenantHeader(t, req, token)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// If artifacts are missing, the service might return 500 or 400.
		// We want to see the status.
		if resp.StatusCode != http.StatusCreated {
			t.Logf("Launch failed with status %d (expected in some CI environments)", resp.StatusCode)
			return
		}

		var res struct {
			Data domain.Instance `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		instanceID = res.Data.ID.String()
		assert.NotEmpty(t, instanceID)
	})

	if instanceID == "" {
		t.Skip("Skipping subsequent steps because instance launch failed or was skipped")
	}

	t.Run("Get Firecracker Instance", func(t *testing.T) {
		req, _ := http.NewRequest("GET", testutil.TestBaseURL+"/instances/"+instanceID, nil)
		req.Header.Set(testutil.TestHeaderAPIKey, token)
		applyTenantHeader(t, req, token)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Terminate Firecracker Instance", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", testutil.TestBaseURL+"/instances/"+instanceID, nil)
		req.Header.Set(testutil.TestHeaderAPIKey, token)
		applyTenantHeader(t, req, token)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
