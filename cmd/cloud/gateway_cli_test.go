package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	gatewayTestAPIKey = "gateway-key"
	gatewayTestID     = "route-1"
)

func TestGatewayCreateRouteCmd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/gateway/routes" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"id":           gatewayTestID,
				"name":         "api",
				"path_prefix":  "/api",
				"target_url":   "https://example.com",
				"strip_prefix": true,
			},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	oldURL := apiURL
	oldKey := apiKey
	apiURL = server.URL
	apiKey = gatewayTestAPIKey
	defer func() {
		apiURL = oldURL
		apiKey = oldKey
	}()

	_ = createRouteCmd.Flags().Set("strip", "true")
	_ = createRouteCmd.Flags().Set("rate-limit", "100")

	out := captureStdout(t, func() {
		createRouteCmd.Run(createRouteCmd, []string{"api", "/api", "https://example.com"})
	})
	if !strings.Contains(out, "Route created") {
		t.Fatalf("expected create output, got: %s", out)
	}
}

func TestGatewayDeleteRouteCmd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/gateway/routes/"+gatewayTestID || r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	oldURL := apiURL
	oldKey := apiKey
	apiURL = server.URL
	apiKey = gatewayTestAPIKey
	defer func() {
		apiURL = oldURL
		apiKey = oldKey
	}()

	out := captureStdout(t, func() {
		deleteRouteCmd.Run(deleteRouteCmd, []string{gatewayTestID})
	})
	if !strings.Contains(out, "Route deleted") {
		t.Fatalf("expected delete output, got: %s", out)
	}
}
