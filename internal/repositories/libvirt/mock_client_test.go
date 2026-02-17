package libvirt

import (
	"context"

	"github.com/digitalocean/go-libvirt"
	"github.com/stretchr/testify/mock"
)

type MockLibvirtClient struct {
	mock.Mock
}

func (m *MockLibvirtClient) Connect(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func (m *MockLibvirtClient) ConnectToURI(ctx context.Context, uri string) error {
	return m.Called(ctx, uri).Error(0)
}

func (m *MockLibvirtClient) Close() error {
	return m.Called().Error(0)
}

// Domain

func (m *MockLibvirtClient) DomainLookupByName(ctx context.Context, name string) (libvirt.Domain, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(libvirt.Domain), args.Error(1)
}

func (m *MockLibvirtClient) DomainDefineXML(ctx context.Context, xml string) (libvirt.Domain, error) {
	args := m.Called(ctx, xml)
	return args.Get(0).(libvirt.Domain), args.Error(1)
}

func (m *MockLibvirtClient) DomainCreate(ctx context.Context, dom libvirt.Domain) error {
	return m.Called(ctx, dom).Error(0)
}

func (m *MockLibvirtClient) DomainDestroy(ctx context.Context, dom libvirt.Domain) error {
	return m.Called(ctx, dom).Error(0)
}

func (m *MockLibvirtClient) DomainUndefine(ctx context.Context, dom libvirt.Domain) error {
	return m.Called(ctx, dom).Error(0)
}

func (m *MockLibvirtClient) DomainGetState(ctx context.Context, dom libvirt.Domain, flags uint32) (int32, int32, error) {
	args := m.Called(ctx, dom, flags)
	return args.Get(0).(int32), args.Get(1).(int32), args.Error(2)
}

func (m *MockLibvirtClient) DomainGetXMLDesc(ctx context.Context, dom libvirt.Domain, flags libvirt.DomainXMLFlags) (string, error) {
	args := m.Called(ctx, dom, flags)
	return args.String(0), args.Error(1)
}

func (m *MockLibvirtClient) DomainAttachDevice(ctx context.Context, dom libvirt.Domain, xml string) error {
	return m.Called(ctx, dom, xml).Error(0)
}

func (m *MockLibvirtClient) DomainDetachDevice(ctx context.Context, dom libvirt.Domain, xml string) error {
	return m.Called(ctx, dom, xml).Error(0)
}

func (m *MockLibvirtClient) DomainMemoryStats(ctx context.Context, dom libvirt.Domain, maxStats uint32, flags uint32) ([]libvirt.DomainMemoryStat, error) {
	args := m.Called(ctx, dom, maxStats, flags)
	return args.Get(0).([]libvirt.DomainMemoryStat), args.Error(1)
}

// Network

func (m *MockLibvirtClient) NetworkLookupByName(ctx context.Context, name string) (libvirt.Network, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(libvirt.Network), args.Error(1)
}

func (m *MockLibvirtClient) NetworkDefineXML(ctx context.Context, xml string) (libvirt.Network, error) {
	args := m.Called(ctx, xml)
	return args.Get(0).(libvirt.Network), args.Error(1)
}

func (m *MockLibvirtClient) NetworkCreate(ctx context.Context, net libvirt.Network) error {
	return m.Called(ctx, net).Error(0)
}

func (m *MockLibvirtClient) NetworkDestroy(ctx context.Context, net libvirt.Network) error {
	return m.Called(ctx, net).Error(0)
}

func (m *MockLibvirtClient) NetworkUndefine(ctx context.Context, net libvirt.Network) error {
	return m.Called(ctx, net).Error(0)
}

func (m *MockLibvirtClient) NetworkGetDhcpLeases(ctx context.Context, net libvirt.Network, mac libvirt.OptString, needResults uint32, flags uint32) ([]libvirt.NetworkDhcpLease, uint32, error) {
	args := m.Called(ctx, net, mac, needResults, flags)
	return args.Get(0).([]libvirt.NetworkDhcpLease), args.Get(1).(uint32), args.Error(2)
}

// Storage

func (m *MockLibvirtClient) StoragePoolLookupByName(ctx context.Context, name string) (libvirt.StoragePool, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(libvirt.StoragePool), args.Error(1)
}

func (m *MockLibvirtClient) StoragePoolRefresh(ctx context.Context, pool libvirt.StoragePool, flags uint32) error {
	return m.Called(ctx, pool, flags).Error(0)
}

func (m *MockLibvirtClient) StorageVolLookupByName(ctx context.Context, pool libvirt.StoragePool, name string) (libvirt.StorageVol, error) {
	args := m.Called(ctx, pool, name)
	return args.Get(0).(libvirt.StorageVol), args.Error(1)
}

func (m *MockLibvirtClient) StorageVolCreateXML(ctx context.Context, pool libvirt.StoragePool, xml string, flags uint32) (libvirt.StorageVol, error) {
	args := m.Called(ctx, pool, xml, flags)
	return args.Get(0).(libvirt.StorageVol), args.Error(1)
}

func (m *MockLibvirtClient) StorageVolDelete(ctx context.Context, vol libvirt.StorageVol, flags uint32) error {
	return m.Called(ctx, vol, flags).Error(0)
}

func (m *MockLibvirtClient) StorageVolGetPath(ctx context.Context, vol libvirt.StorageVol) (string, error) {
	args := m.Called(ctx, vol)
	return args.String(0), args.Error(1)
}
