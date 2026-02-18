package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestSSHKeyList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/ssh-keys" || r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		payload := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":   "key-1",
					"name": "my-laptop",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	oldURL := apiURL
	apiURL = server.URL
	apiKey = "test-key"
	defer func() { 
		apiURL = oldURL 
		apiKey = ""
	}()

	out := captureStdout(t, func() {
		listKeysCmd.Run(listKeysCmd, nil)
	})
	if !strings.Contains(out, "key-1") || !strings.Contains(out, "my-laptop") {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestSSHKeyRegister(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/ssh-keys" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusCreated)
		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"id":   "new-key-id",
				"name": "test-key",
			},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	oldURL := apiURL
	apiURL = server.URL
	apiKey = "test-key"
	defer func() { 
		apiURL = oldURL 
		apiKey = ""
	}()

	// Create a dummy key file
	tmpFile := "test_pub_key.pub"
	_ = os.WriteFile(tmpFile, []byte("ssh-rsa AAA..."), 0600)
	defer func() { _ = os.Remove(tmpFile) }()

	out := captureStdout(t, func() {
		registerKeyCmd.Run(registerKeyCmd, []string{"test-key", tmpFile})
	})
	if !strings.Contains(out, "[SUCCESS]") || !strings.Contains(out, "new-key-id") {
		t.Fatalf("unexpected output: %s", out)
	}
}
