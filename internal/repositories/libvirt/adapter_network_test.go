//go:build integration

package libvirt

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/digitalocean/go-libvirt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testLocalIP     = "192.168.122.100" // Private IP for testing
	invalidPortName = "invalid-port"
	invalidIPName   = "invalid-ip"
)

func TestSetupPortForwardingSuccess(t *testing.T) {
	t.Parallel()
	// Mock execCommand
	oldExec := execCommand
	defer func() { execCommand = oldExec }()
	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("true")
	}
	oldLook := lookPath
	defer func() { lookPath = oldLook }()
	lookPath = func(file string) (string, error) {
		return "/usr/sbin/iptables", nil
	}

	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	a.ipWaitInterval = 1 * time.Microsecond
	dom := libvirt.Domain{Name: testInstanceName}
	netw := libvirt.Network{Name: "default"}
	mac := "52:54:00:12:34:56"
	xml := "<domain><devices><interface type='network'><mac address='" + mac + "'/></interface></devices></domain>"
	ip := testLocalIP

	m.On("DomainLookupByName", mock.Anything, testInstanceName).Return(dom, nil)
	m.On("DomainGetXMLDesc", mock.Anything, dom, libvirt.DomainXMLFlags(0)).Return(xml, nil)
	m.On("NetworkLookupByName", mock.Anything, "default").Return(netw, nil)
	m.On("NetworkGetDhcpLeases", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]libvirt.NetworkDhcpLease{
		{Mac: []string{mac}, Ipaddr: ip},
	}, uint32(1), nil)

	// Call setupPortForwarding directly (synchronously for test)
	a.setupPortForwarding(testInstanceName, []string{"8080:80"})

	a.mu.Lock()
	p, ok := a.portMappings[testInstanceName]["80"]
	a.mu.Unlock()

	assert.True(t, ok)
	assert.Equal(t, 8080, p)
	m.AssertExpectations(t)
}

func TestSetupPortForwardingErrors(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	a.ipWaitInterval = 1 * time.Microsecond

	// 1. IP Wait Failure
	m.On("DomainLookupByName", mock.Anything, "fail").Return(libvirt.Domain{}, libvirt.Error{Code: 1})
	a.setupPortForwarding("fail", []string{"80:80"})

	// 2. Invalid port format
	m.On("DomainLookupByName", mock.Anything, invalidPortName).Return(libvirt.Domain{Name: invalidPortName}, nil)
	m.On("DomainGetXMLDesc", mock.Anything, mock.Anything, mock.Anything).Return("<mac address='00:11:22:33:44:55'/>", nil)
	m.On("NetworkLookupByName", mock.Anything, "default").Return(libvirt.Network{Name: "default"}, nil)
	m.On("NetworkGetDhcpLeases", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]libvirt.NetworkDhcpLease{
		{Mac: []string{"00:11:22:33:44:55"}, Ipaddr: "1.2.3.4"},
	}, uint32(1), nil)

	a.setupPortForwarding(invalidPortName, []string{"invalid"})

	// 3. Invalid IP format
	m.On("DomainLookupByName", mock.Anything, invalidIPName).Return(libvirt.Domain{Name: invalidIPName}, nil)
	m.On("DomainGetXMLDesc", mock.Anything, mock.Anything, mock.Anything).Return("<mac address='00:11:22:33:44:66'/>", nil)
	m.On("NetworkLookupByName", mock.Anything, "default").Return(libvirt.Network{Name: "default"}, nil)
	m.On("NetworkGetDhcpLeases", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]libvirt.NetworkDhcpLease{
		{Mac: []string{"00:11:22:33:44:66"}, Ipaddr: "not-an-ip"},
	}, uint32(1), nil)

	a.setupPortForwarding(invalidIPName, []string{"80:80"})
}

func TestConfigureIptablesNoPath(t *testing.T) {
	t.Parallel()
	// This test depends on environment, but we can verify it doesn't crash
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)

	a.configureIptables("test", "1.2.3.4", "80", 8080, 80)
	// Should log error if iptables not found, but not crash
}

func TestGetInstanceIPErrors(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	ctx := context.Background()

	t.Run("DomainNotFound", func(t *testing.T) {
		t.Parallel()
		m.On("DomainLookupByName", ctx, "missing").Return(libvirt.Domain{}, libvirt.Error{Code: 1})
		_, err := a.GetInstanceIP(ctx, "missing")
		assert.Error(t, err)
	})

	t.Run("NoMAC", func(t *testing.T) {
		t.Parallel()
		dom := libvirt.Domain{Name: "nomac"}
		m.On("DomainLookupByName", ctx, "nomac").Return(dom, nil)
		m.On("DomainGetXMLDesc", ctx, dom, libvirt.DomainXMLFlags(0)).Return("<domain></domain>", nil)
		_, err := a.GetInstanceIP(ctx, "nomac")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no mac address")
	})
}
