package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	iacTestAPIKey = "iac-key"
	iacTestID     = "11111111-1111-1111-1111-111111111111"
)

func TestIACListJSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/iac/stacks" || r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		payload := []map[string]interface{}{
			{
				"id":         iacTestID,
				"name":       "demo",
				"status":     "CREATE_COMPLETE",
				"created_at": time.Now().UTC().Format(time.RFC3339),
			},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	oldURL := apiURL
	oldKey := apiKey
	apiURL = server.URL
	apiKey = iacTestAPIKey
	outputJSON = true
	defer func() {
		apiURL = oldURL
		apiKey = oldKey
		outputJSON = false
	}()

	out := captureStdout(t, func() {
		iacListCmd.Run(iacListCmd, nil)
	})
	if !strings.Contains(out, iacTestID) {
		t.Fatalf("expected JSON output to include stack id, got: %s", out)
	}
}

func TestIACValidateTemplateSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/iac/validate" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		payload := map[string]interface{}{
			"valid":  true,
			"errors": []string{},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	templatePath := filepath.Join(t.TempDir(), "template.yaml")
	if err := os.WriteFile(templatePath, []byte("resources: []"), 0644); err != nil {
		t.Fatalf("write template: %v", err)
	}

	oldURL := apiURL
	oldKey := apiKey
	apiURL = server.URL
	apiKey = iacTestAPIKey
	defer func() {
		apiURL = oldURL
		apiKey = oldKey
	}()

	out := captureStdout(t, func() {
		iacValidateCmd.Run(iacValidateCmd, []string{templatePath})
	})
	if !strings.Contains(out, "Template is valid") {
		t.Fatalf("expected valid template output, got: %s", out)
	}
}
