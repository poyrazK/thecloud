package node

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"sync"
)

// LocalStore manages file storage on the local disk.
type LocalStore struct {
	rootDir string
	mu      sync.RWMutex
}

// NewLocalStore initializes a new local storage backend.
func NewLocalStore(dataDir string) (*LocalStore, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	return &LocalStore{rootDir: dataDir}, nil
}

// Write saves data to disk. Overwrites if exists.
func (s *LocalStore) Write(bucket, key string, data []byte, timestamp int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.getObjectPath(bucket, key)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	// Write metadata (timestamp)
	metaPath := path + ".meta"
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(timestamp))
	return os.WriteFile(metaPath, buf, 0644)
}

// Read retrieves data from disk.
func (s *LocalStore) Read(bucket, key string) ([]byte, int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := s.getObjectPath(bucket, key)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, err
	}

	// Read metadata
	var timestamp int64
	metaBytes, err := os.ReadFile(path + ".meta")
	if err == nil && len(metaBytes) >= 8 {
		timestamp = int64(binary.LittleEndian.Uint64(metaBytes))
	} else {
		// Fallback to file mtime if meta missing
		info, statErr := os.Stat(path)
		if statErr == nil {
			timestamp = info.ModTime().UnixNano()
		}
	}

	return data, timestamp, nil
}

// Delete removes data from disk.
func (s *LocalStore) Delete(bucket, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.getObjectPath(bucket, key)
	_ = os.Remove(path + ".meta")
	return os.Remove(path)
}

func (s *LocalStore) getObjectPath(bucket, key string) string {
	return filepath.Join(s.rootDir, bucket, key)
}
