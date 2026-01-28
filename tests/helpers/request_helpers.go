// Package helpers provides shared test helpers for E2E tests.
package helpers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/require"
)

// SendMalformedJSON sends a request with invalid JSON syntax
func SendMalformedJSON(t *testing.T, client *http.Client, url, method, token string) *http.Response {
	t.Helper()
	malformedJSON := `{"name": "test", "incomplete": ` // Missing closing brace and value
	req, err := http.NewRequest(method, url, bytes.NewBufferString(malformedJSON))
	require.NoError(t, err)

	req.Header.Set("Content-Type", testutil.TestContentTypeAppJSON)
	if token != "" {
		req.Header.Set(testutil.TestHeaderAPIKey, token)
	}

	resp, err := client.Do(req)
	require.NoError(t, err)
	return resp
}

// SendOversizedPayload sends a request with a large body
func SendOversizedPayload(t *testing.T, client *http.Client, url, method, token string, sizeMB int) *http.Response {
	t.Helper()
	data := make([]byte, sizeMB*1024*1024)
	for i := range data {
		data[i] = 'a'
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/octet-stream")
	if token != "" {
		req.Header.Set(testutil.TestHeaderAPIKey, token)
	}

	resp, err := client.Do(req)
	require.NoError(t, err)
	return resp
}

// SendWithContentType sends a request with a specific Content-Type header
func SendWithContentType(t *testing.T, client *http.Client, url, method, token, contentType string, body io.Reader) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, url, body)
	require.NoError(t, err)

	req.Header.Set("Content-Type", contentType)
	if token != "" {
		req.Header.Set(testutil.TestHeaderAPIKey, token)
	}

	resp, err := client.Do(req)
	require.NoError(t, err)
	return resp
}

// GetBaseURL returns the test base URL
func GetBaseURL() string {
	return testutil.TestBaseURL
}

// FormatURL joins base URL with path
func FormatURL(path string) string {
	return fmt.Sprintf("%s%s", testutil.TestBaseURL, path)
}
