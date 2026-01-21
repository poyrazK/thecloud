package services_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
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
	svc := services.NewStorageService(repo, store, auditSvc, nil)
	return repo, store, auditSvc, svc
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
	repo.On("SaveMeta", ctx, mock.AnythingOfType("*domain.Object")).Return(nil)
	auditSvc.On("Log", ctx, userID, "storage.object_upload", "storage", mock.Anything, mock.Anything).Return(nil)

	obj, err := svc.Upload(ctx, bucket, key, reader)

	assert.NoError(t, err)
	assert.NotNil(t, obj)
	assert.Equal(t, bucket, obj.Bucket)
	assert.Equal(t, key, obj.Key)
	assert.Equal(t, int64(len(content)), obj.SizeBytes)
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
