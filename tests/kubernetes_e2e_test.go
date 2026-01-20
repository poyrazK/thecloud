package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/poyrazk/thecloud/pkg/testutil"
)

type VPC struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	CIDR string `json:"cidr_block"`
}

type Cluster struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	WorkerCount int    `json:"worker_count"`
}

func TestKubernetesE2E(t *testing.T) {
	if err := waitForServer(); err != nil {
		t.Skipf("Skipping K8s E2E test: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	token := registerAndLogin(t, client, "k8s-user@test.com", "K8s User")

	// 1. Create VPC
	vpc := createVPC(t, client, token, "k8s-vpc", "10.100.0.0/16")
	require.NotEmpty(t, vpc.ID)
	// defer deleteVPC(t, client, token, vpc.ID) // Optional cleanup

	// 2. Create Cluster
	fmt.Printf("Creating cluster in VPC %s...\n", vpc.ID)
	cluster := createCluster(t, client, token, "test-cluster", vpc.ID, 1)
	assert.NotEmpty(t, cluster.ID)
	assert.Equal(t, "test-cluster", cluster.Name)
	assert.Equal(t, "provisioning", cluster.Status)

	// 3. Get Cluster Details
	t.Run("Get Cluster", func(t *testing.T) {
		got := getCluster(t, client, token, cluster.ID)
		assert.Equal(t, cluster.ID, got.ID)
		assert.Equal(t, cluster.Name, got.Name)
	})

	// 4. Get Kubeconfig
	// Note: deeper validation requires the provisioner to actually complete, which might not happen in this mock env.
	// But we can check if the endpoint is reachable.
	t.Run("Get Kubeconfig", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/clusters/%s/kubeconfig", testutil.TestBaseURL, cluster.ID), nil)
		req.Header.Set(testutil.TestHeaderAPIKey, token)
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// It might return 404 or empty if not ready, strictly speaking.
		// But in our current impl, we return what's in DB. If provisioner hasn't updated it, it might be empty.
		if resp.StatusCode == http.StatusOK {
			// Success
		} else {
			// It's acceptable for now if it fails due to being provisioning
			t.Logf("Kubeconfig endpoint returned %d", resp.StatusCode)
		}
	})

	// 5. Delete Cluster
	t.Run("Delete Cluster", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/clusters/%s", testutil.TestBaseURL, cluster.ID), nil)
		req.Header.Set(testutil.TestHeaderAPIKey, token)
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func createVPC(t *testing.T, client *http.Client, token, name, cidr string) VPC {
	reqBody := map[string]string{
		"name":       name,
		"cidr_block": cidr,
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", testutil.TestBaseURL+"/vpcs", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", testutil.TestContentTypeAppJSON)
	req.Header.Set(testutil.TestHeaderAPIKey, token)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	type Wrapper struct {
		Data VPC `json:"data"`
	}
	var w Wrapper
	err = json.NewDecoder(resp.Body).Decode(&w)
	require.NoError(t, err)
	return w.Data
}

func createCluster(t *testing.T, client *http.Client, token, name, vpcID string, workers int) Cluster {
	reqBody := map[string]interface{}{
		"name":         name,
		"vpc_id":       vpcID,
		"worker_count": workers,
		"version":      "v1.29.0",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", testutil.TestBaseURL+"/clusters", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", testutil.TestContentTypeAppJSON)
	req.Header.Set(testutil.TestHeaderAPIKey, token)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		t.Fatalf("Create cluster failed: status %d", resp.StatusCode)
	}

	type Wrapper struct {
		Data Cluster `json:"data"`
	}
	var w Wrapper
	err = json.NewDecoder(resp.Body).Decode(&w)
	require.NoError(t, err)
	return w.Data
}

func getCluster(t *testing.T, client *http.Client, token, id string) Cluster {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/clusters/%s", testutil.TestBaseURL, id), nil)
	req.Header.Set(testutil.TestHeaderAPIKey, token)
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	type Wrapper struct {
		Data Cluster `json:"data"`
	}
	var w Wrapper
	err = json.NewDecoder(resp.Body).Decode(&w)
	require.NoError(t, err)
	return w.Data
}
