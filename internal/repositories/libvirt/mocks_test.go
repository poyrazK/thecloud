package libvirt

import (
	"context"

	"github.com/digitalocean/go-libvirt"
	"github.com/stretchr/testify/mock"
)

type mockLibvirtClient struct {
	mock.Mock
}

func (m *mockLibvirtClient) Connect(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockLibvirtClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockLibvirtClient) DomainLookupByName(ctx context.Context, name string) (libvirt.Domain, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(libvirt.Domain), args.Error(1)
}

func (m *mockLibvirtClient) DomainDefineXML(ctx context.Context, xml string) (libvirt.Domain, error) {
	args := m.Called(ctx, xml)
	return args.Get(0).(libvirt.Domain), args.Error(1)
}

func (m *mockLibvirtClient) DomainCreate(ctx context.Context, dom libvirt.Domain) error {
	args := m.Called(ctx, dom)
	return args.Error(0)
}

func (m *mockLibvirtClient) DomainDestroy(ctx context.Context, dom libvirt.Domain) error {
	args := m.Called(ctx, dom)
	return args.Error(0)
}

func (m *mockLibvirtClient) DomainUndefine(ctx context.Context, dom libvirt.Domain) error {
	args := m.Called(ctx, dom)
	return args.Error(0)
}

func (m *mockLibvirtClient) DomainGetState(ctx context.Context, dom libvirt.Domain, flags uint32) (int32, int32, error) {
	args := m.Called(ctx, dom, flags)
	return args.Get(0).(int32), args.Get(1).(int32), args.Error(2)
}

func (m *mockLibvirtClient) DomainGetXMLDesc(ctx context.Context, dom libvirt.Domain, flags libvirt.DomainXMLFlags) (string, error) {
	args := m.Called(ctx, dom, flags)
	return args.String(0), args.Error(1)
}

func (m *mockLibvirtClient) DomainAttachDevice(ctx context.Context, dom libvirt.Domain, xml string) error {
	args := m.Called(ctx, dom, xml)
	return args.Error(0)
}

func (m *mockLibvirtClient) DomainDetachDevice(ctx context.Context, dom libvirt.Domain, xml string) error {
	args := m.Called(ctx, dom, xml)
	return args.Error(0)
}

func (m *mockLibvirtClient) DomainMemoryStats(ctx context.Context, dom libvirt.Domain, maxStats uint32, flags uint32) ([]libvirt.DomainMemoryStat, error) {
	args := m.Called(ctx, dom, maxStats, flags)
	return args.Get(0).([]libvirt.DomainMemoryStat), args.Error(1)
}

func (m *mockLibvirtClient) NetworkLookupByName(ctx context.Context, name string) (libvirt.Network, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(libvirt.Network), args.Error(1)
}

func (m *mockLibvirtClient) NetworkDefineXML(ctx context.Context, xml string) (libvirt.Network, error) {
	args := m.Called(ctx, xml)
	return args.Get(0).(libvirt.Network), args.Error(1)
}

func (m *mockLibvirtClient) NetworkCreate(ctx context.Context, net libvirt.Network) error {
	args := m.Called(ctx, net)
	return args.Error(0)
}

func (m *mockLibvirtClient) NetworkDestroy(ctx context.Context, net libvirt.Network) error {
	args := m.Called(ctx, net)
	return args.Error(0)
}

func (m *mockLibvirtClient) NetworkUndefine(ctx context.Context, net libvirt.Network) error {
	args := m.Called(ctx, net)
	return args.Error(0)
}

func (m *mockLibvirtClient) NetworkGetDhcpLeases(ctx context.Context, net libvirt.Network, mac libvirt.OptString, needResults uint32, flags uint32) ([]libvirt.NetworkDhcpLease, uint32, error) {
	args := m.Called(ctx, net, mac, needResults, flags)
	return args.Get(0).([]libvirt.NetworkDhcpLease), args.Get(1).(uint32), args.Error(2)
}

func (m *mockLibvirtClient) StoragePoolLookupByName(ctx context.Context, name string) (libvirt.StoragePool, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(libvirt.StoragePool), args.Error(1)
}

func (m *mockLibvirtClient) StoragePoolRefresh(ctx context.Context, pool libvirt.StoragePool, flags uint32) error {
	args := m.Called(ctx, pool, flags)
	return args.Error(0)
}

func (m *mockLibvirtClient) StorageVolLookupByName(ctx context.Context, pool libvirt.StoragePool, name string) (libvirt.StorageVol, error) {
	args := m.Called(ctx, pool, name)
	return args.Get(0).(libvirt.StorageVol), args.Error(1)
}

func (m *mockLibvirtClient) StorageVolCreateXML(ctx context.Context, pool libvirt.StoragePool, xml string, flags uint32) (libvirt.StorageVol, error) {
	args := m.Called(ctx, pool, xml, flags)
	return args.Get(0).(libvirt.StorageVol), args.Error(1)
}

func (m *mockLibvirtClient) StorageVolDelete(ctx context.Context, vol libvirt.StorageVol, flags uint32) error {
	args := m.Called(ctx, vol, flags)
	return args.Error(0)
}

func (m *mockLibvirtClient) StorageVolGetPath(ctx context.Context, vol libvirt.StorageVol) (string, error) {
	args := m.Called(ctx, vol)
	return args.String(0), args.Error(1)
}
