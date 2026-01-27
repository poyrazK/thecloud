package sdk_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/poyrazk/thecloud/pkg/sdk"
	"github.com/stretchr/testify/assert"
)

func TestFunctionSDK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/api/v1/functions" && r.Method == "POST" {
			// Check multipart form data
			err := r.ParseMultipartForm(10 << 20)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			name := r.FormValue("name")
			runtime := r.FormValue("runtime")
			handler := r.FormValue("handler")
			file, _, err := r.FormFile("code")
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			defer file.Close()
			content, _ := io.ReadAll(file)

			if name == "my-fn" && runtime == "nodejs20" && handler == "index.js" && string(content) == "console.log('hi')" {
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{
						"id":      "fn-1",
						"name":    "my-fn",
						"runtime": "nodejs20",
					},
				})
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if r.URL.Path == "/api/v1/functions" && r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []map[string]interface{}{
					{"id": "fn-1", "name": "my-fn"},
				},
			})
			return
		}

		if r.URL.Path == "/api/v1/functions/fn-1" && r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"id":   "fn-1",
					"name": "my-fn",
				},
			})
			return
		}

		if r.URL.Path == "/api/v1/functions/fn-1" && r.Method == "DELETE" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.URL.Path == "/api/v1/functions/fn-1/invoke" && r.Method == "POST" {
			body, _ := io.ReadAll(r.Body)
			if string(body) == `{"key":"value"}` {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{
						"id":          "inv-1",
						"function_id": "fn-1",
						"status":      "completed",
					},
				})
				return
			}
		}

		if r.URL.Path == "/api/v1/functions/fn-1/logs" && r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []map[string]interface{}{
					{"id": "inv-1", "logs": "hello"},
				},
			})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	client := sdk.NewClient(ts.URL+"/api/v1", "test-api-key")

	t.Run("CreateFunction", func(t *testing.T) {
		fn, err := client.CreateFunction("my-fn", "nodejs20", "index.js", []byte("console.log('hi')"))
		assert.NoError(t, err)
		if fn != nil {
			assert.Equal(t, "my-fn", fn.Name)
		}
	})

	t.Run("ListFunctions", func(t *testing.T) {
		fns, err := client.ListFunctions()
		assert.NoError(t, err)
		if fns != nil {
			assert.Len(t, fns, 1)
		}
	})

	t.Run("GetFunction", func(t *testing.T) {
		fn, err := client.GetFunction("fn-1")
		assert.NoError(t, err)
		if fn != nil {
			assert.Equal(t, "fn-1", fn.ID)
		}
	})

	t.Run("InvokeFunction", func(t *testing.T) {
		inv, err := client.InvokeFunction("fn-1", []byte(`{"key":"value"}`), false)
		assert.NoError(t, err)
		if inv != nil {
			assert.Equal(t, "inv-1", inv.ID)
		}
	})

	t.Run("GetFunctionLogs", func(t *testing.T) {
		logs, err := client.GetFunctionLogs("fn-1")
		assert.NoError(t, err)
		if logs != nil {
			assert.Len(t, logs, 1)
		}
	})

	t.Run("DeleteFunction", func(t *testing.T) {
		err := client.DeleteFunction("fn-1")
		assert.NoError(t, err)
	})
}

func TestFunctionSDK_Errors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("boom"))
	}))
	defer server.Close()

	client := sdk.NewClient(server.URL+"/api/v1", "test-api-key")

	_, err := client.CreateFunction("my-fn", "nodejs20", "index.js", []byte("code"))
	assert.Error(t, err)

	_, err = client.ListFunctions()
	assert.Error(t, err)

	_, err = client.GetFunction("fn-1")
	assert.Error(t, err)

	err = client.DeleteFunction("fn-1")
	assert.Error(t, err)

	_, err = client.InvokeFunction("fn-1", []byte(`{"key":"value"}`), true)
	assert.Error(t, err)

	_, err = client.GetFunctionLogs("fn-1")
	assert.Error(t, err)
}

func TestFunctionSDK_RequestError(t *testing.T) {
	client := sdk.NewClient("http://127.0.0.1:0/api/v1", "test-api-key")
	_, err := client.CreateFunction("my-fn", "nodejs20", "index.js", []byte("code"))

	assert.Error(t, err)
}
