package sdk_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/poyrazk/thecloud/pkg/sdk"
	"github.com/stretchr/testify/assert"
)

func TestCacheSDK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/api/v1/caches" && r.Method == "POST" {
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["name"] == "my-cache" {
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{
						"id":   "c-1",
						"name": "my-cache",
					},
				})
				return
			}
		}

		if r.URL.Path == "/api/v1/caches" && r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []map[string]interface{}{
					{"id": "c-1", "name": "my-cache"},
				},
			})
			return
		}

		if r.URL.Path == "/api/v1/caches/c-1" && r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"id":   "c-1",
					"name": "my-cache",
				},
			})
			return
		}

		if r.URL.Path == "/api/v1/caches/c-1" && r.Method == "DELETE" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.URL.Path == "/api/v1/caches/c-1/connection" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"connection_string": "redis://localhost:6379",
				},
			})
			return
		}

		if r.URL.Path == "/api/v1/caches/c-1/flush" && r.Method == "POST" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]string{"result": "OK"},
			})
			return
		}

		if r.URL.Path == "/api/v1/caches/c-1/stats" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"total_keys": 10,
				},
			})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	client := sdk.NewClient(ts.URL+"/api/v1", "test-api-key")

	t.Run("CreateCache", func(t *testing.T) {
		cache, err := client.CreateCache("my-cache", "7.0", 128, nil)
		assert.NoError(t, err)
		if cache != nil {
			assert.Equal(t, "my-cache", cache.Name)
		}
	})

	t.Run("ListCaches", func(t *testing.T) {
		caches, err := client.ListCaches()
		assert.NoError(t, err)
		if caches != nil {
			assert.Len(t, caches, 1)
		}
	})

	t.Run("GetCache", func(t *testing.T) {
		cache, err := client.GetCache("c-1")
		assert.NoError(t, err)
		if cache != nil {
			assert.Equal(t, "c-1", cache.ID)
		}
	})

	t.Run("GetConnectionString", func(t *testing.T) {
		conn, err := client.GetCacheConnectionString("c-1")
		assert.NoError(t, err)
		assert.Equal(t, "redis://localhost:6379", conn)
	})

	t.Run("FlushCache", func(t *testing.T) {
		err := client.FlushCache("c-1")
		assert.NoError(t, err)
	})

	t.Run("GetStats", func(t *testing.T) {
		stats, err := client.GetCacheStats("c-1")
		assert.NoError(t, err)
		if stats != nil {
			assert.Equal(t, int64(10), stats.TotalKeys)
		}
	})

	t.Run("DeleteCache", func(t *testing.T) {
		err := client.DeleteCache("c-1")
		assert.NoError(t, err)
	})
}
