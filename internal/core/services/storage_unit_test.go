package services_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStorageService_Unit(t *testing.T) {
	mockRepo := new(MockStorageRepo)
	mockStore := new(MockFileStore)
	mockAuditSvc := new(MockAuditService)
	cfg := &platform.Config{SecretsEncryptionKey: "test-secret-key-32-chars-long-!!!"}
	svc := services.NewStorageService(mockRepo, mockStore, mockAuditSvc, nil, cfg)

	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)

	t.Run("CreateBucket", func(t *testing.T) {
		mockRepo.On("CreateBucket", mock.Anything, mock.Anything).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "storage.bucket_create", "bucket", mock.Anything, mock.Anything).Return(nil).Once()

		bucket, err := svc.CreateBucket(ctx, "my-bucket", false)
		assert.NoError(t, err)
		assert.NotNil(t, bucket)
		assert.Equal(t, "my-bucket", bucket.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Upload", func(t *testing.T) {
		bucket := &domain.Bucket{Name: "my-bucket", VersioningEnabled: false}
		mockRepo.On("GetBucket", mock.Anything, "my-bucket").Return(bucket, nil).Once()
		mockStore.On("Write", mock.Anything, "my-bucket", "test.txt", mock.Anything).Return(int64(12), nil).Once()
		mockRepo.On("SaveMeta", mock.Anything, mock.Anything).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "storage.object_upload", "storage", mock.Anything, mock.Anything).Return(nil).Once()

		obj, err := svc.Upload(ctx, "my-bucket", "test.txt", strings.NewReader("hello world!"))
		assert.NoError(t, err)
		assert.NotNil(t, obj)
		assert.Equal(t, int64(12), obj.SizeBytes)
	})
}
