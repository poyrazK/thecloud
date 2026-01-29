package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	notifyTestAPIKey  = "notify-key"
	notifyTestTopicID = "topic-1"
	notifyTestTopic   = "alerts"
)

func TestCreateTopicCmd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/notify/topics" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"id":   notifyTestTopicID,
				"name": notifyTestTopic,
				"arn":  "arn:cloud:notify:topic-1",
			},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	oldURL := apiURL
	oldKey := apiKey
	apiURL = server.URL
	apiKey = notifyTestAPIKey
	defer func() {
		apiURL = oldURL
		apiKey = oldKey
	}()

	out := captureStdout(t, func() {
		createTopicCmd.Run(createTopicCmd, []string{notifyTestTopic})
	})
	if !strings.Contains(out, "Topic created") || !strings.Contains(out, notifyTestTopicID) {
		t.Fatalf("expected success output, got: %s", out)
	}
}

func TestPublishCmd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/notify/topics/"+notifyTestTopicID+"/publish" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	oldURL := apiURL
	oldKey := apiKey
	apiURL = server.URL
	apiKey = notifyTestAPIKey
	defer func() {
		apiURL = oldURL
		apiKey = oldKey
	}()

	out := captureStdout(t, func() {
		publishCmd.Run(publishCmd, []string{notifyTestTopicID, "hello"})
	})
	if !strings.Contains(out, "Message published") {
		t.Fatalf("expected success output, got: %s", out)
	}
}
