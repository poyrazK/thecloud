package workers

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/mock"
)

type mockStorageService struct {
	mock.Mock
}

func (m *mockStorageService) CleanupDeleted(ctx context.Context, limit int) (int, error) {
	args := m.Called(ctx, limit)
	return args.Int(0), args.Error(1)
}

func (m *mockStorageService) Upload(ctx context.Context, bucket, key string, r io.Reader) (*domain.Object, error) { return nil, nil }
func (m *mockStorageService) Download(ctx context.Context, bucket, key string) (io.ReadCloser, *domain.Object, error) { return nil, nil, nil }
func (m *mockStorageService) ListObjects(ctx context.Context, bucket string) ([]*domain.Object, error) { return nil, nil }
func (m *mockStorageService) DeleteObject(ctx context.Context, bucket, key string) error { return nil }
func (m *mockStorageService) DownloadVersion(ctx context.Context, bucket, key, versionID string) (io.ReadCloser, *domain.Object, error) { return nil, nil, nil }
func (m *mockStorageService) ListVersions(ctx context.Context, bucket, key string) ([]*domain.Object, error) {
	return nil, nil
}
func (m *mockStorageService) DeleteVersion(ctx context.Context, bucket, key, versionID string) error {
	return nil
}
func (m *mockStorageService) CreateBucket(ctx context.Context, name string, isPublic bool) (*domain.Bucket, error) {
	return nil, nil
}
func (m *mockStorageService) GetBucket(ctx context.Context, name string) (*domain.Bucket, error) {
	return nil, nil
}
func (m *mockStorageService) DeleteBucket(ctx context.Context, name string, force bool) error { return nil }
func (m *mockStorageService) ListBuckets(ctx context.Context) ([]*domain.Bucket, error) {
	return nil, nil
}

func (m *mockStorageService) SetBucketVersioning(ctx context.Context, name string, enabled bool) error { return nil }
func (m *mockStorageService) GetClusterStatus(ctx context.Context) (*domain.StorageCluster, error) { return nil, nil }
func (m *mockStorageService) CreateMultipartUpload(ctx context.Context, bucket, key string) (*domain.MultipartUpload, error) { return nil, nil }
func (m *mockStorageService) UploadPart(ctx context.Context, uploadID uuid.UUID, partNumber int, r io.Reader) (*domain.Part, error) { return nil, nil }
func (m *mockStorageService) CompleteMultipartUpload(ctx context.Context, uploadID uuid.UUID) (*domain.Object, error) { return nil, nil }
func (m *mockStorageService) AbortMultipartUpload(ctx context.Context, uploadID uuid.UUID) error { return nil }
func (m *mockStorageService) GeneratePresignedURL(ctx context.Context, bucket, key, method string, expiry time.Duration) (*domain.PresignedURL, error) { return nil, nil }

func TestStorageCleanupWorker_Cleanup(t *testing.T) {
	svc := new(mockStorageService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	worker := NewStorageCleanupWorker(svc, logger)
	worker.batchSize = 2

	t.Run("SingleBatch", func(t *testing.T) {
		svc.On("CleanupDeleted", mock.Anything, 2).Return(1, nil).Once()
		worker.cleanup(context.Background())
		svc.AssertExpectations(t)
	})

	t.Run("MultipleBatches", func(t *testing.T) {
		svc.On("CleanupDeleted", mock.Anything, 2).Return(2, nil).Once()
		svc.On("CleanupDeleted", mock.Anything, 2).Return(1, nil).Once()
		worker.cleanup(context.Background())
		svc.AssertExpectations(t)
	})
}
