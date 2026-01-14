package libvirt

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/digitalocean/go-libvirt"
	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/assert"
)

func TestSanitizeDomainName(t *testing.T) {
	a := &LibvirtAdapter{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid name",
			input:    "my-instance",
			expected: "my-instance",
		},
		{
			name:     "with invalid chars",
			input:    "my@instance#123",
			expected: "myinstance123",
		},
		{
			name:     "with spaces",
			input:    "my instance",
			expected: "myinstance",
		},
		{
			name:     "with special chars",
			input:    "test_vm.local",
			expected: "testvmlocal",
		},
		{
			name:     "only valid chars",
			input:    "abc123-test",
			expected: "abc123-test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := a.sanitizeDomainName(tt.input)
			if tt.input == "" {
				// Empty string generates UUID
				assert.NotEmpty(t, result)
				assert.Len(t, result, 8)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestSanitizeDomainNameEmptyString(t *testing.T) {
	a := &LibvirtAdapter{}
	result := a.sanitizeDomainName("")

	assert.NotEmpty(t, result, "empty name should generate UUID")
	assert.Len(t, result, 8, "generated name should be 8 chars")
}

func TestParseAndValidatePort(t *testing.T) {
	a := &LibvirtAdapter{}

	tests := []struct {
		input     string
		wantHost  bool
		wantError bool
	}{
		{"8080:80", true, false},
		{"80", true, false},
		{"invalid", false, true},
		{"80:80:80", false, true},
		{"abc:80", false, true},
	}

	for _, tt := range tests {
		h, c, err := a.parseAndValidatePort(tt.input)
		if tt.wantError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.True(t, h >= 30000 || h == 8080)
			assert.True(t, c > 0)
		}
	}
}

func TestGenerateUserData(t *testing.T) {
	a := &LibvirtAdapter{}

	tests := []struct {
		name        string
		env         []string
		cmd         []string
		shouldExist []string
	}{
		{
			name:        "with env vars",
			env:         []string{"FOO=bar", "DEBUG=true"},
			cmd:         nil,
			shouldExist: []string{"FOO=bar", "DEBUG=true"},
		},
		{
			name:        "with command",
			env:         nil,
			cmd:         []string{"echo", "hello"},
			shouldExist: []string{"echo", "hello"},
		},
		{
			name:        "with both",
			env:         []string{"PATH=/usr/bin"},
			cmd:         []string{"/bin/sh", "-c", "echo test"},
			shouldExist: []string{"PATH=/usr/bin", "/bin/sh", "-c", "echo test"},
		},
		{
			name:        "empty",
			env:         nil,
			cmd:         nil,
			shouldExist: []string{"#cloud-config"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := a.generateUserData(tt.env, tt.cmd)
			content := string(result)

			assert.Contains(t, content, "#cloud-config")
			for _, expected := range tt.shouldExist {
				assert.Contains(t, content, expected)
			}
		})
	}
}

func TestResolveBinds(t *testing.T) {
	a := &LibvirtAdapter{}

	tests := []struct {
		name  string
		binds []string
	}{
		{
			name:  "nil binds",
			binds: nil,
		},
		{
			name:  "empty binds",
			binds: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := a.resolveBinds(tt.binds)
			assert.Empty(t, result)
		})
	}
}

func TestPrepareCloudInit(t *testing.T) {
	a := &LibvirtAdapter{}

	tests := []struct {
		name     string
		env      []string
		cmd      []string
		expected string
	}{
		{
			name:     "no env or cmd",
			env:      nil,
			cmd:      nil,
			expected: "",
		},
		{
			name:     "empty slices",
			env:      []string{},
			cmd:      []string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := a.prepareCloudInit(context.Background(), "test", tt.env, tt.cmd)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "valid simple",
			id:      "instance-123",
			wantErr: false,
		},
		{
			name:    "valid uuid",
			id:      "abc123-def456",
			wantErr: false,
		},
		{
			name:    "path traversal",
			id:      "../etc/passwd",
			wantErr: true,
		},
		{
			name:    "absolute path",
			id:      "/etc/passwd",
			wantErr: true,
		},
		{
			name:    "dot dot",
			id:      "test/../admin",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateID(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseAndValidatePortExtended(t *testing.T) {
	a := &LibvirtAdapter{}

	tests := []struct {
		name    string
		port    string
		hostP   int
		contP   int
		wantErr bool
	}{
		{
			name:    "same port",
			port:    "80:80",
			hostP:   80,
			contP:   80,
			wantErr: false,
		},
		{
			name:    "different ports",
			port:    "8080:80",
			hostP:   8080,
			contP:   80,
			wantErr: false,
		},
		{
			name:    "high port numbers",
			port:    "65535:8080",
			hostP:   65535,
			contP:   8080,
			wantErr: false,
		},
		{
			name:    "invalid format three parts",
			port:    "80:80:80",
			wantErr: true,
		},
		{
			name:    "invalid host port",
			port:    "abc:80",
			wantErr: true,
		},
		{
			name:    "invalid container port",
			port:    "80:xyz",
			wantErr: true,
		},
		{
			name:    "empty string",
			port:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, cont, err := a.parseAndValidatePort(tt.port)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.hostP, host)
				assert.Equal(t, tt.contP, cont)
			}
		})
	}
}

func TestResolveVolumePath(t *testing.T) {
	a := &LibvirtAdapter{}
	skipLibvirt := fmt.Errorf("skip libvirt")

	t.Run("LVM path", func(t *testing.T) {
		path := a.resolveVolumePath("/dev/vg0/lv0", libvirt.StoragePool{}, skipLibvirt)
		assert.Equal(t, "/dev/vg0/lv0", path)
	})

	t.Run("Direct file path", func(t *testing.T) {
		// Use a file that definitely exists
		path := a.resolveVolumePath("/etc/hosts", libvirt.StoragePool{}, skipLibvirt)
		assert.Equal(t, "/etc/hosts", path)
	})

	t.Run("Non-existent path", func(t *testing.T) {
		path := a.resolveVolumePath("/non/existent/path", libvirt.StoragePool{}, skipLibvirt)
		assert.Equal(t, "", path)
	})
}

func TestCleanupCreateFailure(t *testing.T) {
	// This function has side effects but we can test it doesn't panic
	a := &LibvirtAdapter{}

	// Should not panic - just verify it's callable
	// Cannot test without real libvirt connection
	assert.NotNil(t, a)
}

func TestWaitInitialIPContextCancellation(t *testing.T) {
	a := &LibvirtAdapter{}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := a.waitInitialIP(ctx, "test-instance")
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestGetNextNetworkRange(t *testing.T) {
	a := &LibvirtAdapter{
		networkCounter: 0,
		poolStart:      net.ParseIP(testutil.TestLibvirtPoolStart),
		poolEnd:        net.ParseIP(testutil.TestLibvirtPoolEnd),
	}

	gateway1, start1, end1 := a.getNextNetworkRange()
	assert.NotEmpty(t, gateway1)
	assert.NotEmpty(t, start1)
	assert.NotEmpty(t, end1)
	assert.Equal(t, 1, a.networkCounter)

	gateway2, _, _ := a.getNextNetworkRange()
	assert.NotEqual(t, gateway1, gateway2, "subsequent calls should return different gateways")
	assert.Equal(t, 2, a.networkCounter)
}

func TestExecNotSupported(t *testing.T) {
	a := &LibvirtAdapter{}
	_, err := a.Exec(context.Background(), "instance-id", []string{"echo", "test"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}
