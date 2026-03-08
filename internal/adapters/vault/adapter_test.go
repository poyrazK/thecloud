// Package vault provides a HashiCorp Vault adapter for secret management.
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

func TestAdapter(t *testing.T) {
	// Mock Vault Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/secret/data/test":
			switch r.Method {
			case http.MethodGet:
				w.WriteHeader(http.StatusOK)
				resp := map[string]interface{}{
					"data": map[string]interface{}{
						"data": map[string]interface{}{"password": "secret-pass"},
					},
				}
				if err := json.NewEncoder(w).Encode(resp); err != nil {
					t.Errorf("json encode failed: %v", err)
				}
			case http.MethodPost, http.MethodPut:
				// Verify KV v2 wrap
				var body map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				if _, ok := body["data"]; !ok {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusOK)
			case http.MethodDelete:
				w.WriteHeader(http.StatusNoContent)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		case "/v1/sys/health":
			w.WriteHeader(http.StatusOK)
			resp := map[string]interface{}{"initialized": true, "sealed": false}
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				t.Errorf("json encode failed: %v", err)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	adapter, err := NewVaultAdapter(server.URL, "test-token", slog.Default())
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name    string
		fn      func(t *testing.T)
	}{
		{
			name: "StoreSecret",
			fn: func(t *testing.T) {
				err := adapter.StoreSecret(ctx, "secret/data/test", map[string]interface{}{"password": "new-pass"})
				assert.NoError(t, err)
			},
		},
		{
			name: "GetSecret",
			fn: func(t *testing.T) {
				data, err := adapter.GetSecret(ctx, "secret/data/test")
				require.NoError(t, err)
				assert.Equal(t, "secret-pass", data["password"])
			},
		},
		{
			name: "DeleteSecret",
			fn: func(t *testing.T) {
				err := adapter.DeleteSecret(ctx, "secret/data/test")
				assert.NoError(t, err)
			},
		},
		{
			name: "Ping",
			fn: func(t *testing.T) {
				err := adapter.Ping(ctx)
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}
