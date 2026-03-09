package sdk_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/poyrazk/thecloud/pkg/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	functionContentType     = "Content-Type"
	functionApplicationJSON = "application/json"
	functionAPIKey          = "test-api-key"
	functionID              = "fn-1"
	functionName            = "my-fn"
	functionRuntime         = "nodejs20"
	functionHandler         = "index.js"
	functionCode            = "console.log('hi')"
	functionInvokePayload   = `{"key":"value"}`
	functionInvocationID    = "inv-1"
	functionPath            = "/api/v1/functions"
	functionPathPrefix      = "/api/v1/functions/"
)

func newFunctionTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(functionContentType, functionApplicationJSON)

		if handleFunctionCreate(w, r) {
			return
		}
		if handleFunctionList(w, r) {
			return
		}
		if handleFunctionGet(w, r) {
			return
		}
		if handleFunctionDelete(w, r) {
			return
		}
		if handleFunctionInvoke(w, r) {
			return
		}
		if handleFunctionLogs(w, r) {
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func handleFunctionCreate(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Path != functionPath || r.Method != http.MethodPost {
		return false
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return true
	}

	name := r.FormValue("name")
	runtime := r.FormValue("runtime")
	handler := r.FormValue("handler")
	file, _, err := r.FormFile("code")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return true
	}
	defer func() { _ = file.Close() }()
	content, _ := io.ReadAll(file)

	if name == functionName && runtime == functionRuntime && handler == functionHandler && string(content) == functionCode {
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":      functionID,
				"name":    functionName,
				"runtime": functionRuntime,
			},
		})
		return true
	}

	w.WriteHeader(http.StatusBadRequest)
	return true
}

func handleFunctionList(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Path != functionPath || r.Method != http.MethodGet {
		return false
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data": []map[string]interface{}{
			{"id": functionID, "name": functionName},
		},
	})
	return true
}

func handleFunctionGet(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Path != functionPathPrefix+functionID || r.Method != http.MethodGet {
		return false
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data": map[string]interface{}{
			"id":   functionID,
			"name": functionName,
		},
	})
	return true
}

func handleFunctionDelete(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Path != functionPathPrefix+functionID || r.Method != http.MethodDelete {
		return false
	}
	w.WriteHeader(http.StatusNoContent)
	return true
}

func handleFunctionInvoke(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Path != functionPathPrefix+functionID+"/invoke" || r.Method != http.MethodPost {
		return false
	}

	body, _ := io.ReadAll(r.Body)
	if string(body) == functionInvokePayload {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":          functionInvocationID,
				"function_id": functionID,
				"status":      "completed",
			},
		})
		return true
	}
	return true
}

func handleFunctionLogs(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Path != functionPathPrefix+functionID+"/logs" || r.Method != http.MethodGet {
		return false
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data": []map[string]interface{}{
			{"id": functionInvocationID, "logs": "hello"},
		},
	})
	return true
}

func TestFunctionSDK(t *testing.T) {
	ts := newFunctionTestServer(t)
	defer ts.Close()

	client := sdk.NewClient(ts.URL+"/api/v1", functionAPIKey)

	t.Run("CreateFunction", func(t *testing.T) {
		fn, err := client.CreateFunction(functionName, functionRuntime, functionHandler, []byte(functionCode))
		require.NoError(t, err)
		if fn != nil {
			assert.Equal(t, functionName, fn.Name)
		}
	})

	t.Run("ListFunctions", func(t *testing.T) {
		fns, err := client.ListFunctions()
		require.NoError(t, err)
		if fns != nil {
			assert.Len(t, fns, 1)
		}
	})

	t.Run("GetFunction", func(t *testing.T) {
		fn, err := client.GetFunction(functionID)
		require.NoError(t, err)
		if fn != nil {
			assert.Equal(t, functionID, fn.ID)
		}
	})

	t.Run("InvokeFunction", func(t *testing.T) {
		inv, err := client.InvokeFunction(functionID, []byte(functionInvokePayload), false)
		require.NoError(t, err)
		if inv != nil {
			assert.Equal(t, functionInvocationID, inv.ID)
		}
	})

	t.Run("GetFunctionLogs", func(t *testing.T) {
		logs, err := client.GetFunctionLogs(functionID)
		require.NoError(t, err)
		if logs != nil {
			assert.Len(t, logs, 1)
		}
	})

	t.Run("DeleteFunction", func(t *testing.T) {
		err := client.DeleteFunction(functionID)
		require.NoError(t, err)
	})
}

func TestFunctionSDKErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	defer server.Close()

	client := sdk.NewClient(server.URL+"/api/v1", functionAPIKey)

	_, err := client.CreateFunction(functionName, functionRuntime, functionHandler, []byte("code"))
	require.Error(t, err)

	_, err = client.ListFunctions()
	require.Error(t, err)

	_, err = client.GetFunction(functionID)
	require.Error(t, err)

	err = client.DeleteFunction(functionID)
	require.Error(t, err)

	_, err = client.InvokeFunction(functionID, []byte(functionInvokePayload), true)
	require.Error(t, err)

	_, err = client.GetFunctionLogs(functionID)
	require.Error(t, err)
}

func TestFunctionSDKRequestError(t *testing.T) {
	client := sdk.NewClient("http://127.0.0.1:0/api/v1", functionAPIKey)
	_, err := client.CreateFunction(functionName, functionRuntime, functionHandler, []byte("code"))

	require.Error(t, err)
}
