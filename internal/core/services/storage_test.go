package services_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStorageUpload_Success(t *testing.T) {
	repo := new(MockStorageRepo)
	store := new(MockFileStore)
	auditSvc := new(services.MockAuditService)
	svc := services.NewStorageService(repo, store, auditSvc)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	bucket := "test-bucket"
	key := "test-key"
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

	repo.AssertExpectations(t)
	store.AssertExpectations(t)
}

func TestStorageDownload_Success(t *testing.T) {
	repo := new(MockStorageRepo)
	store := new(MockFileStore)
	auditSvc := new(services.MockAuditService)
	svc := services.NewStorageService(repo, store, auditSvc)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	bucket := "test-bucket"
	key := "test-key"
	meta := &domain.Object{Bucket: bucket, Key: key}
	content := io.NopCloser(strings.NewReader("data"))

	repo.On("GetMeta", ctx, bucket, key).Return(meta, nil)
	store.On("Read", ctx, bucket, key).Return(content, nil)
	auditSvc.On("Log", ctx, userID, "storage.object_download", "storage", mock.Anything, mock.Anything).Return(nil)

	r, obj, err := svc.Download(ctx, bucket, key)

	assert.NoError(t, err)
	assert.Equal(t, meta, obj)
	assert.NotNil(t, r)

	repo.AssertExpectations(t)
	store.AssertExpectations(t)
}

func TestStorageDelete_Success(t *testing.T) {
	repo := new(MockStorageRepo)
	store := new(MockFileStore)
	auditSvc := new(services.MockAuditService)
	svc := services.NewStorageService(repo, store, auditSvc)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	bucket := "test-bucket"
	key := "test-key"

	repo.On("SoftDelete", ctx, bucket, key).Return(nil)
	auditSvc.On("Log", ctx, userID, "storage.object_delete", "storage", mock.Anything, mock.Anything).Return(nil)

	err := svc.DeleteObject(ctx, bucket, key)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestStorageList_Success(t *testing.T) {
	repo := new(MockStorageRepo)
	store := new(MockFileStore)
	auditSvc := new(services.MockAuditService)
	svc := services.NewStorageService(repo, store, auditSvc)

	ctx := context.Background()
	bucket := "test-bucket"
	expected := []*domain.Object{{Key: "k1"}, {Key: "k2"}}

	repo.On("List", ctx, bucket).Return(expected, nil)

	list, err := svc.ListObjects(ctx, bucket)

	assert.NoError(t, err)
	assert.Equal(t, expected, list)
	repo.AssertExpectations(t)
}
