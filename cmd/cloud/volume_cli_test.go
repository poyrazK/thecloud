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
	volumeTestAPIKey = "volume-key"
	volumeTestID     = "volume-1"
)

func TestVolumeListJSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/volumes" || r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		payload := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":         uuid.New().String(),
					"name":       volumeTestID,
					"size_gb":    20,
					"status":     "available",
					"created_at": time.Now().UTC().Format(time.RFC3339),
					"updated_at": time.Now().UTC().Format(time.RFC3339),
				},
			},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	oldURL := apiURL
	oldKey := apiKey
	apiURL = server.URL
	apiKey = volumeTestAPIKey
	outputJSON = true
	defer func() {
		apiURL = oldURL
		apiKey = oldKey
		outputJSON = false
	}()

	out := captureStdout(t, func() {
		volumeListCmd.Run(volumeListCmd, nil)
	})
	if !strings.Contains(out, volumeTestID) {
		t.Fatalf("expected JSON output to include volume name, got: %s", out)
	}
}
