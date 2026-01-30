package tests

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/poyrazk/thecloud/tests/helpers"
	"github.com/stretchr/testify/assert"
)

func TestConcurrency(t *testing.T) {
	t.Parallel()
	if err := waitForServer(); err != nil {
		t.Fatalf("Failing Concurrency test: %v", err)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	token := registerAndLogin(t, client, "concurrent-tester@thecloud.local", "Concurrent Tester")

	t.Run("Duplicate Resources Simultaneously", func(t *testing.T) {
		vpcName := fmt.Sprintf("concurrent-vpc-%d", time.Now().UnixNano())
		payload := map[string]string{
			"name":       vpcName,
			"cidr_block": "10.10.0.0/16",
		}

		results := helpers.RunConcurrently(5, func(i int) error {
			resp := postRequest(t, client, testutil.TestBaseURL+"/vpcs", token, payload)
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusAccepted {
				return nil
			}
			return fmt.Errorf("Status %d", resp.StatusCode)
		})

		// One should succeed, others should fail (ideally with 409 Conflict)
		successCount := 0
		for _, err := range results {
			if err == nil {
				successCount++
			}
		}

		assert.Equal(t, 1, successCount, "Exactly one VPC creation should succeed")
	})

	t.Run("Concurrent Reads", func(t *testing.T) {
		// Read instances list concurrently
		results := helpers.RunConcurrently(10, func(i int) error {
			resp := getRequest(t, client, testutil.TestBaseURL+"/instances", token)
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode == http.StatusOK {
				return nil
			}
			return fmt.Errorf("Status %d", resp.StatusCode)
		})

		for i, err := range results {
			assert.NoError(t, err, "Request %d failed", i)
		}
	})
}
