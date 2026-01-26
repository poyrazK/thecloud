package tests

import (
	"io"
	"net/http"
	"os/exec"
	"testing"
	"time"

	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChaos(t *testing.T) {
	if err := waitForServer(); err != nil {
		t.Skipf("Skipping Chaos test: %v", err)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	token := registerAndLogin(t, client, "chaos-tester@thecloud.local", "Chaos Tester")

	t.Run("Redis Failure Resilience", func(t *testing.T) {
		// 1. Verify system works initially
		resp := getRequest(t, client, testutil.TestBaseURL+"/instances", token)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()

		// 2. Kill Redis
		t.Log("Killing Redis container...")
		cmd := exec.Command("docker", "stop", "cloud-redis")
		err := cmd.Run()
		require.NoError(t, err, "Failed to stop Redis container")

		// Ensure Redis is stopped
		defer func() {
			t.Log("Restarting Redis container...")
			_ = exec.Command("docker", "start", "cloud-redis").Run()
			// Wait for Redis to be ready again
			time.Sleep(2 * time.Second)
		}()

		// 3. Verify API behavior without Redis
		resp = getRequest(t, client, testutil.TestBaseURL+"/instances", token)
		defer resp.Body.Close()

		t.Logf("API response without Redis: %d", resp.StatusCode)
		assert.NotEqual(t, 0, resp.StatusCode, "API should still respond")

		// 4. Verify Health Check
		respH, err := http.Get(testutil.TestBaseURL + "/health")
		if err == nil {
			defer respH.Body.Close()
			body, _ := io.ReadAll(respH.Body)
			t.Logf("Health check status without Redis: %d, body: %s", respH.StatusCode, string(body))
		}
	})

	t.Run("API Restart mid-operation", func(t *testing.T) {
		// 1. Start a goroutine that tries to reach the API
		recoveryChan := make(chan bool)
		errorCount := 0

		go func() {
			for i := 0; i < 60; i++ {
				resp, err := http.Get(testutil.TestBaseURL + "/health")
				if err == nil && resp.StatusCode == http.StatusOK {
					resp.Body.Close()
					if errorCount > 0 {
						// We recovered!
						recoveryChan <- true
						return
					}
				} else {
					errorCount++
				}
				time.Sleep(500 * time.Millisecond)
			}
			recoveryChan <- false
		}()

		// 2. Sudden Restart
		t.Log("Restarting API container mid-operation...")
		cmd := exec.Command("docker", "restart", "thecloud-api-1")
		err := cmd.Run()
		require.NoError(t, err, "Failed to restart API container")

		// 3. Wait for recovery
		t.Log("Waiting for API recovery...")
		select {
		case recovered := <-recoveryChan:
			assert.True(t, recovered, "API should have recovered after restart")
			t.Logf("API recovered after %d failed health checks", errorCount)
		case <-time.After(30 * time.Second):
			t.Fatal("Timeout waiting for API to recover")
		}

		// 4. Verify system is still functional (new request)
		token2 := registerAndLogin(t, client, "post-restart@thecloud.local", "Post Restart")
		resp := getRequest(t, client, testutil.TestBaseURL+testutil.TestRouteInstances, token2)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode, "API should be functional after restart")
	})
}
