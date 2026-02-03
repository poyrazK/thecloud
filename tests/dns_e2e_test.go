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

func TestDNSE2E(t *testing.T) {
	t.Parallel()
	if err := waitForServer(); err != nil {
		t.Fatalf("Failing DNS E2E test: %v", err)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	token := registerAndLogin(t, client, "dns-tester@thecloud.local", "DNS Tester")

	const recordsRoute = "%s/dns/zones/%s/records"
	const zoneRoute = "%s%s/%s"

	var vpcID string
	vpcName := fmt.Sprintf("dns-e2e-vpc-%d", time.Now().UnixNano())

	// 1. Create VPC
	t.Run("CreateVPC", func(t *testing.T) {
		payload := map[string]string{
			"name":       vpcName,
			"cidr_block": "10.10.0.0/16",
		}
		resp := postRequest(t, client, testutil.TestBaseURL+testutil.TestRouteVpcs, token, payload)
		defer func() { _ = resp.Body.Close() }()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var res struct {
			Data domain.VPC `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		vpcID = res.Data.ID.String()
	})

	if vpcID == "" {
		t.Fatal("VPC creation failed, skipping dependent subtests")
	}

	var zoneID string
	zoneName := fmt.Sprintf("e2e-%d.internal", time.Now().UnixNano())

	// 2. Create DNS Zone
	t.Run("CreateZone", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":        zoneName,
			"description": "Comprehensive E2E Test Zone",
			"vpc_id":      vpcID,
		}
		resp := postRequest(t, client, testutil.TestBaseURL+testutil.TestRouteDNSZones, token, payload)
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("DNS Zone creation failed: status %d, body: %s", resp.StatusCode, string(body))
		}

		var res struct {
			Data domain.DNSZone `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		zoneID = res.Data.ID.String()
		assert.Equal(t, zoneName, res.Data.Name)
	})

	if zoneID == "" {
		t.Fatal("DNS Zone creation failed, skipping dependent subtests")
	}

	// 3. Duplicate Zone Rejection
	t.Run("DuplicateZoneRejected", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":   zoneName,
			"vpc_id": vpcID,
		}
		resp := postRequest(t, client, testutil.TestBaseURL+testutil.TestRouteDNSZones, token, payload)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	// 4. Multiple Record Types
	t.Run("MultipleRecordTypes", func(t *testing.T) {
		recordTypes := []struct {
			Name     string
			Type     domain.RecordType
			Content  string
			Priority *int
		}{
			{"ipv4", domain.RecordTypeA, "10.10.1.1", nil},
			{"ipv6", domain.RecordTypeAAAA, "2001:db8::1", nil},
			{"alias", domain.RecordTypeCNAME, "web.example.com.", nil},
			{"mail", domain.RecordTypeMX, "mail.example.com.", intPtr(10)},
			{"txt", domain.RecordTypeTXT, "\"v=spf1 include:_spf.google.com ~all\"", nil},
		}

		for _, rt := range recordTypes {
			t.Run(string(rt.Type), func(t *testing.T) {
				payload := map[string]interface{}{
					"name":     rt.Name,
					"type":     rt.Type,
					"content":  rt.Content,
					"priority": rt.Priority,
					"ttl":      300,
				}
				resp := postRequest(t, client, fmt.Sprintf(recordsRoute, testutil.TestBaseURL, zoneID), token, payload)
				defer func() { _ = resp.Body.Close() }()

				if resp.StatusCode != http.StatusCreated {
					body, _ := io.ReadAll(resp.Body)
					t.Fatalf("CreateRecord %s failed: status %d, body: %s", rt.Type, resp.StatusCode, string(body))
				}

				var res struct {
					Data domain.DNSRecord `json:"data"`
				}
				require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
				assert.Equal(t, rt.Name, res.Data.Name)
				assert.Equal(t, rt.Content, res.Data.Content)
			})
		}
	})

	// 5. Boundary Testing
	t.Run("BoundaryTesting", func(t *testing.T) {
		t.Run("MinTTLClamping", func(t *testing.T) {
			payload := map[string]interface{}{
				"name":    "lowttl",
				"type":    "A",
				"content": "10.10.1.5",
				"ttl":     10, // Should be clamped to min (e.g. 60)
			}
			resp := postRequest(t, client, fmt.Sprintf(recordsRoute, testutil.TestBaseURL, zoneID), token, payload)
			defer func() { _ = resp.Body.Close() }()
			require.Equal(t, http.StatusCreated, resp.StatusCode)

			var res struct {
				Data domain.DNSRecord `json:"data"`
			}
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
			assert.GreaterOrEqual(t, res.Data.TTL, 60)
		})

		t.Run("InvalidRecordType", func(t *testing.T) {
			payload := map[string]interface{}{
				"name":    "invalid",
				"type":    "UNKNOWN",
				"content": "some content",
			}
			resp := postRequest(t, client, fmt.Sprintf("%s/dns/zones/%s/records", testutil.TestBaseURL, zoneID), token, payload)
			defer func() { _ = resp.Body.Close() }()
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	})

	// 6. Cleanup
	t.Run("Cleanup", func(t *testing.T) {
		// Delete DNS Zone
		resp := deleteRequest(t, client, fmt.Sprintf(zoneRoute, testutil.TestBaseURL, testutil.TestRouteDNSZones, zoneID), token)
		_ = resp.Body.Close()
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		// Verify deletion in PowerDNS (via GET)
		getResp := getRequest(t, client, fmt.Sprintf(zoneRoute, testutil.TestBaseURL, testutil.TestRouteDNSZones, zoneID), token)
		_ = getResp.Body.Close()
		assert.Equal(t, http.StatusNotFound, getResp.StatusCode)

		// Delete VPC
		resp = deleteRequest(t, client, fmt.Sprintf("%s%s/%s", testutil.TestBaseURL, testutil.TestRouteVpcs, vpcID), token)
		_ = resp.Body.Close()
		assert.Contains(t, []int{http.StatusOK, http.StatusNoContent}, resp.StatusCode)
	})
}

func intPtr(i int) *int {
	return &i
}
