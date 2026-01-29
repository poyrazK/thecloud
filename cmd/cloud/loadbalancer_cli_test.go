package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	lbTestAPIKey = "lb-key"
	lbTestID     = "lb-1"
)

func TestLBListJSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/lb" || r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		payload := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":        lbTestID,
					"name":      "public",
					"vpc_id":    "vpc-1",
					"port":      80,
					"algorithm": "round-robin",
					"status":    "ACTIVE",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	oldURL := apiURL
	oldKey := apiKey
	apiURL = server.URL
	apiKey = lbTestAPIKey
	outputJSON = true
	defer func() {
		apiURL = oldURL
		apiKey = oldKey
		outputJSON = false
	}()

	out := captureStdout(t, func() {
		lbListCmd.Run(lbListCmd, nil)
	})
	if !strings.Contains(out, lbTestID) {
		t.Fatalf("expected JSON output to include lb id, got: %s", out)
	}
}

func TestLBAddTargetCmd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/lb/"+lbTestID+"/targets" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	oldURL := apiURL
	oldKey := apiKey
	apiURL = server.URL
	apiKey = lbTestAPIKey
	defer func() {
		apiURL = oldURL
		apiKey = oldKey
	}()

	_ = lbAddTargetCmd.Flags().Set("instance", "inst-1")
	_ = lbAddTargetCmd.Flags().Set("port", "80")
	_ = lbAddTargetCmd.Flags().Set("weight", "1")

	out := captureStdout(t, func() {
		lbAddTargetCmd.Run(lbAddTargetCmd, []string{lbTestID})
	})
	if !strings.Contains(out, "Target") {
		t.Fatalf("expected add target output, got: %s", out)
	}
}
