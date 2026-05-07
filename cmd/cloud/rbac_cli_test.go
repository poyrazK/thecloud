package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

const (
	rbacTestAPIKey   = "rbac-key"
	rbacTestRoleID   = "22222222-2222-2222-2222-222222222222"
	rbacTestRoleName = "viewer"
)

func TestCreateRoleCmd(t *testing.T) {
	fmt.Fprintf(os.Stderr, "DEBUG: TestCreateRoleCmd starting\n")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/rbac/roles" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"id":          rbacTestRoleID,
				"name":        rbacTestRoleName,
				"description": "read-only",
			},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	t.Setenv("HOME", t.TempDir())
	saveConfig(rbacTestAPIKey)

	oldURL := opts.APIURL
	opts.APIURL = server.URL
	defer func() { opts.APIURL = oldURL }()

	_ = createRoleCmd.Flags().Set("description", "read-only")
	_ = createRoleCmd.Flags().Set("permissions", string(domain.PermissionInstanceRead))

	fmt.Fprintf(os.Stderr, "DEBUG: About to call createRoleCmd.Run\n")
	var out string
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "DEBUG: panic in Run: %v\n", r)
				t.Fatalf("panic in createRoleCmd.Run: %v", r)
			}
		}()
		fmt.Fprintf(os.Stderr, "DEBUG: Before captureStdout\n")
		// Use direct stdout capture with explicit timing
		oldStdout := os.Stdout
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("pipe: %v", err)
		}
		os.Stdout = w
		fmt.Fprintf(os.Stderr, "DEBUG: Pipe created, about to call Run\n")
		createRoleCmd.Run(createRoleCmd, []string{rbacTestRoleName})
		fmt.Fprintf(os.Stderr, "DEBUG: Run completed, closing pipe\n")
		w.Close()
		os.Stdout = oldStdout
		fmt.Fprintf(os.Stderr, "DEBUG: Pipe closed, about to copy\n")
		var buf strings.Builder
		io.Copy(&buf, r)
		out = buf.String()
		fmt.Fprintf(os.Stderr, "DEBUG: output captured: %d bytes\n", len(out))
	}()
	fmt.Fprintf(os.Stderr, "DEBUG: createRoleCmd.Run completed, output length: %d\n", len(out))
	if !strings.Contains(out, "Role created") || !strings.Contains(out, rbacTestRoleID) {
		t.Fatalf("expected success output, got: %s", out)
	}
}

func TestListRolesCmd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/rbac/roles" || r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		payload := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":          rbacTestRoleID,
					"name":        rbacTestRoleName,
					"permissions": []string{"instance:read"},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	t.Setenv("HOME", t.TempDir())
	saveConfig(rbacTestAPIKey)

	oldURL := opts.APIURL
	opts.APIURL = server.URL
	defer func() { opts.APIURL = oldURL }()

	out := captureStdout(t, func() {
		listRolesCmd.Run(listRolesCmd, nil)
	})
	if !strings.Contains(out, rbacTestRoleName) {
		t.Fatalf("expected list output to include role name, got: %s", out)
	}
}

func TestBindRoleCmd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/rbac/bindings" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Setenv("HOME", t.TempDir())
	saveConfig(rbacTestAPIKey)

	oldURL := opts.APIURL
	opts.APIURL = server.URL
	defer func() { opts.APIURL = oldURL }()

	out := captureStdout(t, func() {
		bindRoleCmd.Run(bindRoleCmd, []string{"user@example.com", rbacTestRoleName})
	})
	if !strings.Contains(out, "Role") || !strings.Contains(out, rbacTestRoleName) {
		t.Fatalf("expected success output, got: %s", out)
	}
}

func TestDeleteRoleCmd(t *testing.T) {
	roleID := uuid.New()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/rbac/roles/"+roleID.String() || r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Setenv("HOME", t.TempDir())
	saveConfig(rbacTestAPIKey)

	oldURL := opts.APIURL
	opts.APIURL = server.URL
	defer func() { opts.APIURL = oldURL }()

	out := captureStdout(t, func() {
		deleteRoleCmd.Run(deleteRoleCmd, []string{roleID.String()})
	})
	if !strings.Contains(out, "Role") || !strings.Contains(out, roleID.String()) {
		t.Fatalf("expected delete output, got: %s", out)
	}
}
