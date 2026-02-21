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
	"github.com/stretchr/testify/require"
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
			rrsets, ok := reqBody["rrsets"].([]interface{})
			if !ok {
				t.Errorf("expected rrsets in request body")
				return
			}
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
		rrsets, ok := reqBody["rrsets"].([]interface{})
		if !ok {
			t.Errorf("expected rrsets in request body")
			return
		}
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

func TestPowerDNSUpdateRecords(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/servers/localhost/zones/"+testPDNSZone, r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var reqBody map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&reqBody)
		rrsets, ok := reqBody["rrsets"].([]interface{})
		if !ok {
			t.Errorf("expected rrsets in request body")
			return
		}
		assert.Len(t, rrsets, 1)

		rrset, ok := rrsets[0].(map[string]interface{})
		if !ok {
			t.Errorf("expected rrset to be map[string]interface{}")
			return
		}
		assert.Equal(t, "REPLACE", rrset["changetype"])

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
			Records: []string{"1.2.3.4"},
			TTL:     300,
		},
	}

	err = backend.UpdateRecords(context.Background(), testPDNSZone, records)
	assert.NoError(t, err)
}

func TestPowerDNSDeleteRecords(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/servers/localhost/zones/"+testPDNSZone, r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var reqBody map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&reqBody)
		rrsets, ok := reqBody["rrsets"].([]interface{})
		if !ok {
			t.Errorf("expected rrsets in request body")
			return
		}
		assert.Len(t, rrsets, 1)

		rrset, ok := rrsets[0].(map[string]interface{})
		if !ok {
			t.Errorf("expected rrset to be map[string]interface{}")
			return
		}
		assert.Equal(t, "DELETE", rrset["changetype"])
		assert.Equal(t, "www."+testPDNSZone, rrset["name"])

		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	backend, err := NewPowerDNSBackend(ts.URL, testPDNSKey, "localhost", logger)
	assert.NoError(t, err)

	err = backend.DeleteRecords(context.Background(), testPDNSZone, "www."+testPDNSZone, "A")
	assert.NoError(t, err)
}

func TestPowerDNSGetZone(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/servers/localhost/zones/"+testPDNSZone, r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		resp := map[string]interface{}{
			"name":            testPDNSZone,
			"kind":            "Native",
			"serial":          12345,
			"notified_serial": 12345,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	backend, err := NewPowerDNSBackend(ts.URL, testPDNSKey, "localhost", logger)
	assert.NoError(t, err)

	zone, err := backend.GetZone(context.Background(), testPDNSZone)
	assert.NoError(t, err)
	assert.NotNil(t, zone)
	assert.Equal(t, testPDNSZone, zone.Name)
	assert.Equal(t, "Native", zone.Kind)
}

func TestPowerDNSListRecords(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/servers/localhost/zones/"+testPDNSZone, r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		resp := map[string]interface{}{
			"name": testPDNSZone,
			"rrsets": []map[string]interface{}{
				{
					"name": "www." + testPDNSZone,
					"type": "A",
					"ttl":  3600,
					"records": []map[string]interface{}{
						{"content": "1.2.3.4", "disabled": false},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	backend, err := NewPowerDNSBackend(ts.URL, testPDNSKey, "localhost", logger)
	assert.NoError(t, err)

	records, err := backend.ListRecords(context.Background(), testPDNSZone)
	assert.NoError(t, err)
	assert.Len(t, records, 1)
	assert.Equal(t, "www."+testPDNSZone, records[0].Name)
	assert.Equal(t, "A", records[0].Type)
	assert.Equal(t, "1.2.3.4", records[0].Records[0])
}
