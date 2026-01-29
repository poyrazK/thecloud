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
	vpcTestAPIKey = "vpc-key"
	vpcTestID     = "vpc-1"
)

func TestVPCListJSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/vpcs" || r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		payload := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":         vpcTestID,
					"name":       "main",
					"cidr_block": "10.0.0.0/16",
					"vxlan_id":   1001,
					"status":     "active",
					"created_at": time.Now().UTC().Format(time.RFC3339),
				},
			},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	oldURL := apiURL
	oldKey := apiKey
	apiURL = server.URL
	apiKey = vpcTestAPIKey
	outputJSON = true
	defer func() {
		apiURL = oldURL
		apiKey = oldKey
		outputJSON = false
	}()

	out := captureStdout(t, func() {
		vpcListCmd.Run(vpcListCmd, nil)
	})
	if !strings.Contains(out, vpcTestID) {
		t.Fatalf("expected JSON output to include vpc id, got: %s", out)
	}
}
