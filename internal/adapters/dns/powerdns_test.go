package dns

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/stretchr/testify/assert"
)

const (
	testPDNSKey  = "test-key"
	testPDNSZone = "example.com."
)

func TestPowerDNSCreateZone(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, testPDNSKey, r.Header.Get("X-API-Key"))

		if r.Method == "POST" {
			assert.Equal(t, "/api/v1/servers/localhost/zones", r.URL.Path)

			body, _ := io.ReadAll(r.Body)
			var reqBody map[string]interface{}
			_ = json.Unmarshal(body, &reqBody)
			assert.Equal(t, testPDNSZone, reqBody["name"])

			w.WriteHeader(http.StatusCreated)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id": "example.com."}`))
			return
		}

		if r.Method == "PATCH" {
			assert.Equal(t, "/api/v1/servers/localhost/zones/"+testPDNSZone, r.URL.Path)

			body, _ := io.ReadAll(r.Body)
			var reqBody map[string]interface{}
			_ = json.Unmarshal(body, &reqBody)
			rrsets := reqBody["rrsets"].([]interface{})
			assert.Len(t, rrsets, 1) // Expect SOA record

			w.WriteHeader(http.StatusNoContent)
			return
		}

		t.Errorf("Unexpected method: %s", r.Method)
	}))
	defer ts.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	backend, err := NewPowerDNSBackend(ts.URL, testPDNSKey, "localhost", logger)
	assert.NoError(t, err)

	err = backend.CreateZone(context.Background(), testPDNSZone, []string{"ns1.example.com."})
	assert.NoError(t, err)
}

func TestPowerDNSAddRecords(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/servers/localhost/zones/"+testPDNSZone, r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var reqBody map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&reqBody)
		rrsets := reqBody["rrsets"].([]interface{})
		assert.Len(t, rrsets, 1)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	backend, err := NewPowerDNSBackend(ts.URL, testPDNSKey, "localhost", logger)
	assert.NoError(t, err)

	records := []ports.RecordSet{
		{
			Name:    "www." + testPDNSZone,
			Type:    "A",
			Records: []string{"1.1.1.1"},
			TTL:     3600,
		},
	}

	err = backend.AddRecords(context.Background(), testPDNSZone, records)
	assert.NoError(t, err)
}

func TestPowerDNSDeleteZone(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/servers/localhost/zones/"+testPDNSZone, r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	backend, err := NewPowerDNSBackend(ts.URL, testPDNSKey, "localhost", logger)
	assert.NoError(t, err)

	err = backend.DeleteZone(context.Background(), testPDNSZone)
	assert.NoError(t, err)
}
