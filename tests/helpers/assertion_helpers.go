// Package helpers provides shared test helpers for E2E tests.
package helpers

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertErrorCode verifies that the response status code matches expectation
func AssertErrorCode(t *testing.T, resp *http.Response, expectedCode int) {
	t.Helper()
	assert.Equal(t, expectedCode, resp.StatusCode, "Expected status code %d, got %d", expectedCode, resp.StatusCode)
}

// AssertValidationError verifies that the response contains a validation error for a specific field
func AssertValidationError(t *testing.T, resp *http.Response, field string) {
	t.Helper()
	AssertErrorCode(t, resp, http.StatusBadRequest)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &res))

	// Adjust based on actual error structure of the platform
	// Assuming structure like {"error": {"message": "...", "details": {"field": "..."}}}
	errObj, ok := res["error"].(map[string]interface{})
	if !ok {
		t.Errorf("Unexpected error response format: %s", string(body))
		return
	}

	assert.Contains(t, errObj["message"], field, "Error message should mention field %s", field)
}

// AssertNotEmpty verifies that a field in the JSON response is not empty
func AssertNotEmpty(t *testing.T, body []byte, field string) {
	t.Helper()
	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &res))

	// Traverse data if present (assuming standard wrapper)
	data := res
	if d, ok := res["data"].(map[string]interface{}); ok {
		data = d
	}

	val, exists := data[field]
	assert.True(t, exists, "Field %s should exist in response", field)
	assert.NotEmpty(t, val, "Field %s should not be empty", field)
}
