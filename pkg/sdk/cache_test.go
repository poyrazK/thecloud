package sdk_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/poyrazk/thecloud/pkg/sdk"
	"github.com/stretchr/testify/assert"
)

const (
	cacheTestAPIKey       = "test-api-key"
	cacheTestName         = "my-cache"
	cacheTestID           = "c-1"
	cacheTestBasePath     = "/api/v1/caches"
	cacheTestContentType  = "Content-Type"
	cacheTestAppJSON      = "application/json"
	cacheTestConnString   = "redis://localhost:6379"
	cacheTestVersion      = "7.0"
	cacheTestMemoryMB     = 128
	cacheTestStatsTotal   = int64(10)
)

func newCacheTestServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()
		w.Header().Set(cacheTestContentType, cacheTestAppJSON)

		switch {
		case r.URL.Path == cacheTestBasePath && r.Method == http.MethodPost:
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["name"] == cacheTestName {
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{
						"id":   cacheTestID,
						"name": cacheTestName,
					},
				})
				return
			}
		case r.URL.Path == cacheTestBasePath && r.Method == http.MethodGet:
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []map[string]interface{}{{"id": cacheTestID, "name": cacheTestName}},
			})
			return
		case r.URL.Path == cacheTestBasePath+"/"+cacheTestID && r.Method == http.MethodGet:
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"id":   cacheTestID,
					"name": cacheTestName,
				},
			})
			return
		case r.URL.Path == cacheTestBasePath+"/"+cacheTestID && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
			return
		case r.URL.Path == cacheTestBasePath+"/"+cacheTestID+"/connection":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"connection_string": cacheTestConnString,
				},
			})
			return
		case r.URL.Path == cacheTestBasePath+"/"+cacheTestID+"/flush" && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]string{"result": "OK"},
			})
			return
		case r.URL.Path == cacheTestBasePath+"/"+cacheTestID+"/stats":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"total_keys": cacheTestStatsTotal,
				},
			})
			return
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestCacheSDK(t *testing.T) {
	ts := newCacheTestServer(t)
	defer ts.Close()

	client := sdk.NewClient(ts.URL+"/api/v1", cacheTestAPIKey)

	t.Run("CreateCache", func(t *testing.T) {
		cache, err := client.CreateCache(cacheTestName, cacheTestVersion, cacheTestMemoryMB, nil)
		assert.NoError(t, err)
		if cache != nil {
			assert.Equal(t, cacheTestName, cache.Name)
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
		cache, err := client.GetCache(cacheTestID)
		assert.NoError(t, err)
		if cache != nil {
			assert.Equal(t, cacheTestID, cache.ID)
		}
	})

	t.Run("GetConnectionString", func(t *testing.T) {
		conn, err := client.GetCacheConnectionString(cacheTestID)
		assert.NoError(t, err)
		assert.Equal(t, cacheTestConnString, conn)
	})

	t.Run("FlushCache", func(t *testing.T) {
		err := client.FlushCache(cacheTestID)
		assert.NoError(t, err)
	})

	t.Run("GetStats", func(t *testing.T) {
		stats, err := client.GetCacheStats(cacheTestID)
		assert.NoError(t, err)
		if stats != nil {
			assert.Equal(t, cacheTestStatsTotal, stats.TotalKeys)
		}
	})

	t.Run("DeleteCache", func(t *testing.T) {
		err := client.DeleteCache(cacheTestID)
		assert.NoError(t, err)
	})
}

func TestCacheSDKErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("boom"))
	}))
	defer server.Close()

	client := sdk.NewClient(server.URL+"/api/v1", cacheTestAPIKey)

	_, err := client.CreateCache(cacheTestName, cacheTestVersion, cacheTestMemoryMB, nil)
	assert.Error(t, err)

	_, err = client.ListCaches()
	assert.Error(t, err)

	_, err = client.GetCache(cacheTestID)
	assert.Error(t, err)

	err = client.DeleteCache(cacheTestID)
	assert.Error(t, err)

	_, err = client.GetCacheConnectionString(cacheTestID)
	assert.Error(t, err)

	err = client.FlushCache(cacheTestID)
	assert.Error(t, err)

	_, err = client.GetCacheStats(cacheTestID)
	assert.Error(t, err)
}
