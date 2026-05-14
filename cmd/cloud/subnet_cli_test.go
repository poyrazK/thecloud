package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/poyrazk/thecloud/pkg/sdk"
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

	oldURL := opts.APIURL
	oldKey := opts.APIKey
	opts.APIURL = server.URL
	opts.APIKey = subnetTestAPIKey
	opts.JSON = true
	defer func() {
		opts.APIURL = oldURL
		opts.APIKey = oldKey
		opts.JSON = false
	}()

	out := captureStdout(t, func() {
		subnetListCmd.Run(subnetListCmd, []string{vpcID})
	})
	if !strings.Contains(out, subnetTestID) {
		t.Fatalf("expected JSON output to include subnet id, got: %s", out)
	}
}

func TestResolveSubnetIDByName(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/vpcs/*/subnets" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(sdk.Response[[]sdk.Subnet]{
			Data: []sdk.Subnet{
				{ID: "uuid-subnet-1", Name: "app-subnet", VpcID: "vpc-1", CIDRBlock: "10.0.1.0/24"},
			},
		})
	}))
	defer server.Close()

	client := sdk.NewClient(server.URL, "test-key")
	resolved := resolveSubnetID("app-subnet", client)
	if resolved != "uuid-subnet-1" {
		t.Fatalf("expected uuid-subnet-1, got %s", resolved)
	}
}

func TestResolveSubnetIDByUUID(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound) // Should not be called
	}))
	defer server.Close()

	client := sdk.NewClient(server.URL, "test-key")
	id := "abc123-def456"
	resolved := resolveSubnetID(id, client)
	if resolved != id {
		t.Fatalf("expected %s, got %s", id, resolved)
	}
}

func TestResolveVPCID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantServer bool // true if server should be called
	}{
		{
			name:     "resolve by name",
			input:    "my-vpc",
			expected: "uuid-vpc-1",
			wantServer: true,
		},
		{
			name:     "passthrough UUID",
			input:    "abc123-def456",
			expected: "abc123-def456",
			wantServer: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var server *httptest.Server
			if tt.wantServer {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					if r.URL.Path == "/vpcs" {
						err := json.NewEncoder(w).Encode(sdk.Response[[]sdk.VPC]{
							Data: []sdk.VPC{
								{ID: "uuid-vpc-1", Name: "my-vpc", CIDRBlock: "10.0.0.0/16"},
							},
						})
						if err != nil {
							t.Fatalf("failed to encode response: %v", err)
						}
						return
					}
					w.WriteHeader(http.StatusNotFound)
				}))
				defer server.Close()
			} else {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound) // Should not be called
				}))
				defer server.Close()
			}

			client := sdk.NewClient(server.URL, "test-key")
			resolved := resolveVPCID(tt.input, client)
			if resolved != tt.expected {
				t.Fatalf("expected %s, got %s", tt.expected, resolved)
			}
		})
	}
}
