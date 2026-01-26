package sdk_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/poyrazk/thecloud/pkg/sdk"
	"github.com/stretchr/testify/assert"
)

func TestQueueSDK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/api/v1/queues" && r.Method == "POST" {
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["name"] == "my-queue" || body["name"] == "my-queue-options" {
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{
						"id":   "q-1",
						"name": body["name"],
					},
				})
				return
			}
		}

		if r.URL.Path == "/api/v1/queues" && r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []map[string]interface{}{
					{"id": "q-1", "name": "my-queue"},
				},
			})
			return
		}

		if r.URL.Path == "/api/v1/queues/q-1" && r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"id":   "q-1",
					"name": "my-queue",
				},
			})
			return
		}

		if r.URL.Path == "/api/v1/queues/q-1" && r.Method == "DELETE" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.URL.Path == "/api/v1/queues/q-1/messages" && r.Method == "POST" {
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["body"] == "hello" {
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{
						"id":   "msg-1",
						"body": "hello",
					},
				})
				return
			}
		}

		if r.URL.Path == "/api/v1/queues/q-1/messages" && r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []map[string]interface{}{
					{"id": "msg-1", "body": "hello"},
				},
			})
			return
		}

		if r.URL.Path == "/api/v1/queues/q-1/messages/handle-1" && r.Method == "DELETE" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.URL.Path == "/api/v1/queues/q-1/purge" && r.Method == "POST" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	client := sdk.NewClient(ts.URL+"/api/v1", "test-api-key")

	t.Run("CreateQueue", func(t *testing.T) {
		q, err := client.CreateQueue("my-queue", nil, nil, nil)
		assert.NoError(t, err)
		if q != nil {
			assert.Equal(t, "my-queue", q.Name)
		}
	})

	t.Run("CreateQueueWithOptions", func(t *testing.T) {
		vt := 30
		rd := 7
		mms := 1024
		q, err := client.CreateQueue("my-queue-options", &vt, &rd, &mms)
		assert.NoError(t, err)
		if q != nil {
			assert.Equal(t, "my-queue-options", q.Name)
		}
	})

	t.Run("ListQueues", func(t *testing.T) {
		qs, err := client.ListQueues()
		assert.NoError(t, err)
		assert.Len(t, qs, 1)
	})

	t.Run("GetQueue", func(t *testing.T) {
		q, err := client.GetQueue("q-1")
		assert.NoError(t, err)
		if q != nil {
			assert.Equal(t, "q-1", q.ID)
		}
	})

	t.Run("SendMessage", func(t *testing.T) {
		msg, err := client.SendMessage("q-1", "hello")
		assert.NoError(t, err)
		if msg != nil {
			assert.Equal(t, "hello", msg.Body)
		}
	})

	t.Run("ReceiveMessages", func(t *testing.T) {
		msgs, err := client.ReceiveMessages("q-1", 10)
		assert.NoError(t, err)
		assert.Len(t, msgs, 1)
	})

	t.Run("DeleteMessage", func(t *testing.T) {
		err := client.DeleteMessage("q-1", "handle-1")
		assert.NoError(t, err)
	})

	t.Run("PurgeQueue", func(t *testing.T) {
		err := client.PurgeQueue("q-1")
		assert.NoError(t, err)
	})

	t.Run("DeleteQueue", func(t *testing.T) {
		err := client.DeleteQueue("q-1")
		assert.NoError(t, err)
	})
}

func TestQueueSDK_Errors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("boom"))
	}))
	defer server.Close()

	client := sdk.NewClient(server.URL+"/api/v1", "test-api-key")
	_, err := client.CreateQueue("q", nil, nil, nil)
	assert.Error(t, err)

	_, err = client.ListQueues()
	assert.Error(t, err)

	_, err = client.GetQueue("q-1")
	assert.Error(t, err)

	err = client.DeleteQueue("q-1")
	assert.Error(t, err)

	_, err = client.SendMessage("q-1", "hello")
	assert.Error(t, err)

	_, err = client.ReceiveMessages("q-1", 10)
	assert.Error(t, err)

	err = client.DeleteMessage("q-1", "handle-1")
	assert.Error(t, err)

	err = client.PurgeQueue("q-1")
	assert.Error(t, err)
}
