package libvirt

import (
	"context"
	"testing"

	"github.com/digitalocean/go-libvirt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateVolumeSuccess(t *testing.T) {
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	ctx := context.Background()

	pool := libvirt.StoragePool{Name: "default"}
	vol := libvirt.StorageVol{Name: "test-vol"}

	m.On("StoragePoolLookupByName", mock.Anything, "default").Return(pool, nil)
	m.On("StoragePoolRefresh", mock.Anything, pool, uint32(0)).Return(nil)
	m.On("StorageVolCreateXML", mock.Anything, pool, mock.Anything, uint32(0)).Return(vol, nil)

	err := a.CreateVolume(ctx, "test-vol")
	assert.NoError(t, err)
	m.AssertExpectations(t)
}

func TestDeleteVolumeSuccess(t *testing.T) {
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	ctx := context.Background()

	pool := libvirt.StoragePool{Name: "default"}
	vol := libvirt.StorageVol{Name: "test-vol"}

	m.On("StoragePoolLookupByName", mock.Anything, "default").Return(pool, nil)
	m.On("StorageVolLookupByName", mock.Anything, pool, "test-vol").Return(vol, nil)
	m.On("StorageVolDelete", mock.Anything, vol, uint32(0)).Return(nil)

	err := a.DeleteVolume(ctx, "test-vol")
	assert.NoError(t, err)
	m.AssertExpectations(t)
}

func TestStoragePoolNotFound(t *testing.T) {
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	ctx := context.Background()

	m.On("StoragePoolLookupByName", mock.Anything, "default").Return(libvirt.StoragePool{}, libvirt.Error{Code: 1, Message: "not found"})

	err := a.CreateVolume(ctx, "test-vol")
	assert.Error(t, err)
}
