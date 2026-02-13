//go:build integration

package libvirt

import (
	"fmt"
	"strings"
	"testing"

	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/assert"
)

const libvirtTestDiskPath = "/path/to/disk"

func TestGenerateVolumeXML(t *testing.T) {
	t.Parallel()
	xml := generateVolumeXML("test-vol", 20, "")
	assert.Contains(t, xml, "<name>test-vol</name>")
	assert.Contains(t, xml, "<capacity unit=\"G\">20</capacity>")
	assert.Contains(t, xml, "<format type='qcow2'/>")
}

func TestGenerateDomainXML(t *testing.T) {
	t.Parallel()
	t.Run("basic", func(t *testing.T) {
		t.Parallel()
		xml := generateDomainXML("test-vm", libvirtTestDiskPath, "default", "", 1024, 2, nil, nil, "")
		assert.Contains(t, xml, "<name>test-vm</name>")
		assert.Contains(t, xml, "<memory unit='KiB'>1048576</memory>")
		assert.Contains(t, xml, "<vcpu placement='static'>2</vcpu>")
		assert.Contains(t, xml, "<source file='"+libvirtTestDiskPath+"'/>")
		assert.Contains(t, xml, "<type arch='x86_64' machine='pc'>hvm</type>")
	})

	t.Run("default network when empty", func(t *testing.T) {
		t.Parallel()
		xml := generateDomainXML("vm-default", libvirtTestDiskPath, "", "", 512, 1, nil, nil, "")
		assert.Contains(t, xml, "<interface type='network'>")
		assert.Contains(t, xml, "<source network='default'/>")
	})

	t.Run("with iso and disks", func(t *testing.T) {
		t.Parallel()
		additionalDisks := []string{"/path/to/vdb", "/dev/sdc"}
		xml := generateDomainXML("test-vm", libvirtTestDiskPath, "custom-net", "/path/to/iso", 512, 1, additionalDisks, nil, "")

		assert.Contains(t, xml, "<source file='/path/to/iso'/>")
		assert.Contains(t, xml, "<target dev='hda' bus='ide'/>")
		assert.Contains(t, xml, "<disk type='file' device='disk'>")
		assert.Contains(t, xml, "<source file='/path/to/vdb'/>")
		assert.Contains(t, xml, "<target dev='vdb' bus='virtio'/>")
		assert.Contains(t, xml, "<disk type='block' device='disk'>")
		assert.Contains(t, xml, "<source dev='/dev/sdc'/>")
		assert.Contains(t, xml, "<target dev='vdc' bus='virtio'/>")
	})
	
	t.Run("with port mapping", func(t *testing.T) {
		t.Parallel()
		ports := []string{"8080:80", "443:443"}
		xml := generateDomainXML("test-port-vm", libvirtTestDiskPath, "default", "", 512, 1, nil, ports, "")
		assert.Contains(t, xml, "hostfwd=tcp::8080-:80")
		assert.Contains(t, xml, "hostfwd=tcp::443-:443")
		assert.Contains(t, xml, "virtio-net-pci,netdev=net0,bus=pci.0,addr=0x8")
		// interface XML should be empty when port mapping is used (SLIRP)
		assert.NotContains(t, xml, "<interface type='network'>")
	})

	t.Run("arch auto-detection aarch64", func(t *testing.T) {
		t.Parallel()
		xml := generateDomainXML("test-arm", "ubuntu-arm64.qcow2", "default", "", 1024, 2, nil, nil, "")
		assert.Contains(t, xml, "<type arch='aarch64' machine='virt'>hvm</type>")
	})

	t.Run("explicit arch", func(t *testing.T) {
		t.Parallel()
		xml := generateDomainXML("test-arch", "disk.qcow2", "default", "", 1024, 2, nil, nil, "aarch64")
		assert.Contains(t, xml, "<type arch='aarch64' machine='virt'>hvm</type>")
	})
}

func TestGenerateNetworkXML(t *testing.T) {
	t.Parallel()
	xml := generateNetworkXML("test-net", "test-br", testutil.TestIPHost, testutil.TestLibvirtDHCPStart, testutil.TestLibvirtDHCPEnd)
	assert.Contains(t, xml, "<name>test-net</name>")
	assert.Contains(t, xml, "<bridge name='test-br' stp='on' delay='0'/>")
	assert.Contains(t, xml, fmt.Sprintf("<ip address='%s'", testutil.TestIPHost))
	assert.Contains(t, xml, fmt.Sprintf("<range start='%s' end='%s'/>", testutil.TestLibvirtDHCPStart, testutil.TestLibvirtDHCPEnd))
	assert.True(t, strings.Contains(xml, "<network>"))
}
