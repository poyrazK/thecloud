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

func TestDatabaseE2E(t *testing.T) {
	if err := waitForServer(); err != nil {
		t.Skipf("Skipping Database E2E test: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	token := registerAndLogin(t, client, "db-tester@thecloud.local", "DB Tester")

	var dbID string
	dbName := fmt.Sprintf("e2e-db-%d", time.Now().UnixNano()%1000)

	// 1. Create Database
	t.Run("CreateDatabase", func(t *testing.T) {
		payload := map[string]string{
			"name":    dbName,
			"engine":  "postgres",
			"version": "16",
		}
		resp := postRequest(t, client, testutil.TestBaseURL+"/databases", token, payload)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var res struct {
			Data domain.Database `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		dbID = res.Data.ID.String()
		assert.NotEmpty(t, dbID)
	})

	// 2. Get Connection String
	t.Run("GetConnectionString", func(t *testing.T) {
		resp := getRequest(t, client, fmt.Sprintf("%s/databases/%s/connection", testutil.TestBaseURL, dbID), token)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// 3. Delete Database
	t.Run("DeleteDatabase", func(t *testing.T) {
		resp := deleteRequest(t, client, fmt.Sprintf("%s/databases/%s", testutil.TestBaseURL, dbID), token)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
