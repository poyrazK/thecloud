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
	subnetTestAPIKey = "subnet-key"
	subnetTestID     = "subnet-1"
)

func TestSubnetListJSONOutput(t *testing.T) {
	vpcID := "vpc-1"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/vpcs/"+vpcID+"/subnets" || r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		payload := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":                subnetTestID,
					"vpc_id":            vpcID,
					"name":              "app",
					"cidr_block":        "10.0.1.0/24",
					"availability_zone": "us-east-1a",
					"gateway_ip":        "10.0.1.1",
					"status":            "active",
					"created_at":        time.Now().UTC().Format(time.RFC3339),
				},
			},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	oldURL := apiURL
	oldKey := apiKey
	apiURL = server.URL
	apiKey = subnetTestAPIKey
	outputJSON = true
	defer func() {
		apiURL = oldURL
		apiKey = oldKey
		outputJSON = false
	}()

	out := captureStdout(t, func() {
		subnetListCmd.Run(subnetListCmd, []string{vpcID})
	})
	if !strings.Contains(out, subnetTestID) {
		t.Fatalf("expected JSON output to include subnet id, got: %s", out)
	}
}
