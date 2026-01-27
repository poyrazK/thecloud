package node

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalStore(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewLocalStore(tmpDir)
	require.NoError(t, err)

	bucket := "test-bucket"
	key := "folder/test-object.txt"
	data := []byte("hello world")
	timestamp := time.Now().UnixNano()

	// 1. Test Write
	err = store.Write(bucket, key, data, timestamp)
	require.NoError(t, err)

	// Verify file exists
	path := filepath.Join(tmpDir, bucket, key)
	_, err = os.Stat(path)
	require.NoError(t, err)

	// 2. Test Read
	readData, readTs, err := store.Read(bucket, key)
	require.NoError(t, err)
	assert.Equal(t, data, readData)
	assert.Equal(t, timestamp, readTs)

	// 3. Test Delete
	err = store.Delete(bucket, key)
	require.NoError(t, err)

	_, err = os.Stat(path)
	require.True(t, os.IsNotExist(err))
}

func TestLocalStore_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()
	store, _ := NewLocalStore(tmpDir)

	// Attempt to write outside root
	err := store.Write("bucket", "../outside.txt", []byte("evil"), 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied") // os.ErrPermission

	// Attempt to read outside root
	_, _, err = store.Read("bucket", "../outside.txt")
	require.Error(t, err)
}

func TestLocalStore_Assemble(t *testing.T) {
	tmpDir := t.TempDir()
	store, _ := NewLocalStore(tmpDir)

	bucket := "upload-bucket"
	finalKey := "final.bin"

	// Create parts
	part1 := "parts/1"
	part2 := "parts/2"
	data1 := []byte("part1")
	data2 := []byte("part2")

	require.NoError(t, store.Write(bucket, part1, data1, 0))
	require.NoError(t, store.Write(bucket, part2, data2, 0))

	// Assemble
	totalSize, err := store.Assemble(bucket, finalKey, []string{part1, part2})
	require.NoError(t, err)
	assert.Equal(t, int64(len(data1)+len(data2)), totalSize)

	// Verify final content
	content, _, err := store.Read(bucket, finalKey)
	require.NoError(t, err)
	assert.Equal(t, []byte("part1part2"), content)

	// Verify parts are deleted
	_, _, err = store.Read(bucket, part1)
	require.Error(t, err)
}

func TestLocalStore_ReadFallbackToMtime(t *testing.T) {
	tmpDir := t.TempDir()
	store, _ := NewLocalStore(tmpDir)

	bucket := "test-bucket"
	key := "file.txt"
	data := []byte("data")

	require.NoError(t, store.Write(bucket, key, data, 1))

	path := filepath.Join(tmpDir, bucket, key)
	metaPath := path + ".meta"
	require.NoError(t, os.Remove(metaPath))

	mtime := time.Unix(1700000000, 0)
	require.NoError(t, os.Chtimes(path, mtime, mtime))

	_, ts, err := store.Read(bucket, key)
	require.NoError(t, err)
	assert.Equal(t, mtime.UnixNano(), ts)
}

func TestLocalStore_InvalidAbsolutePath(t *testing.T) {
	tmpDir := t.TempDir()
	store, _ := NewLocalStore(tmpDir)

	err := store.Write("bucket", "/abs/path", []byte("data"), 0)
	require.Error(t, err)
	assert.ErrorIs(t, err, os.ErrInvalid)
}
