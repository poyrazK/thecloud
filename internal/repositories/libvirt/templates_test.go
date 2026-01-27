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
	xml := generateVolumeXML("test-vol", 20)
	assert.Contains(t, xml, "<name>test-vol</name>")
	assert.Contains(t, xml, "<capacity unit=\"G\">20</capacity>")
	assert.Contains(t, xml, "<format type='qcow2'/>")
}

func TestGenerateDomainXML(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		xml := generateDomainXML("test-vm", libvirtTestDiskPath, "default", "", 1024, 2, nil)
		assert.Contains(t, xml, "<name>test-vm</name>")
		assert.Contains(t, xml, "<memory unit='KiB'>1048576</memory>")
		assert.Contains(t, xml, "<vcpu placement='static'>2</vcpu>")
		assert.Contains(t, xml, "<source file='"+libvirtTestDiskPath+"'/>")
		assert.Contains(t, xml, "<source network='default'/>")
	})

	t.Run("default network when empty", func(t *testing.T) {
		xml := generateDomainXML("vm-default", libvirtTestDiskPath, "", "", 512, 1, nil)
		assert.Contains(t, xml, "<source network='default'/>")
	})

	t.Run("with iso and disks", func(t *testing.T) {
		additionalDisks := []string{"/path/to/vdb", "/dev/sdc"}
		xml := generateDomainXML("test-vm", libvirtTestDiskPath, "custom-net", "/path/to/iso", 512, 1, additionalDisks)

		assert.Contains(t, xml, "<source file='/path/to/iso'/>")
		assert.Contains(t, xml, "<target dev='sda' bus='sata'/>")
		assert.Contains(t, xml, "<disk type='file' device='disk'>")
		assert.Contains(t, xml, "<source file='/path/to/vdb'/>")
		assert.Contains(t, xml, "<target dev='vdb' bus='virtio'/>")
		assert.Contains(t, xml, "<disk type='block' device='disk'>")
		assert.Contains(t, xml, "<source dev='/dev/sdc'/>")
		assert.Contains(t, xml, "<target dev='vdc' bus='virtio'/>")
		assert.Contains(t, xml, "<source network='custom-net'/>")
	})
}

func TestGenerateNetworkXML(t *testing.T) {
	xml := generateNetworkXML("test-net", "test-br", testutil.TestIPHost, testutil.TestLibvirtDHCPStart, testutil.TestLibvirtDHCPEnd)
	assert.Contains(t, xml, "<name>test-net</name>")
	assert.Contains(t, xml, "<bridge name='test-br' stp='on' delay='0'/>")
	assert.Contains(t, xml, fmt.Sprintf("<ip address='%s'", testutil.TestIPHost))
	assert.Contains(t, xml, fmt.Sprintf("<range start='%s' end='%s'/>", testutil.TestLibvirtDHCPStart, testutil.TestLibvirtDHCPEnd))
	assert.True(t, strings.Contains(xml, "<network>"))
}
