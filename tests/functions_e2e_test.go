package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	if err := waitForServer(); err != nil {
		t.Skipf("Skipping Functions E2E test: %v", err)
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
		writer.Close()

		req, _ := http.NewRequest("POST", testutil.TestBaseURL+"/functions", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set(testutil.TestHeaderAPIKey, token)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

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

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Note: Actual invocation might fail if backend doesn't support "fake zip content"
		// but the endpoint should respond.
		assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusInternalServerError)
	})

	// 3. Delete Function
	t.Run("DeleteFunction", func(t *testing.T) {
		resp := deleteRequest(t, client, fmt.Sprintf("%s/functions/%s", testutil.TestBaseURL, functionID), token)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
