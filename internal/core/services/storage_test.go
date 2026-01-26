package services_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testBucket = "test-bucket"
	testKey    = "test-key"
)

func setupStorageServiceTest(_ *testing.T) (*MockStorageRepo, *MockFileStore, *MockAuditService, ports.StorageService) {
	repo := new(MockStorageRepo)
	store := new(MockFileStore)
	auditSvc := new(MockAuditService)
	svc := services.NewStorageService(repo, store, auditSvc, nil, nil)
	return repo, store, auditSvc, svc
}

func setupStorageServiceWithEncryption(_ *testing.T) (*MockStorageRepo, *MockFileStore, *MockAuditService, *MockEncryptionService, ports.StorageService) {
	repo := new(MockStorageRepo)
	store := new(MockFileStore)
	auditSvc := new(MockAuditService)
	encryptSvc := new(MockEncryptionService)
	svc := services.NewStorageService(repo, store, auditSvc, encryptSvc, nil)
	return repo, store, auditSvc, encryptSvc, svc
}

func TestStorageUploadSuccess(t *testing.T) {
	repo, store, auditSvc, svc := setupStorageServiceTest(t)
	defer repo.AssertExpectations(t)
	defer store.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	bucket := testBucket
	key := testKey
	content := "hello world"
	reader := strings.NewReader(content)

	store.On("Write", ctx, bucket, key, reader).Return(int64(len(content)), nil)
	repo.On("GetBucket", ctx, bucket).Return(&domain.Bucket{Name: bucket, VersioningEnabled: false, UserID: userID}, nil)
	repo.On("SaveMeta", ctx, mock.AnythingOfType("*domain.Object")).Return(nil)
	auditSvc.On("Log", ctx, userID, "storage.object_upload", "storage", mock.Anything, mock.Anything).Return(nil)

	obj, err := svc.Upload(ctx, bucket, key, reader)

	assert.NoError(t, err)
	assert.NotNil(t, obj)
	assert.Equal(t, bucket, obj.Bucket)
	assert.Equal(t, key, obj.Key)
	assert.Equal(t, int64(len(content)), obj.SizeBytes)
}

func TestStorageUploadBucketNotFound(t *testing.T) {
	repo, _, _, svc := setupStorageServiceTest(t)
	defer repo.AssertExpectations(t)

	ctx := context.Background()
	repo.On("GetBucket", ctx, "missing").Return(nil, errors.New(errors.NotFound, "bucket not found")).Once()

	obj, err := svc.Upload(ctx, "missing", "key", strings.NewReader("data"))
	assert.Error(t, err)
	assert.Nil(t, obj)
}

func TestStorageDownloadSuccess(t *testing.T) {
	repo, store, auditSvc, svc := setupStorageServiceTest(t)
	defer repo.AssertExpectations(t)
	defer store.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	bucket := testBucket
	key := testKey
	meta := &domain.Object{Bucket: bucket, Key: key}
	content := io.NopCloser(strings.NewReader("data"))

	repo.On("GetMeta", ctx, bucket, key).Return(meta, nil)
	store.On("Read", ctx, bucket, key).Return(content, nil)
	repo.On("GetBucket", ctx, bucket).Return(&domain.Bucket{Name: bucket, EncryptionEnabled: false, UserID: userID}, nil)

	r, obj, err := svc.Download(ctx, bucket, key)

	assert.NoError(t, err)
	assert.Equal(t, meta, obj)
	assert.NotNil(t, r)
}

func TestStorageDeleteSuccess(t *testing.T) {
	repo, _, auditSvc, svc := setupStorageServiceTest(t)
	defer repo.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	bucket := testBucket
	key := testKey

	repo.On("SoftDelete", ctx, bucket, key).Return(nil)
	auditSvc.On("Log", ctx, userID, "storage.object_delete", "storage", mock.Anything, mock.Anything).Return(nil)

	err := svc.DeleteObject(ctx, bucket, key)

	assert.NoError(t, err)
}

func TestStorageDownloadNotFound(t *testing.T) {
	repo, _, _, svc := setupStorageServiceTest(t)
	defer repo.AssertExpectations(t)

	ctx := context.Background()
	repo.On("GetMeta", ctx, testBucket, testKey).Return(nil, errors.New(errors.NotFound, "not found")).Once()

	r, obj, err := svc.Download(ctx, testBucket, testKey)
	assert.Error(t, err)
	assert.Nil(t, r)
	assert.Nil(t, obj)
}

func TestStorageListSuccess(t *testing.T) {
	repo, _, _, svc := setupStorageServiceTest(t)
	defer repo.AssertExpectations(t)

	ctx := context.Background()
	bucket := testBucket
	expected := []*domain.Object{{Key: "k1"}, {Key: "k2"}}

	repo.On("List", ctx, bucket).Return(expected, nil)

	list, err := svc.ListObjects(ctx, bucket)

	assert.NoError(t, err)
	assert.Equal(t, expected, list)
}

func TestStorageVersioning(t *testing.T) {
	repo, store, auditSvc, svc := setupStorageServiceTest(t)
	defer repo.AssertExpectations(t)
	defer store.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	bucket := testBucket
	key := testKey
	versionID := "v1"

	t.Run("UploadVersioned", func(t *testing.T) {
		repo.On("GetBucket", mock.Anything, bucket).Return(&domain.Bucket{
			Name:              bucket,
			VersioningEnabled: true,
			UserID:            userID,
		}, nil).Once()

		// Matching the regex/format used in service for versioned store key
		store.On("Write", mock.Anything, bucket, mock.MatchedBy(func(k string) bool {
			return strings.HasPrefix(k, key+"?versionId=")
		}), mock.Anything).Return(int64(5), nil).Once()

		repo.On("SaveMeta", mock.Anything, mock.MatchedBy(func(o *domain.Object) bool {
			return o.Bucket == bucket && o.VersionID != "null"
		})).Return(nil).Once()

		auditSvc.On("Log", mock.Anything, userID, "storage.object_upload", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		obj, err := svc.Upload(ctx, bucket, key, strings.NewReader("hello"))
		assert.NoError(t, err)
		assert.NotEqual(t, "null", obj.VersionID)
	})

	t.Run("DownloadVersion", func(t *testing.T) {
		obj := &domain.Object{Bucket: bucket, Key: key, VersionID: versionID}
		repo.On("GetMetaByVersion", ctx, bucket, key, versionID).Return(obj, nil).Once()

		storeKey := key + "?versionId=" + versionID
		store.On("Read", ctx, bucket, storeKey).Return(io.NopCloser(strings.NewReader("data")), nil).Once()

		r, meta, err := svc.DownloadVersion(ctx, bucket, key, versionID)
		assert.NoError(t, err)
		assert.Equal(t, versionID, meta.VersionID)
		assert.NotNil(t, r)
	})

	t.Run("DownloadVersionNotFound", func(t *testing.T) {
		repo.On("GetMetaByVersion", ctx, bucket, key, "v2").Return(nil, errors.New(errors.NotFound, "not found")).Once()
		_, _, err := svc.DownloadVersion(ctx, bucket, key, "v2")
		assert.Error(t, err)
	})
}

func TestStorageMultipart(t *testing.T) {
	repo, store, auditSvc, svc := setupStorageServiceTest(t)
	defer repo.AssertExpectations(t)
	defer store.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	bucket := testBucket
	key := testKey
	uploadID := uuid.New()

	t.Run("CreateMultipart", func(t *testing.T) {
		repo.On("GetBucket", ctx, bucket).Return(&domain.Bucket{Name: bucket}, nil).Once()
		repo.On("SaveMultipartUpload", ctx, mock.MatchedBy(func(u *domain.MultipartUpload) bool {
			return u.Bucket == bucket && u.Key == key
		})).Return(nil).Once()
		auditSvc.On("Log", ctx, userID, "storage.multipart_init", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		u, err := svc.CreateMultipartUpload(ctx, bucket, key)
		assert.NoError(t, err)
		assert.NotNil(t, u)
	})

	t.Run("UploadPart", func(t *testing.T) {
		upload := &domain.MultipartUpload{ID: uploadID, Bucket: bucket, Key: key}
		repo.On("GetMultipartUpload", ctx, uploadID).Return(upload, nil).Once()

		store.On("Write", ctx, bucket, mock.Anything, mock.Anything).Return(int64(10), nil).Once()
		repo.On("SavePart", ctx, mock.Anything).Return(nil).Once()

		part, err := svc.UploadPart(ctx, uploadID, 1, strings.NewReader("part-data"))
		assert.NoError(t, err)
		assert.Equal(t, 1, part.PartNumber)
	})

	t.Run("CompleteMultipart", func(t *testing.T) {
		upload := &domain.MultipartUpload{ID: uploadID, Bucket: bucket, Key: key, UserID: userID}
		repo.On("GetMultipartUpload", ctx, uploadID).Return(upload, nil).Once()

		parts := []*domain.Part{{PartNumber: 1}, {PartNumber: 2}}
		repo.On("ListParts", ctx, uploadID).Return(parts, nil).Once()

		store.On("Assemble", ctx, bucket, key, mock.Anything).Return(int64(20), nil).Once()
		repo.On("SaveMeta", ctx, mock.Anything).Return(nil).Once()
		repo.On("DeleteMultipartUpload", ctx, uploadID).Return(nil).Once()
		auditSvc.On("Log", ctx, userID, "storage.multipart_complete", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		obj, err := svc.CompleteMultipartUpload(ctx, uploadID)
		assert.NoError(t, err)
		assert.Equal(t, key, obj.Key)
	})

	t.Run("AbortMultipart", func(t *testing.T) {
		upload := &domain.MultipartUpload{ID: uploadID, Bucket: bucket, Key: key}
		repo.On("GetMultipartUpload", ctx, uploadID).Return(upload, nil).Once()
		repo.On("ListParts", ctx, uploadID).Return([]*domain.Part{}, nil).Once()
		repo.On("DeleteMultipartUpload", ctx, uploadID).Return(nil).Once()
		auditSvc.On("Log", ctx, userID, "storage.multipart_abort", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		err := svc.AbortMultipartUpload(ctx, uploadID)
		assert.NoError(t, err)
	})

	t.Run("UploadPartNotFound", func(t *testing.T) {
		repo.On("GetMultipartUpload", ctx, uploadID).Return(nil, errors.New(errors.NotFound, "not found")).Once()
		_, err := svc.UploadPart(ctx, uploadID, 1, nil)
		assert.Error(t, err)
	})
}

func TestStorageBucketOps(t *testing.T) {
	repo, store, auditSvc, svc := setupStorageServiceTest(t)
	defer repo.AssertExpectations(t)
	defer store.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	bucket := "my-bucket"

	t.Run("CreateBucket", func(t *testing.T) {
		repo.On("CreateBucket", ctx, mock.MatchedBy(func(b *domain.Bucket) bool {
			return b.Name == bucket && b.UserID == userID
		})).Return(nil).Once()
		auditSvc.On("Log", ctx, userID, "storage.bucket_create", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		b, err := svc.CreateBucket(ctx, bucket, false)
		assert.NoError(t, err)
		assert.Equal(t, bucket, b.Name)
	})

	t.Run("DeleteBucket", func(t *testing.T) {
		repo.On("DeleteBucket", ctx, bucket).Return(nil).Once()
		err := svc.DeleteBucket(ctx, bucket)
		assert.NoError(t, err)
	})

	t.Run("SetVersioning", func(t *testing.T) {
		repo.On("SetBucketVersioning", ctx, bucket, true).Return(nil).Once()
		err := svc.SetBucketVersioning(ctx, bucket, true)
		assert.NoError(t, err)
	})

	t.Run("ListBuckets", func(t *testing.T) {
		repo.On("ListBuckets", ctx, userID.String()).Return([]*domain.Bucket{{Name: "b1"}}, nil).Once()
		buckets, err := svc.ListBuckets(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(buckets))
	})

	t.Run("GetClusterStatus", func(t *testing.T) {
		store.On("GetClusterStatus", ctx).Return(&domain.StorageCluster{}, nil).Once()
		status, err := svc.GetClusterStatus(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, status)
	})
}

func TestStoragePresignedURL(t *testing.T) {
	repo, _, _, svc := setupStorageServiceTest(t)
	defer repo.AssertExpectations(t)

	ctx := context.Background()
	bucket := testBucket
	key := testKey

	repo.On("GetBucket", ctx, bucket).Return(&domain.Bucket{Name: bucket}, nil).Once()

	url, err := svc.GeneratePresignedURL(ctx, bucket, key, "GET", 0)
	assert.NoError(t, err)
	assert.NotEmpty(t, url.URL)
	assert.Contains(t, url.URL, bucket)
	assert.Contains(t, url.URL, key)
	assert.Equal(t, "GET", url.Method)
}

func TestStorageEncryption(t *testing.T) {
	repo, store, auditSvc, encryptSvc, svc := setupStorageServiceWithEncryption(t)
	defer repo.AssertExpectations(t)
	defer store.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)
	defer encryptSvc.AssertExpectations(t)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	bucket := testBucket
	key := testKey
	plaintext := []byte("secret")
	encrypted := []byte("encrypted-data")

	t.Run("UploadEncrypted", func(t *testing.T) {
		repo.On("GetBucket", ctx, bucket).Return(&domain.Bucket{
			Name:              bucket,
			EncryptionEnabled: true,
			UserID:            userID,
		}, nil).Once()

		encryptSvc.On("Encrypt", ctx, bucket, plaintext).Return(encrypted, nil).Once()

		store.On("Write", ctx, bucket, key, mock.MatchedBy(func(r io.Reader) bool {
			data, _ := io.ReadAll(r)
			return string(data) == string(encrypted)
		})).Return(int64(len(encrypted)), nil).Once()

		repo.On("SaveMeta", ctx, mock.Anything).Return(nil).Once()
		auditSvc.On("Log", ctx, userID, "storage.object_upload", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		obj, err := svc.Upload(ctx, bucket, key, bytes.NewReader(plaintext))
		assert.NoError(t, err)
		assert.NotNil(t, obj)
	})

	t.Run("DownloadEncrypted", func(t *testing.T) {
		repo.On("GetMeta", ctx, bucket, key).Return(&domain.Object{Bucket: bucket, Key: key}, nil).Once()

		repo.On("GetBucket", ctx, bucket).Return(&domain.Bucket{
			Name:              bucket,
			EncryptionEnabled: true,
		}, nil).Once()

		store.On("Read", ctx, bucket, key).Return(io.NopCloser(bytes.NewReader(encrypted)), nil).Once()

		encryptSvc.On("Decrypt", ctx, bucket, encrypted).Return(plaintext, nil).Once()

		r, _, err := svc.Download(ctx, bucket, key)
		assert.NoError(t, err)
		data, _ := io.ReadAll(r)
		assert.Equal(t, plaintext, data)
	})
}
