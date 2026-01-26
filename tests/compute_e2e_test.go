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

func TestComputeE2E(t *testing.T) {
	if err := waitForServer(); err != nil {
		t.Skipf("Skipping Compute E2E test: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	token := registerAndLogin(t, client, "compute-tester@thecloud.local", "Compute Tester")

	var instanceID string
	instanceName := fmt.Sprintf("e2e-inst-%d", time.Now().UnixNano()%1000)

	// 1. Launch Instance
	t.Run("LaunchInstance", func(t *testing.T) {
		payload := map[string]string{
			"name":  instanceName,
			"image": "nginx:alpine",
			"ports": "80:80",
		}
		resp := postRequest(t, client, testutil.TestBaseURL+"/instances", token, payload)
		defer resp.Body.Close()

		require.Equal(t, http.StatusAccepted, resp.StatusCode)

		var res struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		instanceID = res.Data.ID
		assert.NotEmpty(t, instanceID)
	})

	// 2. Get Instance Details
	t.Run("GetInstance", func(t *testing.T) {
		resp := getRequest(t, client, fmt.Sprintf("%s/instances/%s", testutil.TestBaseURL, instanceID), token)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res struct {
			Data domain.Instance `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		assert.Equal(t, instanceName, res.Data.Name)
	})

	// 3. List Instances
	t.Run("ListInstances", func(t *testing.T) {
		resp := getRequest(t, client, testutil.TestBaseURL+"/instances", token)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res struct {
			Data []domain.Instance `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))

		found := false
		for _, inst := range res.Data {
			if inst.ID.String() == instanceID {
				found = true
				break
			}
		}
		assert.True(t, found, "instance not found in list")
	})

	// 4. Get Logs
	t.Run("GetLogs", func(t *testing.T) {
		// Might be empty initially but endpoint should work
		resp := getRequest(t, client, fmt.Sprintf("%s/instances/%s/logs", testutil.TestBaseURL, instanceID), token)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// 5. Stop Instance
	t.Run("StopInstance", func(t *testing.T) {
		resp := postRequest(t, client, fmt.Sprintf("%s/instances/%s/stop", testutil.TestBaseURL, instanceID), token, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// 6. Terminate Instance
	t.Run("TerminateInstance", func(t *testing.T) {
		resp := deleteRequest(t, client, fmt.Sprintf("%s/instances/%s", testutil.TestBaseURL, instanceID), token)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
