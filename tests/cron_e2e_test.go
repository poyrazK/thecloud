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

func TestCronE2E(t *testing.T) {
	t.Parallel()
	if err := waitForServer(); err != nil {
		t.Fatalf("Failing Cron E2E test: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	token := registerAndLogin(t, client, "cron-tester@thecloud.local", "Cron Tester")

	var jobID string
	jobName := fmt.Sprintf("e2e-job-%d", time.Now().UnixNano()%1000)

	// 1. Create Job
	t.Run("CreateJob", func(t *testing.T) {
		payload := map[string]string{
			"name":          jobName,
			"schedule":      "*/5 * * * *",
			"target_url":    "http://localhost:8080/health",
			"target_method": "GET",
		}
		resp := postRequest(t, client, testutil.TestBaseURL+"/cron/jobs", token, payload)
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var res struct {
			Data domain.CronJob `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		jobID = res.Data.ID.String()
		assert.NotEmpty(t, jobID)
	})

	// 2. Get Job
	t.Run("GetJob", func(t *testing.T) {
		resp := getRequest(t, client, fmt.Sprintf("%s/cron/jobs/%s", testutil.TestBaseURL, jobID), token)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res struct {
			Data domain.CronJob `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		assert.Equal(t, jobName, res.Data.Name)
	})

	// 3. Pause Job
	t.Run("PauseJob", func(t *testing.T) {
		resp := postRequest(t, client, fmt.Sprintf("%s/cron/jobs/%s/pause", testutil.TestBaseURL, jobID), token, nil)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// 4. Resume Job
	t.Run("ResumeJob", func(t *testing.T) {
		resp := postRequest(t, client, fmt.Sprintf("%s/cron/jobs/%s/resume", testutil.TestBaseURL, jobID), token, nil)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// 5. Cleanup
	t.Run("Cleanup", func(t *testing.T) {
		resp := deleteRequest(t, client, fmt.Sprintf("%s/cron/jobs/%s", testutil.TestBaseURL, jobID), token)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
