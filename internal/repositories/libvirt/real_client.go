package libvirt

import (
	"github.com/digitalocean/go-libvirt"
)

// RealLibvirtClient implements LibvirtClient using the actual libvirt library.
type RealLibvirtClient struct {
	conn *libvirt.Libvirt
}

func (r *RealLibvirtClient) Connect() error {
	return r.conn.Connect()
}

// Domain

func (r *RealLibvirtClient) DomainLookupByName(name string) (libvirt.Domain, error) {
	return r.conn.DomainLookupByName(name)
}

func (r *RealLibvirtClient) DomainDefineXML(xml string) (libvirt.Domain, error) {
	return r.conn.DomainDefineXML(xml)
}

func (r *RealLibvirtClient) DomainCreate(dom libvirt.Domain) error {
	return r.conn.DomainCreate(dom)
}

func (r *RealLibvirtClient) DomainDestroy(dom libvirt.Domain) error {
	return r.conn.DomainDestroy(dom)
}

func (r *RealLibvirtClient) DomainUndefine(dom libvirt.Domain) error {
	return r.conn.DomainUndefine(dom)
}

func (r *RealLibvirtClient) DomainGetState(dom libvirt.Domain, flags uint32) (int32, int32, error) {
	return r.conn.DomainGetState(dom, flags)
}

func (r *RealLibvirtClient) DomainGetXMLDesc(dom libvirt.Domain, flags libvirt.DomainXMLFlags) (string, error) {
	return r.conn.DomainGetXMLDesc(dom, flags)
}

func (r *RealLibvirtClient) DomainAttachDevice(dom libvirt.Domain, xml string) error {
	return r.conn.DomainAttachDevice(dom, xml)
}

func (r *RealLibvirtClient) DomainDetachDevice(dom libvirt.Domain, xml string) error {
	return r.conn.DomainDetachDevice(dom, xml)
}

func (r *RealLibvirtClient) DomainMemoryStats(dom libvirt.Domain, maxStats uint32, flags uint32) ([]libvirt.DomainMemoryStat, error) {
	return r.conn.DomainMemoryStats(dom, maxStats, flags)
}

// Network

func (r *RealLibvirtClient) NetworkLookupByName(name string) (libvirt.Network, error) {
	return r.conn.NetworkLookupByName(name)
}

func (r *RealLibvirtClient) NetworkDefineXML(xml string) (libvirt.Network, error) {
	return r.conn.NetworkDefineXML(xml)
}

func (r *RealLibvirtClient) NetworkCreate(net libvirt.Network) error {
	return r.conn.NetworkCreate(net)
}

func (r *RealLibvirtClient) NetworkDestroy(net libvirt.Network) error {
	return r.conn.NetworkDestroy(net)
}

func (r *RealLibvirtClient) NetworkUndefine(net libvirt.Network) error {
	return r.conn.NetworkUndefine(net)
}

func (r *RealLibvirtClient) NetworkGetDhcpLeases(net libvirt.Network, mac libvirt.OptString, needResults uint32, flags uint32) ([]libvirt.NetworkDhcpLease, uint32, error) {
	// The library seems to use different types in different versions or bindings?
	// Based on error: cannot use needResults (variable of type uint32) as int32 value
	// We cast it.
	return r.conn.NetworkGetDhcpLeases(net, mac, int32(needResults), flags)
}

// Storage

func (r *RealLibvirtClient) StoragePoolLookupByName(name string) (libvirt.StoragePool, error) {
	return r.conn.StoragePoolLookupByName(name)
}

func (r *RealLibvirtClient) StoragePoolRefresh(pool libvirt.StoragePool, flags uint32) error {
	return r.conn.StoragePoolRefresh(pool, flags)
}

func (r *RealLibvirtClient) StorageVolLookupByName(pool libvirt.StoragePool, name string) (libvirt.StorageVol, error) {
	return r.conn.StorageVolLookupByName(pool, name)
}

func (r *RealLibvirtClient) StorageVolCreateXML(pool libvirt.StoragePool, xml string, flags uint32) (libvirt.StorageVol, error) {
	return r.conn.StorageVolCreateXML(pool, xml, libvirt.StorageVolCreateFlags(flags))
}

func (r *RealLibvirtClient) StorageVolDelete(vol libvirt.StorageVol, flags uint32) error {
	return r.conn.StorageVolDelete(vol, libvirt.StorageVolDeleteFlags(flags))
}

func (r *RealLibvirtClient) StorageVolGetPath(vol libvirt.StorageVol) (string, error) {
	return r.conn.StorageVolGetPath(vol)
}
