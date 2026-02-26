package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestDNSListZones(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dns/zones" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		payload := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":         uuid.New().String(),
					"name":       "example.com",
					"vpc_id":     uuid.Nil.String(),
					"created_at": time.Now().UTC().Format(time.RFC3339),
				},
			},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	opts.APIURL = server.URL
	opts.APIKey = "test-key"
	opts.JSON = true
	defer func() { opts.JSON = false }()

	out := captureStdout(t, func() {
		dnsListZonesCmd.Run(dnsListZonesCmd, nil)
	})

	if !strings.Contains(out, "\"name\": \"example.com\"") {
		t.Fatalf("expected JSON output to include zone name, got: %s", out)
	}
}

func TestDNSCreateZoneFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"id":         uuid.New().String(),
				"name":       "newzone.com",
				"created_at": time.Now().UTC().Format(time.RFC3339),
			},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	opts.APIURL = server.URL
	opts.APIKey = "test-key"

	out := captureStdout(t, func() {
		dnsCreateZoneCmd.Run(dnsCreateZoneCmd, []string{"newzone.com"})
	})

	if !strings.Contains(out, "[SUCCESS] DNS Zone newzone.com created successfully!") {
		t.Fatalf("expected success message, got: %s", out)
	}
}
