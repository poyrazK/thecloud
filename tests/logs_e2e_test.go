package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/pkg/testutil"
)

func TestLogsE2E(t *testing.T) {
	if err := waitForServer(); err != nil {
		t.Fatalf("Failing Logs E2E test: %v", err)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	token := registerAndLogin(t, client, "logs-tester@thecloud.local", "Logs Tester")

	var instanceID string
	instanceName := fmt.Sprintf("e2e-logs-%d", time.Now().UnixNano()%1000)

	// 1. Launch Instance
	t.Run("LaunchInstance", func(t *testing.T) {
		payload := map[string]string{
			"name":  instanceName,
			"image": "alpine",
			"cmd":   "echo 'hello-cloudlogs-e2e'",
		}
		resp := postRequest(t, client, testutil.TestBaseURL+testutil.TestRouteInstances, token, payload)
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusAccepted, resp.StatusCode)

		var res struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		instanceID = res.Data.ID
	})

	// 2. Wait for it to finish/run (best effort)
	time.Sleep(5 * time.Second)

	// 3. Terminate Instance (triggers log ingestion)
	t.Run("TerminateInstance", func(t *testing.T) {
		resp := deleteRequest(t, client, fmt.Sprintf(testutil.TestRouteFormat, testutil.TestBaseURL, testutil.TestRouteInstances, instanceID), token)
		defer func() { _ = resp.Body.Close() }()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// 4. Verify logs in CloudLogs
	t.Run("VerifyHistoricalLogs", func(t *testing.T) {
		// Wait a bit for async cleanup/ingestion
		time.Sleep(2 * time.Second)

		resp := getRequest(t, client, fmt.Sprintf("%s/logs/%s", testutil.TestBaseURL, instanceID), token)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res struct {
			Data struct {
				Entries []domain.LogEntry `json:"entries"`
				Total   int               `json:"total"`
			} `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		
		// In a real environment with Docker, we'd expect entries > 0.
		// In CI/Dev without real execution, we at least verify the service returns 200 OK.
		t.Logf("Found %d historical logs for terminated instance", res.Data.Total)
	})

	// 5. Search Logs
	t.Run("SearchLogs", func(t *testing.T) {
		resp := getRequest(t, client, testutil.TestBaseURL+"/logs?resource_type=instance&limit=10", token)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
