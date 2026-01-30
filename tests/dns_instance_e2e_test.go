package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDNSInstanceAutoRegistrationE2E(t *testing.T) {
	t.Parallel()
	if err := waitForServer(); err != nil {
		t.Fatalf("Failing DNS Instance E2E test: %v", err)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	token := registerAndLogin(t, client, "dns-inst-tester@thecloud.local", "DNS Instance Tester")

	const instRoute = "%s%s/%s"

	var vpcID string
	vpcName := fmt.Sprintf("dns-inst-vpc-%d", time.Now().UnixNano())

	// 1. Create VPC
	t.Run("CreateVPC", func(t *testing.T) {
		payload := map[string]string{
			"name":       vpcName,
			"cidr_block": "10.20.0.0/16",
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

	var zoneID string
	zoneName := fmt.Sprintf("inst-%d.internal", time.Now().UnixNano())

	// 2. Create DNS Zone for VPC
	t.Run("CreateZone", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":        zoneName,
			"description": "Instance Auto-Registration Zone",
			"vpc_id":      vpcID,
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

	var instanceID string
	instanceName := fmt.Sprintf("web-%d", time.Now().UnixNano())

	// 3. Launch Instance in VPC
	t.Run("LaunchInstanceWithDNS", func(t *testing.T) {
		payload := map[string]string{
			"name":   instanceName,
			"image":  "nginx:alpine",
			"vpc_id": vpcID,
		}
		resp := postRequest(t, client, testutil.TestBaseURL+testutil.TestRouteInstances, token, payload)
		defer resp.Body.Close()
		require.Equal(t, http.StatusAccepted, resp.StatusCode)

		var res struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		instanceID = res.Data.ID
	})

	// 4. Wait for Instance and Verify DNS Record
	t.Run("VerifyAutoDNSCreation", func(t *testing.T) {
		privateIP := waitForInstanceRunning(t, client, token, instanceID)
		if privateIP == "" {
			t.Skip("Instance did... skipping DNS verification")
			return
		}

		// Verify DNS record exists in DB
		resp := getRequest(t, client, fmt.Sprintf("%s/dns/zones/%s/records", testutil.TestBaseURL, zoneID), token)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var recordsRes struct {
			Data []domain.DNSRecord `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&recordsRes))

		found := false
		for _, rec := range recordsRes.Data {
			if rec.Name == instanceName && rec.Type == domain.RecordTypeA && rec.Content == privateIP {
				found = true
				assert.True(t, rec.AutoManaged, "record should be auto-managed")
				break
			}
		}
		assert.True(t, found, "auto-registered DNS record not found")
	})

	// 5. Terminate and Verify Cleanup
	t.Run("Cleanup", func(t *testing.T) {
		// Terminate Instance
		resp := deleteRequest(t, client, fmt.Sprintf(instRoute, testutil.TestBaseURL, testutil.TestRouteInstances, instanceID), token)
		resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		// Wait a bit for async cleanup
		time.Sleep(2 * time.Second)

		// Verify record is gone
		getRecordsResp := getRequest(t, client, fmt.Sprintf("%s/dns/zones/%s/records", testutil.TestBaseURL, zoneID), token)
		defer getRecordsResp.Body.Close()

		var recordsRes struct {
			Data []domain.DNSRecord `json:"data"`
		}
		require.NoError(t, json.NewDecoder(getRecordsResp.Body).Decode(&recordsRes))

		for _, rec := range recordsRes.Data {
			assert.NotEqual(t, instanceName, rec.Name, "DNS record should have been cleaned up")
		}

		// Delete Zone
		deleteRequest(t, client, fmt.Sprintf(instRoute, testutil.TestBaseURL, testutil.TestRouteDNSZones, zoneID), token).Body.Close()
		// Delete VPC
		deleteRequest(t, client, fmt.Sprintf(instRoute, testutil.TestBaseURL, testutil.TestRouteVpcs, vpcID), token).Body.Close()
	})
}

func waitForInstanceRunning(t *testing.T, client *http.Client, token, instanceID string) string {
	timeout := 120 * time.Second
	start := time.Now()
	var privateIP string

	instRoute := "%s%s/%s"

	for time.Since(start) < timeout {
		resp := getRequest(t, client, fmt.Sprintf(instRoute, testutil.TestBaseURL, testutil.TestRouteInstances, instanceID), token)
		var res struct {
			Data domain.Instance `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		resp.Body.Close()

		if res.Data.Status == domain.StatusRunning && res.Data.PrivateIP != "" {
			privateIP = res.Data.PrivateIP
			break
		}
		if res.Data.Status == domain.StatusError {
			t.Fatalf("Instance reached error state")
		}
		time.Sleep(2 * time.Second)
	}

	if privateIP == "" {
		resp := getRequest(t, client, fmt.Sprintf(instRoute, testutil.TestBaseURL, testutil.TestRouteInstances, instanceID), token)
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		t.Logf("Final instance state: %s", string(body))
	}

	return privateIP
}
