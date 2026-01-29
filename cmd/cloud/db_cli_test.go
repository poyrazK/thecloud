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
	dbTestAPIKey = "db-key"
	dbTestID     = "db-1"
	dbTestName   = "appdb"
)

func TestDBListJSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/databases" || r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		payload := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":         dbTestID,
					"name":       dbTestName,
					"engine":     "postgres",
					"version":    "16",
					"status":     "available",
					"port":       5432,
					"username":   "admin",
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
	apiKey = dbTestAPIKey
	outputJSON = true
	defer func() {
		apiURL = oldURL
		apiKey = oldKey
		outputJSON = false
	}()

	out := captureStdout(t, func() {
		dbListCmd.Run(dbListCmd, nil)
	})
	if !strings.Contains(out, dbTestID) {
		t.Fatalf("expected JSON output to include database id, got: %s", out)
	}
}

func TestDBCreateJSONMasksPassword(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/databases" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"id":         dbTestID,
				"name":       dbTestName,
				"engine":     "postgres",
				"version":    "16",
				"status":     "available",
				"port":       5432,
				"username":   "admin",
				"password":   "secret",
				"created_at": time.Now().UTC().Format(time.RFC3339),
				"updated_at": time.Now().UTC().Format(time.RFC3339),
			},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	oldURL := apiURL
	oldKey := apiKey
	apiURL = server.URL
	apiKey = dbTestAPIKey
	outputJSON = true
	defer func() {
		apiURL = oldURL
		apiKey = oldKey
		outputJSON = false
	}()

	_ = dbCreateCmd.Flags().Set("name", dbTestName)
	_ = dbCreateCmd.Flags().Set("engine", "postgres")
	_ = dbCreateCmd.Flags().Set("version", "16")

	out := captureStdout(t, func() {
		dbCreateCmd.Run(dbCreateCmd, nil)
	})
	if strings.Contains(out, "secret") {
		t.Fatalf("expected password to be masked, got: %s", out)
	}
	if !strings.Contains(out, "********") {
		t.Fatalf("expected masked password, got: %s", out)
	}
}

func TestDBConnectionCmd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/databases/"+dbTestID+"/connection" || r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		payload := map[string]interface{}{
			"data": map[string]string{"connection_string": "postgres://user:pass@localhost:5432/db"},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	oldURL := apiURL
	oldKey := apiKey
	apiURL = server.URL
	apiKey = dbTestAPIKey
	defer func() {
		apiURL = oldURL
		apiKey = oldKey
	}()

	out := captureStdout(t, func() {
		dbConnCmd.Run(dbConnCmd, []string{dbTestID})
	})
	if !strings.Contains(out, "Connection String") {
		t.Fatalf("expected connection string output, got: %s", out)
	}
}
