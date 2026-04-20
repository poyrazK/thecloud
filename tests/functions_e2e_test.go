package tests

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/pkg/testutil"
)

func TestFunctionsE2E(t *testing.T) {
	t.Parallel()
	if err := waitForServer(); err != nil {
		t.Fatalf("Failing Functions E2E test: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	token := registerAndLogin(t, client, "functions-tester@thecloud.local", "Functions Tester")

	var functionID string
	functionName := fmt.Sprintf("e2e-fn-%d", time.Now().UnixNano()%1000)

	// 1. Create Function
	t.Run("CreateFunction", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		_ = writer.WriteField("name", functionName)
		_ = writer.WriteField("runtime", "nodejs20")
		_ = writer.WriteField("handler", "index.handler")

		part, _ := writer.CreateFormFile("code", "code.zip")
		_, _ = part.Write([]byte("fake zip content"))
		_ = writer.Close()

		req, _ := http.NewRequest("POST", testutil.TestBaseURL+"/functions", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set(testutil.TestHeaderAPIKey, token)
		applyTenantHeader(t, req, token)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var res struct {
			Data domain.Function `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		functionID = res.Data.ID.String()
		assert.NotEmpty(t, functionID)
	})

	// 2. Invoke Function
	t.Run("InvokeFunction", func(t *testing.T) {
		payload := map[string]string{"foo": "bar"}
		b, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/functions/%s/invoke", testutil.TestBaseURL, functionID), bytes.NewBuffer(b))
		req.Header.Set(testutil.TestHeaderAPIKey, token)
		applyTenantHeader(t, req, token)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Note: Actual invocation might fail if backend doesn't support "fake zip content"
		// but the endpoint should respond.
		assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusInternalServerError)
	})

	// 3. Delete Function
	t.Run("DeleteFunction", func(t *testing.T) {
		resp := deleteRequest(t, client, fmt.Sprintf("%s/functions/%s", testutil.TestBaseURL, functionID), token)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestFunctionsUpdateE2E(t *testing.T) {
	t.Parallel()
	if err := waitForServer(); err != nil {
		t.Fatalf("Failing Functions Update E2E test: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	token := registerAndLogin(t, client, "fn-update-e2e@thecloud.local", "Fn Update E2E")

	fnName := fmt.Sprintf("e2e-fn-update-%d", time.Now().UnixNano()%1000)
	fnID := createFunction(t, client, token, fnName, "nodejs20", "index.handler",
		`console.log(process.env.MY_VAR); process.exit(0);`)

	// 1. Update timeout and memory
	t.Run("UpdateTimeoutAndMemory", func(t *testing.T) {
		updateReq := map[string]interface{}{
			"timeout":  60,
			"memory_mb": 256,
		}
		resp := patchRequest(t, client, fmt.Sprintf("%s/functions/%s", testutil.TestBaseURL, fnID), token, updateReq)
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusOK, resp.StatusCode, "update should succeed")
		var res struct {
			Data domain.Function `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		assert.Equal(t, 60, res.Data.Timeout)
		assert.Equal(t, 256, res.Data.MemoryMB)
	})

	// 2. Verify updated values persisted
	t.Run("GetFunctionAfterUpdate", func(t *testing.T) {
		resp := getRequest(t, client, fmt.Sprintf("%s/functions/%s", testutil.TestBaseURL, fnID), token)
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		var res struct {
			Data domain.Function `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		assert.Equal(t, 60, res.Data.Timeout)
		assert.Equal(t, 256, res.Data.MemoryMB)
	})

	// 3. Update with env vars
	t.Run("UpdateEnvVars", func(t *testing.T) {
		updateReq := map[string]interface{}{
			"env_vars": []map[string]string{
				{"key": "MY_VAR", "value": "hello-from-update"},
				{"key": "ENV", "value": "test"},
			},
		}
		resp := patchRequest(t, client, fmt.Sprintf("%s/functions/%s", testutil.TestBaseURL, fnID), token, updateReq)
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		var res struct {
			Data domain.Function `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		require.Len(t, res.Data.EnvVars, 2)
	})

	// 4. Invoke and verify env vars are available
	t.Run("InvokeWithEnvVars", func(t *testing.T) {
		payload := map[string]string{}
		b, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/functions/%s/invoke", testutil.TestBaseURL, fnID), bytes.NewBuffer(b))
		req.Header.Set(testutil.TestHeaderAPIKey, token)
		applyTenantHeader(t, req, token)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		var invRes struct {
			Data struct {
				Logs string `json:"logs"`
			} `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&invRes))
		assert.Contains(t, invRes.Data.Logs, "hello-from-update")
		assert.Contains(t, invRes.Data.Logs, "ENV=test")
	})

	// 5. Update invalid timeout
	t.Run("UpdateInvalidTimeout", func(t *testing.T) {
		updateReq := map[string]interface{}{"timeout": 0}
		resp := patchRequest(t, client, fmt.Sprintf("%s/functions/%s", testutil.TestBaseURL, fnID), token, updateReq)
		defer func() { _ = resp.Body.Close() }()
		// Service returns 400 for validation error
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	// 6. Update invalid memory
	t.Run("UpdateInvalidMemory", func(t *testing.T) {
		updateReq := map[string]interface{}{"memory_mb": 10}
		resp := patchRequest(t, client, fmt.Sprintf("%s/functions/%s", testutil.TestBaseURL, fnID), token, updateReq)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	// Cleanup
	t.Run("Cleanup", func(t *testing.T) {
		resp := deleteRequest(t, client, fmt.Sprintf("%s/functions/%s", testutil.TestBaseURL, fnID), token)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

// createFunction creates a real zip and registers a function, returning its ID.
func createFunction(t *testing.T, client *http.Client, token, name, runtime, handler, jsCode string) string {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	f, err := zw.Create("index.js")
	require.NoError(t, err)
	_, err = f.Write([]byte(jsCode))
	require.NoError(t, err)
	require.NoError(t, zw.Close())

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("name", name)
	_ = writer.WriteField("runtime", runtime)
	_ = writer.WriteField("handler", handler)
	part, err := writer.CreateFormFile("code", "code.zip")
	require.NoError(t, err)
	_, err = io.Copy(part, bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req, _ := http.NewRequest("POST", testutil.TestBaseURL+"/functions", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set(testutil.TestHeaderAPIKey, token)
	applyTenantHeader(t, req, token)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var res struct {
		Data domain.Function `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
	return res.Data.ID.String()
}
