package workers

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

type fakeStorageService struct {
	clusterStatus *domain.StorageCluster
	statusErr     error
	callsMu       sync.Mutex
	statusCalls   int
}

func (f *fakeStorageService) GetClusterStatus(ctx context.Context) (*domain.StorageCluster, error) {
	f.callsMu.Lock()
	defer f.callsMu.Unlock()
	f.statusCalls++
	return f.clusterStatus, f.statusErr
}

func (f *fakeStorageService) StatusCalls() int {
	f.callsMu.Lock()
	defer f.callsMu.Unlock()
	return f.statusCalls
}

func (f *fakeStorageService) Upload(ctx context.Context, bucket, key string, r io.Reader) (*domain.Object, error) {
	return nil, nil
}
func (f *fakeStorageService) Download(ctx context.Context, bucket, key string) (io.ReadCloser, *domain.Object, error) {
	return nil, nil, nil
}
func (f *fakeStorageService) ListObjects(ctx context.Context, bucket string) ([]*domain.Object, error) {
	return nil, nil
}
func (f *fakeStorageService) DeleteObject(ctx context.Context, bucket, key string) error {
	return nil
}
func (f *fakeStorageService) DownloadVersion(ctx context.Context, bucket, key, versionID string) (io.ReadCloser, *domain.Object, error) {
	return nil, nil, nil
}
func (f *fakeStorageService) ListVersions(ctx context.Context, bucket, key string) ([]*domain.Object, error) {
	return nil, nil
}
func (f *fakeStorageService) DeleteVersion(ctx context.Context, bucket, key, versionID string) error {
	return nil
}
func (f *fakeStorageService) CreateBucket(ctx context.Context, name string, isPublic bool) (*domain.Bucket, error) {
	return nil, nil
}
func (f *fakeStorageService) GetBucket(ctx context.Context, name string) (*domain.Bucket, error) {
	return nil, nil
}
func (f *fakeStorageService) DeleteBucket(ctx context.Context, name string) error {
	return nil
}
func (f *fakeStorageService) ListBuckets(ctx context.Context) ([]*domain.Bucket, error) {
	return nil, nil
}
func (f *fakeStorageService) SetBucketVersioning(ctx context.Context, name string, enabled bool) error {
	return nil
}
func (f *fakeStorageService) CreateMultipartUpload(ctx context.Context, bucket, key string) (*domain.MultipartUpload, error) {
	return nil, nil
}
func (f *fakeStorageService) UploadPart(ctx context.Context, uploadID uuid.UUID, partNumber int, r io.Reader) (*domain.Part, error) {
	return nil, nil
}
func (f *fakeStorageService) CompleteMultipartUpload(ctx context.Context, uploadID uuid.UUID) (*domain.Object, error) {
	return nil, nil
}
func (f *fakeStorageService) AbortMultipartUpload(ctx context.Context, uploadID uuid.UUID) error {
	return nil
}
func (f *fakeStorageService) GeneratePresignedURL(ctx context.Context, bucket, key, method string, expiry time.Duration) (*domain.PresignedURL, error) {
	return nil, nil
}

func TestMetricsCollectorWorkerRun(t *testing.T) {
	svc := &fakeStorageService{clusterStatus: &domain.StorageCluster{}}
	worker := &MetricsCollectorWorker{
		storageRepo: nil,
		storageSvc:  svc,
		logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
		interval:    10 * time.Millisecond,
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go worker.Run(ctx, &wg)

	time.Sleep(35 * time.Millisecond)
	cancel()
	wg.Wait()

	if svc.StatusCalls() == 0 {
		t.Fatalf("expected GetClusterStatus to be called")
	}
}

func TestMetricsCollectorWorkerCollectMetricsHandlesError(t *testing.T) {
	svc := &fakeStorageService{statusErr: io.EOF}
	worker := &MetricsCollectorWorker{
		storageRepo: nil,
		storageSvc:  svc,
		logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	worker.collectMetrics(context.Background())

	if svc.StatusCalls() != 1 {
		t.Fatalf("expected GetClusterStatus to be called once")
	}
}

func TestMetricsCollectorWorkerCollectMetricsCountsUpNodes(t *testing.T) {
	svc := &fakeStorageService{clusterStatus: &domain.StorageCluster{
		Nodes: []domain.StorageNode{
			{Status: "up"},
			{Status: "alive"},
			{Status: "down"},
		},
	}}
	worker := &MetricsCollectorWorker{
		storageRepo: nil,
		storageSvc:  svc,
		logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	worker.collectMetrics(context.Background())

	if svc.StatusCalls() != 1 {
		t.Fatalf("expected GetClusterStatus to be called once")
	}
}
