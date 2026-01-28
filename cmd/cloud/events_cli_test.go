package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const (
	eventsTestAPIKey = "events-key"
)

func TestEventsListJSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/events" || r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		payload := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":            "evt-1",
					"action":        "create",
					"resource_id":   "res-1",
					"resource_type": "instance",
					"metadata":      "details",
					"created_at":    time.Now().UTC().Format(time.RFC3339),
				},
			},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	oldURL := apiURL
	oldKey := apiKey
	apiURL = server.URL
	apiKey = eventsTestAPIKey
	outputJSON = true
	defer func() {
		apiURL = oldURL
		apiKey = oldKey
		outputJSON = false
	}()

	out := captureStdout(t, func() {
		listEventsCmd.Run(listEventsCmd, nil)
	})
	if !strings.Contains(out, "evt-1") {
		t.Fatalf("expected JSON output to include event id, got: %s", out)
	}
}
