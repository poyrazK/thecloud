package libvirt

import (
	"context"
	"log/slog"
	"net"
	"os"
	"testing"

	"github.com/digitalocean/go-libvirt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupLibvirtAdapterTest() (*LibvirtAdapter, *mockLibvirtClient) {
	client := new(mockLibvirtClient)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	adapter := &LibvirtAdapter{
		client:       client,
		logger:       logger,
		poolStart:    net.ParseIP("192.168.100.0"),
		poolEnd:      net.ParseIP("192.168.200.255"),
		portMappings: make(map[string]map[string]int),
	}
	return adapter, client
}

func TestLibvirtAdapter_InstanceLifecycle(t *testing.T) {
	ctx := context.Background()
	adapter, client := setupLibvirtAdapterTest()

	t.Run("StartInstance_Success", func(t *testing.T) {
		id := "test-vm"
		dom := libvirt.Domain{Name: id}
		client.On("DomainLookupByName", ctx, id).Return(dom, nil)
		client.On("DomainCreate", ctx, dom).Return(nil)

		err := adapter.StartInstance(ctx, id)
		assert.NoError(t, err)
	})

	t.Run("StartInstance_NotFound", func(t *testing.T) {
		id := "missing-vm"
		client.On("DomainLookupByName", ctx, id).Return(libvirt.Domain{}, os.ErrNotExist)

		err := adapter.StartInstance(ctx, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "domain not found")
	})

	t.Run("StopInstance_Success", func(t *testing.T) {
		id := "test-vm"
		dom := libvirt.Domain{Name: id}
		client.On("DomainLookupByName", ctx, id).Return(dom, nil)
		client.On("DomainDestroy", ctx, dom).Return(nil)

		err := adapter.StopInstance(ctx, id)
		assert.NoError(t, err)
	})

	t.Run("DeleteInstance_Success", func(t *testing.T) {
		id := "test-vm"
		dom := libvirt.Domain{Name: id}
		client.On("DomainLookupByName", ctx, id).Return(dom, nil)
		client.On("DomainGetState", ctx, dom, uint32(0)).Return(int32(1), int32(0), nil) // Running
		client.On("DomainDestroy", ctx, dom).Return(nil)
		client.On("DomainUndefine", ctx, dom).Return(nil)

		// Mock cleanupRootVolume
		pool := libvirt.StoragePool{Name: defaultPoolName}
		vol := libvirt.StorageVol{Name: id + "-root"}
		client.On("StoragePoolLookupByName", ctx, defaultPoolName).Return(pool, nil)
		client.On("StorageVolLookupByName", ctx, pool, id+"-root").Return(vol, nil)
		client.On("StorageVolDelete", ctx, vol, uint32(0)).Return(nil)

		err := adapter.DeleteInstance(ctx, id)
		assert.NoError(t, err)
	})
}

func TestLibvirtAdapter_SanitizeDomainName(t *testing.T) {
	adapter, _ := setupLibvirtAdapterTest()

	tests := []struct {
		input    string
		expected string
	}{
		{"safe-name", "safe-name"},
		{"Unsafe!Name@123", "UnsafeName123"},
		{"multiple---dashes", "multiple---dashes"},
		{"", ""}, // Should return partial UUID which we'll check with length
	}

	for _, tt := range tests {
		result := adapter.sanitizeDomainName(tt.input)
		if tt.expected == "" {
			assert.Len(t, result, 8)
		} else {
			assert.Equal(t, tt.expected, result)
		}
	}
}

func TestLibvirtAdapter_Networking(t *testing.T) {
	ctx := context.Background()
	adapter, client := setupLibvirtAdapterTest()

	t.Run("GetInstanceIP_Success", func(t *testing.T) {
		id := "test-vm"
		mac := "52:54:00:12:34:56"
		dom := libvirt.Domain{Name: id}
		xml := `<domain><devices><interface type='network'><mac address='` + mac + `'/></interface></devices></domain>`
		net := libvirt.Network{Name: "default"}
		leases := []libvirt.NetworkDhcpLease{
			{Mac: []string{mac}, Ipaddr: "192.168.122.10"},
		}

		client.On("DomainLookupByName", ctx, id).Return(dom, nil)
		client.On("DomainGetXMLDesc", ctx, dom, libvirt.DomainXMLFlags(0)).Return(xml, nil)
		client.On("NetworkLookupByName", ctx, "default").Return(net, nil)
		client.On("NetworkGetDhcpLeases", ctx, net, mock.Anything, uint32(0), uint32(0)).Return(leases, uint32(1), nil)

		ip, err := adapter.GetInstanceIP(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, "192.168.122.10", ip)
	})

	t.Run("CreateNetwork_Success", func(t *testing.T) {
		name := "test-net"
		net := libvirt.Network{Name: name}
		client.On("NetworkDefineXML", ctx, mock.Anything).Return(net, nil)
		client.On("NetworkCreate", ctx, net).Return(nil)

		id, err := adapter.CreateNetwork(ctx, name)
		assert.NoError(t, err)
		assert.Equal(t, name, id)
	})
}

func TestLibvirtAdapter_Storage(t *testing.T) {
	ctx := context.Background()
	adapter, client := setupLibvirtAdapterTest()

	t.Run("CreateVolume_Success", func(t *testing.T) {
		name := "test-vol"
		pool := libvirt.StoragePool{Name: defaultPoolName}
		vol := libvirt.StorageVol{Name: name}

		client.On("StoragePoolLookupByName", ctx, defaultPoolName).Return(pool, nil)
		client.On("StoragePoolRefresh", ctx, pool, uint32(0)).Return(nil)
		client.On("StorageVolCreateXML", ctx, pool, mock.Anything, uint32(0)).Return(vol, nil)

		err := adapter.CreateVolume(ctx, name)
		assert.NoError(t, err)
	})

	t.Run("AttachVolume_Success", func(t *testing.T) {
		id := "test-vm"
		volPath := "/var/lib/libvirt/images/extra.qcow2"
		dom := libvirt.Domain{Name: id}

		client.On("DomainLookupByName", ctx, id).Return(dom, nil)
		client.On("DomainAttachDevice", ctx, dom, mock.Anything).Return(nil)

		err := adapter.AttachVolume(ctx, id, volPath)
		assert.NoError(t, err)
	})
}

func TestLibvirtAdapter_StatsAndLogs(t *testing.T) {
	ctx := context.Background()
	adapter, client := setupLibvirtAdapterTest()

	t.Run("GetInstanceStats_Success", func(t *testing.T) {
		id := "test-vm"
		dom := libvirt.Domain{Name: id}
		memStats := []libvirt.DomainMemoryStat{
			{Tag: 6, Val: 1024}, // RSS
			{Tag: 5, Val: 2048}, // Actual
		}

		client.On("DomainLookupByName", ctx, id).Return(dom, nil)
		client.On("DomainMemoryStats", ctx, dom, uint32(10), uint32(0)).Return(memStats, nil)

		r, err := adapter.GetInstanceStats(ctx, id)
		assert.NoError(t, err)
		assert.NotNil(t, r)
	})

	t.Run("GetConsoleURL_Success", func(t *testing.T) {
		id := "test-vm"
		dom := libvirt.Domain{Name: id}
		xml := `<domain><devices><graphics type='vnc' port='5900'/></devices></domain>`

		client.On("DomainLookupByName", ctx, id).Return(dom, nil)
		client.On("DomainGetXMLDesc", ctx, dom, libvirt.DomainXMLFlags(0)).Return(xml, nil)

		url, err := adapter.GetConsoleURL(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, "vnc://127.0.0.1:5900", url)
	})
}
