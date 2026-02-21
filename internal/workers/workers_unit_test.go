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
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Minimal local mocks
type mockAccountingSvc struct{ mock.Mock }
func (m *mockAccountingSvc) TrackUsage(ctx context.Context, record domain.UsageRecord) error { return nil }
func (m *mockAccountingSvc) GetSummary(ctx context.Context, userID uuid.UUID, start, end time.Time) (*domain.BillSummary, error) { return nil, nil }
func (m *mockAccountingSvc) ListUsage(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]domain.UsageRecord, error) { return nil, nil }
func (m *mockAccountingSvc) ProcessHourlyBilling(ctx context.Context) error { return m.Called(ctx).Error(0) }

type mockStorageRepo struct{ mock.Mock }
func (m *mockStorageRepo) SaveMeta(ctx context.Context, obj *domain.Object) error { return nil }
func (m *mockStorageRepo) GetMeta(ctx context.Context, bucket, key string) (*domain.Object, error) { return nil, nil }
func (m *mockStorageRepo) List(ctx context.Context, bucket string) ([]*domain.Object, error) { return nil, nil }
func (m *mockStorageRepo) SoftDelete(ctx context.Context, bucket, key string) error { return nil }
func (m *mockStorageRepo) DeleteVersion(ctx context.Context, bucket, key, versionID string) error { return nil }
func (m *mockStorageRepo) GetMetaByVersion(ctx context.Context, bucket, key, versionID string) (*domain.Object, error) { return nil, nil }
func (m *mockStorageRepo) ListVersions(ctx context.Context, bucket, key string) ([]*domain.Object, error) { return nil, nil }
func (m *mockStorageRepo) ListDeleted(ctx context.Context, limit int) ([]*domain.Object, error) { return nil, nil }
func (m *mockStorageRepo) HardDelete(ctx context.Context, bucket, key, versionID string) error { return nil }
func (m *mockStorageRepo) CreateBucket(ctx context.Context, bucket *domain.Bucket) error { return nil }
func (m *mockStorageRepo) GetBucket(ctx context.Context, name string) (*domain.Bucket, error) { return nil, nil }
func (m *mockStorageRepo) DeleteBucket(ctx context.Context, name string) error { return nil }
func (m *mockStorageRepo) ListBuckets(ctx context.Context, userID string) ([]*domain.Bucket, error) { return nil, nil }
func (m *mockStorageRepo) SetBucketVersioning(ctx context.Context, name string, enabled bool) error { return nil }
func (m *mockStorageRepo) SaveMultipartUpload(ctx context.Context, upload *domain.MultipartUpload) error { return nil }
func (m *mockStorageRepo) GetMultipartUpload(ctx context.Context, uploadID uuid.UUID) (*domain.MultipartUpload, error) { return nil, nil }
func (m *mockStorageRepo) DeleteMultipartUpload(ctx context.Context, uploadID uuid.UUID) error { return nil }
func (m *mockStorageRepo) SavePart(ctx context.Context, part *domain.Part) error { return nil }
func (m *mockStorageRepo) ListParts(ctx context.Context, uploadID uuid.UUID) ([]*domain.Part, error) { return nil, nil }

type mockStorageSvc struct{ mock.Mock }
func (m *mockStorageSvc) Upload(ctx context.Context, bucket, key string, r io.Reader) (*domain.Object, error) { return nil, nil }
func (m *mockStorageSvc) Download(ctx context.Context, bucket, key string) (io.ReadCloser, *domain.Object, error) { return nil, nil, nil }
func (m *mockStorageSvc) ListObjects(ctx context.Context, bucket string) ([]*domain.Object, error) { return nil, nil }
func (m *mockStorageSvc) DeleteObject(ctx context.Context, bucket, key string) error { return nil }
func (m *mockStorageSvc) DownloadVersion(ctx context.Context, bucket, key, versionID string) (io.ReadCloser, *domain.Object, error) { return nil, nil, nil }
func (m *mockStorageSvc) ListVersions(ctx context.Context, bucket, key string) ([]*domain.Object, error) { return nil, nil }
func (m *mockStorageSvc) DeleteVersion(ctx context.Context, bucket, key, versionID string) error { return nil }
func (m *mockStorageSvc) CreateBucket(ctx context.Context, name string, isPublic bool) (*domain.Bucket, error) { return nil, nil }
func (m *mockStorageSvc) GetBucket(ctx context.Context, name string) (*domain.Bucket, error) { return nil, nil }
func (m *mockStorageSvc) DeleteBucket(ctx context.Context, name string) error { return nil }
func (m *mockStorageSvc) ListBuckets(ctx context.Context) ([]*domain.Bucket, error) { return nil, nil }
func (m *mockStorageSvc) SetBucketVersioning(ctx context.Context, name string, enabled bool) error { return nil }
func (m *mockStorageSvc) GetClusterStatus(ctx context.Context) (*domain.StorageCluster, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*domain.StorageCluster), args.Error(1)
}
func (m *mockStorageSvc) CreateMultipartUpload(ctx context.Context, bucket, key string) (*domain.MultipartUpload, error) { return nil, nil }
func (m *mockStorageSvc) UploadPart(ctx context.Context, uploadID uuid.UUID, partNumber int, r io.Reader) (*domain.Part, error) { return nil, nil }
func (m *mockStorageSvc) CompleteMultipartUpload(ctx context.Context, uploadID uuid.UUID) (*domain.Object, error) { return nil, nil }
func (m *mockStorageSvc) AbortMultipartUpload(ctx context.Context, uploadID uuid.UUID) error { return nil }
func (m *mockStorageSvc) CleanupDeleted(ctx context.Context, limit int) (int, error) { return 0, nil }
func (m *mockStorageSvc) GeneratePresignedURL(ctx context.Context, bucket, key, method string, expiry time.Duration) (*domain.PresignedURL, error) { return nil, nil }

type mockClusterRepo struct{ mock.Mock }
func (m *mockClusterRepo) Create(ctx context.Context, cluster *domain.Cluster) error { return nil }
func (m *mockClusterRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Cluster, error) { return nil, nil }
func (m *mockClusterRepo) Update(ctx context.Context, cluster *domain.Cluster) error { return nil }
func (m *mockClusterRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockClusterRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Cluster, error) { return nil, nil }
func (m *mockClusterRepo) ListAll(ctx context.Context) ([]*domain.Cluster, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).([]*domain.Cluster), args.Error(1)
}
func (m *mockClusterRepo) AddNode(ctx context.Context, node *domain.ClusterNode) error { return nil }
func (m *mockClusterRepo) GetNodes(ctx context.Context, clusterID uuid.UUID) ([]*domain.ClusterNode, error) { return nil, nil }
func (m *mockClusterRepo) UpdateNode(ctx context.Context, node *domain.ClusterNode) error { return nil }
func (m *mockClusterRepo) DeleteNode(ctx context.Context, nodeID uuid.UUID) error { return nil }

type mockClusterProv struct{ mock.Mock }
func (m *mockClusterProv) Provision(ctx context.Context, cluster *domain.Cluster) error { return nil }
func (m *mockClusterProv) Deprovision(ctx context.Context, cluster *domain.Cluster) error { return nil }
func (m *mockClusterProv) Upgrade(ctx context.Context, cluster *domain.Cluster, version string) error { return nil }
func (m *mockClusterProv) GetKubeconfig(ctx context.Context, cluster *domain.Cluster, role string) (string, error) { return "", nil }
func (m *mockClusterProv) GetHealth(ctx context.Context, cluster *domain.Cluster) (*ports.ClusterHealth, error) {
	args := m.Called(ctx, cluster)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*ports.ClusterHealth), args.Error(1)
}
func (m *mockClusterProv) Repair(ctx context.Context, cluster *domain.Cluster) error { return m.Called(ctx, cluster).Error(0) }
func (m *mockClusterProv) GetStatus(ctx context.Context, cluster *domain.Cluster) (domain.ClusterStatus, error) { return domain.ClusterStatusRunning, nil }
func (m *mockClusterProv) Scale(ctx context.Context, cluster *domain.Cluster) error { return nil }
func (m *mockClusterProv) RotateSecrets(ctx context.Context, cluster *domain.Cluster) error { return nil }
func (m *mockClusterProv) CreateBackup(ctx context.Context, cluster *domain.Cluster) error { return nil }
func (m *mockClusterProv) Restore(ctx context.Context, cluster *domain.Cluster, backupPath string) error { return nil }

func TestNewAccountingWorker(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	w := NewAccountingWorker(new(mockAccountingSvc), logger)
	assert.NotNil(t, w)
}

func TestNewMetricsCollectorWorker(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	w := NewMetricsCollectorWorker(new(mockStorageRepo), new(mockStorageSvc), logger)
	assert.NotNil(t, w)
}

func TestClusterReconciler_Run(t *testing.T) {
	mockRepo := new(mockClusterRepo)
	mockProv := new(mockClusterProv)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	
	r := NewClusterReconciler(mockRepo, mockProv, logger)
	
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately to stop after first reconcile
	
	var wg sync.WaitGroup
	wg.Add(1)
	
	mockRepo.On("ListAll", mock.Anything).Return([]*domain.Cluster{}, nil).Once()
	
	r.Run(ctx, &wg)
	wg.Wait()
	
	mockRepo.AssertExpectations(t)
}
