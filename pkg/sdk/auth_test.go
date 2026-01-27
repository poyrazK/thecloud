package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientCreateKeySuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/auth/keys", r.URL.Path)
		assert.Equal(t, testAPIKey, r.Header.Get("X-API-Key"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response[struct {
			Key string `json:"key"`
		}]{Data: struct {
			Key string `json:"key"`
		}{Key: "new-key"}})
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	key, err := client.CreateKey("bootstrap")

	assert.NoError(t, err)
	assert.Equal(t, "new-key", key)
}

func TestClientCreateKeyStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/auth/keys", r.URL.Path)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("boom"))
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	_, err := client.CreateKey("bootstrap")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api error")
}

func TestClientCreateKeyRequestError(t *testing.T) {
	client := NewClient("http://127.0.0.1:0", testAPIKey)
	_, err := client.CreateKey("bootstrap")

	assert.Error(t, err)
}
