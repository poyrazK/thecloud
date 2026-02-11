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

func TestElasticIPE2E(t *testing.T) {
	t.Parallel()
	if err := waitForServer(); err != nil {
		t.Fatalf("Failing Elastic IP E2E test: %v", err)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	token := registerAndLogin(t, client, "eip-tester@thecloud.local", "EIP Tester")

	var eipID string
	var publicIP string

	// 1. Allocate Elastic IP
	t.Run("AllocateIP", func(t *testing.T) {
		resp := postRequest(t, client, testutil.TestBaseURL+"/elastic-ips", token, nil)
		defer func() { _ = resp.Body.Close() }()

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
		resp := getRequest(t, client, testutil.TestBaseURL+"/elastic-ips", token)
		defer func() { _ = resp.Body.Close() }()

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
		resp := getRequest(t, client, testutil.TestBaseURL+"/elastic-ips/"+eipID, token)
		defer func() { _ = resp.Body.Close() }()

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
		instanceName := fmt.Sprintf("eip-inst-%d-%s", time.Now().UnixNano()%1000, uuid.New().String()[:8])
		payload := map[string]string{
			"name":  instanceName,
			"image": "nginx:alpine",
		}
		resp := postRequest(t, client, testutil.TestBaseURL+testutil.TestRouteInstances, token, payload)
		defer func() { _ = resp.Body.Close() }()

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

		// Wait briefly for instance record to be available (E2E async launch)
		time.Sleep(2 * time.Second)

		// Associate EIP
		assocPayload := map[string]string{
			"instance_id": instanceID,
		}
		resp = postRequest(t, client, testutil.TestBaseURL+"/elastic-ips/"+eipID+"/associate", token, assocPayload)
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var eipRes struct {
			Data domain.ElasticIP `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&eipRes))
		assert.Equal(t, domain.EIPStatusAssociated, eipRes.Data.Status)
		require.NotNil(t, eipRes.Data.InstanceID)
		assert.Equal(t, instanceID, eipRes.Data.InstanceID.String())

		// Disassociate
		resp = postRequest(t, client, testutil.TestBaseURL+"/elastic-ips/"+eipID+"/disassociate", token, nil)
		defer func() { _ = resp.Body.Close() }()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		// Cleanup: terminate instance
		termResp := deleteRequest(t, client, testutil.TestBaseURL+testutil.TestRouteInstances+"/"+instanceID, token)
		_ = termResp.Body.Close()
	})

	// 5. Release Elastic IP
	t.Run("ReleaseIP", func(t *testing.T) {
		resp := deleteRequest(t, client, testutil.TestBaseURL+"/elastic-ips/"+eipID, token)
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify deleted
		resp = getRequest(t, client, testutil.TestBaseURL+"/elastic-ips/"+eipID, token)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
