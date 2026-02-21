package lvm

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

const testVolumeName = "test-vol"

// fakeExecer for testing
type fakeExecer struct {
	// Map of command name to its behavior
	commands map[string]func(args ...string) ([]byte, error)
}

func newFakeExecer() *fakeExecer {
	return &fakeExecer{
		commands: make(map[string]func(args ...string) ([]byte, error)),
	}
}

func (f *fakeExecer) Run(name string, args ...string) ([]byte, error) {
	if fn, ok := f.commands[name]; ok {
		return fn(args...)
	}
	return nil, fmt.Errorf("command not found: %s", name)
}

func (f *fakeExecer) addCommand(name string, fn func(args ...string) ([]byte, error)) {
	f.commands[name] = fn
}

func TestLvmAdapterType(t *testing.T) {
	adapter := NewLvmAdapter("testvg")
	if adapter.Type() != "lvm" {
		t.Errorf("expected type 'lvm', got %s", adapter.Type())
	}
}

func TestLvmAdapterPing(t *testing.T) {
	adapter := NewLvmAdapter("testvg")
	err := adapter.Ping(context.Background())
	if err == nil {
		// If no error, perhaps lvm is available, but unlikely in test
		t.Log("lvm ping succeeded")
	} else {
		// Expected if lvm not available
		t.Logf("lvm ping failed as expected: %v", err)
	}
}

func TestLvmAdapterCreateVolumeInvalidName(t *testing.T) {
	adapter := NewLvmAdapter("testvg")
	_, err := adapter.CreateVolume(context.Background(), "", 10)
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestLvmAdapterDeleteVolumeInvalidName(t *testing.T) {
	adapter := NewLvmAdapter("testvg")
	err := adapter.DeleteVolume(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestLvmAdapterCreateVolumeSuccess(t *testing.T) {
	fake := newFakeExecer()
	fake.addCommand("lvcreate", func(args ...string) ([]byte, error) {
		assert.Equal(t, []string{"-L", "10G", "-n", testVolumeName, "vg0"}, args)
		return []byte("Logical volume \"" + testVolumeName + "\" created"), nil
	})

	adapter := &LvmAdapter{vgName: "vg0", execer: fake}
	path, err := adapter.CreateVolume(context.Background(), testVolumeName, 10)

	require.NoError(t, err)
	assert.Equal(t, "/dev/vg0/"+testVolumeName, path)
}

func TestLvmAdapterCreateVolumeFailure(t *testing.T) {
	fake := newFakeExecer()
	fake.addCommand("lvcreate", func(args ...string) ([]byte, error) {
		return []byte("Volume group \"vg0\" not found"), errors.New("exit status 5")
	})

	adapter := &LvmAdapter{vgName: "vg0", execer: fake}
	_, err := adapter.CreateVolume(context.Background(), "bad-vol", 5)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create logical volume")
}

func TestLvmAdapterDeleteVolumeSuccess(t *testing.T) {
	fake := newFakeExecer()
	fake.addCommand("lvremove", func(args ...string) ([]byte, error) {
		assert.Equal(t, []string{"-f", "vg0/" + testVolumeName}, args)
		return nil, nil
	})

	adapter := &LvmAdapter{vgName: "vg0", execer: fake}
	err := adapter.DeleteVolume(context.Background(), testVolumeName)

	require.NoError(t, err)
}

func TestLvmAdapterCreateSnapshotSuccess(t *testing.T) {
	fake := newFakeExecer()
	fake.addCommand("lvcreate", func(args ...string) ([]byte, error) {
		expected := []string{"-s", "-n", "snap1", "-L", "1G", "/dev/vg0/data-vol"}
		assert.Equal(t, expected, args)
		return nil, nil
	})

	adapter := &LvmAdapter{vgName: "vg0", execer: fake}
	err := adapter.CreateSnapshot(context.Background(), "data-vol", "snap1")

	require.NoError(t, err)
}

func TestLvmAdapterRestoreSnapshotSuccess(t *testing.T) {
	fake := newFakeExecer()
	fake.addCommand("lvconvert", func(args ...string) ([]byte, error) {
		expected := []string{"--merge", "vg0/snap1"}
		assert.Equal(t, expected, args)
		return nil, nil
	})

	adapter := &LvmAdapter{vgName: "vg0", execer: fake}
	err := adapter.RestoreSnapshot(context.Background(), "data-vol", "snap1")

	require.NoError(t, err)
}

func TestLvmAdapterDeleteSnapshotSuccess(t *testing.T) {
	fake := newFakeExecer()
	fake.addCommand("lvremove", func(args ...string) ([]byte, error) {
		expected := []string{"-f", "vg0/snap1"}
		assert.Equal(t, expected, args)
		return nil, nil
	})

	adapter := &LvmAdapter{vgName: "vg0", execer: fake}
	err := adapter.DeleteSnapshot(context.Background(), "snap1")

	require.NoError(t, err)
}

func TestLvmAdapterAttachVolume(t *testing.T) {
	adapter := NewLvmAdapter("vg0")
	err := adapter.AttachVolume(context.Background(), "vol1", "inst1")
	require.NoError(t, err)
}

func TestLvmAdapterDetachVolume(t *testing.T) {
	adapter := NewLvmAdapter("vg0")
	err := adapter.DetachVolume(context.Background(), "vol1", "inst1")
	require.NoError(t, err)
}
