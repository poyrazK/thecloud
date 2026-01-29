package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	cronTestAPIKey   = "cron-key"
	cronTestJobID    = "cron-1"
	cronTestJobName  = "nightly"
	cronTestSchedule = "0 0 * * *"
	cronTestURL      = "https://example.com/hook"
)

func TestCreateCronCmd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/cron/jobs" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		job := map[string]interface{}{
			"id":       cronTestJobID,
			"name":     cronTestJobName,
			"schedule": cronTestSchedule,
			"url":      cronTestURL,
			"status":   "active",
			"next_run": "2024-01-01T00:00:00Z",
		}
		_ = json.NewEncoder(w).Encode(job)
	}))
	defer server.Close()

	oldURL := apiURL
	oldKey := apiKey
	apiURL = server.URL
	apiKey = cronTestAPIKey
	defer func() {
		apiURL = oldURL
		apiKey = oldKey
	}()

	_ = createCronCmd.Flags().Set("method", "POST")
	_ = createCronCmd.Flags().Set("payload", "{}")

	out := captureStdout(t, func() {
		createCronCmd.Run(createCronCmd, []string{cronTestJobName, cronTestSchedule, cronTestURL})
	})
	if !strings.Contains(out, "Cron job created") || !strings.Contains(out, cronTestJobID) {
		t.Fatalf("expected success output, got: %s", out)
	}
}

func TestPauseCronCmd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/cron/jobs/"+cronTestJobID+"/pause" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	oldURL := apiURL
	oldKey := apiKey
	apiURL = server.URL
	apiKey = cronTestAPIKey
	defer func() {
		apiURL = oldURL
		apiKey = oldKey
	}()

	out := captureStdout(t, func() {
		pauseCronCmd.Run(pauseCronCmd, []string{cronTestJobID})
	})
	if !strings.Contains(out, "Job paused") {
		t.Fatalf("expected success output, got: %s", out)
	}
}

func TestDeleteCronCmd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/cron/jobs/"+cronTestJobID || r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	oldURL := apiURL
	oldKey := apiKey
	apiURL = server.URL
	apiKey = cronTestAPIKey
	defer func() {
		apiURL = oldURL
		apiKey = oldKey
	}()

	out := captureStdout(t, func() {
		deleteCronCmd.Run(deleteCronCmd, []string{cronTestJobID})
	})
	if !strings.Contains(out, "Job deleted") {
		t.Fatalf("expected success output, got: %s", out)
	}
}
