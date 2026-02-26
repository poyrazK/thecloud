package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

const (
	kubernetesTestContentType = "Content-Type"
	kubernetesTestAppJSON     = "application/json"
	kubernetesTestAPIKey = "kube-test-key"
)

func TestListClustersJSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/clusters" || r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set(kubernetesTestContentType, kubernetesTestAppJSON)
		payload := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":           uuid.New().String(),
					"name":         "kube-1",
					"version":      "v1.29.0",
					"worker_count": 2,
					"status":       "running",
					"created_at":   time.Now().UTC().Format(time.RFC3339),
					"updated_at":   time.Now().UTC().Format(time.RFC3339),
				},
			},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	oldURL := opts.APIURL
	oldKey := opts.APIKey
	opts.APIURL = server.URL
	opts.APIKey = kubernetesTestAPIKey
	opts.JSON = true
	defer func() {
		opts.APIURL = oldURL
		opts.APIKey = oldKey
		opts.JSON = false
	}()

	out := captureStdout(t, func() {
		listClustersCmd.Run(listClustersCmd, nil)
	})
	if !containsAll(out, []string{"kube-1", "worker_count"}) {
		t.Fatalf("expected JSON output to include cluster data, got: %s", out)
	}
}

func TestCreateClusterSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/clusters" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set(kubernetesTestContentType, kubernetesTestAppJSON)
		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"id":           uuid.New().String(),
				"name":         "dev-cluster",
				"version":      "v1.29.0",
				"worker_count": 2,
				"status":       "provisioning",
				"created_at":   time.Now().UTC().Format(time.RFC3339),
				"updated_at":   time.Now().UTC().Format(time.RFC3339),
			},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	oldURL := opts.APIURL
	oldKey := opts.APIKey
	opts.APIURL = server.URL
	opts.APIKey = kubernetesTestAPIKey
	defer func() {
		opts.APIURL = oldURL
		opts.APIKey = oldKey
	}()

	clusterID := uuid.New()
	_ = createClusterCmd.Flags().Set("name", "dev-cluster")
	_ = createClusterCmd.Flags().Set("vpc", clusterID.String())
	_ = createClusterCmd.Flags().Set("version", "v1.29.0")
	_ = createClusterCmd.Flags().Set("workers", "2")
	_ = createClusterCmd.Flags().Set("isolate", "true")
	_ = createClusterCmd.Flags().Set("ha", "false")

	out := captureStdout(t, func() {
		createClusterCmd.Run(createClusterCmd, nil)
	})
	if !containsAll(out, []string{"Cluster creation initiated", "dev-cluster"}) {
		t.Fatalf("expected success output, got: %s", out)
	}
}

func containsAll(out string, values []string) bool {
	for _, v := range values {
		if !strings.Contains(out, v) {
			return false
		}
	}
	return true
}
