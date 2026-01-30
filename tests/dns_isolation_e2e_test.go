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

func TestDNSIsolationE2E(t *testing.T) {
	t.Parallel()
	if err := waitForServer(); err != nil {
		t.Fatalf("Failing DNS Isolation E2E test: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	token := registerAndLogin(t, client, "dns-iso-tester@thecloud.local", "DNS Isolation Tester")

	// Create VPC A
	vpcAID := createVPCForDNS(t, client, token, "vpc-a")
	zoneAID := createZoneForDNS(t, client, token, "a.internal", vpcAID)

	// Create VPC B
	vpcBID := createVPCForDNS(t, client, token, "vpc-b")
	zoneBID := createZoneForDNS(t, client, token, "b.internal", vpcBID)

	// 1. Create record in Zone A
	t.Run("CreateRecordInZoneA", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":    "web",
			"type":    "A",
			"content": "10.10.1.1",
		}
		resp := postRequest(t, client, fmt.Sprintf("%s/dns/zones/%s/records", testutil.TestBaseURL, zoneAID), token, payload)
		defer func() { _ = resp.Body.Close() }()
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	// 2. Verify Zone B is empty/doesn't have the record
	t.Run("VerifyIsolationInZoneB", func(t *testing.T) {
		resp := getRequest(t, client, fmt.Sprintf("%s/dns/zones/%s/records", testutil.TestBaseURL, zoneBID), token)
		defer func() { _ = resp.Body.Close() }()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var res struct {
			Data []domain.DNSRecord `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))

		for _, rec := range res.Data {
			assert.NotEqual(t, "web", rec.Name, "Zone B should not contain records from Zone A")
		}
	})

	// Cleanup
	t.Cleanup(func() {
		_ = deleteRequest(t, client, fmt.Sprintf("%s/dns/zones/%s", testutil.TestBaseURL, zoneAID), token).Body.Close()
		_ = deleteRequest(t, client, fmt.Sprintf("%s/dns/zones/%s", testutil.TestBaseURL, zoneBID), token).Body.Close()
		_ = deleteRequest(t, client, fmt.Sprintf("%s/vpcs/%s", testutil.TestBaseURL, vpcAID), token).Body.Close()
		_ = deleteRequest(t, client, fmt.Sprintf("%s/vpcs/%s", testutil.TestBaseURL, vpcBID), token).Body.Close()
	})
}

func createVPCForDNS(t *testing.T, client *http.Client, token, name string) string {
	payload := map[string]string{
		"name":       fmt.Sprintf("%s-%d", name, time.Now().UnixNano()),
		"cidr_block": "10.0.0.0/16",
	}
	resp := postRequest(t, client, testutil.TestBaseURL+testutil.TestRouteVpcs, token, payload)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var res struct {
		Data domain.VPC `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
	return res.Data.ID.String()
}

func createZoneForDNS(t *testing.T, client *http.Client, token, name, vpcID string) string {
	payload := map[string]interface{}{
		"name":   name,
		"vpc_id": vpcID,
	}
	resp := postRequest(t, client, testutil.TestBaseURL+testutil.TestRouteDNSZones, token, payload)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var res struct {
		Data domain.DNSZone `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
	return res.Data.ID.String()
}
