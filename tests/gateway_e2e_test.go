package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const gatewayRoutesPath = "/gateway/routes"
const httpbinAnything = "https://httpbin.org/anything"

func TestGatewayE2E(t *testing.T) {
	if err := waitForServer(); err != nil {
		t.Fatalf("Failing Gateway E2E test: %v", err)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	token := registerAndLogin(t, client, "gateway-tester@thecloud.local", "Gateway Tester")

	// Use a unique suffix for route names to avoid collisions in E2E environment
	ts := time.Now().UnixNano() % 100000

	t.Run("CreateAndListPatternRoute", func(t *testing.T) {
		// 1. Create a pattern-based route
		// We'll use httpbin.org to verify the proxying works
		pattern := fmt.Sprintf("/httpbin-%d/{method}", ts)
		routeName := fmt.Sprintf("httpbin-pattern-%d", ts)
		targetURL := "https://httpbin.org"

		payload := map[string]interface{}{
			"name":         routeName,
			"path_prefix":  pattern, // API currently uses path_prefix field for the pattern
			"target_url":   targetURL,
			"strip_prefix": true,
			"rate_limit":   100,
		}

		resp := postRequest(t, client, testutil.TestBaseURL+gatewayRoutesPath, token, payload)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		defer func() { _ = resp.Body.Close() }()

		var res struct {
			Data domain.GatewayRoute `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		assert.Equal(t, "pattern", res.Data.PatternType)
		assert.Equal(t, pattern, res.Data.PathPattern)

		// 2. List routes and verify it's there
		listResp := getRequest(t, client, testutil.TestBaseURL+gatewayRoutesPath, token)
		require.Equal(t, http.StatusOK, listResp.StatusCode)
		defer func() { _ = listResp.Body.Close() }()

		var listRes struct {
			Data []domain.GatewayRoute `json:"data"`
		}
		require.NoError(t, json.NewDecoder(listResp.Body).Decode(&listRes))

		found := false
		for _, r := range listRes.Data {
			if r.Name == routeName {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("VerifyPatternProxying", func(t *testing.T) {
		// Give the gateway a moment to refresh routes
		time.Sleep(2 * time.Second)

		// Test GET request through gateway
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/gw/httpbin-%d/get", testutil.TestBaseURL, ts), nil)
		req.Header.Set(testutil.TestHeaderAPIKey, token)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var httpbinResp struct {
			URL string `json:"url"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&httpbinResp))
		assert.Contains(t, httpbinResp.URL, "/get")
	})

	t.Run("VerifyRegexPatternProxying", func(t *testing.T) {
		pattern := fmt.Sprintf("/status-%d/{code:[0-9]+}", ts)
		routeName := fmt.Sprintf("status-code-%d", ts)
		targetURL := "https://httpbin.org/status"

		payload := map[string]interface{}{
			"name":         routeName,
			"path_prefix":  pattern,
			"target_url":   targetURL,
			"strip_prefix": true,
			"rate_limit":   100,
		}

		resp := postRequest(t, client, testutil.TestBaseURL+gatewayRoutesPath, token, payload)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		_ = resp.Body.Close()

		time.Sleep(2 * time.Second)

		// This should match
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/gw/status-%d/201", testutil.TestBaseURL, ts), nil)
		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		_ = resp.Body.Close()

		// This should NOT match (letters instead of numbers)
		req, _ = http.NewRequest("GET", fmt.Sprintf("%s/gw/status-%d/abc", testutil.TestBaseURL, ts), nil)
		resp, err = client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		_ = resp.Body.Close()
	})

	t.Run("VerifyWildcardProxying", func(t *testing.T) {
		pattern := fmt.Sprintf("/wild-%d/*", ts)
		routeName := fmt.Sprintf("wildcard-route-%d", ts)
		targetURL := httpbinAnything

		payload := map[string]interface{}{
			"name":         routeName,
			"path_prefix":  pattern,
			"target_url":   targetURL,
			"strip_prefix": true,
			"rate_limit":   100,
		}

		resp := postRequest(t, client, testutil.TestBaseURL+gatewayRoutesPath, token, payload)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		_ = resp.Body.Close()

		time.Sleep(2 * time.Second)

		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/gw/wild-%d/foo/bar", testutil.TestBaseURL, ts), nil)
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var httpbinResp struct {
			URL string `json:"url"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&httpbinResp))
		assert.Contains(t, httpbinResp.URL, "/anything/foo/bar")
	})

	t.Run("VerifyMultiParamProxying", func(t *testing.T) {
		pattern := fmt.Sprintf("/orgs-%d/{org}/projects/{project}", ts)
		routeName := fmt.Sprintf("multi-param-%d", ts)
		targetURL := httpbinAnything

		payload := map[string]interface{}{
			"name":         routeName,
			"path_prefix":  pattern,
			"target_url":   targetURL,
			"strip_prefix": true,
			"rate_limit":   100,
		}

		resp := postRequest(t, client, testutil.TestBaseURL+gatewayRoutesPath, token, payload)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		_ = resp.Body.Close()

		time.Sleep(2 * time.Second)

		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/gw/orgs-%d/google/projects/chrome", testutil.TestBaseURL, ts), nil)
		req.Header.Set(testutil.TestHeaderAPIKey, token)
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var httpbinResp struct {
			URL string `json:"url"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&httpbinResp))
		// httpbin /anything echoes everything. The path should be /google/projects/chrome if stripped correctly
		assert.Contains(t, httpbinResp.URL, "/google/projects/chrome")
	})

	t.Run("VerifyPriorityMatching", func(t *testing.T) {
		// 1. General pattern route (Low priority)
		patternGen := fmt.Sprintf("/users-%d/{id}", ts)
		payloadGen := map[string]interface{}{
			"name":         fmt.Sprintf("user-gen-%d", ts),
			"path_prefix":  patternGen,
			"target_url":   httpbinAnything + "/general",
			"strip_prefix": true,
			"priority":     1,
		}
		resGen := postRequest(t, client, testutil.TestBaseURL+gatewayRoutesPath, token, payloadGen)
		_ = resGen.Body.Close()

		// 2. Specific exact route (High priority implied by length or explicit)
		patternSpec := fmt.Sprintf("/users-%d/me", ts)
		payloadSpec := map[string]interface{}{
			"name":         fmt.Sprintf("user-spec-%d", ts),
			"path_prefix":  patternSpec,
			"target_url":   httpbinAnything + "/specific",
			"strip_prefix": true,
			"priority":     10,
		}
		resSpec := postRequest(t, client, testutil.TestBaseURL+gatewayRoutesPath, token, payloadSpec)
		_ = resSpec.Body.Close()

		time.Sleep(2 * time.Second)

		// Test matching "me" -> should hit the specific route because of higher priority
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/gw/users-%d/me", testutil.TestBaseURL, ts), nil)
		req.Header.Set(testutil.TestHeaderAPIKey, token)
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		var httpbinResp struct {
			URL string `json:"url"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&httpbinResp))
		assert.Contains(t, httpbinResp.URL, "/specific")
	})

	t.Run("VerifyExtensionMatching", func(t *testing.T) {
		pattern := fmt.Sprintf("/static-%d/*.{ext}", ts)
		routeName := fmt.Sprintf("extension-route-%d", ts)
		targetURL := httpbinAnything

		payload := map[string]interface{}{
			"name":         routeName,
			"path_prefix":  pattern,
			"target_url":   targetURL,
			"strip_prefix": true,
		}

		resExt := postRequest(t, client, testutil.TestBaseURL+gatewayRoutesPath, token, payload)
		_ = resExt.Body.Close()

		time.Sleep(2 * time.Second)

		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/gw/static-%d/image.png", testutil.TestBaseURL, ts), nil)
		req.Header.Set(testutil.TestHeaderAPIKey, token)
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var httpbinResp struct {
			URL string `json:"url"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&httpbinResp))
		assert.Contains(t, httpbinResp.URL, "/image.png")
	})

	t.Run("VerifyMethodMatching", func(t *testing.T) {
		pattern := fmt.Sprintf("/method-test-%d", ts)

		// 1. GET only route
		payloadGet := map[string]interface{}{
			"name":        fmt.Sprintf("get-route-%d", ts),
			"path_prefix": pattern,
			"target_url":  httpbinAnything + "/get-only",
			"methods":     []string{"GET"},
		}
		resMethodGet := postRequest(t, client, testutil.TestBaseURL+gatewayRoutesPath, token, payloadGet)
		_ = resMethodGet.Body.Close()

		// 2. POST only route
		payloadPost := map[string]interface{}{
			"name":        fmt.Sprintf("post-route-%d", ts),
			"path_prefix": pattern,
			"target_url":  httpbinAnything + "/post-only",
			"methods":     []string{"POST"},
		}
		resMethodPost := postRequest(t, client, testutil.TestBaseURL+gatewayRoutesPath, token, payloadPost)
		_ = resMethodPost.Body.Close()

		time.Sleep(2 * time.Second)

		// Test GET request
		reqGet, _ := http.NewRequest("GET", testutil.TestBaseURL+"/gw"+pattern, nil)
		reqGet.Header.Set(testutil.TestHeaderAPIKey, token)
		respGet, err := client.Do(reqGet)
		require.NoError(t, err)
		defer func() { _ = respGet.Body.Close() }()
		var resGet struct {
			URL string `json:"url"`
		}
		err = json.NewDecoder(respGet.Body).Decode(&resGet)
		require.NoError(t, err)
		assert.Contains(t, resGet.URL, "/get-only")

		// Test POST request
		reqPost, _ := http.NewRequest("POST", testutil.TestBaseURL+"/gw"+pattern, nil)
		reqPost.Header.Set(testutil.TestHeaderAPIKey, token)
		respPost, err := client.Do(reqPost)
		require.NoError(t, err)
		defer func() { _ = respPost.Body.Close() }()
		var resPost struct {
			URL string `json:"url"`
		}
		err = json.NewDecoder(respPost.Body).Decode(&resPost)
		require.NoError(t, err)
		assert.Contains(t, resPost.URL, "/post-only")

		// Test DELETE request (should fail)
		reqDel, _ := http.NewRequest("DELETE", testutil.TestBaseURL+"/gw"+pattern, nil)
		reqDel.Header.Set(testutil.TestHeaderAPIKey, token)
		respDel, err := client.Do(reqDel)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, respDel.StatusCode)
		_ = respDel.Body.Close()
	})
}
