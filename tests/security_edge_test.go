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

const instancesPath = "/instances"

func TestSecurityEdge(t *testing.T) {
	t.Parallel()
	if err := waitForServer(); err != nil {
		t.Fatalf("Failing Security Edge test: %v", err)
	}

	client := &http.Client{Timeout: 5 * time.Second}

	t.Run("Authentication Bypass Attempts", func(t *testing.T) {
		// 1. Missing API Key
		req, _ := http.NewRequest("GET", testutil.TestBaseURL+instancesPath, nil)
		resp, err := client.Do(req)
		if err == nil && resp != nil {
			defer func() { _ = resp.Body.Close() }()
			assert.Contains(t, []int{http.StatusUnauthorized, http.StatusForbidden}, resp.StatusCode)
		} else {
			assert.NoError(t, err)
		}

		// 2. Invalid API Key format
		req, _ = http.NewRequest("GET", testutil.TestBaseURL+instancesPath, nil)
		req.Header.Set(testutil.TestHeaderAPIKey, "invalid-key-format-123")
		resp, err = client.Do(req)
		if err == nil && resp != nil {
			defer func() { _ = resp.Body.Close() }()
			assert.Contains(t, []int{http.StatusUnauthorized, http.StatusForbidden}, resp.StatusCode)
		} else {
			assert.NoError(t, err)
		}
	})

	t.Run("Cross-Tenant Access Attempt", func(t *testing.T) {
		// 1. Register User A and create a resource
		tokenA := registerAndLogin(t, client, "userA-sec@thecloud.local", "User A Sec")
		payload := map[string]string{
			"name":  "user-a-instance",
			"image": "alpine",
		}
		respA := postRequest(t, client, testutil.TestBaseURL+instancesPath, tokenA, payload)
		defer func() { _ = respA.Body.Close() }()

		type Wrapper struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		var w Wrapper
		_ = json.NewDecoder(respA.Body).Decode(&w)
		instID := w.Data.ID

		// 2. Register User B and try to access User A's resource
		tokenB := registerAndLogin(t, client, "userB-sec@thecloud.local", "User B Sec")
		reqB, _ := http.NewRequest("GET", fmt.Sprintf("%s%s/%s", testutil.TestBaseURL, instancesPath, instID), nil)
		reqB.Header.Set(testutil.TestHeaderAPIKey, tokenB)
		applyTenantHeader(t, reqB, tokenB)
		respB, err := client.Do(reqB)
		if err == nil && respB != nil {
			defer func() { _ = respB.Body.Close() }()
			// Should be 404 (preferred for security to avoid ID discovery) or 403
			assert.Contains(t, []int{http.StatusNotFound, http.StatusForbidden}, respB.StatusCode)
		} else {
			assert.NoError(t, err)
		}
	})
}
