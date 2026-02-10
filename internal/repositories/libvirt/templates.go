// Package libvirt provides XML template generation for libvirt domain definitions.
package libvirt

import (
	"fmt"
	"strings"
)

func generateVolumeXML(name string, sizeGB int, backingStorePath string) string {
	backingXML := ""
	if backingStorePath != "" {
		backingXML = fmt.Sprintf(`
  <backingStore>
    <path>%s</path>
    <format type='qcow2'/>
  </backingStore>`, backingStorePath)
	}

	return fmt.Sprintf(`
<volume>
  <name>%s</name>
  <capacity unit="G">%d</capacity>
  <target>
    <format type='qcow2'/>
  </target>%s
</volume>`, name, sizeGB, backingXML)
}

func generateDomainXML(name, diskPath, networkID, isoPath string, memoryMB, vcpu int, additionalDisks []string) string {
	if networkID == "" {
		networkID = "default"
	}
	// Convert MB to KB for libvirt
	memoryKB := memoryMB * 1024

	var isoXML string
	if isoPath != "" {
		isoXML = fmt.Sprintf(`
    <disk type='file' device='cdrom'>
      <driver name='qemu' type='raw'/>
      <source file='%s'/>
      <target dev='hda' bus='ide'/>
      <readonly/>
    </disk>`, isoPath)
	}

	additionalDisksXML := ""
	for i, dPath := range additionalDisks {
		dev := fmt.Sprintf("vd%c", 'b'+i)
		diskType := "file"
		driverType := "qcow2"
		sourceAttr := "file"

		if strings.HasPrefix(dPath, "/dev/") {
			diskType = "block"
			driverType = "raw"
			sourceAttr = "dev"
		}

		additionalDisksXML += fmt.Sprintf(`
    <disk type='%s' device='disk'>
      <driver name='qemu' type='%s'/>
      <source %s='%s'/>
      <target dev='%s' bus='virtio'/>
    </disk>`, diskType, driverType, sourceAttr, dPath, dev)
	}

	return fmt.Sprintf(`
<domain type='kvm'>
  <name>%s</name>
  <memory unit='KiB'>%d</memory>
  <vcpu placement='static'>%d</vcpu>
  <os>
    <type arch='x86_64' machine='pc'>hvm</type>
    <boot dev='hd'/>
  </os>
  <features>
    <acpi/>
    <apic/>
  </features>
  <devices>
    <disk type='file' device='disk'>
      <driver name='qemu' type='qcow2'/>
      <source file='%s'/>
      <target dev='vda' bus='virtio'/>
    </disk>%s%s
    <interface type='network'>
      <source network='%s'/>
      <model type='virtio'/>
    </interface>
    <graphics type='vnc' port='-1' autoport='yes' listen='0.0.0.0'>
      <listen type='address' address='0.0.0.0'/>
    </graphics>
    <serial type='pty'>
      <target port='0'/>
    </serial>
    <console type='pty'>
      <target type='serial' port='0'/>
    </console>
    <rng model='virtio'>
      <backend model='random'>/dev/urandom</backend>
    </rng>
  </devices>
</domain>`, name, memoryKB, vcpu, diskPath, isoXML, additionalDisksXML, networkID)
}

func generateNetworkXML(name, bridgeName, gatewayIP, rangeStart, rangeEnd string) string {
	return fmt.Sprintf(`
<network>
  <name>%s</name>
  <forward mode='nat'/>
  <bridge name='%s' stp='on' delay='0'/>
  <ip address='%s' netmask='255.255.255.0'>
    <dhcp>
      <range start='%s' end='%s'/>
    </dhcp>
  </ip>
</network>`, name, bridgeName, gatewayIP, rangeStart, rangeEnd)
}
