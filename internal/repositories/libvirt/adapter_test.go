package libvirt

import (
	"context"
	"fmt"
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

func TestLibvirtAdapterType(t *testing.T) {
	a := &LibvirtAdapter{}
	assert.Equal(t, "libvirt", a.Type())
}

func TestLibvirtAdapterValidateID(t *testing.T) {
	assert.NoError(t, validateID("valid-id"))
	assert.Error(t, validateID("../traversal"))
}

func TestLibvirtAdapterParseAndValidatePort(t *testing.T) {
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
	a := &LibvirtAdapter{}
	// Test empty binds
	resolved := a.resolveBinds(nil)
	assert.Empty(t, resolved)

	// We cannot test full resolveBinds without a libvirt connection mock (complex)
}

func TestGenerateNginxConfig(t *testing.T) {
	t.Run("round-robin with targets", func(t *testing.T) {
		mockCompute := new(mockComputeBackend)
		a := NewLBProxyAdapter(mockCompute)

		lb := &domain.LoadBalancer{
			ID:        uuid.New(),
			Port:      80,
			Algorithm: "round-robin",
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
		mockCompute := new(mockComputeBackend)
		a := NewLBProxyAdapter(mockCompute)

		lb := &domain.LoadBalancer{
			ID:        uuid.New(),
			Port:      80,
			Algorithm: "round-robin",
		}

		config, err := a.generateNginxConfig(context.Background(), lb, nil)
		assert.NoError(t, err)
		assert.Contains(t, config, "return 503")
		assert.NotContains(t, config, "upstream backend")
	})
}
