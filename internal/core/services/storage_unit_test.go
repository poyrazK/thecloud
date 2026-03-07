package services_test

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

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

		obj, err := svc.Upload(ctx, "my-bucket", "test.txt", strings.NewReader("hello world!"))
		require.NoError(t, err)
		assert.NotNil(t, obj)
		assert.Equal(t, int64(12), obj.SizeBytes)
		assert.Equal(t, "text/plain; charset=utf-8", obj.ContentType)
		assert.NotEmpty(t, obj.Checksum)

		mockRepo.AssertExpectations(t)
		mockStore.AssertExpectations(t)
		mockAuditSvc.AssertExpectations(t)
	})
}
