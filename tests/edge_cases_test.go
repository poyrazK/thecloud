package tests

import (
	"bytes"
	"net/http"
	"testing"
	"time"

	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/poyrazk/thecloud/tests/helpers"
	"github.com/stretchr/testify/assert"
)

func TestEdgeCases(t *testing.T) {
	t.Parallel()
	if err := waitForServer(); err != nil {
		t.Fatalf("Failing Edge Cases test: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	token := registerAndLogin(t, client, "edge-tester@thecloud.local", "Edge Tester")

	t.Run("Empty/Null Inputs", func(t *testing.T) {
		// 1. Create instance with empty name
		payload := map[string]string{
			"name":  "",
			"image": "alpine",
		}
		resp := postRequest(t, client, testutil.TestBaseURL+testutil.TestRouteInstances, token, payload)
		defer func() { _ = resp.Body.Close() }()
		helpers.AssertErrorCode(t, resp, http.StatusBadRequest)

		// 2. Create VPC with empty name
		vpcPayload := map[string]string{
			"name":       "",
			"cidr_block": "10.0.0.0/16",
		}
		resp = postRequest(t, client, testutil.TestBaseURL+testutil.TestRouteVpcs, token, vpcPayload)
		defer func() { _ = resp.Body.Close() }()
		helpers.AssertErrorCode(t, resp, http.StatusBadRequest)
	})

	t.Run("Boundary Values", func(t *testing.T) {
		// 1. Instance name exactly 64 chars (max allowed)
		longName := ""
		for i := 0; i < 64; i++ {
			longName += "a"
		}
		payload := map[string]string{
			"name":  longName,
			"image": "alpine",
		}
		resp := postRequest(t, client, testutil.TestBaseURL+testutil.TestRouteInstances, token, payload)
		defer func() { _ = resp.Body.Close() }()
		assert.Contains(t, []int{http.StatusCreated, http.StatusAccepted}, resp.StatusCode)

		// 2. Instance name 65 chars (overflow)
		tooLongName := longName + "a"
		payload["name"] = tooLongName
		resp = postRequest(t, client, testutil.TestBaseURL+testutil.TestRouteInstances, token, payload)
		defer func() { _ = resp.Body.Close() }()
		helpers.AssertErrorCode(t, resp, http.StatusBadRequest)
	})

	t.Run("Special Characters", func(t *testing.T) {
		// 1. Unicode in instance name (should fail validation)
		payload := map[string]string{
			"name":  "test-🚀-instance",
			"image": "alpine",
		}
		resp := postRequest(t, client, testutil.TestBaseURL+testutil.TestRouteInstances, token, payload)
		defer func() { _ = resp.Body.Close() }()
		helpers.AssertErrorCode(t, resp, http.StatusBadRequest)

		// 2. SQL Injection attempt in bucket name
		bucketPayload := map[string]string{
			"name": "bucket'; DROP TABLE users;--",
		}
		resp = postRequest(t, client, testutil.TestBaseURL+testutil.TestRouteStorageBuckets, token, bucketPayload)
		defer func() { _ = resp.Body.Close() }()

		// If it returns 500, it might be vulnerable to SQL injection or at least unhandled database error
		if resp.StatusCode == http.StatusInternalServerError {
			t.Errorf("VULNERABILITY: SQL Injection attempt returned 500 Internal Server Error (Request ID: %s)", resp.Header.Get("X-Request-ID"))
		} else {
			assert.NotEqual(t, http.StatusInternalServerError, resp.StatusCode, "Should not return 500 for special characters")
		}
	})

	t.Run("Malformed JSON", func(t *testing.T) {
		malformedJSON := `{"name": "test", "incomplete": `
		req, _ := http.NewRequest("POST", testutil.TestBaseURL+testutil.TestRouteInstances, bytes.NewBufferString(malformedJSON))
		req.Header.Set("Content-Type", testutil.TestContentTypeAppJSON)
		req.Header.Set(testutil.TestHeaderAPIKey, token)
		applyTenantHeader(t, req, token)
		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
