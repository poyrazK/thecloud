package sdk_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/poyrazk/thecloud/pkg/sdk"
	"github.com/stretchr/testify/assert"
)

const (
	queueTestAPIKey      = "test-api-key"
	queueTestName        = "my-queue"
	queueTestOptionsName = "my-queue-options"
	queueTestID          = "q-1"
	queueTestMessageID   = "msg-1"
	queueTestHandleID    = "handle-1"
	queueTestMessageBody = "hello"
	queueTestBasePath    = "/api/v1/queues"
	queueTestContentType = "Content-Type"
	queueTestAppJSON     = "application/json"
)

func newQueueTestServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()
		w.Header().Set(queueTestContentType, queueTestAppJSON)

		if handleQueueBase(w, r) {
			return
		}
		if handleQueueItem(w, r) {
			return
		}
		if handleQueueMessages(w, r) {
			return
		}
		if handleQueuePurge(w, r) {
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func handleQueueBase(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Path != queueTestBasePath {
		return false
	}
	if r.Method == http.MethodPost {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["name"] == queueTestName || body["name"] == queueTestOptionsName {
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"id":   queueTestID,
					"name": body["name"],
				},
			})
			return true
		}
		return false
	}
	if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{{"id": queueTestID, "name": queueTestName}},
		})
		return true
	}
	return false
}

func handleQueueItem(w http.ResponseWriter, r *http.Request) bool {
	path := queueTestBasePath + "/" + queueTestID
	if r.URL.Path != path {
		return false
	}
	if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":   queueTestID,
				"name": queueTestName,
			},
		})
		return true
	}
	if r.Method == http.MethodDelete {
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	return false
}

func handleQueueMessages(w http.ResponseWriter, r *http.Request) bool {
	messagesPath := queueTestBasePath + "/" + queueTestID + "/messages"
	if r.URL.Path == messagesPath && r.Method == http.MethodPost {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["body"] == queueTestMessageBody {
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"id":   queueTestMessageID,
					"body": queueTestMessageBody,
				},
			})
			return true
		}
		return false
	}
	if r.URL.Path == messagesPath && r.Method == http.MethodGet {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{{"id": queueTestMessageID, "body": queueTestMessageBody}},
		})
		return true
	}
	if r.URL.Path == messagesPath+"/"+queueTestHandleID && r.Method == http.MethodDelete {
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	return false
}

func handleQueuePurge(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Path == queueTestBasePath+"/"+queueTestID+"/purge" && r.Method == http.MethodPost {
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	return false
}

func TestQueueSDK(t *testing.T) {
	ts := newQueueTestServer(t)
	defer ts.Close()

	client := sdk.NewClient(ts.URL+"/api/v1", queueTestAPIKey)

	t.Run("CreateQueue", func(t *testing.T) {
		q, err := client.CreateQueue(queueTestName, nil, nil, nil)
		assert.NoError(t, err)
		if q != nil {
			assert.Equal(t, queueTestName, q.Name)
		}
	})

	t.Run("CreateQueueWithOptions", func(t *testing.T) {
		vt := 30
		rd := 7
		mms := 1024
		q, err := client.CreateQueue(queueTestOptionsName, &vt, &rd, &mms)
		assert.NoError(t, err)
		if q != nil {
			assert.Equal(t, queueTestOptionsName, q.Name)
		}
	})

	t.Run("ListQueues", func(t *testing.T) {
		qs, err := client.ListQueues()
		assert.NoError(t, err)
		assert.Len(t, qs, 1)
	})

	t.Run("GetQueue", func(t *testing.T) {
		q, err := client.GetQueue(queueTestID)
		assert.NoError(t, err)
		if q != nil {
			assert.Equal(t, queueTestID, q.ID)
		}
	})

	t.Run("SendMessage", func(t *testing.T) {
		msg, err := client.SendMessage(queueTestID, queueTestMessageBody)
		assert.NoError(t, err)
		if msg != nil {
			assert.Equal(t, queueTestMessageBody, msg.Body)
		}
	})

	t.Run("ReceiveMessages", func(t *testing.T) {
		msgs, err := client.ReceiveMessages(queueTestID, 10)
		assert.NoError(t, err)
		assert.Len(t, msgs, 1)
	})

	t.Run("DeleteMessage", func(t *testing.T) {
		err := client.DeleteMessage(queueTestID, queueTestHandleID)
		assert.NoError(t, err)
	})

	t.Run("PurgeQueue", func(t *testing.T) {
		err := client.PurgeQueue(queueTestID)
		assert.NoError(t, err)
	})

	t.Run("DeleteQueue", func(t *testing.T) {
		err := client.DeleteQueue(queueTestID)
		assert.NoError(t, err)
	})
}

func TestQueueSDKErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	defer server.Close()

	client := sdk.NewClient(server.URL+"/api/v1", queueTestAPIKey)
	_, err := client.CreateQueue("q", nil, nil, nil)
	assert.Error(t, err)

	_, err = client.ListQueues()
	assert.Error(t, err)

	_, err = client.GetQueue(queueTestID)
	assert.Error(t, err)

	err = client.DeleteQueue(queueTestID)
	assert.Error(t, err)

	_, err = client.SendMessage(queueTestID, queueTestMessageBody)
	assert.Error(t, err)

	_, err = client.ReceiveMessages(queueTestID, 10)
	assert.Error(t, err)

	err = client.DeleteMessage(queueTestID, queueTestHandleID)
	assert.Error(t, err)

	err = client.PurgeQueue(queueTestID)
	assert.Error(t, err)
}
