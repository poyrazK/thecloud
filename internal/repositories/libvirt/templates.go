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

func generateDomainXML(name, diskPath, networkID, isoPath string, memoryMB, vcpu int, additionalDisks []string, ports []string) string {
	if networkID == "" {
		networkID = "default"
	}
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

	qemuArgsXML := ""
	hasNetworkMapping := false
	for _, p := range ports {
		parts := strings.Split(p, ":")
		if len(parts) == 2 {
			hPort := parts[0]
			cPort := parts[1]
			qemuArgsXML += fmt.Sprintf(`
    <qemu:arg value='-netdev'/>
    <qemu:arg value='user,id=net0,hostfwd=tcp::%s-:%s'/>
    <qemu:arg value='-device'/>
    <qemu:arg value='virtio-net-pci,netdev=net0,bus=pci.0,addr=0x8'/>`, hPort, cPort)
			hasNetworkMapping = true
			break 
		}
	}
	
	qemuArgsXML += fmt.Sprintf(`
    <qemu:arg value='-serial'/>
    <qemu:arg value='file:/tmp/%s-console.log'/>`, name)

	interfaceXML := ""
	if !hasNetworkMapping {
		interfaceXML = fmt.Sprintf(`
    <interface type='network'>
      <source network='%s'/>
      <model type='virtio'/>
    </interface>`, networkID)
	}

	arch := "x86_64"
	machine := "pc"
	if strings.Contains(isoPath, "arm64") || strings.Contains(diskPath, "arm64") {
		arch = "aarch64"
		machine = "virt"
	}

	return fmt.Sprintf(`
<domain type='qemu' xmlns:qemu='http://libvirt.org/schemas/domain/qemu/1.0'>
  <name>%s</name>
  <memory unit='KiB'>%d</memory>
  <vcpu placement='static'>%d</vcpu>
  <os>
    <type arch='%s' machine='%s'>hvm</type>
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
    </disk>%s%s%s
    <graphics type='vnc' port='-1' autoport='yes' listen='0.0.0.0'>
      <listen type='address' address='0.0.0.0'/>
    </graphics>
    <rng model='virtio'>
      <backend model='random'>/dev/urandom</backend>
    </rng>
  </devices>
  <qemu:commandline>%s
  </qemu:commandline>
</domain>`, name, memoryKB, vcpu, arch, machine, diskPath, isoXML, additionalDisksXML, interfaceXML, qemuArgsXML)
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
