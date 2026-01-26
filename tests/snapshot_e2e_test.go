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

func TestSnapshotE2E(t *testing.T) {
	if err := waitForServer(); err != nil {
		t.Skipf("Skipping Snapshot E2E test: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	token := registerAndLogin(t, client, "snapshot-tester@thecloud.local", "Snapshot Tester")

	var volumeID string
	var snapshotID string
	volName := fmt.Sprintf("e2e-vol-%s", uuid.New().String())

	// 1. Create Volume
	t.Run("CreateVolume", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":    volName,
			"size_gb": 10,
		}
		resp := postRequest(t, client, testutil.TestBaseURL+"/volumes", token, payload)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var res struct {
			Data domain.Volume `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		volumeID = res.Data.ID.String()
		assert.NotEmpty(t, volumeID)
	})

	// 2. Create Snapshot
	t.Run("CreateSnapshot", func(t *testing.T) {
		payload := map[string]string{
			"volume_id":   volumeID,
			"description": "E2E snapshot",
		}
		resp := postRequest(t, client, testutil.TestBaseURL+"/snapshots", token, payload)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var res struct {
			Data domain.Snapshot `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		snapshotID = res.Data.ID.String()
		assert.NotEmpty(t, snapshotID)
	})

	// 3. Wait for Snapshot to be AVAILABLE
	t.Run("WaitAvailable", func(t *testing.T) {
		timeout := time.After(30 * time.Second)
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-timeout:
				t.Fatal("Timeout waiting for snapshot to be available")
			case <-ticker.C:
				resp := getRequest(t, client, fmt.Sprintf("%s/snapshots/%s", testutil.TestBaseURL, snapshotID), token)
				var res struct {
					Data domain.Snapshot `json:"data"`
				}
				_ = json.NewDecoder(resp.Body).Decode(&res)
				resp.Body.Close()

				if res.Data.Status == domain.SnapshotStatusAvailable {
					return
				}
				if res.Data.Status == domain.SnapshotStatusError {
					t.Fatal("Snapshot creation failed with error status")
				}
			}
		}
	})

	// 4. Restore Snapshot
	t.Run("RestoreSnapshot", func(t *testing.T) {
		payload := map[string]string{
			"new_volume_name": fmt.Sprintf("restored-vol-%s", uuid.New().String()),
		}
		resp := postRequest(t, client, fmt.Sprintf("%s/snapshots/%s/restore", testutil.TestBaseURL, snapshotID), token, payload)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			var body map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&body)
			t.Logf("RestoreSnapshot failed with %d: %v", resp.StatusCode, body)
		}
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	// 4. Cleanup
	t.Run("Cleanup", func(t *testing.T) {
		resp := deleteRequest(t, client, fmt.Sprintf("%s/snapshots/%s", testutil.TestBaseURL, snapshotID), token)
		resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		resp = deleteRequest(t, client, fmt.Sprintf("%s/volumes/%s", testutil.TestBaseURL, volumeID), token)
		resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
