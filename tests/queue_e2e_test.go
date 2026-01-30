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

func TestQueueE2E(t *testing.T) {
	t.Parallel()
	if err := waitForServer(); err != nil {
		t.Fatalf("Failing Queue E2E test: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	token := registerAndLogin(t, client, "queue-tester@thecloud.local", "Queue Tester")

	var queueID string
	queueName := fmt.Sprintf("e2e-queue-%d", time.Now().UnixNano()%1000)

	// 1. Create Queue
	t.Run("CreateQueue", func(t *testing.T) {
		payload := map[string]string{"name": queueName}
		resp := postRequest(t, client, testutil.TestBaseURL+"/queues", token, payload)
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var res struct {
			Data domain.Queue `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		queueID = res.Data.ID.String()
		assert.NotEmpty(t, queueID)
	})

	// 2. Send Message
	t.Run("SendMessage", func(t *testing.T) {
		payload := map[string]string{"body": "hello e2e"}
		resp := postRequest(t, client, fmt.Sprintf("%s/queues/%s/messages", testutil.TestBaseURL, queueID), token, payload)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	// 3. Receive Messages
	var receiptHandle string
	t.Run("ReceiveMessages", func(t *testing.T) {
		resp := getRequest(t, client, fmt.Sprintf("%s/queues/%s/messages", testutil.TestBaseURL, queueID), token)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res struct {
			Data []domain.Message `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		assert.True(t, len(res.Data) > 0)
		assert.Equal(t, "hello e2e", res.Data[0].Body)
		receiptHandle = res.Data[0].ReceiptHandle
	})

	// 4. Delete Message
	t.Run("DeleteMessage", func(t *testing.T) {
		resp := deleteRequest(t, client, fmt.Sprintf("%s/queues/%s/messages/%s", testutil.TestBaseURL, queueID, receiptHandle), token)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	// 5. Delete Queue
	t.Run("DeleteQueue", func(t *testing.T) {
		resp := deleteRequest(t, client, fmt.Sprintf("%s/queues/%s", testutil.TestBaseURL, queueID), token)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})
}
