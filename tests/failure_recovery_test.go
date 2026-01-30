package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/assert"
)

func TestFailureRecovery(t *testing.T) {
	t.Parallel()
	if err := waitForServer(); err != nil {
		t.Fatalf("Failing Failure Recovery test: %v", err)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	token := registerAndLogin(t, client, "failure-tester@thecloud.local", "Failure Tester")

	t.Run("Delete while Launching", func(t *testing.T) {
		// Launch instance
		payload := map[string]string{
			"name":  "del-while-launch",
			"image": "alpine",
		}
		resp := postRequest(t, client, testutil.TestBaseURL+"/instances", token, payload)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusAccepted, resp.StatusCode)

		// I'll use the existing helper pattern:
		type Wrapper struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		var w Wrapper
		_ = json.NewDecoder(resp.Body).Decode(&w)
		instID := w.Data.ID

		// Immediately delete it
		delResp := deleteRequest(t, client, fmt.Sprintf("%s/instances/%s", testutil.TestBaseURL, instID), token)
		defer func() { _ = delResp.Body.Close() }()

		// Should succeed (200/204) or handle gracefully
		assert.Contains(t, []int{http.StatusOK, http.StatusNoContent, http.StatusAccepted}, delResp.StatusCode)
	})

	t.Run("Invalid Service Endpoint", func(t *testing.T) {
		// Test a hypothetical endpoint that might be broken
		resp := getRequest(t, client, testutil.TestBaseURL+"/non-existent-service", token)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
