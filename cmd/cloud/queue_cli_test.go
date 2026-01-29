package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const (
	queueTestAPIKey = "queue-key"
	queueTestID     = "queue-1"
)

func TestQueueListJSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/queues" || r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		payload := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":                 queueTestID,
					"name":               "jobs",
					"arn":                "arn:cloud:queue:jobs",
					"visibility_timeout": 30,
					"retention_days":     4,
					"max_message_size":   262144,
					"status":             "active",
					"created_at":         time.Now().UTC().Format(time.RFC3339),
					"updated_at":         time.Now().UTC().Format(time.RFC3339),
				},
			},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	oldURL := apiURL
	oldKey := apiKey
	apiURL = server.URL
	apiKey = queueTestAPIKey
	outputJSON = true
	defer func() {
		apiURL = oldURL
		apiKey = oldKey
		outputJSON = false
	}()

	out := captureStdout(t, func() {
		listQueuesCmd.Run(listQueuesCmd, nil)
	})
	if !strings.Contains(out, queueTestID) {
		t.Fatalf("expected JSON output to include queue id, got: %s", out)
	}
}

func TestQueueCreateCmd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/queues" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"id":         queueTestID,
				"name":       "jobs",
				"arn":        "arn:cloud:queue:jobs",
				"status":     "active",
				"created_at": time.Now().UTC().Format(time.RFC3339),
				"updated_at": time.Now().UTC().Format(time.RFC3339),
			},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	oldURL := apiURL
	oldKey := apiKey
	apiURL = server.URL
	apiKey = queueTestAPIKey
	defer func() {
		apiURL = oldURL
		apiKey = oldKey
	}()

	out := captureStdout(t, func() {
		createQueueCmd.Run(createQueueCmd, []string{"jobs"})
	})
	if !strings.Contains(out, "Queue created") || !strings.Contains(out, queueTestID) {
		t.Fatalf("expected create output, got: %s", out)
	}
}

func TestQueueReceiveNoMessages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/queues/"+queueTestID+"/messages" || r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		payload := map[string]interface{}{
			"data": []map[string]interface{}{},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	oldURL := apiURL
	oldKey := apiKey
	apiURL = server.URL
	apiKey = queueTestAPIKey
	defer func() {
		apiURL = oldURL
		apiKey = oldKey
	}()

	_ = receiveMessagesCmd.Flags().Set("max", "1")
	out := captureStdout(t, func() {
		receiveMessagesCmd.Run(receiveMessagesCmd, []string{queueTestID})
	})
	if !strings.Contains(out, "No messages available") {
		t.Fatalf("expected no messages output, got: %s", out)
	}
}

func TestQueueSendMessageCmd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/queues/"+queueTestID+"/messages" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"id":             "msg-1",
				"queue_id":       queueTestID,
				"body":           "hello",
				"receipt_handle": "handle-1",
				"created_at":     time.Now().UTC().Format(time.RFC3339),
			},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	oldURL := apiURL
	oldKey := apiKey
	apiURL = server.URL
	apiKey = queueTestAPIKey
	defer func() {
		apiURL = oldURL
		apiKey = oldKey
	}()

	out := captureStdout(t, func() {
		sendMessageCmd.Run(sendMessageCmd, []string{queueTestID, "hello"})
	})
	if !strings.Contains(out, "Message sent") {
		t.Fatalf("expected send message output, got: %s", out)
	}
}

func TestQueueAckCmd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/queues/"+queueTestID+"/messages/handle-1" || r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	oldURL := apiURL
	oldKey := apiKey
	apiURL = server.URL
	apiKey = queueTestAPIKey
	defer func() {
		apiURL = oldURL
		apiKey = oldKey
	}()

	out := captureStdout(t, func() {
		ackMessageCmd.Run(ackMessageCmd, []string{queueTestID, "handle-1"})
	})
	if !strings.Contains(out, "Message acknowledged") {
		t.Fatalf("expected ack output, got: %s", out)
	}
}

func TestQueuePurgeCmd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/queues/"+queueTestID+"/purge" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	oldURL := apiURL
	oldKey := apiKey
	apiURL = server.URL
	apiKey = queueTestAPIKey
	defer func() {
		apiURL = oldURL
		apiKey = oldKey
	}()

	out := captureStdout(t, func() {
		purgeQueueCmd.Run(purgeQueueCmd, []string{queueTestID})
	})
	if !strings.Contains(out, "Queue purged") {
		t.Fatalf("expected purge output, got: %s", out)
	}
}
