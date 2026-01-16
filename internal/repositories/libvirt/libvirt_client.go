package libvirt

import (
	"github.com/digitalocean/go-libvirt"
)

// LibvirtClient defines the interface for libvirt operations.
// This allows mocking the underlying libvirt connection for testing.
type LibvirtClient interface {
	Connect() error

	// Domain
	DomainLookupByName(name string) (libvirt.Domain, error)
	DomainDefineXML(xml string) (libvirt.Domain, error)
	DomainCreate(dom libvirt.Domain) error
	DomainDestroy(dom libvirt.Domain) error
	DomainUndefine(dom libvirt.Domain) error
	DomainGetState(dom libvirt.Domain, flags uint32) (int32, int32, error)
	DomainGetXMLDesc(dom libvirt.Domain, flags libvirt.DomainXMLFlags) (string, error)
	DomainAttachDevice(dom libvirt.Domain, xml string) error
	DomainDetachDevice(dom libvirt.Domain, xml string) error
	DomainMemoryStats(dom libvirt.Domain, maxStats uint32, flags uint32) ([]libvirt.DomainMemoryStat, error)

	// Network
	NetworkLookupByName(name string) (libvirt.Network, error)
	NetworkDefineXML(xml string) (libvirt.Network, error)
	NetworkCreate(net libvirt.Network) error
	NetworkDestroy(net libvirt.Network) error
	NetworkUndefine(net libvirt.Network) error
	NetworkGetDhcpLeases(net libvirt.Network, mac libvirt.OptString, needResults uint32, flags uint32) ([]libvirt.NetworkDhcpLease, uint32, error)

	// Storage
	StoragePoolLookupByName(name string) (libvirt.StoragePool, error)
	StoragePoolRefresh(pool libvirt.StoragePool, flags uint32) error
	StorageVolLookupByName(pool libvirt.StoragePool, name string) (libvirt.StorageVol, error)
	StorageVolCreateXML(pool libvirt.StoragePool, xml string, flags uint32) (libvirt.StorageVol, error)
	StorageVolDelete(vol libvirt.StorageVol, flags uint32) error
	StorageVolGetPath(vol libvirt.StorageVol) (string, error)
}
