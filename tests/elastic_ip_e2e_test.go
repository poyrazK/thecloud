package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/pkg/testutil"
)

const (
	eipTestHTTPTimeout      = 60 * time.Second
	eipInstanceNameSalt     = 1000
	eipInstanceIDSliceLen   = 8
	eipInstanceReadyTimeout = 30 * time.Second
	eipInstancePollInterval = 1 * time.Second
	eipBaseURL              = "/elastic-ips"
)

func closeResponse(t *testing.T, resp *http.Response) {
	if resp == nil {
		return
	}
	if err := resp.Body.Close(); err != nil {
		t.Errorf("failed to close response body: %v", err)
	}
}

func waitForInstanceReady(t *testing.T, client *http.Client, token, instanceID string) bool {
	deadline := time.Now().Add(eipInstanceReadyTimeout)
	for time.Now().Before(deadline) {
		checkResp := getRequest(t, client, testutil.TestBaseURL+testutil.TestRouteInstances+"/"+instanceID, token)
		var statusRes struct {
			Data struct {
				Status domain.InstanceStatus `json:"status"`
			} `json:"data"`
		}
		if err := json.NewDecoder(checkResp.Body).Decode(&statusRes); err != nil {
			t.Logf("Failed to decode instance status response: %v (Status: %d)", err, checkResp.StatusCode)
		}
		closeResponse(t, checkResp)

		if statusRes.Data.Status == domain.StatusRunning {
			return true
		}
		time.Sleep(eipInstancePollInterval)
	}
	return false
}

func TestElasticIPE2E(t *testing.T) {
	t.Parallel()
	if err := waitForServer(); err != nil {
		t.Fatalf("Failing Elastic IP E2E test: %v", err)
	}

	client := &http.Client{Timeout: eipTestHTTPTimeout}
	token := registerAndLogin(t, client, "eip-tester@thecloud.local", "EIP Tester")

	var eipID string
	var publicIP string

	// 1. Allocate Elastic IP
	t.Run("AllocateIP", func(t *testing.T) {
		resp := postRequest(t, client, testutil.TestBaseURL+eipBaseURL, token, nil)
		defer closeResponse(t, resp)

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var res struct {
			Data domain.ElasticIP `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		eipID = res.Data.ID.String()
		publicIP = res.Data.PublicIP
		assert.NotEmpty(t, eipID)
		assert.NotEmpty(t, publicIP)
		assert.Equal(t, domain.EIPStatusAllocated, res.Data.Status)
	})

	// 2. List Elastic IPs
	t.Run("ListIPs", func(t *testing.T) {
		resp := getRequest(t, client, testutil.TestBaseURL+eipBaseURL, token)
		defer closeResponse(t, resp)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var res struct {
			Data []domain.ElasticIP `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))

		found := false
		for _, eip := range res.Data {
			if eip.ID.String() == eipID {
				found = true
				break
			}
		}
		assert.True(t, found, "allocated EIP not found in list")
	})

	// 3. Get Elastic IP
	t.Run("GetIP", func(t *testing.T) {
		resp := getRequest(t, client, testutil.TestBaseURL+eipBaseURL+"/"+eipID, token)
		defer closeResponse(t, resp)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var res struct {
			Data domain.ElasticIP `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		assert.Equal(t, eipID, res.Data.ID.String())
	})

	// 4. Launch Instance and Associate
	t.Run("AssociateIP", func(t *testing.T) {
		// Launch instance
		instanceName := fmt.Sprintf("eip-inst-%d-%s", time.Now().UnixNano()%eipInstanceNameSalt, uuid.New().String()[:eipInstanceIDSliceLen])
		payload := map[string]string{
			"name":  instanceName,
			"image": "nginx:alpine",
		}
		resp := postRequest(t, client, testutil.TestBaseURL+testutil.TestRouteInstances, token, payload)
		defer closeResponse(t, resp)

		if resp.StatusCode != http.StatusAccepted {
			t.Skip("Skipping association test: failed to launch instance (Docker might be unavailable)")
			return
		}

		var instRes struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&instRes))
		instanceID := instRes.Data.ID

		// Poll for instance readiness
		if !waitForInstanceReady(t, client, token, instanceID) {
			t.Logf("Instance %s did not become ready in time, skipping association test", instanceID)
			// Still try to cleanup
			termResp := deleteRequest(t, client, testutil.TestBaseURL+testutil.TestRouteInstances+"/"+instanceID, token)
			closeResponse(t, termResp)
			t.SkipNow()
		}

		// Associate EIP
		assocPayload := map[string]string{
			"instance_id": instanceID,
		}
		// Create new response variable to avoid double-closing deferred 'resp'
		assocResp := postRequest(t, client, testutil.TestBaseURL+eipBaseURL+"/"+eipID+"/associate", token, assocPayload)
		defer closeResponse(t, assocResp)

		require.Equal(t, http.StatusOK, assocResp.StatusCode)

		var eipRes struct {
			Data domain.ElasticIP `json:"data"`
		}
		require.NoError(t, json.NewDecoder(assocResp.Body).Decode(&eipRes))
		assert.Equal(t, domain.EIPStatusAssociated, eipRes.Data.Status)
		// Handle potential nil
		if assert.NotNil(t, eipRes.Data.InstanceID) {
			assert.Equal(t, instanceID, eipRes.Data.InstanceID.String())
		}

		// Disassociate
		disassocResp := postRequest(t, client, testutil.TestBaseURL+eipBaseURL+"/"+eipID+"/disassociate", token, nil)
		defer closeResponse(t, disassocResp)
		require.Equal(t, http.StatusOK, disassocResp.StatusCode)

		// Cleanup: terminate instance
		termResp := deleteRequest(t, client, testutil.TestBaseURL+testutil.TestRouteInstances+"/"+instanceID, token)
		closeResponse(t, termResp)
	})

	// 5. Release Elastic IP
	t.Run("ReleaseIP", func(t *testing.T) {
		resp := deleteRequest(t, client, testutil.TestBaseURL+eipBaseURL+"/"+eipID, token)
		defer closeResponse(t, resp)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify deleted
		checkResp := getRequest(t, client, testutil.TestBaseURL+eipBaseURL+"/"+eipID, token)
		defer closeResponse(t, checkResp)
		assert.Equal(t, http.StatusNotFound, checkResp.StatusCode)
	})
}
