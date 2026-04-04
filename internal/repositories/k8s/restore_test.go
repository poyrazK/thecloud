package k8s

import (
	"context"
	"encoding/base64"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockNodeExecutor struct {
	mock.Mock
}

func (m *mockNodeExecutor) Run(ctx context.Context, cmd string) (string, error) {
	args := m.Called(ctx, cmd)
	return args.String(0), args.Error(1)
}

func (m *mockNodeExecutor) WriteFile(ctx context.Context, path string, data io.Reader) error {
	args := m.Called(ctx, path, data)
	return args.Error(0)
}

func (m *mockNodeExecutor) WaitForReady(ctx context.Context, timeout time.Duration) error {
	args := m.Called(ctx, timeout)
	return args.Error(0)
}

type mockStorageService struct {
	mock.Mock
}

func (m *mockStorageService) Upload(ctx context.Context, bucket, key string, r io.Reader, providedChecksum string) (*domain.Object, error) {
	args := m.Called(ctx, bucket, key, r, providedChecksum)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Object), args.Error(1)
}

func (m *mockStorageService) Download(ctx context.Context, bucket, key string) (io.ReadCloser, *domain.Object, error) {
	args := m.Called(ctx, bucket, key)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).(io.ReadCloser), nil, args.Error(2)
}

func (m *mockStorageService) ListObjects(ctx context.Context, bucket string) ([]*domain.Object, error) {
	args := m.Called(ctx, bucket)
	return args.Get(0).([]*domain.Object), args.Error(1)
}

func (m *mockStorageService) DeleteObject(ctx context.Context, bucket, key string) error {
	args := m.Called(ctx, bucket, key)
	return args.Error(0)
}

func (m *mockStorageService) DownloadVersion(ctx context.Context, bucket, key, versionID string) (io.ReadCloser, *domain.Object, error) {
	args := m.Called(ctx, bucket, key, versionID)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).(io.ReadCloser), nil, args.Error(2)
}

func (m *mockStorageService) ListVersions(ctx context.Context, bucket, key string) ([]*domain.Object, error) {
	args := m.Called(ctx, bucket, key)
	return args.Get(0).([]*domain.Object), args.Error(1)
}

func (m *mockStorageService) DeleteVersion(ctx context.Context, bucket, key, versionID string) error {
	args := m.Called(ctx, bucket, key, versionID)
	return args.Error(0)
}

func (m *mockStorageService) CreateBucket(ctx context.Context, name string, isPublic bool) (*domain.Bucket, error) {
	args := m.Called(ctx, name, isPublic)
	return args.Get(0).(*domain.Bucket), args.Error(1)
}

func (m *mockStorageService) GetBucket(ctx context.Context, name string) (*domain.Bucket, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(*domain.Bucket), args.Error(1)
}

func (m *mockStorageService) DeleteBucket(ctx context.Context, name string, force bool) error {
	args := m.Called(ctx, name, force)
	return args.Error(0)
}

func (m *mockStorageService) ListBuckets(ctx context.Context) ([]*domain.Bucket, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*domain.Bucket), args.Error(1)
}

func (m *mockStorageService) SetBucketVersioning(ctx context.Context, name string, enabled bool) error {
	args := m.Called(ctx, name, enabled)
	return args.Error(0)
}

func (m *mockStorageService) GetClusterStatus(ctx context.Context) (*domain.StorageCluster, error) {
	args := m.Called(ctx)
	return args.Get(0).(*domain.StorageCluster), args.Error(1)
}

func (m *mockStorageService) CreateMultipartUpload(ctx context.Context, bucket, key string) (*domain.MultipartUpload, error) {
	args := m.Called(ctx, bucket, key)
	return args.Get(0).(*domain.MultipartUpload), args.Error(1)
}

func (m *mockStorageService) UploadPart(ctx context.Context, uploadID uuid.UUID, partNumber int, r io.Reader, providedChecksum string) (*domain.Part, error) {
	args := m.Called(ctx, uploadID, partNumber, r, providedChecksum)
	return args.Get(0).(*domain.Part), args.Error(1)
}

func (m *mockStorageService) CompleteMultipartUpload(ctx context.Context, uploadID uuid.UUID) (*domain.Object, error) {
	args := m.Called(ctx, uploadID)
	return args.Get(0).(*domain.Object), args.Error(1)
}

func (m *mockStorageService) AbortMultipartUpload(ctx context.Context, uploadID uuid.UUID) error {
	args := m.Called(ctx, uploadID)
	return args.Error(0)
}

func (m *mockStorageService) CleanupDeleted(ctx context.Context, limit int) (int, error) {
	args := m.Called(ctx, limit)
	return args.Int(0), args.Error(1)
}

func (m *mockStorageService) GeneratePresignedURL(ctx context.Context, bucket, key, method string, expiry time.Duration) (*domain.PresignedURL, error) {
	args := m.Called(ctx, bucket, key, method, expiry)
	return args.Get(0).(*domain.PresignedURL), args.Error(1)
}

func (m *mockStorageService) CleanupPendingUploads(ctx context.Context, olderThan time.Duration, limit int) (int, error) {
	args := m.Called(ctx, olderThan, limit)
	return args.Int(0), args.Error(1)
}

func TestRestore(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	clusterID := uuid.New()
	cluster := &domain.Cluster{
		ID:              clusterID,
		ControlPlaneIPs: []string{"10.0.0.1"},
	}

	t.Run("Successful Restore", func(t *testing.T) {
		executor := new(mockNodeExecutor)
		storage := new(mockStorageService)

		p := &KubeadmProvisioner{
			storageSvc: storage,
			logger:     logger,
			executorFactory: func(ctx context.Context, c *domain.Cluster, ip string) (NodeExecutor, error) {
				return executor, nil
			},
		}

		backupData := "fake-etcd-data"
		backupPath := "clusters/" + clusterID.String() + "/backup.db"
		storage.On("Download", mock.Anything, "k8s-backups", backupPath).
			Return(io.NopCloser(strings.NewReader(backupData)), nil, nil)

		// Expect node preparation
		executor.On("Run", mock.Anything, "mkdir -p /tmp/manifests-backup").Return("", nil)
		executor.On("Run", mock.Anything, "rm -rf /tmp/manifests-backup/*").Return("", nil)
		executor.On("Run", mock.Anything, "mv /etc/kubernetes/manifests/*.yaml /tmp/manifests-backup/").Return("", nil)

		// Expect file upload
		executor.On("WriteFile", mock.Anything, "/tmp/restore-snapshot.db", mock.Anything).Return(nil)

		// Expect etcd restore
		executor.On("Run", mock.Anything, mock.MatchedBy(func(cmd string) bool {
			return strings.Contains(cmd, "etcdctl snapshot restore")
		})).Return("", nil)

		// Expect swap directories
		executor.On("Run", mock.Anything, mock.MatchedBy(func(cmd string) bool {
			return strings.Contains(cmd, "mv /var/lib/etcd /var/lib/etcd-backup")
		})).Return("", nil)
		executor.On("Run", mock.Anything, "mv /var/lib/etcd-restored /var/lib/etcd").Return("", nil)
		executor.On("Run", mock.Anything, "chown -R 0:0 /var/lib/etcd").Return("", nil)

		// Expect restart pods
		executor.On("Run", mock.Anything, "mv /tmp/manifests-backup/*.yaml /etc/kubernetes/manifests/").Return("", nil)

		// Expect cleanup
		executor.On("Run", mock.Anything, "rm /tmp/restore-snapshot.db").Return("", nil)

		err := p.Restore(ctx, cluster, backupPath)
		require.NoError(t, err)

		executor.AssertExpectations(t)
		storage.AssertExpectations(t)
	})

	t.Run("Successful Restore Base64", func(t *testing.T) {
		executor := new(mockNodeExecutor)
		storage := new(mockStorageService)

		p := &KubeadmProvisioner{
			storageSvc: storage,
			logger:     logger,
			executorFactory: func(ctx context.Context, c *domain.Cluster, ip string) (NodeExecutor, error) {
				return executor, nil
			},
		}

		backupData := base64.StdEncoding.EncodeToString([]byte("fake-etcd-data"))
		backupPath := "clusters/" + clusterID.String() + "/backup.db.b64"
		storage.On("Download", mock.Anything, "k8s-backups", backupPath).
			Return(io.NopCloser(strings.NewReader(backupData)), nil, nil)

		executor.On("Run", mock.Anything, mock.Anything).Return("", nil)
		executor.On("WriteFile", mock.Anything, "/tmp/restore-snapshot.db", mock.Anything).Return(nil)

		err := p.Restore(ctx, cluster, backupPath)
		require.NoError(t, err)
	})

	t.Run("Restore Failure - No Control Plane", func(t *testing.T) {
		p := &KubeadmProvisioner{
			logger: logger,
		}
		err := p.Restore(ctx, &domain.Cluster{}, "path")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no control plane node")
	})

	t.Run("Restore Failure - Storage Error", func(t *testing.T) {
		storage := new(mockStorageService)
		p := &KubeadmProvisioner{
			storageSvc: storage,
			logger:     logger,
			executorFactory: func(ctx context.Context, c *domain.Cluster, ip string) (NodeExecutor, error) {
				return new(mockNodeExecutor), nil
			},
		}

		storage.On("Download", mock.Anything, "k8s-backups", "bad-path").
			Return(nil, nil, os.ErrNotExist)

		err := p.Restore(ctx, cluster, "bad-path")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to download backup")
	})
}

func TestCreateBackup_DR(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	clusterID := uuid.New()
	cluster := &domain.Cluster{
		ID:              clusterID,
		ControlPlaneIPs: []string{"10.0.0.1"},
	}

	t.Run("Successful Backup", func(t *testing.T) {
		executor := new(mockNodeExecutor)
		storage := new(mockStorageService)

		p := &KubeadmProvisioner{
			storageSvc: storage,
			logger:     logger,
			executorFactory: func(ctx context.Context, c *domain.Cluster, ip string) (NodeExecutor, error) {
				return executor, nil
			},
		}

		executor.On("Run", mock.Anything, mock.MatchedBy(func(cmd string) bool {
			return strings.Contains(cmd, "snapshot save")
		})).Return("", nil).Once()

		executor.On("Run", mock.Anything, "base64 /tmp/snapshot.db").Return("YmFja3VwLWRhdGE=", nil).Once()

		storage.On("Upload", mock.Anything, "k8s-backups", mock.Anything, mock.Anything, mock.Anything).
			Return(&domain.Object{}, nil).Once()

		err := p.CreateBackup(ctx, cluster)
		require.NoError(t, err)

		executor.AssertExpectations(t)
		storage.AssertExpectations(t)
	})

	t.Run("Backup Failure - No Control Plane", func(t *testing.T) {
		p := &KubeadmProvisioner{
			logger: logger,
		}
		err := p.CreateBackup(ctx, &domain.Cluster{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no control plane node")
	})

	t.Run("Backup Failure - Etcd Error", func(t *testing.T) {
		executor := new(mockNodeExecutor)
		p := &KubeadmProvisioner{
			logger: logger,
			executorFactory: func(ctx context.Context, c *domain.Cluster, ip string) (NodeExecutor, error) {
				return executor, nil
			},
		}

		executor.On("Run", mock.Anything, mock.Anything).Return("", os.ErrPermission)

		err := p.CreateBackup(ctx, cluster)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "etcd snapshot failed")
	})
}
