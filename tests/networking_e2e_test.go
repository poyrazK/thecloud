package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/pkg/testutil"
)

func TestNetworkingE2E(t *testing.T) {
	if err := waitForServer(); err != nil {
		t.Skipf("Skipping Networking E2E test: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	token := registerAndLogin(t, client, "network-tester@thecloud.local", "Network Tester")

	var vpcID string
	vpcName := fmt.Sprintf("e2e-vpc-%d", time.Now().UnixNano()%1000)

	// 1. Create VPC
	t.Run("CreateVPC", func(t *testing.T) {
		payload := map[string]string{
			"name":       vpcName,
			"cidr_block": "10.0.0.0/16",
		}
		resp := postRequest(t, client, testutil.TestBaseURL+"/vpcs", token, payload)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var res struct {
			Data domain.VPC `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		vpcID = res.Data.ID.String()
		assert.NotEmpty(t, vpcID)
		assert.Equal(t, vpcName, res.Data.Name)
	})

	// 2. Create Subnet
	var subnetID string
	t.Run("CreateSubnet", func(t *testing.T) {
		payload := map[string]string{
			"name":       "e2e-subnet",
			"cidr_block": "10.0.1.0/24",
		}
		resp := postRequest(t, client, fmt.Sprintf("%s/vpcs/%s/subnets", testutil.TestBaseURL, vpcID), token, payload)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var res struct {
			Data domain.Subnet `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		subnetID = res.Data.ID.String()
		assert.NotEmpty(t, subnetID)
	})

	// 3. List Subnets
	t.Run("ListSubnets", func(t *testing.T) {
		resp := getRequest(t, client, fmt.Sprintf("%s/vpcs/%s/subnets", testutil.TestBaseURL, vpcID), token)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res struct {
			Data []domain.Subnet `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		assert.True(t, len(res.Data) >= 1)
	})

	// 4. Create Security Group
	var sgID string
	t.Run("CreateSecurityGroup", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":        "e2e-sg",
			"description": "E2E security group",
			"vpc_id":      vpcID,
		}
		resp := postRequest(t, client, testutil.TestBaseURL+"/security-groups", token, payload)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var res struct {
			Data domain.SecurityGroup `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		sgID = res.Data.ID.String()
	})

	// 5. Load Balancer
	var lbID string
	t.Run("CreateLoadBalancer", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":      "e2e-lb",
			"vpc_id":    vpcID,
			"port":      80,
			"algorithm": "round-robin",
		}
		resp := postRequest(t, client, testutil.TestBaseURL+"/lb", token, payload)
		defer resp.Body.Close()

		require.Equal(t, http.StatusAccepted, resp.StatusCode)

		var res struct {
			Data domain.LoadBalancer `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		lbID = res.Data.ID.String()
		assert.NotEmpty(t, lbID)
	})

	// 6. Cleanup
	t.Run("Cleanup", func(t *testing.T) {
		// Delete LB
		resp := deleteRequest(t, client, fmt.Sprintf("%s/lb/%s", testutil.TestBaseURL, lbID), token)
		resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Delete Security Group
		resp = deleteRequest(t, client, fmt.Sprintf("%s/security-groups/%s", testutil.TestBaseURL, sgID), token)
		resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Delete Subnet
		resp = deleteRequest(t, client, fmt.Sprintf("%s/subnets/%s", testutil.TestBaseURL, subnetID), token)
		resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Delete VPC
		resp = deleteRequest(t, client, fmt.Sprintf("%s/vpcs/%s", testutil.TestBaseURL, vpcID), token)
		resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
