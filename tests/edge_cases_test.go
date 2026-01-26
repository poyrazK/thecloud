package tests

import (
	"net/http"
	"testing"
	"time"

	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/poyrazk/thecloud/tests/helpers"
	"github.com/stretchr/testify/assert"
)

func TestEdgeCases(t *testing.T) {
	if err := waitForServer(); err != nil {
		t.Skipf("Skipping Edge Cases test: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	token := registerAndLogin(t, client, "edge-tester@thecloud.local", "Edge Tester")

	t.Run("Empty/Null Inputs", func(t *testing.T) {
		// 1. Create instance with empty name
		payload := map[string]string{
			"name":  "",
			"image": "alpine",
		}
		resp := postRequest(t, client, testutil.TestBaseURL+"/instances", token, payload)
		defer resp.Body.Close()
		helpers.AssertErrorCode(t, resp, http.StatusBadRequest)

		// 2. Create VPC with empty name
		vpcPayload := map[string]string{
			"name":       "",
			"cidr_block": "10.0.0.0/16",
		}
		resp = postRequest(t, client, testutil.TestBaseURL+"/vpcs", token, vpcPayload)
		defer resp.Body.Close()
		helpers.AssertErrorCode(t, resp, http.StatusBadRequest)
	})

	t.Run("Boundary Values", func(t *testing.T) {
		// 1. Instance name exactly 255 chars
		longName := ""
		for i := 0; i < 255; i++ {
			longName += "a"
		}
		payload := map[string]string{
			"name":  longName,
			"image": "alpine",
		}
		resp := postRequest(t, client, testutil.TestBaseURL+"/instances", token, payload)
		defer resp.Body.Close()
		// Depending on implementation, 255 might be OK or redirected to async launch (202)
		assert.Contains(t, []int{http.StatusCreated, http.StatusAccepted}, resp.StatusCode)

		// 2. Instance name 256 chars (overflow)
		tooLongName := longName + "a"
		payload["name"] = tooLongName
		resp = postRequest(t, client, testutil.TestBaseURL+"/instances", token, payload)
		defer resp.Body.Close()
		helpers.AssertErrorCode(t, resp, http.StatusBadRequest)
	})

	t.Run("Special Characters", func(t *testing.T) {
		// 1. Unicode in instance name
		payload := map[string]string{
			"name":  "test-🚀-instance",
			"image": "alpine",
		}
		resp := postRequest(t, client, testutil.TestBaseURL+"/instances", token, payload)
		defer resp.Body.Close()
		assert.Contains(t, []int{http.StatusCreated, http.StatusAccepted}, resp.StatusCode)

		// 2. SQL Injection attempt in bucket name
		bucketPayload := map[string]string{
			"name": "bucket'; DROP TABLE users;--",
		}
		resp = postRequest(t, client, testutil.TestBaseURL+"/storage/buckets", token, bucketPayload)
		defer resp.Body.Close()

		// If it returns 500, it might be vulnerable to SQL injection or at least unhandled database error
		if resp.StatusCode == http.StatusInternalServerError {
			t.Errorf("VULNERABILITY: SQL Injection attempt returned 500 Internal Server Error (Request ID: %s)", resp.Header.Get("X-Request-ID"))
		} else {
			assert.NotEqual(t, http.StatusInternalServerError, resp.StatusCode, "Should not return 500 for special characters")
		}
	})

	t.Run("Malformed JSON", func(t *testing.T) {
		resp := helpers.SendMalformedJSON(t, client, testutil.TestBaseURL+"/instances", "POST", token)
		defer resp.Body.Close()
		helpers.AssertErrorCode(t, resp, http.StatusBadRequest)
	})
}
