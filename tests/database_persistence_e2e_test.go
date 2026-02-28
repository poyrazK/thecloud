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

func TestDatabasePersistenceE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping database persistence E2E test in short mode")
	}

	if err := waitForServer(); err != nil {
		t.Fatalf("Failing Database Persistence E2E test: %v", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	token := registerAndLogin(t, client, "db-persistence@thecloud.local", "Persistence Tester")

	var dbID string
	dbName := fmt.Sprintf("e2e-persistent-db-%d", time.Now().UnixNano()%1000)

	// 1. Create Database
	t.Run("CreateDatabase_ProvisionsVolume", func(t *testing.T) {
		payload := map[string]string{
			"name":    dbName,
			"engine":  "postgres",
			"version": "16",
		}
		resp := postRequest(t, client, testutil.TestBaseURL+"/databases", token, payload)
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var res struct {
			Data domain.Database `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		dbID = res.Data.ID.String()
		assert.NotEmpty(t, dbID)

		// 2. Verify Volume exists
		respVols := getRequest(t, client, testutil.TestBaseURL+"/volumes", token)
		defer func() { _ = respVols.Body.Close() }()
		require.Equal(t, http.StatusOK, respVols.StatusCode)

		var volsRes struct {
			Data []domain.Volume `json:"data"`
		}
		require.NoError(t, json.NewDecoder(respVols.Body).Decode(&volsRes))

		found := false
		expectedPrefix := fmt.Sprintf("db-vol-%s", dbID[:8])
		for _, v := range volsRes.Data {
			if strings.HasPrefix(v.Name, expectedPrefix) {
				found = true
				break
			}
		}
		assert.True(t, found, fmt.Sprintf("Expected volume with prefix %s not found", expectedPrefix))
	})

	// 3. Delete Database and verify Volume cleanup
	t.Run("DeleteDatabase_CleansUpVolume", func(t *testing.T) {
		resp := deleteRequest(t, client, fmt.Sprintf("%s/databases/%s", testutil.TestBaseURL, dbID), token)
		defer func() { _ = resp.Body.Close() }()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		// Wait a bit for async cleanup if any
		time.Sleep(1 * time.Second)

		// Verify Volume is gone
		respVols := getRequest(t, client, testutil.TestBaseURL+"/volumes", token)
		defer func() { _ = respVols.Body.Close() }()
		require.Equal(t, http.StatusOK, respVols.StatusCode)

		var volsRes struct {
			Data []domain.Volume `json:"data"`
		}
		require.NoError(t, json.NewDecoder(respVols.Body).Decode(&volsRes))

		found := false
		expectedPrefix := fmt.Sprintf("db-vol-%s", dbID[:8])
		for _, v := range volsRes.Data {
			if strings.HasPrefix(v.Name, expectedPrefix) {
				found = true
				break
			}
		}
		assert.False(t, found, fmt.Sprintf("Volume with prefix %s should have been deleted", expectedPrefix))
	})
}
