// Package node implements storage node services.
package node

import (
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LocalStore manages file storage on the local disk.
type LocalStore struct {
	rootDir string
	mu      sync.RWMutex
}

// NewLocalStore initializes a new local storage backend.
func NewLocalStore(dataDir string) (*LocalStore, error) {
	if err := os.MkdirAll(dataDir, 0750); err != nil {
		return nil, err
	}
	return &LocalStore{rootDir: dataDir}, nil
}

// Write saves data to disk. Overwrites if exists.
func (s *LocalStore) Write(bucket, key string, data []byte, timestamp int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path, err := s.getObjectPath(bucket, key)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return err
	}

	// Write metadata (timestamp)
	metaPath := path + ".meta"
	buf := make([]byte, 8)

	var uTimestamp uint64
	if timestamp < 0 {
		uTimestamp = 0
	} else {
		uTimestamp = uint64(timestamp)
	}

	binary.LittleEndian.PutUint64(buf, uTimestamp)
	return os.WriteFile(metaPath, buf, 0600)
}

// Read retrieves data from disk.
func (s *LocalStore) Read(bucket, key string) ([]byte, int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path, err := s.getObjectPath(bucket, key)
	if err != nil {
		return nil, 0, err
	}

	// filepath.Clean is already done in getObjectPath, gosec might still warn
	// so we use the path directly as it is sanitized.
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, 0, err
	}

	// Read metadata
	var timestamp int64
	metaPath := filepath.Clean(path + ".meta")
	metaBytes, err := os.ReadFile(metaPath)
	if err == nil && len(metaBytes) >= 8 {
		uVal := binary.LittleEndian.Uint64(metaBytes)
		if uVal > math.MaxInt64 {
			timestamp = math.MaxInt64
		} else {
			timestamp = int64(uVal)
		}
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

	path, err := s.getObjectPath(bucket, key)
	if err != nil {
		return err
	}

	_ = os.Remove(path + ".meta")
	return os.Remove(path)
}

// Assemble combines multiple parts into a single object.
func (s *LocalStore) Assemble(bucket, key string, parts []string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	destPath, err := s.getObjectPath(bucket, key)
	if err != nil {
		return 0, err
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0750); err != nil {
		return 0, err
	}

	f, err := os.OpenFile(filepath.Clean(destPath), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return 0, err
	}
	defer func() { _ = f.Close() }()

	var totalSize int64
	for _, partKey := range parts {
		partPath, err := s.getObjectPath(bucket, partKey)
		if err != nil {
			return 0, err
		}

		data, err := os.ReadFile(filepath.Clean(partPath))
		if err != nil {
			return 0, err
		}
		n, err := f.Write(data)
		if err != nil {
			return 0, err
		}
		totalSize += int64(n)
		_ = os.Remove(partPath)
		_ = os.Remove(partPath + ".meta")
	}

	// Write final meta with current timestamp
	metaPath := destPath + ".meta"
	buf := make([]byte, 8)

	now := time.Now().UnixNano()
	var uNow uint64
	if now < 0 {
		uNow = 0
	} else {
		uNow = uint64(now)
	}

	binary.LittleEndian.PutUint64(buf, uNow)
	_ = os.WriteFile(metaPath, buf, 0600)

	return totalSize, nil
}

func (s *LocalStore) getObjectPath(bucket, key string) (string, error) {
	// Clean the inputs
	cleanBucket := filepath.Base(filepath.Clean(bucket))
	cleanKey := filepath.Clean(key)

	if filepath.IsAbs(cleanKey) {
		return "", os.ErrInvalid
	}

	bucketDir := filepath.Join(s.rootDir, cleanBucket)
	fullPath := filepath.Join(bucketDir, cleanKey)

	// Verify it's within bucketDir (strict isolation)
	rel, err := filepath.Rel(bucketDir, fullPath)
	if err != nil || len(rel) < 2 && rel == ".." || (len(rel) >= 2 && rel[:2] == "..") {
		return "", os.ErrPermission
	}

	return fullPath, nil
}
