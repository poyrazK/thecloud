//go:build integration

package libvirt

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/digitalocean/go-libvirt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testVolName = "test-vol"
)

func TestCreateVolumeSuccess(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	ctx := context.Background()

	pool := libvirt.StoragePool{Name: "default"}
	vol := libvirt.StorageVol{Name: testVolName}

	m.On("StoragePoolLookupByName", mock.Anything, "default").Return(pool, nil)
	m.On("StoragePoolRefresh", mock.Anything, pool, uint32(0)).Return(nil)
	m.On("StorageVolCreateXML", mock.Anything, pool, mock.Anything, uint32(0)).Return(vol, nil)

	err := a.CreateVolume(ctx, testVolName)
	assert.NoError(t, err)
	m.AssertExpectations(t)
}

func TestDeleteVolumeSuccess(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	ctx := context.Background()

	pool := libvirt.StoragePool{Name: "default"}
	vol := libvirt.StorageVol{Name: testVolName}

	m.On("StoragePoolLookupByName", mock.Anything, "default").Return(pool, nil)
	m.On("StorageVolLookupByName", mock.Anything, pool, testVolName).Return(vol, nil)
	m.On("StorageVolDelete", mock.Anything, vol, uint32(0)).Return(nil)

	err := a.DeleteVolume(ctx, testVolName)
	assert.NoError(t, err)
	m.AssertExpectations(t)
}

func TestStoragePoolNotFound(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	ctx := context.Background()

	m.On("StoragePoolLookupByName", mock.Anything, "default").Return(libvirt.StoragePool{}, libvirt.Error{Code: 1, Message: "not found"})

	err := a.CreateVolume(ctx, testVolName)
	assert.Error(t, err)
}

func TestCreateVolumeSnapshotSuccess(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" {
		t.Skip("Snapshot tests require Linux with QEMU/KVM")
	}
	// Mock execCommand
	oldExec := execCommand
	defer func() { execCommand = oldExec }()
	execCommand = func(name string, arg ...string) *exec.Cmd {
		// Use absolute path to avoid PATH variable security hotspots
		return exec.Command("/usr/bin/true")
	}

	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	ctx := context.Background()

	pool := libvirt.StoragePool{Name: "default"}
	vol := libvirt.StorageVol{Name: testVolName}

	m.On("StoragePoolLookupByName", mock.Anything, "default").Return(pool, nil)
	m.On("StorageVolLookupByName", mock.Anything, pool, testVolName).Return(vol, nil)
	m.On("StorageVolGetPath", mock.Anything, vol).Return("/path/to/vol", nil)

	err := a.CreateVolumeSnapshot(ctx, testVolName, "/tmp/snap")
	assert.NoError(t, err)
	m.AssertExpectations(t)
}

func TestRestoreVolumeSnapshotSuccess(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" {
		t.Skip("Snapshot tests require Linux with QEMU/KVM")
	}
	// Mock execCommand
	oldExec := execCommand
	defer func() { execCommand = oldExec }()
	execCommand = func(name string, arg ...string) *exec.Cmd {
		// Use absolute path to avoid PATH variable security hotspots
		return exec.Command("/usr/bin/true")
	}

	oldMkdir := mkdirTemp
	defer func() { mkdirTemp = oldMkdir }()
	mkdirTemp = func(dir, pattern string) (string, error) {
		tmp, err := os.MkdirTemp(dir, pattern)
		if err == nil {
			writeErr := os.WriteFile(filepath.Join(tmp, "dummy.qcow2"), []byte("data"), 0644)
			if writeErr != nil {
				return tmp, writeErr
			}
		}
		return tmp, err
	}

	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	ctx := context.Background()

	pool := libvirt.StoragePool{Name: "default"}
	vol := libvirt.StorageVol{Name: testVolName}

	m.On("StoragePoolLookupByName", mock.Anything, "default").Return(pool, nil)
	m.On("StorageVolLookupByName", mock.Anything, pool, testVolName).Return(vol, nil)
	m.On("StorageVolGetPath", mock.Anything, vol).Return("/path/to/vol", nil)

	err := a.RestoreVolumeSnapshot(ctx, testVolName, "/tmp/snap")
	assert.NoError(t, err)
	m.AssertExpectations(t)
}

func TestDeleteVolumeFailures(t *testing.T) {
	t.Parallel()
	m := new(MockLibvirtClient)
	a := newTestAdapter(m)
	ctx := context.Background()
	pool := libvirt.StoragePool{Name: "default"}
	vol := libvirt.StorageVol{Name: testVolName}

	// 1. Pool not found
	m.On("StoragePoolLookupByName", ctx, "default").Return(libvirt.StoragePool{}, libvirt.Error{Code: 1}).Once()
	err := a.DeleteVolume(ctx, testVolName)
	assert.Error(t, err)

	// 2. Delete error
	m.On("StoragePoolLookupByName", ctx, "default").Return(pool, nil).Once()
	m.On("StorageVolLookupByName", ctx, pool, testVolName).Return(vol, nil).Once()
	m.On("StorageVolDelete", ctx, vol, uint32(0)).Return(libvirt.Error{Code: 1}).Once()
	err = a.DeleteVolume(ctx, testVolName)
	assert.Error(t, err)
}
