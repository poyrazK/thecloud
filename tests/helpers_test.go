package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/require"
)

const (
	headerTenantID  = "X-Tenant-ID"
	errTenantNotSet = "tenant ID not set for request"
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

var (
	tenantIDByToken = map[string]string{}
	tenantMu        sync.RWMutex
)

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
	tenantID := createTenant(t, client, authResp.Data.APIKey, name)
	setTenantIDForToken(authResp.Data.APIKey, tenantID)
	switchTenant(t, client, authResp.Data.APIKey, tenantID)
	return authResp.Data.APIKey
}

func createTenant(t *testing.T, client *http.Client, token, name string) string {
	payload := map[string]string{
		"name": name,
		"slug": fmt.Sprintf("%s-%d", name, time.Now().UnixNano()),
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", testutil.TestBaseURL+"/tenants", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", testutil.TestContentTypeAppJSON)
	req.Header.Set(testutil.TestHeaderAPIKey, token)
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Create tenant failed: status %d", resp.StatusCode)
	}

	var tenantResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&tenantResp))
	require.NotEmpty(t, tenantResp.Data.ID)
	return tenantResp.Data.ID
}

func switchTenant(t *testing.T, client *http.Client, token, tenantID string) {
	if tenantID == "" {
		t.Fatalf("tenant ID not set before switch")
	}
	req, _ := http.NewRequest("POST", testutil.TestBaseURL+"/tenants/"+tenantID+"/switch", nil)
	req.Header.Set(testutil.TestHeaderAPIKey, token)
	req.Header.Set(headerTenantID, tenantID)
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Switch tenant failed: status %d", resp.StatusCode)
	}
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
	if requiresTenantHeader(url) {
		tenantID := tenantIDForToken(token)
		if tenantID == "" {
			t.Fatal(errTenantNotSet)
		}
		req.Header.Set(headerTenantID, tenantID)
	}
	resp, err := client.Do(req)
	require.NoError(t, err)
	return resp
}

func getRequest(t *testing.T, client *http.Client, url, token string) *http.Response {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set(testutil.TestHeaderAPIKey, token)
	if requiresTenantHeader(url) {
		tenantID := tenantIDForToken(token)
		if tenantID == "" {
			t.Fatal(errTenantNotSet)
		}
		req.Header.Set(headerTenantID, tenantID)
	}
	resp, err := client.Do(req)
	require.NoError(t, err)
	return resp
}

func deleteRequest(t *testing.T, client *http.Client, url, token string) *http.Response {
	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set(testutil.TestHeaderAPIKey, token)
	if requiresTenantHeader(url) {
		tenantID := tenantIDForToken(token)
		if tenantID == "" {
			t.Fatal(errTenantNotSet)
		}
		req.Header.Set(headerTenantID, tenantID)
	}
	resp, err := client.Do(req)
	require.NoError(t, err)
	return resp
}

func applyTenantHeader(t *testing.T, req *http.Request, token string) {
	if !requiresTenantHeader(req.URL.Path) {
		return
	}
	tenantID := tenantIDForToken(token)
	if tenantID == "" {
		t.Fatal(errTenantNotSet)
	}
	req.Header.Set(headerTenantID, tenantID)
}

func requiresTenantHeader(url string) bool {
	return !strings.Contains(url, "/auth/")
}

func setTenantIDForToken(token, tenantID string) {
	tenantMu.Lock()
	defer tenantMu.Unlock()
	tenantIDByToken[token] = tenantID
}

func tenantIDForToken(token string) string {
	tenantMu.RLock()
	defer tenantMu.RUnlock()
	return tenantIDByToken[token]
}
