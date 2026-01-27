package k8s_test

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/repositories/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Re-defining mocks correctly to use mock.Mock.Called
type MockStorageServiceV2 struct{ mock.Mock }

func (m *MockStorageServiceV2) Upload(ctx context.Context, bucket, key string, r io.Reader) (*domain.Object, error) {
	args := m.Called(ctx, bucket, key, r)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Object), args.Error(1)
}
func (m *MockStorageServiceV2) Download(ctx context.Context, bucket, key string) (io.ReadCloser, *domain.Object, error) {
	args := m.Called(ctx, bucket, key)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).(io.ReadCloser), args.Get(1).(*domain.Object), args.Error(2)
}
func (m *MockStorageServiceV2) ListObjects(ctx context.Context, bucket string) ([]*domain.Object, error) {
	return nil, nil
}
func (m *MockStorageServiceV2) DeleteObject(ctx context.Context, bucket, key string) error {
	return nil
}
func (m *MockStorageServiceV2) DownloadVersion(ctx context.Context, bucket, key, versionID string) (io.ReadCloser, *domain.Object, error) {
	return nil, nil, nil
}
func (m *MockStorageServiceV2) ListVersions(ctx context.Context, bucket, key string) ([]*domain.Object, error) {
	return nil, nil
}
func (m *MockStorageServiceV2) DeleteVersion(ctx context.Context, bucket, key, versionID string) error {
	return nil
}
func (m *MockStorageServiceV2) CreateBucket(ctx context.Context, name string, isPublic bool) (*domain.Bucket, error) {
	return &domain.Bucket{Name: name}, nil
}
func (m *MockStorageServiceV2) GetBucket(ctx context.Context, name string) (*domain.Bucket, error) {
	return &domain.Bucket{Name: name}, nil
}
func (m *MockStorageServiceV2) DeleteBucket(ctx context.Context, name string) error {
	return nil
}
func (m *MockStorageServiceV2) ListBuckets(ctx context.Context) ([]*domain.Bucket, error) {
	return nil, nil
}
func (m *MockStorageServiceV2) GetClusterStatus(ctx context.Context) (*domain.StorageCluster, error) {
	return nil, nil
}
func (m *MockStorageServiceV2) SetBucketVersioning(ctx context.Context, name string, enabled bool) error {
	return nil
}
func (m *MockStorageServiceV2) GeneratePresignedURL(ctx context.Context, bucket, key, method string, expiry time.Duration) (*domain.PresignedURL, error) {
	return nil, nil
}
func (m *MockStorageServiceV2) CreateMultipartUpload(ctx context.Context, bucket, key string) (*domain.MultipartUpload, error) {
	return nil, nil
}
func (m *MockStorageServiceV2) UploadPart(ctx context.Context, uploadID uuid.UUID, partNumber int, r io.Reader) (*domain.Part, error) {
	return nil, nil
}
func (m *MockStorageServiceV2) CompleteMultipartUpload(ctx context.Context, uploadID uuid.UUID) (*domain.Object, error) {
	return nil, nil
}
func (m *MockStorageServiceV2) AbortMultipartUpload(ctx context.Context, uploadID uuid.UUID) error {
	return nil
}

const (
	testIPMaster   = "10.0.0.1"
	testIPWorker   = "10.0.0.2"
	testBackupPath = "path/to/backup"
)

func TestKubeadmProvisionerUpgrade(t *testing.T) {
	ctx := context.Background()
	instSvc := new(MockInstanceService)
	repo := new(MockClusterRepo)
	secretSvc := new(MockSecretService)
	sgSvc := new(MockSecurityGroupService)
	storageSvc := new(MockStorageServiceV2)
	lbSvc := new(MockLBService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	p := k8s.NewKubeadmProvisioner(instSvc, repo, secretSvc, sgSvc, storageSvc, lbSvc, logger)
	clusterID := uuid.New()
	cluster := &domain.Cluster{
		ID:              clusterID,
		ControlPlaneIPs: []string{testIPMaster},
	}

	t.Run("Success", func(t *testing.T) {
		masterNode := &domain.ClusterNode{InstanceID: uuid.New(), Role: domain.NodeRoleControlPlane}
		workerID := uuid.New()
		workerNode := &domain.ClusterNode{ID: uuid.New(), InstanceID: workerID, Role: domain.NodeRoleWorker}
		workerInst := &domain.Instance{ID: workerID, PrivateIP: testIPWorker}

		repo.On("GetNodes", ctx, clusterID).Return([]*domain.ClusterNode{masterNode, workerNode}, nil)
		instSvc.On("GetInstance", ctx, workerID.String()).Return(workerInst, nil)
		instSvc.On("GetInstance", ctx, masterNode.InstanceID.String()).Return(&domain.Instance{ID: masterNode.InstanceID, PrivateIP: testIPMaster}, nil)

		instSvc.On("Exec", ctx, masterNode.InstanceID.String(), mock.Anything).Return("success", nil).Once()
		instSvc.On("Exec", ctx, workerID.String(), mock.Anything).Return("success", nil).Once()

		err := p.Upgrade(ctx, cluster, "v1.30.0")
		assert.NoError(t, err)
	})

	t.Run("No Control Plane IPs", func(t *testing.T) {
		badCluster := &domain.Cluster{ID: clusterID}
		err := p.Upgrade(ctx, badCluster, "v1.30.0")
		assert.Error(t, err)
	})
}

func TestKubeadmProvisionerRotateSecrets(t *testing.T) {
	ctx := context.Background()
	instSvc := new(MockInstanceService)
	repo := new(MockClusterRepo)
	secretSvc := new(MockSecretService)
	sgSvc := new(MockSecurityGroupService)
	storageSvc := new(MockStorageServiceV2)
	lbSvc := new(MockLBService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	p := k8s.NewKubeadmProvisioner(instSvc, repo, secretSvc, sgSvc, storageSvc, lbSvc, logger)
	clusterID := uuid.New()
	cluster := &domain.Cluster{
		ID:              clusterID,
		ControlPlaneIPs: []string{testIPMaster},
	}

	masterID := uuid.New()
	repo.On("GetNodes", ctx, clusterID).Return([]*domain.ClusterNode{{InstanceID: masterID, Role: domain.NodeRoleControlPlane}}, nil)
	instSvc.On("GetInstance", ctx, masterID.String()).Return(&domain.Instance{ID: masterID, PrivateIP: testIPMaster}, nil)

	t.Run("Success", func(t *testing.T) {
		instSvc.On("Exec", ctx, masterID.String(), mock.MatchedBy(func(args []string) bool {
			return args[len(args)-1] == "kubeadm certs renew all"
		})).Return("success", nil).Once()

		instSvc.On("Exec", ctx, masterID.String(), mock.MatchedBy(func(args []string) bool {
			return args[len(args)-1] == "cat /etc/kubernetes/admin.conf"
		})).Return("kubeconfig-data", nil).Once()

		repo.On("Update", ctx, cluster).Return(nil).Once()

		err := p.RotateSecrets(ctx, cluster)
		assert.NoError(t, err)
		assert.Equal(t, "kubeconfig-data", cluster.Kubeconfig)
	})
}

func TestKubeadmProvisionerCreateBackup(t *testing.T) {
	ctx := context.Background()
	instSvc := new(MockInstanceService)
	repo := new(MockClusterRepo)
	secretSvc := new(MockSecretService)
	sgSvc := new(MockSecurityGroupService)
	storageSvc := new(MockStorageServiceV2)
	lbSvc := new(MockLBService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	p := k8s.NewKubeadmProvisioner(instSvc, repo, secretSvc, sgSvc, storageSvc, lbSvc, logger)
	clusterID := uuid.New()
	cluster := &domain.Cluster{
		ID:              clusterID,
		ControlPlaneIPs: []string{testIPMaster},
	}

	masterID := uuid.New()
	repo.On("GetNodes", ctx, clusterID).Return([]*domain.ClusterNode{{InstanceID: masterID, Role: domain.NodeRoleControlPlane}}, nil)
	instSvc.On("GetInstance", ctx, masterID.String()).Return(&domain.Instance{ID: masterID, PrivateIP: testIPMaster}, nil)

	t.Run("Success", func(t *testing.T) {
		instSvc.On("Exec", ctx, masterID.String(), mock.Anything).Return("bW9jay1kYXRhCg==", nil)
		storageSvc.On("Upload", ctx, "k8s-backups", mock.Anything, mock.Anything).Return(&domain.Object{}, nil).Once()

		err := p.CreateBackup(ctx, cluster)
		assert.NoError(t, err)
	})
}

func TestKubeadmProvisionerRestore(t *testing.T) {
	ctx := context.Background()
	instSvc := new(MockInstanceService)
	repo := new(MockClusterRepo)
	secretSvc := new(MockSecretService)
	sgSvc := new(MockSecurityGroupService)
	storageSvc := new(MockStorageServiceV2)
	lbSvc := new(MockLBService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	p := k8s.NewKubeadmProvisioner(instSvc, repo, secretSvc, sgSvc, storageSvc, lbSvc, logger)
	clusterID := uuid.New()
	cluster := &domain.Cluster{
		ID:              clusterID,
		ControlPlaneIPs: []string{testIPMaster},
	}

	masterID := uuid.New()
	repo.On("GetNodes", ctx, clusterID).Return([]*domain.ClusterNode{{InstanceID: masterID, Role: domain.NodeRoleControlPlane}}, nil)
	instSvc.On("GetInstance", ctx, masterID.String()).Return(&domain.Instance{ID: masterID, PrivateIP: testIPMaster}, nil)

	t.Run("Success", func(t *testing.T) {
		mockData := []byte("backup-data")
		storageSvc.On("Download", ctx, "k8s-backups", testBackupPath).Return(io.NopCloser(bytes.NewReader(mockData)), &domain.Object{}, nil).Once()
		instSvc.On("Exec", ctx, masterID.String(), mock.Anything).Return("success", nil)

		err := p.Restore(ctx, cluster, testBackupPath)
		assert.NoError(t, err)
	})
}
