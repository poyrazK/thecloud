package services_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestStorageServiceUnit(t *testing.T) {
	mockRepo := new(MockStorageRepo)
	mockStore := new(MockFileStore)
	mockAuditSvc := new(MockAuditService)
	cfg := &platform.Config{SecretsEncryptionKey: "test-secret-key-32-chars-long-!!!"}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := services.NewStorageService(services.StorageServiceParams{
		Repo:     mockRepo,
		Store:    mockStore,
		AuditSvc: mockAuditSvc,
		CFG:      cfg,
		Logger:   logger,
	})

	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)

	t.Run("CreateBucket", func(t *testing.T) {
		mockRepo.On("CreateBucket", mock.Anything, mock.Anything).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "storage.bucket_create", "bucket", mock.Anything, mock.Anything).Return(nil).Once()

		bucket, err := svc.CreateBucket(ctx, "my-bucket", false)
		require.NoError(t, err)
		assert.NotNil(t, bucket)
		assert.Equal(t, "my-bucket", bucket.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("CreateBucket Invalid Names", func(t *testing.T) {
		invalidNames := []string{"a", "ab", "Invalid_Name", "-start-hyphen", "end-dot.", "two..dots"}
		for _, name := range invalidNames {
			_, err := svc.CreateBucket(ctx, name, false)
			assert.Error(t, err, "expected error for bucket name: %s", name)
		}
	})

	t.Run("GetBucket", func(t *testing.T) {
		bucket := &domain.Bucket{Name: "my-bucket"}
		mockRepo.On("GetBucket", mock.Anything, "my-bucket").Return(bucket, nil).Once()

		res, err := svc.GetBucket(ctx, "my-bucket")
		require.NoError(t, err)
		assert.Equal(t, bucket, res)
		mockRepo.AssertExpectations(t)
	})

	t.Run("DeleteBucket", func(t *testing.T) {
		mockRepo.On("List", mock.Anything, "my-bucket").Return([]*domain.Object{}, nil).Once()
		mockRepo.On("DeleteBucket", mock.Anything, "my-bucket").Return(nil).Once()

		err := svc.DeleteBucket(ctx, "my-bucket", false)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("DeleteBucket Not Empty", func(t *testing.T) {
		mockRepo.On("List", mock.Anything, "my-bucket").Return([]*domain.Object{{Key: "k1"}}, nil).Once()

		err := svc.DeleteBucket(ctx, "my-bucket", false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "bucket is not empty")
		mockRepo.AssertExpectations(t)
	})

	t.Run("DeleteBucket Force", func(t *testing.T) {
		mockRepo.On("List", mock.Anything, "my-bucket").Return([]*domain.Object{{Key: "k1"}}, nil).Once()
		mockRepo.On("SoftDelete", mock.Anything, "my-bucket", "k1").Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "storage.object_delete", "storage", mock.Anything, mock.Anything).Return(nil).Once()
		mockRepo.On("DeleteBucket", mock.Anything, "my-bucket").Return(nil).Once()

		err := svc.DeleteBucket(ctx, "my-bucket", true)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ListBuckets", func(t *testing.T) {
		buckets := []*domain.Bucket{{Name: "b1"}, {Name: "b2"}}
		mockRepo.On("ListBuckets", mock.Anything, userID.String()).Return(buckets, nil).Once()

		res, err := svc.ListBuckets(ctx)
		require.NoError(t, err)
		assert.Equal(t, buckets, res)
		mockRepo.AssertExpectations(t)
	})

	t.Run("SetBucketVersioning", func(t *testing.T) {
		mockRepo.On("SetBucketVersioning", mock.Anything, "my-bucket", true).Return(nil).Once()

		err := svc.SetBucketVersioning(ctx, "my-bucket", true)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Upload", func(t *testing.T) {
		bucket := &domain.Bucket{Name: "my-bucket", VersioningEnabled: false}
		mockRepo.On("GetBucket", mock.Anything, "my-bucket").Return(bucket, nil).Once()
		mockStore.On("Write", mock.Anything, "my-bucket", "test.txt", mock.Anything).Return(int64(12), nil).Once()
		
		// First SaveMeta call for PENDING status
		mockRepo.On("SaveMeta", mock.Anything, mock.MatchedBy(func(obj *domain.Object) bool {
			return obj.UploadStatus == domain.UploadStatusPending && obj.SizeBytes == 0
		})).Return(nil).Once()

		// Second SaveMeta call for AVAILABLE status
		mockRepo.On("SaveMeta", mock.Anything, mock.MatchedBy(func(obj *domain.Object) bool {
			return obj.UploadStatus == domain.UploadStatusAvailable && obj.SizeBytes == 12
		})).Return(nil).Once()
		
		mockAuditSvc.On("Log", mock.Anything, userID, "storage.object_upload", "storage", mock.Anything, mock.Anything).Return(nil).Once()

		obj, err := svc.Upload(ctx, "my-bucket", "test.txt", strings.NewReader("hello world!"), "")
		require.NoError(t, err)
		assert.NotNil(t, obj)
		assert.Equal(t, int64(12), obj.SizeBytes)
		assert.Equal(t, "text/plain; charset=utf-8", obj.ContentType)
		assert.NotEmpty(t, obj.Checksum)

		mockRepo.AssertExpectations(t)
		mockStore.AssertExpectations(t)
		mockAuditSvc.AssertExpectations(t)
	})

	t.Run("Upload Checksum Mismatch", func(t *testing.T) {
		bucket := &domain.Bucket{Name: "my-bucket", VersioningEnabled: false}
		mockRepo.On("GetBucket", mock.Anything, "my-bucket").Return(bucket, nil).Once()
		// SaveMeta for PENDING
		mockRepo.On("SaveMeta", mock.Anything, mock.MatchedBy(func(obj *domain.Object) bool {
			return obj.UploadStatus == domain.UploadStatusPending
		})).Return(nil).Once()

		mockStore.On("Write", mock.Anything, "my-bucket", "test.txt", mock.Anything).Return(int64(12), nil).Once()
		mockStore.On("Delete", mock.Anything, "my-bucket", "test.txt").Return(nil).Once()

		providedChecksum := "invalid-checksum"
		_, err := svc.Upload(ctx, "my-bucket", "test.txt", strings.NewReader("hello world!"), providedChecksum)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "data integrity check failed")
		
		mockRepo.AssertExpectations(t)
		mockStore.AssertExpectations(t)
	})

	t.Run("Download", func(t *testing.T) {
		obj := &domain.Object{Bucket: "my-bucket", Key: "test.txt", VersionID: "null", SizeBytes: 12}
		mockRepo.On("GetMeta", mock.Anything, "my-bucket", "test.txt").Return(obj, nil).Once()
		mockStore.On("Read", mock.Anything, "my-bucket", "test.txt").Return(io.NopCloser(strings.NewReader("hello world!")), nil).Once()
		mockRepo.On("GetBucket", mock.Anything, "my-bucket").Return(&domain.Bucket{Name: "my-bucket"}, nil).Once()

		reader, meta, err := svc.Download(ctx, "my-bucket", "test.txt")
		require.NoError(t, err)
		assert.NotNil(t, reader)
		assert.Equal(t, obj, meta)
		
		data, _ := io.ReadAll(reader)
		assert.Equal(t, "hello world!", string(data))
		mockRepo.AssertExpectations(t)
		mockStore.AssertExpectations(t)
	})

	t.Run("ListObjects", func(t *testing.T) {
		objects := []*domain.Object{{Key: "test.txt"}}
		mockRepo.On("List", mock.Anything, "my-bucket").Return(objects, nil).Once()

		res, err := svc.ListObjects(ctx, "my-bucket")
		require.NoError(t, err)
		assert.Equal(t, objects, res)
		mockRepo.AssertExpectations(t)
	})

	t.Run("DeleteObject", func(t *testing.T) {
		mockRepo.On("SoftDelete", mock.Anything, "my-bucket", "test.txt").Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "storage.object_delete", "storage", mock.Anything, mock.Anything).Return(nil).Once()

		err := svc.DeleteObject(ctx, "my-bucket", "test.txt")
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("DownloadVersion", func(t *testing.T) {
		obj := &domain.Object{Bucket: "my-bucket", Key: "test.txt", VersionID: "v1", SizeBytes: 12}
		mockRepo.On("GetMetaByVersion", mock.Anything, "my-bucket", "test.txt", "v1").Return(obj, nil).Once()
		mockStore.On("Read", mock.Anything, "my-bucket", "test.txt?versionId=v1").Return(io.NopCloser(strings.NewReader("hello v1")), nil).Once()

		reader, meta, err := svc.DownloadVersion(ctx, "my-bucket", "test.txt", "v1")
		require.NoError(t, err)
		assert.NotNil(t, reader)
		assert.Equal(t, obj, meta)
		
		data, _ := io.ReadAll(reader)
		assert.Equal(t, "hello v1", string(data))
		mockRepo.AssertExpectations(t)
		mockStore.AssertExpectations(t)
	})

	t.Run("ListVersions", func(t *testing.T) {
		versions := []*domain.Object{{Key: "test.txt", VersionID: "v1"}, {Key: "test.txt", VersionID: "v2"}}
		mockRepo.On("ListVersions", mock.Anything, "my-bucket", "test.txt").Return(versions, nil).Once()

		res, err := svc.ListVersions(ctx, "my-bucket", "test.txt")
		require.NoError(t, err)
		assert.Equal(t, versions, res)
		mockRepo.AssertExpectations(t)
	})

	t.Run("DeleteVersion", func(t *testing.T) {
		obj := &domain.Object{Bucket: "my-bucket", Key: "test.txt", VersionID: "v1"}
		mockRepo.On("GetMetaByVersion", mock.Anything, "my-bucket", "test.txt", "v1").Return(obj, nil).Once()
		mockStore.On("Delete", mock.Anything, "my-bucket", "test.txt?versionId=v1").Return(nil).Once()
		mockRepo.On("DeleteVersion", mock.Anything, "my-bucket", "test.txt", "v1").Return(nil).Once()

		err := svc.DeleteVersion(ctx, "my-bucket", "test.txt", "v1")
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockStore.AssertExpectations(t)
	})

	t.Run("GetClusterStatus", func(t *testing.T) {
		status := &domain.StorageCluster{Nodes: []domain.StorageNode{{ID: "n1"}}}
		mockStore.On("GetClusterStatus", mock.Anything).Return(status, nil).Once()

		res, err := svc.GetClusterStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, status, res)
		mockStore.AssertExpectations(t)
	})

	t.Run("CleanupDeleted", func(t *testing.T) {
		deleted := []*domain.Object{{Bucket: "b1", Key: "k1", VersionID: "null"}}
		mockRepo.On("ListDeleted", mock.Anything, 10).Return(deleted, nil).Once()
		mockStore.On("Delete", mock.Anything, "b1", "k1").Return(nil).Once()
		mockRepo.On("HardDelete", mock.Anything, "b1", "k1", "null").Return(nil).Once()

		count, err := svc.CleanupDeleted(ctx, 10)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		mockRepo.AssertExpectations(t)
		mockStore.AssertExpectations(t)
	})

	t.Run("CleanupPendingUploads", func(t *testing.T) {
		pending := []*domain.Object{{Bucket: "b1", Key: "k1", VersionID: "null"}}
		mockRepo.On("ListPending", mock.Anything, mock.Anything, 10).Return(pending, nil).Once()
		mockStore.On("Delete", mock.Anything, "b1", "k1").Return(nil).Once()
		mockRepo.On("HardDelete", mock.Anything, "b1", "k1", "null").Return(nil).Once()

		count, err := svc.CleanupPendingUploads(ctx, time.Hour, 10)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		mockRepo.AssertExpectations(t)
		mockStore.AssertExpectations(t)
	})

	t.Run("Multipart Lifecycle", func(t *testing.T) {
		uploadID := uuid.New()
		bucket := "my-bucket"
		key := "large.file"

		// 1. Create
		mockRepo.On("GetBucket", mock.Anything, bucket).Return(&domain.Bucket{Name: bucket}, nil).Once()
		mockRepo.On("SaveMultipartUpload", mock.Anything, mock.MatchedBy(func(u *domain.MultipartUpload) bool {
			return u.Bucket == bucket && u.Key == key
		})).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "storage.multipart_init", "storage", mock.Anything, mock.Anything).Return(nil).Once()

		mu, err := svc.CreateMultipartUpload(ctx, bucket, key)
		require.NoError(t, err)
		assert.NotNil(t, mu)

		// 2. Upload Part
		mockRepo.On("GetMultipartUpload", mock.Anything, uploadID).Return(&domain.MultipartUpload{ID: uploadID, Bucket: bucket, Key: key}, nil).Once()
		mockStore.On("Write", mock.Anything, bucket, mock.Anything, mock.Anything).Return(int64(100), nil).Once()
		mockRepo.On("SavePart", mock.Anything, mock.MatchedBy(func(p *domain.Part) bool {
			return p.UploadID == uploadID && p.PartNumber == 1
		})).Return(nil).Once()

		part, err := svc.UploadPart(ctx, uploadID, 1, strings.NewReader("part data"), "")
		require.NoError(t, err)
		assert.NotNil(t, part)

		// 3. Complete
		mockRepo.On("GetMultipartUpload", mock.Anything, uploadID).Return(&domain.MultipartUpload{ID: uploadID, Bucket: bucket, Key: key, UserID: userID}, nil).Once()
		parts := []*domain.Part{{PartNumber: 1, SizeBytes: 100}}
		mockRepo.On("ListParts", mock.Anything, uploadID).Return(parts, nil).Once()
		mockRepo.On("GetBucket", mock.Anything, bucket).Return(&domain.Bucket{Name: bucket}, nil).Once()
		mockStore.On("Assemble", mock.Anything, bucket, key, mock.Anything).Return(int64(100), nil).Once()
		
		// Metadata extraction after assembly
		mockStore.On("Read", mock.Anything, bucket, key).Return(io.NopCloser(strings.NewReader("part data")), nil).Once()
		mockRepo.On("SaveMeta", mock.Anything, mock.MatchedBy(func(obj *domain.Object) bool {
			return obj.UploadStatus == domain.UploadStatusAvailable && obj.SizeBytes == 100
		})).Return(nil).Once()
		mockRepo.On("DeleteMultipartUpload", mock.Anything, uploadID).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "storage.multipart_complete", "storage", mock.Anything, mock.Anything).Return(nil).Once()

		obj, err := svc.CompleteMultipartUpload(ctx, uploadID)
		require.NoError(t, err)
		assert.NotNil(t, obj)
		assert.Equal(t, int64(100), obj.SizeBytes)

		mockRepo.AssertExpectations(t)
		mockStore.AssertExpectations(t)
		mockAuditSvc.AssertExpectations(t)
	})

	t.Run("AbortMultipartUpload", func(t *testing.T) {
		uploadID := uuid.New()
		bucket := "my-bucket"
		mockRepo.On("GetMultipartUpload", mock.Anything, uploadID).Return(&domain.MultipartUpload{ID: uploadID, Bucket: bucket}, nil).Once()
		parts := []*domain.Part{{PartNumber: 1}}
		mockRepo.On("ListParts", mock.Anything, uploadID).Return(parts, nil).Once()
		mockStore.On("Delete", mock.Anything, bucket, mock.Anything).Return(nil).Once()
		mockRepo.On("DeleteMultipartUpload", mock.Anything, uploadID).Return(nil).Once()

		err := svc.AbortMultipartUpload(ctx, uploadID)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockStore.AssertExpectations(t)
	})

	t.Run("GeneratePresignedURL", func(t *testing.T) {
		bucket := "my-bucket"
		key := "file.txt"
		mockRepo.On("GetBucket", mock.Anything, bucket).Return(&domain.Bucket{Name: bucket}, nil).Once()

		res, err := svc.GeneratePresignedURL(ctx, bucket, key, "GET", time.Hour)
		require.NoError(t, err)
		assert.NotNil(t, res)
		assert.Contains(t, res.URL, "/storage/presigned/my-bucket/file.txt")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error Path: Bucket Not Found", func(t *testing.T) {
		mockRepo.On("GetBucket", mock.Anything, "non-existent").Return(nil, fmt.Errorf("not found")).Once()
		_, err := svc.CreateMultipartUpload(ctx, "non-existent", "k")
		assert.Error(t, err)
	})

	t.Run("Error Path: Multipart Not Found", func(t *testing.T) {
		uploadID := uuid.New()
		mockRepo.On("GetMultipartUpload", mock.Anything, uploadID).Return(nil, fmt.Errorf("not found")).Once()
		_, err := svc.UploadPart(ctx, uploadID, 1, strings.NewReader("data"), "")
		assert.Error(t, err)
	})

	t.Run("Error Path: Abort Failure", func(t *testing.T) {
		uploadID := uuid.New()
		mockRepo.On("GetMultipartUpload", mock.Anything, uploadID).Return(&domain.MultipartUpload{ID: uploadID, Bucket: "b"}, nil).Once()
		mockRepo.On("ListParts", mock.Anything, uploadID).Return(nil, fmt.Errorf("list fail")).Once()
		mockRepo.On("DeleteMultipartUpload", mock.Anything, uploadID).Return(fmt.Errorf("delete fail")).Once()
		err := svc.AbortMultipartUpload(ctx, uploadID)
		assert.Error(t, err)
	})

	t.Run("Helper Coverage", func(t *testing.T) {
		// Reset mocks to ensure fresh state for this subtest
		mockRepo.ExpectedCalls = nil
		mockStore.ExpectedCalls = nil

		// Test DownloadVersion error paths
		mockRepo.On("GetMetaByVersion", mock.Anything, "b", "k", "v").Return(nil, fmt.Errorf("meta fail")).Once()
		_, _, err := svc.DownloadVersion(ctx, "b", "k", "v")
		assert.Error(t, err)

		mockRepo.On("GetMetaByVersion", mock.Anything, "b", "k", "v2").Return(&domain.Object{VersionID: "v2"}, nil).Once()
		mockStore.On("Read", mock.Anything, "b", "k?versionId=v2").Return(nil, fmt.Errorf("read fail")).Once()
		_, _, err = svc.DownloadVersion(ctx, "b", "k", "v2")
		assert.Error(t, err)

		// Test GetClusterStatus error path
		mockStore.On("GetClusterStatus", mock.Anything).Return(nil, fmt.Errorf("cluster fail")).Once()
		_, err = svc.GetClusterStatus(ctx)
		assert.Error(t, err)

		// Test CleanupDeleted error path
		mockRepo.On("ListDeleted", mock.Anything, 10).Return(nil, fmt.Errorf("list fail")).Once()
		_, err = svc.CleanupDeleted(ctx, 10)
		assert.Error(t, err)

		// Test CleanupPendingUploads error path
		mockRepo.On("ListPending", mock.Anything, mock.Anything, 10).Return(nil, fmt.Errorf("list fail")).Once()
		_, err = svc.CleanupPendingUploads(ctx, time.Hour, 10)
		assert.Error(t, err)
	})

	t.Run("Additional Error Paths", func(t *testing.T) {
		// Reset mocks to ensure fresh state for this subtest
		mockRepo.ExpectedCalls = nil
		mockStore.ExpectedCalls = nil

		// 1. Download: Bucket Not Found
		mockRepo.On("GetMeta", mock.Anything, "b", "k").Return(&domain.Object{Bucket: "b", Key: "k", VersionID: "null"}, nil).Once()
		mockStore.On("Read", mock.Anything, "b", "k").Return(io.NopCloser(strings.NewReader("data")), nil).Once()
		mockRepo.On("GetBucket", mock.Anything, "b").Return(nil, fmt.Errorf("bucket error")).Once()
		_, _, err := svc.Download(ctx, "b", "k")
		assert.Error(t, err)

		// 2. Download: Store Read Error
		mockRepo.On("GetMeta", mock.Anything, "b", "k").Return(&domain.Object{Bucket: "b", Key: "k", VersionID: "null"}, nil).Once()
		mockRepo.On("GetBucket", mock.Anything, "b").Return(&domain.Bucket{Name: "b"}, nil).Once()
		mockStore.On("Read", mock.Anything, "b", "k").Return(nil, fmt.Errorf("read error")).Once()
		_, _, err = svc.Download(ctx, "b", "k")
		assert.Error(t, err)

		// 3. DeleteVersion: Store Delete Error
		mockRepo.On("GetMetaByVersion", mock.Anything, "b", "k", "v1").Return(&domain.Object{Bucket: "b", Key: "k", VersionID: "v1"}, nil).Once()
		mockStore.On("Delete", mock.Anything, "b", "k?versionId=v1").Return(fmt.Errorf("delete error")).Once()
		err = svc.DeleteVersion(ctx, "b", "k", "v1")
		assert.Error(t, err)

		// 4. DeleteObject: SoftDelete Error
		mockRepo.On("SoftDelete", mock.Anything, "b", "k").Return(fmt.Errorf("repo error")).Once()
		err = svc.DeleteObject(ctx, "b", "k")
		assert.Error(t, err)

		// 5. CreateMultipartUpload: SaveMultipartUpload Error
		mockRepo.On("GetBucket", mock.Anything, "b").Return(&domain.Bucket{Name: "b"}, nil).Once()
		mockRepo.On("SaveMultipartUpload", mock.Anything, mock.Anything).Return(fmt.Errorf("save error")).Once()
		_, err = svc.CreateMultipartUpload(ctx, "b", "k")
		assert.Error(t, err)

		// 6. UploadPart: Checksum Mismatch
		uploadID := uuid.New()
		mockRepo.On("GetMultipartUpload", mock.Anything, uploadID).Return(&domain.MultipartUpload{Bucket: "b", Key: "k"}, nil).Once()
		mockStore.On("Write", mock.Anything, "b", mock.Anything, mock.Anything).Return(int64(10), nil).Once()
		mockStore.On("Delete", mock.Anything, "b", mock.Anything).Return(nil).Once()
		_, err = svc.UploadPart(ctx, uploadID, 1, strings.NewReader("some data"), "wrong-checksum")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "part integrity check failed")

		// 7. CompleteMultipartUpload: Store Assemble Error
		mockRepo.ExpectedCalls = nil
		mockStore.ExpectedCalls = nil
		mockRepo.On("GetMultipartUpload", mock.Anything, uploadID).Return(&domain.MultipartUpload{Bucket: "b", Key: "k", UserID: userID}, nil).Once()
		mockRepo.On("ListParts", mock.Anything, uploadID).Return([]*domain.Part{{PartNumber: 1}}, nil).Once()
		mockRepo.On("GetBucket", mock.Anything, "b").Return(&domain.Bucket{Name: "b"}, nil).Once()
		mockStore.On("Assemble", mock.Anything, "b", "k", mock.Anything).Return(int64(0), fmt.Errorf("assemble error")).Once()
		_, err = svc.CompleteMultipartUpload(ctx, uploadID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to assemble parts")

		// 8. GeneratePresignedURL: GetBucket Error
		mockRepo.On("GetBucket", mock.Anything, "b").Return(nil, fmt.Errorf("repo error")).Once()
		_, err = svc.GeneratePresignedURL(ctx, "b", "k", "GET", time.Hour)
		assert.Error(t, err)
	})

	t.Run("Version Generation", func(t *testing.T) {
		bucket := &domain.Bucket{Name: "v-bucket", VersioningEnabled: true}
		mockRepo.On("GetBucket", mock.Anything, "v-bucket").Return(bucket, nil).Once()
		mockStore.On("Write", mock.Anything, "v-bucket", mock.MatchedBy(func(k string) bool {
			return strings.Contains(k, "?versionId=")
		}), mock.Anything).Return(int64(5), nil).Once()
		mockRepo.On("SaveMeta", mock.Anything, mock.Anything).Return(nil).Twice()
		mockAuditSvc.On("Log", mock.Anything, userID, "storage.object_upload", "storage", mock.Anything, mock.Anything).Return(nil).Once()

		obj, err := svc.Upload(ctx, "v-bucket", "k", strings.NewReader("data"), "")
		require.NoError(t, err)
		assert.NotEmpty(t, obj.VersionID)
		assert.NotEqual(t, "null", obj.VersionID)
	})
}
