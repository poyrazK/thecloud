package sdk

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	clientTestAPIKey = "key-123"
	clientErrorURL   = "http://127.0.0.1:0"
)

func TestNewClientSetsAPIKey(t *testing.T) {
	client := NewClient("http://localhost", clientTestAPIKey)
	assert.Equal(t, clientTestAPIKey, client.resty.Header.Get("X-API-Key"))
}

func TestClientGetError(t *testing.T) {
	client := NewClient(clientErrorURL, clientTestAPIKey)
	var res Response[map[string]string]

	err := client.get("/fail", &res)
	assert.Error(t, err)
}

func TestClientGetStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad"))
	}))
	defer server.Close()

	client := NewClient(server.URL, clientTestAPIKey)
	var res Response[map[string]string]

	err := client.get("/fail", &res)
	assert.Error(t, err)
}

func TestClientPostError(t *testing.T) {
	client := NewClient(clientErrorURL, clientTestAPIKey)
	var res Response[map[string]string]

	err := client.post("/fail", map[string]string{"a": "b"}, &res)
	assert.Error(t, err)
}

func TestClientPostStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	defer server.Close()

	client := NewClient(server.URL, clientTestAPIKey)
	var res Response[map[string]string]

	err := client.post("/fail", map[string]string{"a": "b"}, &res)
	assert.Error(t, err)
}

func TestClientDeleteError(t *testing.T) {
	client := NewClient(clientErrorURL, clientTestAPIKey)
	var res Response[map[string]string]

	err := client.delete("/fail", &res)
	assert.Error(t, err)
}

func TestClientDeleteStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("nope"))
	}))
	defer server.Close()

	client := NewClient(server.URL, clientTestAPIKey)
	var res Response[map[string]string]

	err := client.delete("/fail", &res)
	assert.Error(t, err)
}

func TestClientPutError(t *testing.T) {
	client := NewClient(clientErrorURL, clientTestAPIKey)
	var res Response[map[string]string]

	err := client.put("/fail", map[string]string{"a": "b"}, &res)
	assert.Error(t, err)
}

func TestClientPutStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte("conflict"))
	}))
	defer server.Close()

	client := NewClient(server.URL, clientTestAPIKey)
	var res Response[map[string]string]

	err := client.put("/fail", map[string]string{"a": "b"}, &res)
	assert.Error(t, err)
}

func TestClientPatchError(t *testing.T) {
	client := NewClient(clientErrorURL, clientTestAPIKey)
	var res Response[map[string]string]

	err := client.patch("/fail", map[string]string{"a": "b"}, &res)
	assert.Error(t, err)
}

func TestClientPatchStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("teapot"))
	}))
	defer server.Close()

	client := NewClient(server.URL, clientTestAPIKey)
	var res Response[map[string]string]

	err := client.patch("/fail", map[string]string{"a": "b"}, &res)
	assert.Error(t, err)
}
