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
	"github.com/stretchr/testify/require"
)

func waitForServer() error {
	client := &http.Client{Timeout: 5 * time.Second}
	for i := 0; i < 30; i++ {
		resp, err := client.Get(testutil.TestBaseURL + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				return nil
			}
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("server not ready at %s", testutil.TestBaseURL)
}

type ResponseWrapper struct {
	Data interface{} `json:"data"`
}

func registerAndLogin(t *testing.T, client *http.Client, email, name string) string {
	// Register
	regReq := map[string]string{
		"email":    email,
		"password": testutil.TestPasswordStrong,
		"name":     name,
	}
	body, _ := json.Marshal(regReq)
	resp, err := client.Post(testutil.TestBaseURL+"/auth/register", testutil.TestContentTypeAppJSON, bytes.NewBuffer(body))
	if err == nil {
		_ = resp.Body.Close()
	}

	// Login
	loginReq := map[string]string{
		"email":    email,
		"password": testutil.TestPasswordStrong,
	}
	body, _ = json.Marshal(loginReq)
	resp, err = client.Post(testutil.TestBaseURL+"/auth/login", testutil.TestContentTypeAppJSON, bytes.NewBuffer(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Login failed for %s: status %d", email, resp.StatusCode)
	}

	var authResp struct {
		Data struct {
			APIKey string `json:"api_key"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&authResp))
	return authResp.Data.APIKey
}

func postRequest(t *testing.T, client *http.Client, url, token string, payload interface{}) *http.Response {
	var body io.Reader
	if payload != nil {
		b, _ := json.Marshal(payload)
		body = bytes.NewBuffer(b)
	}
	req, _ := http.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", testutil.TestContentTypeAppJSON)
	req.Header.Set(testutil.TestHeaderAPIKey, token)
	resp, err := client.Do(req)
	require.NoError(t, err)
	return resp
}

func getRequest(t *testing.T, client *http.Client, url, token string) *http.Response {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set(testutil.TestHeaderAPIKey, token)
	resp, err := client.Do(req)
	require.NoError(t, err)
	return resp
}

func deleteRequest(t *testing.T, client *http.Client, url, token string) *http.Response {
	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set(testutil.TestHeaderAPIKey, token)
	resp, err := client.Do(req)
	require.NoError(t, err)
	return resp
}
