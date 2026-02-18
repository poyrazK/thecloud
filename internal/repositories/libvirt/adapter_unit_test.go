package libvirt

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/digitalocean/go-libvirt"
	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testEnvVar  = "FOO=bar"
	testVolPath = "/path/to/vol1"
)

func TestSanitizeDomainName(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
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
	t.Parallel()
	a := &LibvirtAdapter{}
	result := a.sanitizeDomainName("")

	assert.NotEmpty(t, result, "empty name should generate UUID")
	assert.Len(t, result, 8, "generated name should be 8 chars")
}

func TestParseAndValidatePort(t *testing.T) {
	t.Parallel()
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
			assert.True(t, h >= 30000 || h == 8080 || h == 80)
			assert.True(t, c > 0)
		}
	}
}

func TestGenerateUserData(t *testing.T) {
	t.Parallel()
	a := &LibvirtAdapter{}

	tests := []struct {
		name        string
		env         []string
		cmd         []string
		shouldExist []string
	}{
		{
			name:        "with env vars",
			env:         []string{testEnvVar, "DEBUG=true"},
			cmd:         nil,
			shouldExist: []string{testEnvVar, "DEBUG=true"},
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
			t.Parallel()
			result := a.generateUserData(tt.env, tt.cmd, "")
			content := string(result)

			assert.Contains(t, content, "#cloud-config")
			for _, expected := range tt.shouldExist {
				assert.Contains(t, content, expected)
			}
		})
	}
}

func TestResolveBinds(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			result := a.resolveBinds(context.Background(), tt.binds)
			assert.Empty(t, result)
		})
	}

	t.Run("successful pool lookup", func(t *testing.T) {
		t.Parallel()
		m := new(MockLibvirtClient)
		a := &LibvirtAdapter{client: m}
		ctx := context.Background()
		pool := libvirt.StoragePool{Name: "default"}
		vol := libvirt.StorageVol{Name: "vol1"}

		m.On("StoragePoolLookupByName", ctx, "default").Return(pool, nil)
		m.On("StorageVolLookupByName", ctx, pool, "vol1").Return(vol, nil)
		m.On("StorageVolGetPath", ctx, vol).Return(testVolPath, nil)

		result := a.resolveBinds(ctx, []string{"vol1:/data"})
		assert.Len(t, result, 1)
		assert.Equal(t, testVolPath, result[0])
		m.AssertExpectations(t)
	})
}

func TestPrepareCloudInit(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			result := a.prepareCloudInit(context.Background(), "test", tt.env, tt.cmd, "")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateID(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
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
	t.Parallel()
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
			t.Parallel()
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
	t.Parallel()
	a := &LibvirtAdapter{}
	skipLibvirt := fmt.Errorf("skip libvirt")

	t.Run("LVM path", func(t *testing.T) {
		t.Parallel()
		path := a.resolveVolumePath(context.Background(), "/dev/vg0/lv0", libvirt.StoragePool{}, skipLibvirt)
		assert.Equal(t, "/dev/vg0/lv0", path)
	})

	t.Run("Direct file path", func(t *testing.T) {
		t.Parallel()
		// Use a file that definitely exists
		path := a.resolveVolumePath(context.Background(), "/etc/hosts", libvirt.StoragePool{}, skipLibvirt)
		assert.Equal(t, "/etc/hosts", path)
	})

	t.Run("Non-existent path", func(t *testing.T) {
		t.Parallel()
		path := a.resolveVolumePath(context.Background(), "/non/existent/path", libvirt.StoragePool{}, skipLibvirt)
		assert.Equal(t, "", path)
	})

	t.Run("Libvirt pool lookup success", func(t *testing.T) {
		t.Parallel()
		m := new(MockLibvirtClient)
		a := &LibvirtAdapter{client: m}
		ctx := context.Background()
		pool := libvirt.StoragePool{Name: "default"}
		vol := libvirt.StorageVol{Name: "vol1"}

		m.On("StorageVolLookupByName", ctx, pool, "vol1").Return(vol, nil)
		m.On("StorageVolGetPath", ctx, vol).Return(testVolPath, nil)

		path := a.resolveVolumePath(ctx, "vol1", pool, nil)
		assert.Equal(t, testVolPath, path)
		m.AssertExpectations(t)
	})
}

func TestCleanupCreateFailure(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := &LibvirtAdapter{
		client: m,
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	ctx := context.Background()

	// 1. Success path
	vol := libvirt.StorageVol{Name: "test-vol"}
	m.On("StorageVolDelete", ctx, vol, uint32(0)).Return(nil).Once()
	a.cleanupCreateFailure(ctx, vol, "/tmp/iso")

	// 2. Failure path (should not panic)
	m.On("StorageVolDelete", ctx, vol, uint32(0)).Return(fmt.Errorf("delete error")).Once()
	a.cleanupCreateFailure(ctx, vol, "/tmp/iso-fail")

	m.AssertExpectations(t)
}

func TestGetInstanceLogs(t *testing.T) {
	t.Parallel()
	// Mock osOpen
	tmpFile, _ := os.CreateTemp("", "log")
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_, _ = tmpFile.WriteString("log data")
	_ = tmpFile.Close()

	m := new(MockLibvirtClient)
	a := &LibvirtAdapter{
		client: m,
		osOpen: func(name string) (*os.File, error) {
			return os.Open(tmpFile.Name())
		},
	}
	ctx := context.Background()

	rc, err := a.GetInstanceLogs(ctx, "test")
	assert.NoError(t, err)
	assert.NotNil(t, rc)
	_ = rc.Close()
}

func TestGetInstancePort(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := &LibvirtAdapter{
		client:       m,
		portMappings: map[string]map[string]int{"test": {"80": 8080}},
	}
	ctx := context.Background()

	port, err := a.GetInstancePort(ctx, "test", "80")
	assert.NoError(t, err)
	assert.Equal(t, 8080, port)

	_, err = a.GetInstancePort(ctx, "test", "443")
	assert.Error(t, err)
}

func TestNewLibvirtAdapter(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	_, _ = NewLibvirtAdapter(logger, "qemu:///system")
}

func TestClose(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := &LibvirtAdapter{client: m}
	m.On("Close").Return(nil)
	err := a.Close()
	assert.NoError(t, err)
	m.AssertExpectations(t)
}

func TestWaitInitialIPContextCancellation(t *testing.T) {
	t.Parallel()
	a := &LibvirtAdapter{
		ipWaitInterval: time.Millisecond,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := a.waitInitialIP(ctx, "test-instance")
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestGetNextNetworkRange(t *testing.T) {
	t.Parallel()
	a := &LibvirtAdapter{
		networkCounter: 0,
		poolStart:      net.ParseIP(testutil.TestLibvirtPoolStart),
		poolEnd:        net.ParseIP(testutil.TestLibvirtPoolEnd),
		ipWaitInterval: time.Millisecond,
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

func TestGetNextNetworkRangePoolExhaustion(t *testing.T) {
	t.Parallel()
	a := &LibvirtAdapter{
		networkCounter: 254, // The 255th /24 network
		// Test IPs for private network calculation
		poolStart: net.ParseIP("192.168.0.0"),
		poolEnd:   net.ParseIP("192.168.255.255"),
	}

	gateway, _, _ := a.getNextNetworkRange()
	// 192.168.0.0 + (254 * 256) = 192.168.254.0
	// Gateway is .1 -> 192.168.254.1
	assert.Equal(t, "192.168.254.1", gateway, "Should calculate 255th network correctly")

	// Next one should be 192.168.255.1
	gateway2, _, _ := a.getNextNetworkRange()
	assert.Equal(t, "192.168.255.1", gateway2, "Should calculate 256th network correctly")
}

func TestExecNotSupported(t *testing.T) {
	t.Parallel()
	a := &LibvirtAdapter{}
	_, err := a.Exec(context.Background(), "instance-id", []string{"echo", "test"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}

func TestCleanupPortMappings(t *testing.T) {
	t.Parallel()
	a := &LibvirtAdapter{
		portMappings: make(map[string]map[string]int),
	}
	instanceID := "inst-1"
	a.portMappings[instanceID] = map[string]int{"80": 30080}

	a.cleanupPortMappings(instanceID)

	assert.Empty(t, a.portMappings[instanceID])
}

func TestGenerateCloudInitISO(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	a := &LibvirtAdapter{
		logger: logger,
		execCommand: func(name string, arg ...string) *exec.Cmd {
			return exec.Command("true")
		},
		lookPath: func(file string) (string, error) {
			return "/usr/bin/true", nil
		},
	}
	ctx := context.Background()
	name := "test-asg"
	env := []string{testEnvVar}
	cmd := []string{"ls -la"}

	isoPath, err := a.generateCloudInitISO(ctx, name, env, cmd, "")
	assert.NoError(t, err)
	assert.NotEmpty(t, isoPath)
	assert.Contains(t, isoPath, ".iso")

	// Cleanup the generated ISO file if it exists (though it might not if true succeeded)
	if isoPath != "" {
		_ = os.Remove(isoPath)
	}
}

func TestStopInstance_Unit(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := &LibvirtAdapter{client: m, logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	ctx := context.Background()
	dom := libvirt.Domain{Name: "test-vm"}

	m.On("DomainLookupByName", mock.Anything, "test-vm").Return(dom, nil).Once()
	m.On("DomainDestroy", mock.Anything, dom).Return(nil).Once()

	err := a.StopInstance(ctx, "test-vm")
	assert.NoError(t, err)
	m.AssertExpectations(t)
}

func TestDeleteInstance_Unit(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := &LibvirtAdapter{
		client: m, 
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		portMappings: make(map[string]map[string]int),
	}
	ctx := context.Background()
	dom := libvirt.Domain{Name: "test-vm"}

	m.On("DomainLookupByName", mock.Anything, "test-vm").Return(dom, nil).Once()
	m.On("DomainGetState", mock.Anything, dom, uint32(0)).Return(int32(domainStateRunning), int32(0), nil).Once()
	m.On("DomainDestroy", mock.Anything, dom).Return(nil).Once()
	m.On("DomainUndefine", mock.Anything, dom).Return(nil).Once()
	
	// Root volume cleanup mocks
	pool := libvirt.StoragePool{Name: "default"}
	vol := libvirt.StorageVol{Name: "test-vm-root"}
	m.On("StoragePoolLookupByName", mock.Anything, "default").Return(pool, nil).Once()
	m.On("StorageVolLookupByName", mock.Anything, pool, "test-vm-root").Return(vol, nil).Once()
	m.On("StorageVolDelete", mock.Anything, vol, uint32(0)).Return(nil).Once()

	err := a.DeleteInstance(ctx, "test-vm")
	assert.NoError(t, err)
	m.AssertExpectations(t)
}
