//go:build integration

package libvirt

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/digitalocean/go-libvirt"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testInstanceName = "test-instance"
	testIP           = "192.168.122.10"
	testTaskID       = "task-id"
)

func newTestAdapter(m *MockLibvirtClient) *LibvirtAdapter {
	return &LibvirtAdapter{
		client:           m,
		logger:           slog.New(slog.NewTextHandler(io.Discard, nil)),
		uri:              "qemu:///system",
		ipWaitInterval:   1 * time.Millisecond,
		taskWaitInterval: 1 * time.Millisecond,
		poolStart:        net.ParseIP("192.168.100.0"),
		poolEnd:          net.ParseIP("192.168.200.255"),
		portMappings:     make(map[string]map[string]int),
	}
}

func TestStopInstanceSuccess(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	ctx := context.Background()

	dom := libvirt.Domain{Name: testInstanceName, ID: 1, UUID: [16]byte{1}}

	m.On("DomainLookupByName", mock.Anything, testInstanceName).Return(dom, nil)
	m.On("DomainDestroy", mock.Anything, dom).Return(nil)

	err := a.StopInstance(ctx, testInstanceName)
	assert.NoError(t, err)
	m.AssertExpectations(t)
}

func TestStopInstanceNotFound(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	ctx := context.Background()

	m.On("DomainLookupByName", mock.Anything, testInstanceName).Return(libvirt.Domain{}, errors.New("domain not found"))

	err := a.StopInstance(ctx, testInstanceName)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	m.AssertExpectations(t)
}

func TestDeleteInstanceSuccess(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	ctx := context.Background()

	dom := libvirt.Domain{Name: testInstanceName, ID: 1, UUID: [16]byte{1}}
	pool := libvirt.StoragePool{Name: "default"}
	vol := libvirt.StorageVol{Name: testInstanceName + "-root"}

	// 1. Lookup
	m.On("ConnectToURI", mock.Anything, mock.Anything).Return(nil)
	m.On("DomainLookupByName", mock.Anything, testInstanceName).Return(dom, nil)

	// 2. Stop check (Running = 1)
	m.On("DomainGetState", mock.Anything, dom, uint32(0)).Return(int32(1), int32(0), nil)
	m.On("DomainDestroy", mock.Anything, dom).Return(nil)

	// 3. Undefine
	m.On("DomainUndefine", mock.Anything, dom).Return(nil)

	// 4. Cleanup Root Volume
	m.On("StoragePoolLookupByName", mock.Anything, "default").Return(pool, nil)
	m.On("StorageVolLookupByName", mock.Anything, pool, testInstanceName+"-root").Return(vol, nil)
	m.On("StorageVolGetPath", mock.Anything, vol).Return("/path/to/disk", nil)
	m.On("StorageVolDelete", mock.Anything, vol, uint32(0)).Return(nil)

	err := a.DeleteInstance(ctx, testInstanceName)
	assert.NoError(t, err)
	m.AssertExpectations(t)
}

func TestGetInstanceIPSuccess(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	ctx := context.Background()

	dom := libvirt.Domain{Name: testInstanceName, ID: 1}
	net := libvirt.Network{Name: "default"}

	// Mock XML containing MAC
	xmlDesc := `<domain><devices><interface type='network'><mac address='52:54:00:11:22:33'/></interface></devices></domain>`

	m.On("DomainLookupByName", mock.Anything, testInstanceName).Return(dom, nil)
	m.On("DomainGetXMLDesc", mock.Anything, dom, libvirt.DomainXMLFlags(0)).Return(xmlDesc, nil)
	m.On("NetworkLookupByName", mock.Anything, "default").Return(net, nil)

	leases := []libvirt.NetworkDhcpLease{
		{Mac: []string{"52:54:00:11:22:33"}, Ipaddr: testIP},
	}
	m.On("NetworkGetDhcpLeases", mock.Anything, net, mock.Anything, uint32(0), uint32(0)).Return(leases, uint32(1), nil)

	ip, err := a.GetInstanceIP(ctx, testInstanceName)
	assert.NoError(t, err)
	assert.Equal(t, testIP, ip)
	m.AssertExpectations(t)
}

func TestWaitTaskTimeout(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	a.taskWaitInterval = 1 * time.Millisecond
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	dom := libvirt.Domain{Name: testTaskID}
	m.On("DomainLookupByName", mock.Anything, testTaskID).Return(dom, nil)
	// Return a state that is NOT Shutoff
	m.On("DomainGetState", mock.Anything, dom, uint32(0)).Return(int32(libvirt.DomainRunning), int32(1), nil)

	status, err := a.WaitTask(ctx, testTaskID)
	assert.Error(t, err)
	assert.Equal(t, int64(-1), status)
	assert.True(t, errors.Is(err, context.DeadlineExceeded) || err == context.DeadlineExceeded)
}

func TestWaitInitialIPSuccess(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	ctx := context.Background()

	// 4. Get IP (waitInitialIP)
	dom := libvirt.Domain{Name: testInstanceName, ID: 1}
	m.On("DomainLookupByName", mock.Anything, testInstanceName).Return(dom, nil)

	// Get XML for MAC
	xmlDesc := `<domain><devices><interface type='network'><mac address='52:54:00:11:22:33'/></interface></devices></domain>`
	m.On("DomainGetXMLDesc", mock.Anything, dom, libvirt.DomainXMLFlags(0)).Return(xmlDesc, nil)

	// Get Network
	net := libvirt.Network{Name: "default"}
	m.On("NetworkLookupByName", mock.Anything, "default").Return(net, nil)

	// Get Leases
	leases := []libvirt.NetworkDhcpLease{
		{Mac: []string{"52:54:00:11:22:33"}, Ipaddr: testIP},
	}
	m.On("NetworkGetDhcpLeases", mock.Anything, net, mock.Anything, uint32(0), uint32(0)).Return(leases, uint32(1), nil)

	// Execute
	ip, err := a.waitInitialIP(ctx, testInstanceName)
	assert.NoError(t, err)
	assert.Equal(t, testIP, ip)

	m.AssertExpectations(t)
}

func TestAttachVolumeSuccess(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	ctx := context.Background()

	dom := libvirt.Domain{Name: testInstanceName, ID: 1, UUID: [16]byte{1}}
	volumePath := "/var/lib/libvirt/images/vol1.qcow2"

	m.On("DomainLookupByName", mock.Anything, testInstanceName).Return(dom, nil)
	// We expect DomainAttachDevice with an XML for the disk
	m.On("DomainAttachDevice", mock.Anything, dom, mock.AnythingOfType("string")).Return(nil)

	err := a.AttachVolume(ctx, testInstanceName, volumePath)
	assert.NoError(t, err)
	m.AssertExpectations(t)
}

func TestDetachVolumeSuccess(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	ctx := context.Background()

	dom := libvirt.Domain{Name: testInstanceName, ID: 1, UUID: [16]byte{1}}
	volumePath := "/var/lib/libvirt/images/vol1.qcow2"

	m.On("DomainLookupByName", mock.Anything, testInstanceName).Return(dom, nil)
	m.On("DomainDetachDevice", mock.Anything, dom, mock.AnythingOfType("string")).Return(nil)

	err := a.DetachVolume(ctx, testInstanceName, volumePath)
	assert.NoError(t, err)

	m.AssertExpectations(t)
}

func TestGetConsoleURLSuccess(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	ctx := context.Background()

	dom := libvirt.Domain{Name: testInstanceName, ID: 1, UUID: [16]byte{1}}
	xmlDesc := "<domain><devices><graphics type='vnc' port='5900'/></devices></domain>"

	m.On("DomainLookupByName", mock.Anything, testInstanceName).Return(dom, nil)
	m.On("DomainGetXMLDesc", mock.Anything, dom, libvirt.DomainXMLFlags(0)).Return(xmlDesc, nil)

	url, err := a.GetConsoleURL(ctx, testInstanceName)
	assert.NoError(t, err)
	assert.Equal(t, "vnc://127.0.0.1:5900", url)
	m.AssertExpectations(t)
}

func TestGetInstanceStatsSuccess(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	ctx := context.Background()

	dom := libvirt.Domain{Name: testInstanceName, ID: 1, UUID: [16]byte{1}}
	stats := []libvirt.DomainMemoryStat{
		{Tag: 6, Val: 1048576}, // rss
	}

	m.On("DomainLookupByName", mock.Anything, testInstanceName).Return(dom, nil)
	m.On("DomainMemoryStats", mock.Anything, dom, uint32(10), uint32(0)).Return(stats, nil)

	rc, err := a.GetInstanceStats(ctx, testInstanceName)
	assert.NoError(t, err)
	assert.NotNil(t, rc)
	_ = rc.Close()
	m.AssertExpectations(t)
}

func TestCreateInstanceSuccess(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	ctx := context.Background()

	pool := libvirt.StoragePool{Name: "default"}
	vol := libvirt.StorageVol{Name: testInstanceName + "-root"}
	dom := libvirt.Domain{Name: testInstanceName}

	m.On("StoragePoolLookupByName", mock.Anything, "default").Return(pool, nil)
	m.On("StorageVolLookupByName", mock.Anything, pool, "ubuntu").Return(libvirt.StorageVol{}, errors.New("not found"))
	m.On("StorageVolCreateXML", mock.Anything, pool, mock.Anything, uint32(0)).Return(vol, nil)
	m.On("StorageVolGetPath", mock.Anything, vol).Return("/path/to/disk", nil)
	m.On("DomainDefineXML", mock.Anything, mock.Anything).Return(dom, nil)
	m.On("DomainCreate", mock.Anything, dom).Return(nil)

	opts := ports.CreateInstanceOptions{
		Name:      testInstanceName,
		ImageName: "ubuntu",
	}

	id, _, err := a.LaunchInstanceWithOptions(ctx, opts)
	assert.NoError(t, err)
	assert.Equal(t, testInstanceName, id)
	m.AssertExpectations(t)
}

func TestRunTaskSuccess(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	ctx := context.Background()

	pool := libvirt.StoragePool{Name: "default"}
	vol := libvirt.StorageVol{Name: "task-root"}
	dom := libvirt.Domain{Name: "task-instance"}

	m.On("StoragePoolLookupByName", mock.Anything, "default").Return(pool, nil)
	m.On("StorageVolLookupByName", mock.Anything, pool, "ubuntu").Return(libvirt.StorageVol{}, errors.New("not found"))
	m.On("StorageVolCreateXML", mock.Anything, pool, mock.Anything, uint32(0)).Return(vol, nil)
	m.On("StorageVolGetPath", mock.Anything, vol).Return("/path/to/task-disk", nil)
	m.On("DomainDefineXML", mock.Anything, mock.Anything).Return(dom, nil)
	m.On("DomainCreate", mock.Anything, dom).Return(nil)

	opts := ports.RunTaskOptions{
		Image:    "ubuntu",
		Command:  []string{"echo", "hello"},
		MemoryMB: 512,
	}

	id, _, err := a.RunTask(ctx, opts)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)
	m.AssertExpectations(t)
}

func TestCreateNetworkSuccess(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	ctx := context.Background()

	netName := "test-net"
	mockNet := libvirt.Network{Name: netName}

	m.On("NetworkDefineXML", mock.Anything, mock.Anything).Return(mockNet, nil)
	m.On("NetworkCreate", mock.Anything, mockNet).Return(nil)

	id, err := a.CreateNetwork(ctx, netName)
	assert.NoError(t, err)
	assert.Equal(t, netName, id)
	m.AssertExpectations(t)
}

func TestDeleteNetworkSuccess(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	ctx := context.Background()

	netName := "test-net"
	mockNet := libvirt.Network{Name: netName}

	m.On("NetworkLookupByName", mock.Anything, netName).Return(mockNet, nil)
	m.On("NetworkDestroy", mock.Anything, mockNet).Return(nil)
	m.On("NetworkUndefine", mock.Anything, mockNet).Return(nil)

	err := a.DeleteNetwork(ctx, netName)
	assert.NoError(t, err)
	m.AssertExpectations(t)
}

func TestWaitTaskSuccess(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	ctx := context.Background()

	dom := libvirt.Domain{Name: testTaskID}
	m.On("DomainLookupByName", mock.Anything, testTaskID).Return(dom, nil)
	m.On("DomainGetState", mock.Anything, dom, uint32(0)).Return(int32(libvirt.DomainShutoff), int32(1), nil)

	status, err := a.WaitTask(ctx, testTaskID)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), status)
	m.AssertExpectations(t)
}

func TestWaitTaskDomainNotFound(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	ctx := context.Background()

	m.On("DomainLookupByName", mock.Anything, testTaskID).Return(libvirt.Domain{}, errors.New("not found"))

	status, err := a.WaitTask(ctx, testTaskID)
	assert.Error(t, err)
	assert.Equal(t, int64(-1), status)
	m.AssertExpectations(t)
}

func TestRunTaskFailures(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	ctx := context.Background()
	pool := libvirt.StoragePool{Name: "default"}

	opts := ports.RunTaskOptions{
		Image:   "ubuntu",
		Command: []string{"echo"},
	}

	// 1. Pool lookup failure
	m.On("StoragePoolLookupByName", ctx, "default").Return(libvirt.StoragePool{}, libvirt.Error{Code: 1}).Once()
	_, _, err := a.RunTask(ctx, opts)
	assert.Error(t, err)

	// 2. Vol creation failure
	m.On("StoragePoolLookupByName", ctx, "default").Return(pool, nil).Once()
	m.On("StorageVolCreateXML", ctx, pool, mock.Anything, uint32(0)).Return(libvirt.StorageVol{}, libvirt.Error{Code: 1}).Once()
	_, _, err = a.RunTask(ctx, opts)
	assert.Error(t, err)
}
