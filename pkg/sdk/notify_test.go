package sdk_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/poyrazk/thecloud/pkg/sdk"
	"github.com/stretchr/testify/assert"
)

func TestNotifySDK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/api/v1/notify/topics" && r.Method == "POST" {
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["name"] == "my-topic" {
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{
						"id":   "topic-1",
						"name": "my-topic",
					},
				})
				return
			}
		}

		if r.URL.Path == "/api/v1/notify/topics" && r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []map[string]interface{}{
					{"id": "topic-1", "name": "my-topic"},
				},
			})
			return
		}

		if r.URL.Path == "/api/v1/notify/topics/topic-1" && r.Method == "DELETE" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.URL.Path == "/api/v1/notify/topics/topic-1/subscriptions" && r.Method == "POST" {
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["protocol"] == "http" {
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{
						"id":       "sub-1",
						"topic_id": "topic-1",
					},
				})
				return
			}
		}

		if r.URL.Path == "/api/v1/notify/topics/topic-1/subscriptions" && r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []map[string]interface{}{
					{"id": "sub-1", "topic_id": "topic-1"},
				},
			})
			return
		}

		if r.URL.Path == "/api/v1/notify/subscriptions/sub-1" && r.Method == "DELETE" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.URL.Path == "/api/v1/notify/topics/topic-1/publish" && r.Method == "POST" {
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["message"] == "hello" {
				w.WriteHeader(http.StatusOK)
				return
			}
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	client := sdk.NewClient(ts.URL+"/api/v1", "test-api-key")

	t.Run("CreateTopic", func(t *testing.T) {
		topic, err := client.CreateTopic("my-topic")
		assert.NoError(t, err)
		if topic != nil {
			assert.Equal(t, "my-topic", topic.Name)
		}
	})

	t.Run("ListTopics", func(t *testing.T) {
		topics, err := client.ListTopics()
		assert.NoError(t, err)
		assert.Len(t, topics, 1)
	})

	t.Run("DeleteTopic", func(t *testing.T) {
		err := client.DeleteTopic("topic-1")
		assert.NoError(t, err)
	})

	t.Run("Subscribe", func(t *testing.T) {
		sub, err := client.Subscribe("topic-1", "http", "http://example.com")
		assert.NoError(t, err)
		if sub != nil {
			assert.Equal(t, "sub-1", sub.ID)
		}
	})

	t.Run("ListSubscriptions", func(t *testing.T) {
		subs, err := client.ListSubscriptions("topic-1")
		assert.NoError(t, err)
		assert.Len(t, subs, 1)
	})

	t.Run("Unsubscribe", func(t *testing.T) {
		err := client.Unsubscribe("sub-1")
		assert.NoError(t, err)
	})

	t.Run("Publish", func(t *testing.T) {
		err := client.Publish("topic-1", "hello")
		assert.NoError(t, err)
	})
}
