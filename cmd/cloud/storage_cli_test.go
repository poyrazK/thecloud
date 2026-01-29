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
	storageTestAPIKey = "storage-key"
	storageTestBucket = "images"
)

func TestStorageListBucketsJSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/storage/buckets" || r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		payload := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":                 "bucket-1",
					"name":               storageTestBucket,
					"is_public":          true,
					"versioning_enabled": false,
					"created_at":         time.Now().UTC().Format(time.RFC3339),
				},
			},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	oldURL := apiURL
	oldKey := apiKey
	apiURL = server.URL
	apiKey = storageTestAPIKey
	outputJSON = true
	defer func() {
		apiURL = oldURL
		apiKey = oldKey
		outputJSON = false
	}()

	out := captureStdout(t, func() {
		storageListCmd.Run(storageListCmd, nil)
	})
	if !strings.Contains(out, storageTestBucket) {
		t.Fatalf("expected JSON output to include bucket name, got: %s", out)
	}
}

func TestStorageVersioningInvalidStatus(t *testing.T) {
	out := captureStdout(t, func() {
		storageVersioningCmd.Run(storageVersioningCmd, []string{storageTestBucket, "invalid"})
	})
	if !strings.Contains(out, "Invalid status") {
		t.Fatalf("expected invalid status message, got: %s", out)
	}
}
