package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/poyrazk/thecloud/pkg/testutil"
)

func TestAuthE2E(t *testing.T) {
	t.Parallel()
	if err := waitForServer(); err != nil {
		t.Fatalf("Failing Auth E2E test: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	email := fmt.Sprintf("auth-test-%d@thecloud.local", time.Now().UnixNano())
	password := testutil.TestPasswordStrong

	// 1. Register
	t.Run("Register", func(t *testing.T) {
		payload := map[string]string{
			"email":    email,
			"password": password,
			"name":     "Auth Tester",
		}
		resp := postRequest(t, client, testutil.TestBaseURL+"/auth/register", "", payload)
		defer func() { _ = resp.Body.Close() }()

		require.True(t, resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict)
	})

	// 2. Login
	var apiKey string
	t.Run("Login", func(t *testing.T) {
		payload := map[string]string{
			"email":    email,
			"password": password,
		}
		resp := postRequest(t, client, testutil.TestBaseURL+"/auth/login", "", payload)
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var res struct {
			Data struct {
				APIKey string `json:"api_key"`
			} `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		apiKey = res.Data.APIKey
		assert.NotEmpty(t, apiKey)
	})

	// 3. Create Additional API Key
	t.Run("CreateKey", func(t *testing.T) {
		payload := map[string]string{"name": "secondary-key"}
		resp := postRequest(t, client, testutil.TestBaseURL+"/auth/keys", apiKey, payload)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	// 4. List Keys
	t.Run("ListKeys", func(t *testing.T) {
		resp := getRequest(t, client, testutil.TestBaseURL+"/auth/keys", apiKey)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
