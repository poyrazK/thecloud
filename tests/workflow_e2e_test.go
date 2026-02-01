package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const routeFormat = "%s%s/%s"

func TestFullWorkflowE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test")
	}

	if err := waitForServer(); err != nil {
		t.Fatalf("Server not ready: %v", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	token := registerAndLogin(t, client, fmt.Sprintf("workflow-%d@thecloud.local", time.Now().UnixNano()), "Workflow User")

	var vpcID string
	var subnetID string
	var instanceID string
	var zoneID string
	var instanceTypeID string

	// 0. List Instance Types
	t.Run("ListInstanceTypes", func(t *testing.T) {
		resp := getRequest(t, client, testutil.TestBaseURL+"/instance-types", token)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var res struct {
			Data []*domain.InstanceType `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		require.NotEmpty(t, res.Data)
		instanceTypeID = res.Data[0].ID
	})

	// 1. Create VPC
	t.Run("CreateVPC", func(t *testing.T) {
		payload := map[string]string{
			"name":       fmt.Sprintf("workflow-vpc-%d", time.Now().UnixNano()),
			"cidr_block": "10.60.0.0/16",
		}
		resp := postRequest(t, client, testutil.TestBaseURL+testutil.TestRouteVpcs, token, payload)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var res struct {
			Data domain.VPC `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		vpcID = res.Data.ID.String()
	})

	// 2. Create Subnet
	t.Run("CreateSubnet", func(t *testing.T) {
		payload := map[string]string{
			"name":       "workflow-subnet",
			"cidr_block": "10.60.1.0/24",
		}
		resp := postRequest(t, client, fmt.Sprintf("%s/vpcs/%s/subnets", testutil.TestBaseURL, vpcID), token, payload)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var res struct {
			Data domain.Subnet `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		subnetID = res.Data.ID.String()
	})

	// 3. Launch Instance
	t.Run("LaunchInstance", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":             "workflow-inst",
			"image":            "alpine",
			"vpc_id":           vpcID,
			"subnet_id":        subnetID,
			"instance_type_id": instanceTypeID,
		}
		resp := postRequest(t, client, testutil.TestBaseURL+testutil.TestRouteInstances, token, payload)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var res struct {
			Data domain.Instance `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		instanceID = res.Data.ID.String()
	})

	// 4. Create DNS Zone
	t.Run("CreateDNSZone", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":   fmt.Sprintf("workflow-%d.test", time.Now().UnixNano()),
			"vpc_id": vpcID,
		}
		resp := postRequest(t, client, testutil.TestBaseURL+testutil.TestRouteDNSZones, token, payload)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var res struct {
			Data domain.DNSZone `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		zoneID = res.Data.ID.String()
	})

	// 5. Create DNS Record
	t.Run("CreateDNSRecord", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":    "app",
			"type":    "A",
			"content": "10.60.1.10",
			"ttl":     300,
		}
		resp := postRequest(t, client, fmt.Sprintf("%s/dns/zones/%s/records", testutil.TestBaseURL, zoneID), token, payload)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var res struct {
			Data domain.DNSRecord `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		assert.Contains(t, res.Data.Name, "app.")
	})

	// 6. Cleanup
	t.Run("Cleanup", func(t *testing.T) {
		// Delete Instance
		resp := deleteRequest(t, client, fmt.Sprintf(routeFormat, testutil.TestBaseURL, testutil.TestRouteInstances, instanceID), token)
		resp.Body.Close()
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		// Delete Zone
		resp = deleteRequest(t, client, fmt.Sprintf(routeFormat, testutil.TestBaseURL, testutil.TestRouteDNSZones, zoneID), token)
		resp.Body.Close()
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		// Delete VPC
		resp = deleteRequest(t, client, fmt.Sprintf(routeFormat, testutil.TestBaseURL, testutil.TestRouteVpcs, vpcID), token)
		resp.Body.Close()
		assert.Contains(t, []int{http.StatusOK, http.StatusNoContent}, resp.StatusCode)
	})
}
