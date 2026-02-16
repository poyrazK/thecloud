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

func TestDatabaseReplicationE2E(t *testing.T) {
	// Skip if running in short mode as it requires a live server
	if testing.Short() {
		t.Skip("skipping database replication E2E test in short mode")
	}

	if err := waitForServer(); err != nil {
		t.Fatalf("Failing Database Replication E2E test: %v", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	token := registerAndLogin(t, client, "replication-tester@thecloud.local", "Replica Tester")

	var primaryID string
	var replicaID string
	dbName := fmt.Sprintf("e2e-primary-%d", time.Now().UnixNano()%1000)

	// 1. Create Primary Database
	t.Run("CreatePrimary", func(t *testing.T) {
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
		primaryID = res.Data.ID.String()
		assert.Equal(t, domain.RolePrimary, res.Data.Role)
	})

	// 2. Create Replica
	t.Run("CreateReplica", func(t *testing.T) {
		payload := map[string]string{
			"name": dbName + "-replica",
		}
		url := fmt.Sprintf("%s/databases/%s/replicas", testutil.TestBaseURL, primaryID)
		resp := postRequest(t, client, url, token, payload)
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var res struct {
			Data domain.Database `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		replicaID = res.Data.ID.String()
		assert.Equal(t, domain.RoleReplica, res.Data.Role)
		assert.Equal(t, primaryID, res.Data.PrimaryID.String())
	})

	// 3. List Replicas
	t.Run("ListReplicas", func(t *testing.T) {
		// We don't have a direct ListReplicas API endpoint, but we can verify role via GetDatabase
		resp := getRequest(t, client, fmt.Sprintf("%s/databases/%s", testutil.TestBaseURL, replicaID), token)
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		var res struct {
			Data domain.Database `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		assert.Equal(t, domain.RoleReplica, res.Data.Role)
	})

	// 4. Promote Replica to Primary
	t.Run("PromoteReplica", func(t *testing.T) {
		url := fmt.Sprintf("%s/databases/%s/promote", testutil.TestBaseURL, replicaID)
		resp := postRequest(t, client, url, token, nil)
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify role changed to Primary
		resp2 := getRequest(t, client, fmt.Sprintf("%s/databases/%s", testutil.TestBaseURL, replicaID), token)
		defer func() { _ = resp2.Body.Close() }()

		var res struct {
			Data domain.Database `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp2.Body).Decode(&res))
		assert.Equal(t, domain.RolePrimary, res.Data.Role)
		assert.Nil(t, res.Data.PrimaryID)
	})

	// Cleanup
	t.Run("Cleanup", func(t *testing.T) {
		resp1 := deleteRequest(t, client, fmt.Sprintf("%s/databases/%s", testutil.TestBaseURL, primaryID), token)
		if resp1 != nil {
			defer func() { _ = resp1.Body.Close() }()
		}
		resp2 := deleteRequest(t, client, fmt.Sprintf("%s/databases/%s", testutil.TestBaseURL, replicaID), token)
		if resp2 != nil {
			defer func() { _ = resp2.Body.Close() }()
		}
	})
}
