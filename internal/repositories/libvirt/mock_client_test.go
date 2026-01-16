package libvirt

import (
	"github.com/digitalocean/go-libvirt"
	"github.com/stretchr/testify/mock"
)

type MockLibvirtClient struct {
	mock.Mock
}

func (m *MockLibvirtClient) Connect() error {
	return m.Called().Error(0)
}

// Domain

func (m *MockLibvirtClient) DomainLookupByName(name string) (libvirt.Domain, error) {
	args := m.Called(name)
	return args.Get(0).(libvirt.Domain), args.Error(1)
}

func (m *MockLibvirtClient) DomainDefineXML(xml string) (libvirt.Domain, error) {
	args := m.Called(xml)
	return args.Get(0).(libvirt.Domain), args.Error(1)
}

func (m *MockLibvirtClient) DomainCreate(dom libvirt.Domain) error {
	return m.Called(dom).Error(0)
}

func (m *MockLibvirtClient) DomainDestroy(dom libvirt.Domain) error {
	return m.Called(dom).Error(0)
}

func (m *MockLibvirtClient) DomainUndefine(dom libvirt.Domain) error {
	return m.Called(dom).Error(0)
}

func (m *MockLibvirtClient) DomainGetState(dom libvirt.Domain, flags uint32) (int32, int32, error) {
	args := m.Called(dom, flags)
	return args.Get(0).(int32), args.Get(1).(int32), args.Error(2)
}

func (m *MockLibvirtClient) DomainGetXMLDesc(dom libvirt.Domain, flags libvirt.DomainXMLFlags) (string, error) {
	args := m.Called(dom, flags)
	return args.String(0), args.Error(1)
}

func (m *MockLibvirtClient) DomainAttachDevice(dom libvirt.Domain, xml string) error {
	return m.Called(dom, xml).Error(0)
}

func (m *MockLibvirtClient) DomainDetachDevice(dom libvirt.Domain, xml string) error {
	return m.Called(dom, xml).Error(0)
}

func (m *MockLibvirtClient) DomainMemoryStats(dom libvirt.Domain, maxStats uint32, flags uint32) ([]libvirt.DomainMemoryStat, error) {
	args := m.Called(dom, maxStats, flags)
	return args.Get(0).([]libvirt.DomainMemoryStat), args.Error(1)
}

// Network

func (m *MockLibvirtClient) NetworkLookupByName(name string) (libvirt.Network, error) {
	args := m.Called(name)
	return args.Get(0).(libvirt.Network), args.Error(1)
}

func (m *MockLibvirtClient) NetworkDefineXML(xml string) (libvirt.Network, error) {
	args := m.Called(xml)
	return args.Get(0).(libvirt.Network), args.Error(1)
}

func (m *MockLibvirtClient) NetworkCreate(net libvirt.Network) error {
	return m.Called(net).Error(0)
}

func (m *MockLibvirtClient) NetworkDestroy(net libvirt.Network) error {
	return m.Called(net).Error(0)
}

func (m *MockLibvirtClient) NetworkUndefine(net libvirt.Network) error {
	return m.Called(net).Error(0)
}

func (m *MockLibvirtClient) NetworkGetDhcpLeases(net libvirt.Network, mac libvirt.OptString, needResults uint32, flags uint32) ([]libvirt.NetworkDhcpLease, uint32, error) {
	args := m.Called(net, mac, needResults, flags)
	return args.Get(0).([]libvirt.NetworkDhcpLease), args.Get(1).(uint32), args.Error(2)
}

// Storage

func (m *MockLibvirtClient) StoragePoolLookupByName(name string) (libvirt.StoragePool, error) {
	args := m.Called(name)
	return args.Get(0).(libvirt.StoragePool), args.Error(1)
}

func (m *MockLibvirtClient) StoragePoolRefresh(pool libvirt.StoragePool, flags uint32) error {
	return m.Called(pool, flags).Error(0)
}

func (m *MockLibvirtClient) StorageVolLookupByName(pool libvirt.StoragePool, name string) (libvirt.StorageVol, error) {
	args := m.Called(pool, name)
	return args.Get(0).(libvirt.StorageVol), args.Error(1)
}

func (m *MockLibvirtClient) StorageVolCreateXML(pool libvirt.StoragePool, xml string, flags uint32) (libvirt.StorageVol, error) {
	args := m.Called(pool, xml, flags)
	return args.Get(0).(libvirt.StorageVol), args.Error(1)
}

func (m *MockLibvirtClient) StorageVolDelete(vol libvirt.StorageVol, flags uint32) error {
	return m.Called(vol, flags).Error(0)
}

func (m *MockLibvirtClient) StorageVolGetPath(vol libvirt.StorageVol) (string, error) {
	args := m.Called(vol)
	return args.String(0), args.Error(1)
}
