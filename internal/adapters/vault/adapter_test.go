package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"log/slog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVaultAdapter(t *testing.T) {
	// Mock Vault Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/secret/data/test":
			if r.Method == http.MethodGet {
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{
						"data": map[string]interface{}{"password": "secret-pass"},
					},
				})
			} else if r.Method == http.MethodPost || r.Method == http.MethodPut {
				w.WriteHeader(http.StatusOK)
			} else if r.Method == http.MethodDelete {
				w.WriteHeader(http.StatusNoContent)
			}
		case "/v1/sys/health":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"initialized": true, "sealed": false})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	adapter, err := NewVaultAdapter(server.URL, "test-token", slog.Default())
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("StoreSecret", func(t *testing.T) {
		err := adapter.StoreSecret(ctx, "secret/data/test", map[string]interface{}{"password": "new-pass"})
		assert.NoError(t, err)
	})

	t.Run("GetSecret", func(t *testing.T) {
		data, err := adapter.GetSecret(ctx, "secret/data/test")
		require.NoError(t, err)
		// Vault KV v2 returns data nested under "data"
		innerData, ok := data["data"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "secret-pass", innerData["password"])
	})

	t.Run("DeleteSecret", func(t *testing.T) {
		err := adapter.DeleteSecret(ctx, "secret/data/test")
		assert.NoError(t, err)
	})

	t.Run("Ping", func(t *testing.T) {
		err := adapter.Ping(ctx)
		assert.NoError(t, err)
	})
}
