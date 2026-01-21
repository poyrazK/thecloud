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

func TestAutoScalingE2E(t *testing.T) {
	if err := waitForServer(); err != nil {
		t.Skipf("Skipping AutoScaling E2E test: %v", err)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	token := registerAndLogin(t, client, "asg-tester@thecloud.local", "ASG Tester")

	// 1. Setup VPC
	var vpcID string
	t.Run("SetupVPC", func(t *testing.T) {
		payload := map[string]string{
			"name":       "asg-vpc",
			"cidr_block": "10.10.0.0/16",
		}
		resp := postRequest(t, client, testutil.TestBaseURL+"/vpcs", token, payload)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var res struct{ Data domain.VPC }
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		vpcID = res.Data.ID.String()
	})

	var groupID string
	groupName := fmt.Sprintf("e2e-asg-%d", time.Now().UnixNano()%1000)

	// 2. Create Scaling Group
	t.Run("CreateGroup", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":          groupName,
			"vpc_id":        vpcID,
			"image":         "alpine",
			"ports":         "80:80",
			"min_instances": 1,
			"max_instances": 5,
			"desired_count": 2,
		}
		resp := postRequest(t, client, testutil.TestBaseURL+"/autoscaling/groups", token, payload)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var res struct {
			Data domain.ScalingGroup `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		groupID = res.Data.ID.String()
		assert.NotEmpty(t, groupID)
	})

	// 3. Create Scaling Policy
	var policyID string
	t.Run("CreatePolicy", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":           "scale-up-cpu",
			"metric_type":    "cpu_utilization",
			"target_value":   70.0,
			"scale_out_step": 1,
			"scale_in_step":  1,
			"cooldown_sec":   300,
		}
		resp := postRequest(t, client, fmt.Sprintf("%s/autoscaling/groups/%s/policies", testutil.TestBaseURL, groupID), token, payload)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var res struct {
			Data domain.ScalingPolicy `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		policyID = res.Data.ID.String()
		assert.NotEmpty(t, policyID)
	})

	// 4. Cleanup
	t.Run("Cleanup", func(t *testing.T) {
		// Delete Policy
		resp := deleteRequest(t, client, fmt.Sprintf("%s/autoscaling/policies/%s", testutil.TestBaseURL, policyID), token)
		resp.Body.Close()
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		// Delete Group
		resp = deleteRequest(t, client, fmt.Sprintf("%s/autoscaling/groups/%s", testutil.TestBaseURL, groupID), token)
		resp.Body.Close()
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		// Delete VPC
		resp = deleteRequest(t, client, fmt.Sprintf("%s/vpcs/%s", testutil.TestBaseURL, vpcID), token)
		resp.Body.Close()
	})
}
