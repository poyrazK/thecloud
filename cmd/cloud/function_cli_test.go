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
	functionTestAPIKey = "function-key"
	functionTestID     = "fn-1"
	functionTestName   = "hello"
)

func TestFunctionCreateCmd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/functions" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"id":         functionTestID,
				"name":       functionTestName,
				"runtime":    "nodejs20",
				"status":     "active",
				"created_at": time.Now().UTC().Format(time.RFC3339),
			},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	codePath := filepath.Join(t.TempDir(), "fn.zip")
	if err := os.WriteFile(codePath, []byte("zip"), 0644); err != nil {
		t.Fatalf("write code: %v", err)
	}

	oldURL := apiURL
	oldKey := apiKey
	apiURL = server.URL
	apiKey = functionTestAPIKey
	defer func() {
		apiURL = oldURL
		apiKey = oldKey
	}()

	_ = createFnCmd.Flags().Set("name", functionTestName)
	_ = createFnCmd.Flags().Set("runtime", "nodejs20")
	_ = createFnCmd.Flags().Set("handler", "index.handler")
	_ = createFnCmd.Flags().Set("code", codePath)

	out := captureStdout(t, func() {
		if err := createFnCmd.RunE(createFnCmd, nil); err != nil {
			t.Fatalf("create function: %v", err)
		}
	})
	if !strings.Contains(out, "Function") || !strings.Contains(out, functionTestID) {
		t.Fatalf("expected create output, got: %s", out)
	}
}

func TestFunctionListEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/functions" || r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		payload := map[string]interface{}{
			"data": []map[string]interface{}{},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	oldURL := apiURL
	oldKey := apiKey
	apiURL = server.URL
	apiKey = functionTestAPIKey
	defer func() {
		apiURL = oldURL
		apiKey = oldKey
	}()

	out := captureStdout(t, func() {
		if err := listFnCmd.RunE(listFnCmd, nil); err != nil {
			t.Fatalf("list function: %v", err)
		}
	})
	if !strings.Contains(out, "No functions found") {
		t.Fatalf("expected empty list output, got: %s", out)
	}
}
