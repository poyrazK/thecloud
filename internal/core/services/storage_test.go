package services_test

import (
	"context"
	"strings"
	"testing"

	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/platform"
	"github.com/poyrazk/thecloud/internal/repositories/noop"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
)

func setupStorageServiceIntegrationTest(t *testing.T) (ports.StorageService, ports.StorageRepository, context.Context) {
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewStorageRepository(db)
	store := &noop.NoopFileStore{}
	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(auditRepo)
	cfg := &platform.Config{SecretsEncryptionKey: "test-secret", Port: "8080"}

	svc := services.NewStorageService(repo, store, auditSvc, nil, cfg)

	return svc, repo, ctx
}

func TestStorageService_Integration(t *testing.T) {
	svc, _, ctx := setupStorageServiceIntegrationTest(t)
	userID := appcontext.UserIDFromContext(ctx)

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

		// Delete
		err = svc.DeleteObject(ctx, bucketName, key)
		assert.NoError(t, err)
	})

	t.Run("MultipartUpload", func(t *testing.T) {
		bucketName := "mp-bucket"
		_, _ = svc.CreateBucket(ctx, bucketName, false)
		key := "large.zip"

		// Init
		upload, err := svc.CreateMultipartUpload(ctx, bucketName, key)
		assert.NoError(t, err)
		assert.NotNil(t, upload)

		// Upload Parts
		part1, err := svc.UploadPart(ctx, upload.ID, 1, strings.NewReader("part1"))
		assert.NoError(t, err)
		assert.Equal(t, 1, part1.PartNumber)

		part2, err := svc.UploadPart(ctx, upload.ID, 2, strings.NewReader("part2"))
		assert.NoError(t, err)
		assert.Equal(t, 2, part2.PartNumber)

		// Complete
		obj, err := svc.CompleteMultipartUpload(ctx, upload.ID)
		assert.NoError(t, err)
		assert.NotNil(t, obj)
		assert.Equal(t, key, obj.Key)
	})

	t.Run("PresignedURL", func(t *testing.T) {
		bucketName := "url-bucket"
		_, _ = svc.CreateBucket(ctx, bucketName, false)

		url, err := svc.GeneratePresignedURL(ctx, bucketName, "file.txt", "GET", 0)
		assert.NoError(t, err)
		assert.NotEmpty(t, url.URL)
		assert.Equal(t, "GET", url.Method)
	})
}
