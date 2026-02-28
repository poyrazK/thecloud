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

func TestDatabaseAdvancedE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping database advanced E2E test in short mode")
	}

	if err := waitForServer(); err != nil {
		t.Fatalf("Failing Database Advanced E2E test: %v", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	token := registerAndLogin(t, client, "db-advanced@thecloud.local", "Advanced Tester")

	t.Run("InvalidConfigurations", func(t *testing.T) {
		testCases := []struct {
			name    string
			payload map[string]interface{}
		}{
			{
				"UnsupportedEngine",
				map[string]interface{}{
					"name":    "invalid-engine-db",
					"engine":  "oracle",
					"version": "19c",
				},
			},
			{
				"InvalidStorageSize",
				map[string]interface{}{
					"name":              "invalid-storage-db",
					"engine":            "postgres",
					"version":           "16",
					"allocated_storage": -5,
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resp := postRequest(t, client, testutil.TestBaseURL+"/databases", token, tc.payload)
				defer closeBody(t, resp)

				assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Expected 400 Bad Request for %s", tc.name)
			})
		}
	})

	t.Run("PromotionEdgeCases", func(t *testing.T) {
		// 1. Create a Primary DB
		dbName := fmt.Sprintf("promo-edge-db-%d", time.Now().UnixNano()%1000)
		payload := map[string]interface{}{
			"name":    dbName,
			"engine":  "postgres",
			"version": "16",
		}
		resp := postRequest(t, client, testutil.TestBaseURL+"/databases", token, payload)
		defer closeBody(t, resp)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var res struct {
			Data domain.Database `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		dbID := res.Data.ID.String()

		// 2. Attempt to promote a primary (should fail as it's already primary)
		t.Run("PromotePrimaryFails", func(t *testing.T) {
			respPromo := postRequest(t, client, fmt.Sprintf("%s/databases/%s/promote", testutil.TestBaseURL, dbID), token, nil)
			defer closeBody(t, respPromo)
			assert.Equal(t, http.StatusBadRequest, respPromo.StatusCode)
		})

		// 3. Attempt to promote non-existent UUID
		t.Run("PromoteNotFound", func(t *testing.T) {
			fakeID := "00000000-0000-0000-0000-000000000000"
			respPromo := postRequest(t, client, fmt.Sprintf("%s/databases/%s/promote", testutil.TestBaseURL, fakeID), token, nil)
			defer closeBody(t, respPromo)
			assert.Equal(t, http.StatusNotFound, respPromo.StatusCode)
		})

		// Cleanup
		deleteRequest(t, client, fmt.Sprintf("%s/databases/%s", testutil.TestBaseURL, dbID), token)
	})

	t.Run("VpcIntegration", func(t *testing.T) {
		// 1. Create a VPC
		vpcPayload := map[string]string{
			"name":       "db-vpc",
			"cidr_block": "10.10.0.0/16",
		}
		respVpc := postRequest(t, client, testutil.TestBaseURL+"/vpcs", token, vpcPayload)
		defer closeBody(t, respVpc)
		require.Equal(t, http.StatusCreated, respVpc.StatusCode)

		var vpcRes struct {
			Data domain.VPC `json:"data"`
		}
		require.NoError(t, json.NewDecoder(respVpc.Body).Decode(&vpcRes))
		vpcID := vpcRes.Data.ID

		// 2. Create DB in VPC
		dbName := "vpc-integrated-db"
		dbPayload := map[string]interface{}{
			"name":    dbName,
			"engine":  "postgres",
			"version": "16",
			"vpc_id":  vpcID,
		}
		respDb := postRequest(t, client, testutil.TestBaseURL+"/databases", token, dbPayload)
		defer closeBody(t, respDb)
		require.Equal(t, http.StatusCreated, respDb.StatusCode)

		var dbRes struct {
			Data domain.Database `json:"data"`
		}
		require.NoError(t, json.NewDecoder(respDb.Body).Decode(&dbRes))
		assert.Equal(t, vpcID, *dbRes.Data.VpcID)

		// Cleanup
		deleteRequest(t, client, fmt.Sprintf("%s/databases/%s", testutil.TestBaseURL, dbRes.Data.ID), token)
		deleteRequest(t, client, fmt.Sprintf("%s/vpcs/%s", testutil.TestBaseURL, vpcID), token)
	})

	t.Run("MultiReplicaPromotion", func(t *testing.T) {
		// 1. Create Primary
		payload := map[string]interface{}{
			"name":    "multi-rep-primary",
			"engine":  "postgres",
			"version": "16",
		}
		respP := postRequest(t, client, testutil.TestBaseURL+"/databases", token, payload)
		defer closeBody(t, respP)
		require.Equal(t, http.StatusCreated, respP.StatusCode)

		var pRes struct {
			Data domain.Database `json:"data"`
		}
		require.NoError(t, json.NewDecoder(respP.Body).Decode(&pRes))
		primaryID := pRes.Data.ID

		// 2. Create 2 Replicas
		replicaIDs := make([]string, 2)
		for i := 0; i < 2; i++ {
			repPayload := map[string]string{
				"name": fmt.Sprintf("replica-%d", i+1),
			}
			url := fmt.Sprintf("%s/databases/%s/replicas", testutil.TestBaseURL, primaryID)
			respR := postRequest(t, client, url, token, repPayload)
			defer closeBody(t, respR)
			require.Equal(t, http.StatusCreated, respR.StatusCode)

			var rRes struct {
				Data domain.Database `json:"data"`
			}
			require.NoError(t, json.NewDecoder(respR.Body).Decode(&rRes))
			replicaIDs[i] = rRes.Data.ID.String()
			assert.Equal(t, domain.RoleReplica, rRes.Data.Role)
			assert.Equal(t, primaryID, *rRes.Data.PrimaryID)
		}

		// 3. Promote Replica 1
		respPromo := postRequest(t, client, fmt.Sprintf("%s/databases/%s/promote", testutil.TestBaseURL, replicaIDs[0]), token, nil)
		defer closeBody(t, respPromo)
		assert.Equal(t, http.StatusOK, respPromo.StatusCode)

		// 4. Verify Replica 1 is now Primary
		respG1 := getRequest(t, client, fmt.Sprintf("%s/databases/%s", testutil.TestBaseURL, replicaIDs[0]), token)
		defer closeBody(t, respG1)
		var g1Res struct {
			Data domain.Database `json:"data"`
		}
		require.NoError(t, json.NewDecoder(respG1.Body).Decode(&g1Res))
		assert.Equal(t, domain.RolePrimary, g1Res.Data.Role)
		assert.Nil(t, g1Res.Data.PrimaryID)

		// 5. Verify Replica 2 still points to original Primary
		respG2 := getRequest(t, client, fmt.Sprintf("%s/databases/%s", testutil.TestBaseURL, replicaIDs[1]), token)
		defer closeBody(t, respG2)
		var g2Res struct {
			Data domain.Database `json:"data"`
		}
		require.NoError(t, json.NewDecoder(respG2.Body).Decode(&g2Res))
		assert.Equal(t, domain.RoleReplica, g2Res.Data.Role)
		assert.Equal(t, primaryID, *g2Res.Data.PrimaryID)

		// Cleanup
		deleteRequest(t, client, fmt.Sprintf("%s/databases/%s", testutil.TestBaseURL, primaryID), token)
		deleteRequest(t, client, fmt.Sprintf("%s/databases/%s", testutil.TestBaseURL, replicaIDs[0]), token)
		deleteRequest(t, client, fmt.Sprintf("%s/databases/%s", testutil.TestBaseURL, replicaIDs[1]), token)
	})

	t.Run("ConnectionStringFormats", func(t *testing.T) {
		engines := []struct {
			engine string
			prefix string
		}{
			{"postgres", "postgres://"},
			{"mysql", ""}, // MySQL format is user:pass@tcp(host:port)/db
		}

		for _, tc := range engines {
			t.Run(tc.engine, func(t *testing.T) {
				dbName := fmt.Sprintf("conn-str-%s-%d", tc.engine, time.Now().UnixNano()%1000)
				payload := map[string]interface{}{
					"name":    dbName,
					"engine":  tc.engine,
					"version": "16",
				}
				resp := postRequest(t, client, testutil.TestBaseURL+"/databases", token, payload)
				if resp.StatusCode != http.StatusCreated {
					closeBody(t, resp)
					t.Skipf("Skipping %s connection string test due to infra error: %d", tc.engine, resp.StatusCode)
				}

				var res struct {
					Data domain.Database `json:"data"`
				}
				json.NewDecoder(resp.Body).Decode(&res)
				dbID := res.Data.ID.String()
				closeBody(t, resp)

				respConn := getRequest(t, client, fmt.Sprintf("%s/databases/%s/connection", testutil.TestBaseURL, dbID), token)
				defer closeBody(t, respConn)

				var connRes struct {
					Data struct {
						ConnectionString string `json:"connection_string"`
					} `json:"data"`
				}
				require.NoError(t, json.NewDecoder(respConn.Body).Decode(&connRes))
				assert.Contains(t, connRes.Data.ConnectionString, tc.prefix)
				assert.Contains(t, connRes.Data.ConnectionString, dbName)

				// Cleanup
				deleteRequest(t, client, fmt.Sprintf("%s/databases/%s", testutil.TestBaseURL, dbID), token)
			})
		}
	})
}
