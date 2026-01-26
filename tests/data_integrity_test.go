package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataIntegrity(t *testing.T) {
	if err := waitForServer(); err != nil {
		t.Skipf("Skipping Data Integrity test: %v", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	token := registerAndLogin(t, client, "integrity-tester@thecloud.local", "Integrity Tester")

	t.Run("Encoding Edge Cases", func(t *testing.T) {

		// Use emojis and RTL characters in resource names
		fancyName := "test-🚀-עברית-عربي-123"
		payload := map[string]string{
			"name":  fancyName,
			"image": "alpine",
		}
		resp := postRequest(t, client, testutil.TestBaseURL+testutil.TestRouteInstances, token, payload)
		defer resp.Body.Close()

		// Should receive 400 Bad Request due to strict name validation
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Large Payload Handling", func(t *testing.T) {
		// Send a moderately large payload (5MB) to a JSON endpoint
		// Most APIs have a limit around 1MB-10MB
		largePayload := make(map[string]string)
		for i := 0; i < 10000; i++ {
			largePayload[fmt.Sprintf("key_%d", i)] = "value_which_is_quite_long_to_fill_up_memory_and_test_parsing_performance"
		}

		resp := postRequest(t, client, testutil.TestBaseURL+testutil.TestRouteInstances, token, largePayload)
		defer resp.Body.Close()

		// Should either reject (413 Payload Too Large) or handle correctly (400 if missing required fields, but not 500)
		assert.NotEqual(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("Wrong Content-Type", func(t *testing.T) {
		payload := bytes.NewBufferString("name=test-instance&image=alpine")
		req, _ := http.NewRequest("POST", testutil.TestBaseURL+testutil.TestRouteInstances, payload)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded") // API expects JSON
		req.Header.Set(testutil.TestHeaderAPIKey, token)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should be 400 Bad Request or 415 Unsupported Media Type
		assert.Contains(t, []int{http.StatusBadRequest, http.StatusUnsupportedMediaType}, resp.StatusCode)
	})

	t.Run("Data Consistency Lifecycle", func(t *testing.T) {
		// Create -> Read -> Update -> Read -> Delete -> Read
		name := "consistency-test"
		payload := map[string]string{"name": name, "cidr_block": "10.20.0.0/16"}

		// 1. Create
		resp := postRequest(t, client, testutil.TestBaseURL+testutil.TestRouteVpcs, token, payload)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var w struct {
			Data struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"data"`
		}
		body, _ := io.ReadAll(resp.Body)
		require.NoError(t, json.Unmarshal(body, &w))
		id := w.Data.ID

		// 2. Read
		vpcPath := fmt.Sprintf(testutil.TestRouteFormat, testutil.TestBaseURL, testutil.TestRouteVpcs, id)
		resp = getRequest(t, client, vpcPath, token)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&w))
		assert.Equal(t, name, w.Data.Name)

		// 3. Update (if supported, using PUT/PATCH)
		// Assuming instances or VPCs support updates - checking VPC handler
		newName := "consistency-test-updated"
		updatePayload := map[string]string{"name": newName}
		// Some endpoints might not support name updates, testing common pattern
		b, _ := json.Marshal(updatePayload)
		req, _ := http.NewRequest("PATCH", vpcPath, bytes.NewBuffer(b))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set(testutil.TestHeaderAPIKey, token)
		resp, _ = client.Do(req)
		if resp != nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
				// 4. Verify Update
				resp = getRequest(t, client, vpcPath, token)
				defer resp.Body.Close()
				require.NoError(t, json.NewDecoder(resp.Body).Decode(&w))
				assert.Equal(t, newName, w.Data.Name)
			}
		}

		// 5. Delete
		resp = deleteRequest(t, client, vpcPath, token)
		resp.Body.Close()
		assert.Contains(t, []int{http.StatusOK, http.StatusNoContent}, resp.StatusCode)

		// 6. Verify Deleted
		resp = getRequest(t, client, vpcPath, token)
		resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
