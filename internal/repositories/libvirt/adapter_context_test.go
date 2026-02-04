//go:build integration

package libvirt

import (
	"context"
	"testing"

	"github.com/digitalocean/go-libvirt"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateInstanceContextCancelled(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// The adapter calls prepareRootVolume -> StoragePoolLookupByName
	m.On("StoragePoolLookupByName", ctx, mock.Anything).Return(libvirt.StoragePool{}, context.Canceled)

	opts := ports.CreateInstanceOptions{
		Name:      testInstanceName,
		ImageName: "ubuntu",
	}

	_, err := a.LaunchInstanceWithOptions(ctx, opts)
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestStopInstanceContextCancelled(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	m.On("DomainLookupByName", ctx, testInstanceName).Return(libvirt.Domain{}, context.Canceled)

	err := a.StopInstance(ctx, testInstanceName)
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestGetInstanceIPContextCancelled(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	m.On("DomainLookupByName", ctx, testInstanceName).Return(libvirt.Domain{}, context.Canceled)

	_, err := a.GetInstanceIP(ctx, testInstanceName)
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestCreateNetworkContextCancelled(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	m.On("NetworkDefineXML", ctx, mock.Anything).Return(libvirt.Network{}, context.Canceled)

	_, err := a.CreateNetwork(ctx, "test-net")
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestCreateVolumeContextCancelled(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	m.On("StoragePoolLookupByName", ctx, defaultPoolName).Return(libvirt.StoragePool{}, context.Canceled)

	err := a.CreateVolume(ctx, "test-vol")
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestWaitTaskContextCancelled(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := a.WaitTask(ctx, "test-task")
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestPingContextCancelled(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	m.On("Connect", ctx).Return(context.Canceled)

	err := a.Ping(ctx)
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}
