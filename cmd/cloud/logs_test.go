package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLogsSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Corrected path from /logs/search to /logs
		if r.URL.Path != "/logs" || r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"entries": []map[string]interface{}{
					{
						"timestamp":     time.Now().Format(time.RFC3339),
						"level":         "INFO",
						"resource_type": "instance",
						"resource_id":   "inst-1",
						"message":       "System started",
					},
				},
				"total": 1,
			},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	oldURL := apiURL
	apiURL = server.URL
	apiKey = "test-key"
	defer func() { 
		apiURL = oldURL 
		apiKey = ""
	}()

	out := captureStdout(t, func() {
		logsSearchCmd.Run(logsSearchCmd, nil)
	})
	if !strings.Contains(out, "INFO") || !strings.Contains(out, "System started") {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestLogsShow(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Corrected path from /logs/resource/ to just /logs/
		if !strings.HasPrefix(r.URL.Path, "/logs/") || r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"entries": []map[string]interface{}{
					{
						"timestamp":     time.Now().Format(time.RFC3339),
						"level":         "ERROR",
						"resource_type": "instance",
						"resource_id":   "inst-1",
						"message":       "Critical failure",
					},
				},
				"total": 1,
			},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	oldURL := apiURL
	apiURL = server.URL
	apiKey = "test-key"
	defer func() { 
		apiURL = oldURL 
		apiKey = ""
	}()

	out := captureStdout(t, func() {
		logsShowCmd.Run(logsShowCmd, []string{"inst-1"})
	})
	if !strings.Contains(out, "ERROR") || !strings.Contains(out, "Critical failure") {
		t.Fatalf("unexpected output: %s", out)
	}
}
