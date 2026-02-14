package services_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"

	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/platform"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// InMemFileStore is a simple in-memory implementation of ports.FileStore.
type InMemFileStore struct {
	mu       sync.RWMutex
	files    map[string][]byte
	failNext bool
}

func NewInMemFileStore() *InMemFileStore {
	return &InMemFileStore{files: make(map[string][]byte)}
}

func (s *InMemFileStore) key(bucket, key string) string {
	return bucket + "/" + key
}

func (s *InMemFileStore) Write(ctx context.Context, bucket, key string, r io.Reader) (int64, error) {
	if s.failNext {
		s.failNext = false
		return 0, fmt.Errorf("injected failure")
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return 0, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.files[s.key(bucket, key)] = data
	return int64(len(data)), nil
}

func (s *InMemFileStore) Read(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	if s.failNext {
		s.failNext = false
		return nil, fmt.Errorf("injected failure")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, ok := s.files[s.key(bucket, key)]
	if !ok {
		return nil, fmt.Errorf("file not found")
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (s *InMemFileStore) Delete(ctx context.Context, bucket, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.files, s.key(bucket, key))
	return nil
}

func (s *InMemFileStore) GetClusterStatus(ctx context.Context) (*domain.StorageCluster, error) {
	return &domain.StorageCluster{Nodes: []domain.StorageNode{{ID: "mem-1", Status: "online"}}}, nil
}

func (s *InMemFileStore) Assemble(ctx context.Context, bucket, key string, parts []string) (int64, error) {
	if s.failNext {
		s.failNext = false
		return 0, fmt.Errorf("injected failure")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	var buf bytes.Buffer
	for _, pk := range parts {
		data, ok := s.files[s.key(bucket, pk)]
		if !ok {
			return 0, fmt.Errorf("part not found: %s", pk)
		}
		buf.Write(data)
		delete(s.files, s.key(bucket, pk))
	}
	s.files[s.key(bucket, key)] = buf.Bytes()
	return int64(buf.Len()), nil
}

// FailingEncryptionService wraps a real one but can fail.
type FailingEncryptionService struct {
	ports.EncryptionService
	failNext bool
}

func (f *FailingEncryptionService) Encrypt(ctx context.Context, bucket string, data []byte) ([]byte, error) {
	if f.failNext {
		f.failNext = false
		return nil, fmt.Errorf("injected failure")
	}
	return f.EncryptionService.Encrypt(ctx, bucket, data)
}

func (f *FailingEncryptionService) Decrypt(ctx context.Context, bucket string, data []byte) ([]byte, error) {
	if f.failNext {
		f.failNext = false
		return nil, fmt.Errorf("injected failure")
	}
	return f.EncryptionService.Decrypt(ctx, bucket, data)
}

func setupStorageServiceIntegrationTest(t *testing.T) (ports.StorageService, ports.StorageRepository, *InMemFileStore, *FailingEncryptionService, postgres.DB, context.Context) {
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewStorageRepository(db)
	store := NewInMemFileStore()
	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(auditRepo)
	cfg := &platform.Config{SecretsEncryptionKey: "test-secret-32-chars-long-needed-!!", Port: "8080"}

	// Setup encryption service
	masterKeyHex := "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"
	encRepo := postgres.NewEncryptionRepository(db)
	realEncSvc, err := services.NewEncryptionService(encRepo, masterKeyHex)
	require.NoError(t, err)
	require.NotNil(t, realEncSvc)
	encSvc := &FailingEncryptionService{EncryptionService: realEncSvc}

	svc := services.NewStorageService(repo, store, auditSvc, encSvc, cfg)

	return svc, repo, store, encSvc, db, ctx
}

// FailingReader returns an error on Read.
type FailingReader struct{}

func (f *FailingReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("read failure")
}

func TestStorageService_Integration(t *testing.T) {
	svc, _, store, encSvc, db, ctx := setupStorageServiceIntegrationTest(t)
	userID := appcontext.UserIDFromContext(ctx)
	masterKeyHex := "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"

	t.Run("BucketLifecycle", func(t *testing.T) {
		name := "my-integration-bucket"
		bucket, err := svc.CreateBucket(ctx, name, false)
		assert.NoError(t, err)
		assert.NotNil(t, bucket)
		assert.Equal(t, name, bucket.Name)
		assert.Equal(t, userID, bucket.UserID)

		// Get
		fetched, err := svc.GetBucket(ctx, name)
		assert.NoError(t, err)
		assert.Equal(t, name, fetched.Name)

		// List
		buckets, err := svc.ListBuckets(ctx)
		assert.NoError(t, err)
		assert.Len(t, buckets, 1)

		// Set Versioning
		err = svc.SetBucketVersioning(ctx, name, true)
		assert.NoError(t, err)

		updated, _ := svc.GetBucket(ctx, name)
		assert.True(t, updated.VersioningEnabled)

		// Delete
		err = svc.DeleteBucket(ctx, name)
		assert.NoError(t, err)
	})

	t.Run("ObjectOps", func(t *testing.T) {
		bucketName := "obj-bucket"
		_, _ = svc.CreateBucket(ctx, bucketName, false)

		key := "test.txt"
		content := "integration test data"

		// Upload
		obj, err := svc.Upload(ctx, bucketName, key, strings.NewReader(content))
		assert.NoError(t, err)
		assert.NotNil(t, obj)
		assert.Equal(t, key, obj.Key)

		// Meta
		list, err := svc.ListObjects(ctx, bucketName)
		assert.NoError(t, err)
		assert.Len(t, list, 1)

		// Download
		r, meta, err := svc.Download(ctx, bucketName, key)
		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.Equal(t, key, meta.Key)
		data, _ := io.ReadAll(r)
		assert.Equal(t, content, string(data))

		// Delete
		err = svc.DeleteObject(ctx, bucketName, key)
		assert.NoError(t, err)

		// Should fail to download after delete (soft delete in metadata)
		_, _, err = svc.Download(ctx, bucketName, key)
		assert.Error(t, err)
	})

	t.Run("MultipartUpload", func(t *testing.T) {
		bucketName := "mp-bucket"
		_, _ = svc.CreateBucket(ctx, bucketName, false)
		key := "large.zip"

		// Success Case
		upload, err := svc.CreateMultipartUpload(ctx, bucketName, key)
		assert.NoError(t, err)

		_, err = svc.UploadPart(ctx, upload.ID, 1, strings.NewReader("part1"))
		assert.NoError(t, err)
		_, err = svc.UploadPart(ctx, upload.ID, 2, strings.NewReader("part2"))
		assert.NoError(t, err)

		obj, err := svc.CompleteMultipartUpload(ctx, upload.ID)
		assert.NoError(t, err)
		assert.Equal(t, key, obj.Key)
		assert.Equal(t, int64(10), obj.SizeBytes) // "part1" (5) + "part2" (5) = 10

		// Verify content
		r, _, _ := svc.Download(ctx, bucketName, key)
		data, _ := io.ReadAll(r)
		assert.Equal(t, "part1part2", string(data))

		// Error: Complete with no parts
		uploadErr, _ := svc.CreateMultipartUpload(ctx, bucketName, "no-parts")
		_, err = svc.CompleteMultipartUpload(ctx, uploadErr.ID)
		assert.Error(t, err)

		// Abort Case
		abortKey := "abort.me"
		upload2, _ := svc.CreateMultipartUpload(ctx, bucketName, abortKey)
		err = svc.AbortMultipartUpload(ctx, upload2.ID)
		assert.NoError(t, err)

		// Should fail to complete after abort
		_, err = svc.CompleteMultipartUpload(ctx, upload2.ID)
		assert.Error(t, err)
	})

	t.Run("Encryption", func(t *testing.T) {
		bucketName := "encrypted-bucket"
		_, err := svc.CreateBucket(ctx, bucketName, false)
		require.NoError(t, err)

		// Enable encryption manually in DB for testing
		_, err = db.Exec(ctx, "UPDATE buckets SET encryption_enabled = TRUE WHERE name = $1", bucketName)
		require.NoError(t, err)

		// Initialize encryption key for this bucket
		encRepo := postgres.NewEncryptionRepository(db)
		masterKeyHex := "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"
		encSvc, _ := services.NewEncryptionService(encRepo, masterKeyHex)
		_, err = encSvc.CreateKey(ctx, bucketName)
		require.NoError(t, err)

		key := "secret.data"
		content := "very sensitive information"

		// Upload (should be encrypted transparently)
		obj, err := svc.Upload(ctx, bucketName, key, strings.NewReader(content))
		require.NoError(t, err)
		assert.NotNil(t, obj)

		// Download (should be decrypted transparently)
		r, _, err := svc.Download(ctx, bucketName, key)
		assert.NoError(t, err)
		data, _ := io.ReadAll(r)
		assert.Equal(t, content, string(data))
	})

	t.Run("Versioning", func(t *testing.T) {
		bucketName := "version-bucket"
		_, _ = svc.CreateBucket(ctx, bucketName, false)
		err := svc.SetBucketVersioning(ctx, bucketName, true)
		assert.NoError(t, err)

		// Verify versioning is enabled
		b, err := svc.GetBucket(ctx, bucketName)
		assert.NoError(t, err)
		assert.True(t, b.VersioningEnabled)

		key := "v.txt"

		// Upload v1
		v1, err := svc.Upload(ctx, bucketName, key, strings.NewReader("v1"))
		assert.NoError(t, err)
		assert.NotEqual(t, "null", v1.VersionID)

		// Upload v2
		v2, err := svc.Upload(ctx, bucketName, key, strings.NewReader("v2"))
		assert.NoError(t, err)
		assert.NotEqual(t, "null", v2.VersionID)
		assert.NotEqual(t, v1.VersionID, v2.VersionID)

		// List Versions
		versions, err := svc.ListVersions(ctx, bucketName, key)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(versions), 2)

		// Download v1
		r, meta, err := svc.DownloadVersion(ctx, bucketName, key, v1.VersionID)
		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.Equal(t, v1.VersionID, meta.VersionID)
		d1, _ := io.ReadAll(r)
		assert.Equal(t, "v1", string(d1))

		// Download v2
		r2, meta2, _ := svc.DownloadVersion(ctx, bucketName, key, v2.VersionID)
		assert.Equal(t, v2.VersionID, meta2.VersionID)
		d2, _ := io.ReadAll(r2)
		assert.Equal(t, "v2", string(d2))

		// Delete v1
		err = svc.DeleteVersion(ctx, bucketName, key, v1.VersionID)
		assert.NoError(t, err)

		// Should fail to download after delete
		_, _, err = svc.DownloadVersion(ctx, bucketName, key, v1.VersionID)
		assert.Error(t, err)
	})

	t.Run("ErrorPaths", func(t *testing.T) {
		// Invalid bucket names
		_, err := svc.CreateBucket(ctx, "INVALID NAME", false)
		assert.Error(t, err)
		_, err = svc.CreateBucket(ctx, "", false)
		assert.Error(t, err)
		_, err = svc.CreateBucket(ctx, strings.Repeat("a", 65), false)
		assert.Error(t, err)
		_, err = svc.CreateBucket(ctx, "-invalid", false)
		assert.Error(t, err)

		// Bucket not found
		_, err = svc.Upload(ctx, "non-existent", "key", strings.NewReader("data"))
		assert.Error(t, err)

		_, _, err = svc.Download(ctx, "non-existent", "key")
		assert.Error(t, err)

		// Object not found
		_, _, err = svc.Download(ctx, "obj-bucket", "missing-key")
		assert.Error(t, err)

		// Multipart not found
		_, err = svc.UploadPart(ctx, [16]byte{1}, 1, strings.NewReader("data"))
		assert.Error(t, err)

		err = svc.AbortMultipartUpload(ctx, [16]byte{1})
		assert.Error(t, err)

		_, err = svc.CompleteMultipartUpload(ctx, [16]byte{1})
		assert.Error(t, err)

		// CreateMultipart with missing bucket
		_, err = svc.CreateMultipartUpload(ctx, "missing-bucket", "key")
		assert.Error(t, err)

		// Download store read error
		// 1. Create a fresh object
		_, err = svc.Upload(ctx, "obj-bucket", "fail-store", strings.NewReader("data"))
		if err != nil {
			t.Fatalf("Upload failed: %v", err)
		}
		// 2. Force store failure
		store.failNext = true
		_, _, err = svc.Download(ctx, "obj-bucket", "fail-store")
		assert.Error(t, err)

		// Download decryption error
		// 1. Enable encryption on a bucket
		encBucket := "fail-decrypt"
		_, _ = svc.CreateBucket(ctx, encBucket, false)
		_, _ = db.Exec(ctx, "UPDATE buckets SET encryption_enabled = TRUE WHERE name = $1", encBucket)
		encRepo := postgres.NewEncryptionRepository(db)
		tempEncSvc, _ := services.NewEncryptionService(encRepo, masterKeyHex)
		_, _ = tempEncSvc.CreateKey(ctx, encBucket)

		// 2. Upload valid encrypted file
		_, _ = svc.Upload(ctx, encBucket, "secret", strings.NewReader("top-secret"))

		// 3. Force decryption failure
		encSvc.failNext = true
		_, _, err = svc.Download(ctx, encBucket, "secret")
		assert.Error(t, err)

		// Multipart Assemble failure
		up3, _ := svc.CreateMultipartUpload(ctx, "obj-bucket", "fail-assemble")
		_, _ = svc.UploadPart(ctx, up3.ID, 1, strings.NewReader("p1"))
		store.failNext = true
		_, err = svc.CompleteMultipartUpload(ctx, up3.ID)
		assert.Error(t, err)

		// UploadPart store failure
		up4, _ := svc.CreateMultipartUpload(ctx, "obj-bucket", "fail-part")
		store.failNext = true
		_, err = svc.UploadPart(ctx, up4.ID, 1, strings.NewReader("data"))
		assert.Error(t, err)

		// Upload Reader error (with encryption)
		encReadBucket := "fail-read-enc"
		_, _ = svc.CreateBucket(ctx, encReadBucket, false)
		_, _ = db.Exec(ctx, "UPDATE buckets SET encryption_enabled = TRUE WHERE name = $1", encReadBucket)
		tempEncSvc2, _ := services.NewEncryptionService(encRepo, masterKeyHex)
		_, _ = tempEncSvc2.CreateKey(ctx, encReadBucket)

		_, err = svc.Upload(ctx, encReadBucket, "fail", &FailingReader{})
		assert.Error(t, err)

		// GeneratePresignedURL secret missing
		badCfg := &platform.Config{SecretsEncryptionKey: "", Port: "8080"}
		badSvc := services.NewStorageService(postgres.NewStorageRepository(db), store, services.NewAuditService(postgres.NewAuditRepository(db)), encSvc, badCfg)
		_, err = badSvc.GeneratePresignedURL(ctx, "obj-bucket", "f.txt", "GET", 0)
		assert.Error(t, err)
	})

	t.Run("ClusterStatus", func(t *testing.T) {
		status, err := svc.GetClusterStatus(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, status)
		assert.NotEmpty(t, status.Nodes)
	})

	t.Run("PresignedURL", func(t *testing.T) {
		bucketName := "url-bucket"
		_, _ = svc.CreateBucket(ctx, bucketName, false)

		url, err := svc.GeneratePresignedURL(ctx, bucketName, "file.txt", "GET", 0)
		assert.NoError(t, err)
		assert.NotEmpty(t, url.URL)
		assert.Equal(t, "GET", url.Method)

		// Error: bucket not found
		_, err = svc.GeneratePresignedURL(ctx, "ghost", "file.txt", "GET", 0)
		assert.Error(t, err)
	})
}
