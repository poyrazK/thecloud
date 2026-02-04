//go:build integration

package libvirt

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockComputeBackend struct {
	ports.ComputeBackend
	mock.Mock
}

func (m *mockComputeBackend) GetInstanceIP(ctx context.Context, id string) (string, error) {
	args := m.Called(ctx, id)
	return args.String(0), args.Error(1)
}

const (
	algoRoundRobin = "round-robin"
)

func TestLibvirtAdapterType(t *testing.T) {
	t.Parallel()
	a := &LibvirtAdapter{}
	assert.Equal(t, "libvirt", a.Type())
}

func TestLibvirtAdapterValidateID(t *testing.T) {
	t.Parallel()
	assert.NoError(t, validateID("valid-id"))
	assert.Error(t, validateID("../traversal"))
}

func TestLibvirtAdapterParseAndValidatePort(t *testing.T) {
	t.Parallel()
	a := &LibvirtAdapter{}

	tests := []struct {
		input    string
		wantHost int
		wantCont int
		wantErr  bool
	}{
		{"80:80", 80, 80, false},
		{"8080:80", 8080, 80, false},
		{"80:80:80", 0, 0, true},
		{"abc:80", 0, 0, true},
	}

	for _, tt := range tests {
		h, c, err := a.parseAndValidatePort(tt.input)
		if tt.wantErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.wantHost, h)
			assert.Equal(t, tt.wantCont, c)
		}
	}
}

func TestLibvirtAdapterResolveBinds(t *testing.T) {
	t.Parallel()
	a := &LibvirtAdapter{}
	// Test empty binds
	resolved := a.resolveBinds(context.Background(), nil)
	assert.Empty(t, resolved)

	// We cannot test full resolveBinds without a libvirt connection mock (complex)
}

func TestGenerateNginxConfig(t *testing.T) {
	t.Parallel()
	t.Run("round-robin with targets", func(t *testing.T) {
		t.Parallel()
		mockCompute := new(mockComputeBackend)
		a := NewLBProxyAdapter(mockCompute)

		lb := &domain.LoadBalancer{
			ID:        uuid.New(),
			Port:      80,
			Algorithm: algoRoundRobin,
		}
		instanceID := uuid.New()
		targets := []*domain.LBTarget{
			{InstanceID: instanceID, Port: 8080, Weight: 1},
		}
		mockCompute.On("GetInstanceIP", mock.Anything, instanceID.String()).Return(testutil.TestLibvirtInstanceIP, nil)

		config, err := a.generateNginxConfig(context.Background(), lb, targets)
		assert.NoError(t, err)
		assert.Contains(t, config, "listen 80;")
		assert.Contains(t, config, fmt.Sprintf("server %s:8080 weight=1;", testutil.TestLibvirtInstanceIP))
		assert.NotContains(t, config, "least_conn;")
	})

	t.Run("least-conn with targets", func(t *testing.T) {
		t.Parallel()
		mockCompute := new(mockComputeBackend)
		a := NewLBProxyAdapter(mockCompute)

		lb := &domain.LoadBalancer{
			ID:        uuid.New(),
			Port:      80,
			Algorithm: "least-conn",
		}
		instanceID := uuid.New()
		targets := []*domain.LBTarget{
			{InstanceID: instanceID, Port: 8080, Weight: 1},
		}

		mockCompute.On("GetInstanceIP", mock.Anything, instanceID.String()).Return(testutil.TestLibvirtInstanceIP, nil)

		config, err := a.generateNginxConfig(context.Background(), lb, targets)
		assert.NoError(t, err)
		assert.Contains(t, config, "least_conn;")
	})

	t.Run("no targets", func(t *testing.T) {
		t.Parallel()
		mockCompute := new(mockComputeBackend)
		a := NewLBProxyAdapter(mockCompute)

		lb := &domain.LoadBalancer{
			ID:        uuid.New(),
			Port:      80,
			Algorithm: algoRoundRobin,
		}

		config, err := a.generateNginxConfig(context.Background(), lb, nil)
		assert.NoError(t, err)
		assert.Contains(t, config, "return 503")
		assert.NotContains(t, config, "upstream backend")
	})

	t.Run("compute error", func(t *testing.T) {
		t.Parallel()
		mockCompute := new(mockComputeBackend)
		a := NewLBProxyAdapter(mockCompute)

		lb := &domain.LoadBalancer{
			ID:        uuid.New(),
			Port:      80,
			Algorithm: algoRoundRobin,
		}
		instanceID := uuid.New()
		targets := []*domain.LBTarget{
			{InstanceID: instanceID, Port: 8080, Weight: 1},
		}

		mockCompute.On("GetInstanceIP", mock.Anything, instanceID.String()).Return("", fmt.Errorf("compute error"))

		config, err := a.generateNginxConfig(context.Background(), lb, targets)
		assert.NoError(t, err)
		assert.Contains(t, config, "return 503")
		assert.NotContains(t, config, "upstream backend")
	})
}

func TestLBProxyAdapterOperations(t *testing.T) {
	t.Parallel()
	// Mock execCommandContext using a hermetic helper process
	oldExec := execCommandContext
	defer func() { execCommandContext = oldExec }()
	execCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", name}
		cs = append(cs, args...)
		cmd := exec.CommandContext(ctx, os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}

	mockCompute := new(mockComputeBackend)
	a := NewLBProxyAdapter(mockCompute)
	ctx := context.Background()

	lb := &domain.LoadBalancer{
		ID:        uuid.New(),
		Port:      80,
		Algorithm: algoRoundRobin,
	}

	t.Run("DeployProxy", func(t *testing.T) {
		t.Parallel()
		id, err := a.DeployProxy(ctx, lb, nil)
		assert.NoError(t, err)
		assert.Equal(t, lb.ID.String(), id)
	})

	t.Run("UpdateProxyConfig", func(t *testing.T) {
		t.Parallel()
		err := a.UpdateProxyConfig(ctx, lb, nil)
		assert.NoError(t, err)
	})

	t.Run("RemoveProxy", func(t *testing.T) {
		t.Parallel()
		err := a.RemoveProxy(ctx, lb.ID)
		assert.NoError(t, err)
	})
}

// TestHelperProcess isn't a real test. It's used as a helper process for TestLBProxyAdapterOperations.
func TestHelperProcess(t *testing.T) {
	t.Parallel()
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	// Exit with 0 (success) which approximates "true" behavior
	os.Exit(0)
}
