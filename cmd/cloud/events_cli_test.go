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

const (
	eventsTestAPIKey = "events-key"
)

func TestEventsListJSONOutput(t *testing.T) {
	eventID := uuid.New().String()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/events" || r.Method != http.MethodGet || r.URL.Query().Get("limit") != "50" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		payload := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":            eventID,
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
	if !strings.Contains(out, eventID) {
		t.Fatalf("expected JSON output to include event id, got: %s", out)
	}
}
