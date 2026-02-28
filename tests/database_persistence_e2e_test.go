package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/pkg/testutil"
)

const (
	dbSuffixMod  = 1000
	dbPrefixLen  = 8
	pollInterval = 200 * time.Millisecond
	pollTimeout  = 5 * time.Second
)

func safePrefix(id string, n int) string {
	if len(id) < n {
		return id
	}
	return id[:n]
}

func closeBody(t *testing.T, resp *http.Response) {
	t.Helper()
	if resp != nil && resp.Body != nil {
		err := resp.Body.Close()
		require.NoError(t, err, "failed to close response body")
	}
}

func TestDatabasePersistenceE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping database persistence E2E test in short mode")
	}

	if err := waitForServer(); err != nil {
		t.Fatalf("Failing Database Persistence E2E test: %v", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	token := registerAndLogin(t, client, "db-persistence@thecloud.local", "Persistence Tester")

	testCases := []struct {
		engine  string
		version string
	}{
		{"postgres", "16"},
		{"mysql", "8.0"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Engine/%s", tc.engine), func(t *testing.T) {
			var dbID string
			dbName := fmt.Sprintf("e2e-persistent-%s-%d", tc.engine, time.Now().UnixNano()%dbSuffixMod)

			// 1. Create Database
			t.Run("CreateDatabase_ProvisionsVolume", func(t *testing.T) {
				payload := map[string]interface{}{
					"name":              dbName,
					"engine":            tc.engine,
					"version":           tc.version,
					"allocated_storage": 20,
				}
				resp := postRequest(t, client, testutil.TestBaseURL+"/databases", token, payload)
				defer closeBody(t, resp)

				require.Equal(t, http.StatusCreated, resp.StatusCode)

				var res struct {
					Data domain.Database `json:"data"`
				}
				require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
				dbID = res.Data.ID.String()
				assert.NotEmpty(t, dbID)
				assert.Equal(t, 20, res.Data.AllocatedStorage)

				// 2. Verify Volume exists
				respVols := getRequest(t, client, testutil.TestBaseURL+"/volumes", token)
				defer closeBody(t, respVols)
				require.Equal(t, http.StatusOK, respVols.StatusCode)

				var volsRes struct {
					Data []domain.Volume `json:"data"`
				}
				require.NoError(t, json.NewDecoder(respVols.Body).Decode(&volsRes))

				found := false
				expectedPrefix := fmt.Sprintf("db-vol-%s", safePrefix(dbID, dbPrefixLen))
				for _, v := range volsRes.Data {
					if strings.HasPrefix(v.Name, expectedPrefix) {
						found = true
						assert.Equal(t, 20, v.SizeGB)
						break
					}
				}
				assert.True(t, found, "Expected volume with prefix %s not found", expectedPrefix)
			})

			// 3. Create Replica and verify its Volume
			t.Run("CreateReplica_ProvisionsVolume", func(t *testing.T) {
				if dbID == "" {
					t.Skip("skipping replica test as primary dbID is empty")
				}

				replicaName := fmt.Sprintf("e2e-persistent-rep-%s-%d", tc.engine, time.Now().UnixNano()%dbSuffixMod)
				payload := map[string]string{
					"name": replicaName,
				}
				resp := postRequest(t, client, fmt.Sprintf("%s/databases/%s/replicas", testutil.TestBaseURL, dbID), token, payload)
				defer closeBody(t, resp)

				require.Equal(t, http.StatusCreated, resp.StatusCode)

				var res struct {
					Data domain.Database `json:"data"`
				}
				require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
				replicaID := res.Data.ID.String()
				assert.NotEmpty(t, replicaID)

				// Verify Replica Volume exists
				respVols := getRequest(t, client, testutil.TestBaseURL+"/volumes", token)
				defer closeBody(t, respVols)
				require.Equal(t, http.StatusOK, respVols.StatusCode)

				var volsRes struct {
					Data []domain.Volume `json:"data"`
				}
				require.NoError(t, json.NewDecoder(respVols.Body).Decode(&volsRes))

				found := false
				expectedPrefix := fmt.Sprintf("db-replica-vol-%s", safePrefix(replicaID, dbPrefixLen))
				for _, v := range volsRes.Data {
					if strings.HasPrefix(v.Name, expectedPrefix) {
						found = true
						assert.Equal(t, 20, v.SizeGB) // Should inherit 20GB from primary in previous subtest
						break
					}
				}
				assert.True(t, found, "Expected replica volume with prefix %s not found", expectedPrefix)

				// Cleanup replica
				respDel := deleteRequest(t, client, fmt.Sprintf("%s/databases/%s", testutil.TestBaseURL, replicaID), token)
				defer closeBody(t, respDel)
				assert.Equal(t, http.StatusOK, respDel.StatusCode)
			})

			// 4. Delete Database and verify Volume cleanup
			t.Run("DeleteDatabase_CleansUpVolume", func(t *testing.T) {
				if dbID == "" {
					t.Skip("skipping delete test as dbID is empty")
				}

				resp := deleteRequest(t, client, fmt.Sprintf("%s/databases/%s", testutil.TestBaseURL, dbID), token)
				defer closeBody(t, resp)
				require.Equal(t, http.StatusOK, resp.StatusCode)

				// Polling verify Volume is gone
				require.Eventually(t, func() bool {
					respVols := getRequest(t, client, testutil.TestBaseURL+"/volumes", token)
					defer closeBody(t, respVols)
					if respVols.StatusCode != http.StatusOK {
						return false
					}

					var volsRes struct {
						Data []domain.Volume `json:"data"`
					}
					if err := json.NewDecoder(respVols.Body).Decode(&volsRes); err != nil {
						return false
					}

					expectedPrefix := fmt.Sprintf("db-vol-%s", safePrefix(dbID, dbPrefixLen))
					for _, v := range volsRes.Data {
						if strings.HasPrefix(v.Name, expectedPrefix) {
							return false
						}
					}
					return true
				}, pollTimeout, pollInterval, "Volume with prefix db-vol-%s should have been deleted", safePrefix(dbID, dbPrefixLen))
			})
		})
	}
}
