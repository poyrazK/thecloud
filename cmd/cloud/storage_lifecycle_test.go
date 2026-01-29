package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	lifecycleTestBucket    = "logs"
	lifecycleTestContent   = "Content-Type"
	lifecycleTestAppJSON   = "application/json"
	lifecycleTestAPIKey    = "lifecycle-key"
	lifecycleTestRuleID    = "rule-1"
	lifecycleTestCreatedAt = "2024-01-01T00:00:00Z"
)

func TestLifecycleSetRejectsInvalidDays(t *testing.T) {
	out := captureStdout(t, func() {
		_ = lifecycleSetCmd.Flags().Set("prefix", "logs/")
		_ = lifecycleSetCmd.Flags().Set("days", "0")
		_ = lifecycleSetCmd.Flags().Set("enabled", "true")
		lifecycleSetCmd.Run(lifecycleSetCmd, []string{lifecycleTestBucket})
	})
	if !strings.Contains(out, "--days must be at least 1") {
		t.Fatalf("expected validation error, got: %s", out)
	}
}

func TestLifecycleListJSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/storage/buckets/"+lifecycleTestBucket+"/lifecycle" || r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set(lifecycleTestContent, lifecycleTestAppJSON)
		payload := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":              lifecycleTestRuleID,
					"bucket_name":     lifecycleTestBucket,
					"prefix":          "logs/",
					"expiration_days": 30,
					"enabled":         true,
					"created_at":      lifecycleTestCreatedAt,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	oldURL := apiURL
	oldKey := apiKey
	apiURL = server.URL
	apiKey = lifecycleTestAPIKey
	outputJSON = true
	defer func() {
		apiURL = oldURL
		apiKey = oldKey
		outputJSON = false
	}()

	out := captureStdout(t, func() {
		lifecycleListCmd.Run(lifecycleListCmd, []string{lifecycleTestBucket})
	})
	if !strings.Contains(out, lifecycleTestRuleID) {
		t.Fatalf("expected JSON output to include rule id, got: %s", out)
	}
}

func TestLifecycleDeleteSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/storage/buckets/"+lifecycleTestBucket+"/lifecycle/"+lifecycleTestRuleID || r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set(lifecycleTestContent, lifecycleTestAppJSON)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	oldURL := apiURL
	oldKey := apiKey
	apiURL = server.URL
	apiKey = lifecycleTestAPIKey
	defer func() {
		apiURL = oldURL
		apiKey = oldKey
	}()

	out := captureStdout(t, func() {
		lifecycleDeleteCmd.Run(lifecycleDeleteCmd, []string{lifecycleTestBucket, lifecycleTestRuleID})
	})
	if !strings.Contains(out, "Deleted lifecycle rule") {
		t.Fatalf("expected success output, got: %s", out)
	}
}
