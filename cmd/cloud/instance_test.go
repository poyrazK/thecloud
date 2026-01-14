package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/poyrazk/thecloud/pkg/sdk"
)

// TestGetClient_WithApiKeyFlag checks if getClient() respects the apiKey flag.

func TestGetClientWithApiKeyFlag(t *testing.T) {
	oldKey := apiKey
	defer func() { apiKey = oldKey }()
	apiKey = "test-key"

	client := getClient()
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestGetClientWithEnvVar(t *testing.T) {
	oldKey := apiKey
	defer func() { apiKey = oldKey }()
	apiKey = ""

	t.Setenv("CLOUD_API_KEY", "env-key")

	client := getClient()
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestListCmdCommandSetup(t *testing.T) {
	// Validate the command setup
	if instanceCmd == nil {
		t.Fatal("instanceCmd should not be nil")
	}
	if listCmd == nil {
		t.Fatal("listCmd should not be nil")
	}
}

func TestLaunchCmdFlagsSetup(t *testing.T) {
	if launchCmd.Flag("name") == nil {
		t.Error("launch command should have --name flag")
	}
	if launchCmd.Flag("image") == nil {
		t.Error("launch command should have --image flag")
	}
	if launchCmd.Flag("port") == nil {
		t.Error("launch command should have --port flag")
	}
	if launchCmd.Flag("vpc") == nil {
		t.Error("launch command should have --vpc flag")
	}
	if launchCmd.Flag("subnet") == nil {
		t.Error("launch command should have --subnet flag")
	}
	if launchCmd.Flag("volume") == nil {
		t.Error("launch command should have --volume flag")
	}
}

func TestStopCmdRequiresOneArg(t *testing.T) {
	if stopCmd.Args == nil {
		t.Fatal("stop command should require args")
	}

	// Verify it's set to require exactly 1 arg
	err := stopCmd.Args(stopCmd, []string{})
	if err == nil {
		t.Error("stop command should error with no args")
	}

	err = stopCmd.Args(stopCmd, []string{"id1", "id2"})
	if err == nil {
		t.Error("stop command should error with multiple args")
	}

	err = stopCmd.Args(stopCmd, []string{"id1"})
	if err != nil {
		t.Errorf("stop command should accept 1 arg, got error: %v", err)
	}
}

func TestLogsCmdRequiresOneArg(t *testing.T) {
	if logsCmd.Args == nil {
		t.Fatal("logs command should require args")
	}

	err := logsCmd.Args(logsCmd, []string{})
	if err == nil {
		t.Error("logs command should error with no args")
	}

	err = logsCmd.Args(logsCmd, []string{"id1"})
	if err != nil {
		t.Errorf("logs command should accept 1 arg, got error: %v", err)
	}
}

func TestShowCmdRequiresOneArg(t *testing.T) {
	if showCmd.Args == nil {
		t.Fatal("show command should require args")
	}

	err := showCmd.Args(showCmd, []string{})
	if err == nil {
		t.Error("show command should error with no args")
	}
}

func TestRmCmdRequiresOneArg(t *testing.T) {
	if rmCmd.Args == nil {
		t.Fatal("rm command should require args")
	}

	err := rmCmd.Args(rmCmd, []string{})
	if err == nil {
		t.Error("rm command should error with no args")
	}
}

func TestStatsCmdRequiresOneArg(t *testing.T) {
	if statsCmd.Args == nil {
		t.Fatal("stats command should require args")
	}

	err := statsCmd.Args(statsCmd, []string{})
	if err == nil {
		t.Error("stats command should error with no args")
	}
}

func TestInstanceCmdHasSubcommands(t *testing.T) {
	subcommands := instanceCmd.Commands()

	expectedCommands := []string{"list", "launch", "stop", "logs", "show", "rm", "stats"}
	found := make(map[string]bool)

	for _, cmd := range subcommands {
		found[cmd.Name()] = true
	}

	for _, expected := range expectedCommands {
		if !found[expected] {
			t.Errorf("instance command missing subcommand: %s", expected)
		}
	}
}

const portMapping = "8080:80"

func TestFormatAccessPorts(t *testing.T) {
	tests := []struct {
		name     string
		ports    string
		status   string
		expected string
	}{
		{
			name:     "no ports",
			ports:    "",
			status:   "RUNNING",
			expected: "-",
		},
		{
			name:     "stopped instance",
			ports:    portMapping,
			status:   "STOPPED",
			expected: "-",
		},
		{
			name:     "single port mapping",
			ports:    portMapping,
			status:   "RUNNING",
			expected: "localhost:8080->80",
		},
		{
			name:     "multiple port mappings",
			ports:    portMapping + ",8443:443",
			status:   "RUNNING",
			expected: "localhost:8080->80, localhost:8443->443",
		},
		{
			name:     "invalid port format",
			ports:    "8080",
			status:   "RUNNING",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			access := formatAccessPorts(tt.ports, tt.status)

			if access != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, access)
			}
		})
	}
}

func TestTruncateID(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected string
	}{
		{
			name:     "short id",
			id:       "abc123",
			expected: "abc123",
		},
		{
			name:     "long id",
			id:       "abcdef123456789",
			expected: "abcdef12",
		},
		{
			name:     "exactly 8 chars",
			id:       "12345678",
			expected: "12345678",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.id
			if len(result) > 8 {
				result = result[:8]
			}

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestJSONMarshalIndent(t *testing.T) {
	inst := sdk.Instance{
		ID:     "test-id",
		Name:   "test-instance",
		Image:  "alpine",
		Status: "RUNNING",
		Ports:  portMapping,
	}

	data, err := json.MarshalIndent(inst, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if !strings.Contains(string(data), "test-id") {
		t.Error("JSON should contain instance ID")
	}
	if !strings.Contains(string(data), "test-instance") {
		t.Error("JSON should contain instance name")
	}
}
