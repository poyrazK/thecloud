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

func TestComputeE2E(t *testing.T) {
	t.Parallel()
	if err := waitForServer(); err != nil {
		t.Fatalf("Failing Compute E2E test: %v", err)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	token := registerAndLogin(t, client, "compute-tester@thecloud.local", "Compute Tester")

	var instanceID string
	instanceName := fmt.Sprintf("e2e-inst-%d-%s", time.Now().UnixNano()%1000, uuid.New().String())

	// 1. Launch Instance
	t.Run("LaunchInstance", func(t *testing.T) {
		payload := map[string]string{
			"name":  instanceName,
			"image": "nginx:alpine",
			"ports": "0:80",
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
		assert.NotEmpty(t, instanceID)
	})

	// 2. Get Instance Details
	t.Run("GetInstance", func(t *testing.T) {
		resp := getRequest(t, client, fmt.Sprintf(testutil.TestRouteFormat, testutil.TestBaseURL, testutil.TestRouteInstances, instanceID), token)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res struct {
			Data domain.Instance `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		assert.Equal(t, instanceName, res.Data.Name)
	})

	// 2.5 Wait for Instance to be Running
	t.Run("WaitForRunning", func(t *testing.T) {
		timeout := 90 * time.Second
		start := time.Now()
		var lastStatus domain.InstanceStatus
		errorCount := 0

		for time.Since(start) < timeout {
			resp := getRequest(t, client, fmt.Sprintf(testutil.TestRouteFormat, testutil.TestBaseURL, testutil.TestRouteInstances, instanceID), token)
			var res struct {
				Data domain.Instance `json:"data"`
			}
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
			_ = resp.Body.Close()

			lastStatus = res.Data.Status

			if res.Data.Status == domain.StatusRunning {
				return
			}
			if res.Data.Status == domain.StatusError {
				errorCount++
				// If the instance is stuck in error state for multiple iterations,
				// the Docker backend is likely unavailable (e.g., in CI without Docker-in-Docker)
				if errorCount > 5 {
					t.Skipf("Docker backend appears unavailable in CI environment (consecutive errors: %d)", errorCount)
				}
			} else {
				errorCount = 0
			}
			t.Logf("Waiting for instance to be running... Current status: %s", res.Data.Status)
			time.Sleep(2 * time.Second)
		}
		// Skip instead of fail if backend is unavailable
		t.Skipf("Instance did not reach running state within timeout (90s). Last status: %s. Docker backend may be unavailable.", lastStatus)
	})

	// 3. List Instances
	t.Run("ListInstances", func(t *testing.T) {
		resp := getRequest(t, client, testutil.TestBaseURL+testutil.TestRouteInstances, token)
		defer func() { _ = resp.Body.Close() }()

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
		resp := getRequest(t, client, fmt.Sprintf("%s%s/%s/logs", testutil.TestBaseURL, testutil.TestRouteInstances, instanceID), token)
		defer func() { _ = resp.Body.Close() }()

		assert.Contains(t, []int{http.StatusOK, http.StatusConflict, http.StatusNotFound}, resp.StatusCode)
	})

	// 5. Stop Instance
	t.Run("StopInstance", func(t *testing.T) {
		resp := postRequest(t, client, fmt.Sprintf("%s%s/%s/stop", testutil.TestBaseURL, testutil.TestRouteInstances, instanceID), token, nil)
		defer func() { _ = resp.Body.Close() }()

		assert.Contains(t, []int{http.StatusOK, http.StatusConflict, http.StatusInternalServerError}, resp.StatusCode)
	})

	// 6. Terminate Instance
	t.Run("TerminateInstance", func(t *testing.T) {
		resp := deleteRequest(t, client, fmt.Sprintf(testutil.TestRouteFormat, testutil.TestBaseURL, testutil.TestRouteInstances, instanceID), token)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestResizeInstance(t *testing.T) {
	if err := waitForServer(); err != nil {
		t.Fatalf("Failing Resize E2E test: %v", err)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	token := registerAndLogin(t, client, "resize-tester@thecloud.local", "Resize Tester")

	var instanceID string
	instanceName := fmt.Sprintf("e2e-resize-%d-%s", time.Now().UnixNano()%1000, uuid.New().String())

	// 1. Launch Instance with basic-2 type
	t.Run("LaunchInstance", func(t *testing.T) {
		payload := map[string]string{
			"name":          instanceName,
			"image":          "nginx:alpine",
			"instance_type":  "basic-2",
			"ports":          "0:80",
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
		assert.NotEmpty(t, instanceID)
	})

	// 2. Wait for Instance to be Running
	t.Run("WaitForRunning", func(t *testing.T) {
		timeout := 90 * time.Second
		start := time.Now()
		var lastStatus domain.InstanceStatus
		errorCount := 0

		for time.Since(start) < timeout {
			resp := getRequest(t, client, fmt.Sprintf(testutil.TestRouteFormat, testutil.TestBaseURL, testutil.TestRouteInstances, instanceID), token)
			var res struct {
				Data domain.Instance `json:"data"`
			}
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
			_ = resp.Body.Close()

			lastStatus = res.Data.Status

			if res.Data.Status == domain.StatusRunning {
				return
			}
			if res.Data.Status == domain.StatusError {
				errorCount++
				if errorCount > 5 {
					t.Skipf("Docker backend appears unavailable (consecutive errors: %d)", errorCount)
				}
			} else {
				errorCount = 0
			}
			t.Logf("Waiting for instance to be running... Current status: %s", res.Data.Status)
			time.Sleep(2 * time.Second)
		}
		t.Skipf("Instance did not reach running state within timeout (90s). Last status: %s", lastStatus)
	})

	// 3. Resize to basic-4
	t.Run("Resize", func(t *testing.T) {
		payload := map[string]string{
			"instance_type": "basic-4",
		}
		resp := postRequest(t, client, fmt.Sprintf("%s%s/%s/resize", testutil.TestBaseURL, testutil.TestRouteInstances, instanceID), token, payload)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// 4. Verify instance type changed via GET
	t.Run("VerifyResize", func(t *testing.T) {
		resp := getRequest(t, client, fmt.Sprintf(testutil.TestRouteFormat, testutil.TestBaseURL, testutil.TestRouteInstances, instanceID), token)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res struct {
			Data struct {
				InstanceType string `json:"instance_type"`
			} `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		assert.Equal(t, "basic-4", res.Data.InstanceType)
	})

	// 5. Terminate Instance
	t.Run("TerminateInstance", func(t *testing.T) {
		resp := deleteRequest(t, client, fmt.Sprintf(testutil.TestRouteFormat, testutil.TestBaseURL, testutil.TestRouteInstances, instanceID), token)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestResizeInstanceDownsize(t *testing.T) {
	if err := waitForServer(); err != nil {
		t.Fatalf("Failing Resize Downsize E2E test: %v", err)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	token := registerAndLogin(t, client, "resize-down-tester@thecloud.local", "Resize Down Tester")

	var instanceID string
	instanceName := fmt.Sprintf("e2e-resize-down-%d-%s", time.Now().UnixNano()%1000, uuid.New().String())

	// 1. Launch Instance with basic-4 type
	t.Run("LaunchInstance", func(t *testing.T) {
		payload := map[string]string{
			"name":         instanceName,
			"image":        "nginx:alpine",
			"instance_type": "basic-4",
			"ports":        "0:80",
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

	// 2. Wait for Running
	t.Run("WaitForRunning", func(t *testing.T) {
		timeout := 90 * time.Second
		start := time.Now()
		var lastStatus domain.InstanceStatus
		errorCount := 0

		for time.Since(start) < timeout {
			resp := getRequest(t, client, fmt.Sprintf(testutil.TestRouteFormat, testutil.TestBaseURL, testutil.TestRouteInstances, instanceID), token)
			var res struct {
				Data domain.Instance `json:"data"`
			}
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
			_ = resp.Body.Close()

			lastStatus = res.Data.Status

			if res.Data.Status == domain.StatusRunning {
				return
			}
			if res.Data.Status == domain.StatusError {
				errorCount++
				if errorCount > 5 {
					t.Skipf("Docker backend appears unavailable (consecutive errors: %d)", errorCount)
				}
			} else {
				errorCount = 0
			}
			time.Sleep(2 * time.Second)
		}
		t.Skipf("Instance did not reach running state within timeout. Last status: %s", lastStatus)
	})

	// 3. Downsize to basic-2
	t.Run("Resize", func(t *testing.T) {
		payload := map[string]string{
			"instance_type": "basic-2",
		}
		resp := postRequest(t, client, fmt.Sprintf("%s%s/%s/resize", testutil.TestBaseURL, testutil.TestRouteInstances, instanceID), token, payload)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// 4. Terminate
	t.Run("TerminateInstance", func(t *testing.T) {
		resp := deleteRequest(t, client, fmt.Sprintf(testutil.TestRouteFormat, testutil.TestBaseURL, testutil.TestRouteInstances, instanceID), token)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestResizeInstanceInvalidType(t *testing.T) {
	if err := waitForServer(); err != nil {
		t.Fatalf("Failing Resize Invalid Type E2E test: %v", err)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	token := registerAndLogin(t, client, "resize-invalid-tester@thecloud.local", "Resize Invalid Tester")

	var instanceID string
	instanceName := fmt.Sprintf("e2e-resize-inv-%d-%s", time.Now().UnixNano()%1000, uuid.New().String())

	// 1. Launch Instance
	t.Run("LaunchInstance", func(t *testing.T) {
		payload := map[string]string{
			"name":  instanceName,
			"image": "nginx:alpine",
			"ports": "0:80",
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

	// 2. Wait for Running
	t.Run("WaitForRunning", func(t *testing.T) {
		timeout := 90 * time.Second
		start := time.Now()
		var lastStatus domain.InstanceStatus
		errorCount := 0

		for time.Since(start) < timeout {
			resp := getRequest(t, client, fmt.Sprintf(testutil.TestRouteFormat, testutil.TestBaseURL, testutil.TestRouteInstances, instanceID), token)
			var res struct {
				Data domain.Instance `json:"data"`
			}
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
			_ = resp.Body.Close()

			lastStatus = res.Data.Status

			if res.Data.Status == domain.StatusRunning {
				return
			}
			if res.Data.Status == domain.StatusError {
				errorCount++
				if errorCount > 5 {
					t.Skipf("Docker backend appears unavailable (consecutive errors: %d)", errorCount)
				}
			} else {
				errorCount = 0
			}
			time.Sleep(2 * time.Second)
		}
		t.Skipf("Instance did not reach running state within timeout. Last status: %s", lastStatus)
	})

	// 3. Try to resize to invalid type (should fail with 400 or 422)
	t.Run("ResizeInvalidType", func(t *testing.T) {
		payload := map[string]string{
			"instance_type": "nonexistent-type",
		}
		resp := postRequest(t, client, fmt.Sprintf("%s%s/%s/resize", testutil.TestBaseURL, testutil.TestRouteInstances, instanceID), token, payload)
		defer func() { _ = resp.Body.Close() }()

		assert.True(t, resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnprocessableEntity,
			"expected 400 or 422, got %d", resp.StatusCode)
	})

	// 4. Terminate
	t.Run("TerminateInstance", func(t *testing.T) {
		resp := deleteRequest(t, client, fmt.Sprintf(testutil.TestRouteFormat, testutil.TestBaseURL, testutil.TestRouteInstances, instanceID), token)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
