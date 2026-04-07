package services_test

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/stretchr/testify/mock"
)

// MockStorageRepo
type MockStorageRepo struct {
	mock.Mock
}

func (m *MockStorageRepo) SaveMeta(ctx context.Context, obj *domain.Object) error {
	args := m.Called(ctx, obj)
	return args.Error(0)
}
func (m *MockStorageRepo) GetMeta(ctx context.Context, bucket, key string) (*domain.Object, error) {
	args := m.Called(ctx, bucket, key)
	r0, _ := args.Get(0).(*domain.Object)
	return r0, args.Error(1)
}
func (m *MockStorageRepo) List(ctx context.Context, bucket string) ([]*domain.Object, error) {
	args := m.Called(ctx, bucket)
	r0, _ := args.Get(0).([]*domain.Object)
	return r0, args.Error(1)
}
func (m *MockStorageRepo) SoftDelete(ctx context.Context, bucket, key string) error {
	args := m.Called(ctx, bucket, key)
	return args.Error(0)
}
func (m *MockStorageRepo) DeleteVersion(ctx context.Context, bucket, key, versionID string) error {
	args := m.Called(ctx, bucket, key, versionID)
	return args.Error(0)
}
func (m *MockStorageRepo) GetMetaByVersion(ctx context.Context, bucket, key, versionID string) (*domain.Object, error) {
	args := m.Called(ctx, bucket, key, versionID)
	r0, _ := args.Get(0).(*domain.Object)
	return r0, args.Error(1)
}
func (m *MockStorageRepo) ListVersions(ctx context.Context, bucket, key string) ([]*domain.Object, error) {
	args := m.Called(ctx, bucket, key)
	r0, _ := args.Get(0).([]*domain.Object)
	return r0, args.Error(1)
}

func (m *MockStorageRepo) ListDeleted(ctx context.Context, limit int) ([]*domain.Object, error) {
	args := m.Called(ctx, limit)
	r0, _ := args.Get(0).([]*domain.Object)
	return r0, args.Error(1)
}

func (m *MockStorageRepo) HardDelete(ctx context.Context, bucket, key, versionID string) error {
	return m.Called(ctx, bucket, key, versionID).Error(0)
}

func (m *MockStorageRepo) ListPending(ctx context.Context, olderThan time.Time, limit int) ([]*domain.Object, error) {
	args := m.Called(ctx, olderThan, limit)
	r0, _ := args.Get(0).([]*domain.Object)
	return r0, args.Error(1)
}

func (m *MockStorageRepo) CreateBucket(ctx context.Context, bucket *domain.Bucket) error {
	return m.Called(ctx, bucket).Error(0)
}
func (m *MockStorageRepo) GetBucket(ctx context.Context, name string) (*domain.Bucket, error) {
	args := m.Called(ctx, name)
	r0, _ := args.Get(0).(*domain.Bucket)
	return r0, args.Error(1)
}
func (m *MockStorageRepo) DeleteBucket(ctx context.Context, name string) error {
	return m.Called(ctx, name).Error(0)
}
func (m *MockStorageRepo) ListBuckets(ctx context.Context, userID string) ([]*domain.Bucket, error) {
	args := m.Called(ctx, userID)
	r0, _ := args.Get(0).([]*domain.Bucket)
	return r0, args.Error(1)
}
func (m *MockStorageRepo) SetBucketVersioning(ctx context.Context, name string, enabled bool) error {
	return m.Called(ctx, name, enabled).Error(0)
}

func (m *MockStorageRepo) SaveMultipartUpload(ctx context.Context, upload *domain.MultipartUpload) error {
	return m.Called(ctx, upload).Error(0)
}
func (m *MockStorageRepo) GetMultipartUpload(ctx context.Context, uploadID uuid.UUID) (*domain.MultipartUpload, error) {
	args := m.Called(ctx, uploadID)
	r0, _ := args.Get(0).(*domain.MultipartUpload)
	return r0, args.Error(1)
}
func (m *MockStorageRepo) DeleteMultipartUpload(ctx context.Context, uploadID uuid.UUID) error {
	return m.Called(ctx, uploadID).Error(0)
}
func (m *MockStorageRepo) SavePart(ctx context.Context, part *domain.Part) error {
	return m.Called(ctx, part).Error(0)
}
func (m *MockStorageRepo) ListParts(ctx context.Context, uploadID uuid.UUID) ([]*domain.Part, error) {
	args := m.Called(ctx, uploadID)
	r0, _ := args.Get(0).([]*domain.Part)
	return r0, args.Error(1)
}

type MockStorageRepository = MockStorageRepo

// MockFileStore
type MockFileStore struct {
	mock.Mock
}

func (m *MockFileStore) Write(ctx context.Context, bucket, key string, r io.Reader) (int64, error) {
	args := m.Called(ctx, bucket, key, r)
	r0, _ := args.Get(0).(int64)
	return r0, args.Error(1)
}
func (m *MockFileStore) Read(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	args := m.Called(ctx, bucket, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).(io.ReadCloser)
	return r0, args.Error(1)
}
func (m *MockFileStore) Delete(ctx context.Context, bucket, key string) error {
	args := m.Called(ctx, bucket, key)
	return args.Error(0)
}
func (m *MockFileStore) List(ctx context.Context, bucket string) ([]string, error) {
	args := m.Called(ctx, bucket)
	return args.Get(0).([]string), args.Error(1)
}
func (m *MockFileStore) Upload(ctx context.Context, bucket string, key string, r io.Reader) error {
	return m.Called(ctx, bucket, key, r).Error(0)
}
func (m *MockFileStore) Download(ctx context.Context, bucket string, key string) (io.ReadCloser, error) {
	args := m.Called(ctx, bucket, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}
func (m *MockFileStore) GetClusterStatus(ctx context.Context) (*domain.StorageCluster, error) {
	args := m.Called(ctx)
	r0, _ := args.Get(0).(*domain.StorageCluster)
	return r0, args.Error(1)
}
func (m *MockFileStore) Assemble(ctx context.Context, bucket, key string, parts []string) (int64, error) {
	args := m.Called(ctx, bucket, key, parts)
	r0, _ := args.Get(0).(int64)
	return r0, args.Error(1)
}

// MockVolumeRepo
type MockVolumeRepo struct {
	mock.Mock
}

func (m *MockVolumeRepo) Create(ctx context.Context, v *domain.Volume) error {
	args := m.Called(ctx, v)
	return args.Error(0)
}

func (m *MockVolumeRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Volume, error) {
	args := m.Called(ctx, id)
	r0, _ := args.Get(0).(*domain.Volume)
	return r0, args.Error(1)
}

func (m *MockVolumeRepo) GetByName(ctx context.Context, name string) (*domain.Volume, error) {
	args := m.Called(ctx, name)
	r0, _ := args.Get(0).(*domain.Volume)
	return r0, args.Error(1)
}

func (m *MockVolumeRepo) List(ctx context.Context) ([]*domain.Volume, error) {
	args := m.Called(ctx)
	r0, _ := args.Get(0).([]*domain.Volume)
	return r0, args.Error(1)
}

func (m *MockVolumeRepo) ListByInstanceID(ctx context.Context, id uuid.UUID) ([]*domain.Volume, error) {
	args := m.Called(ctx, id)
	r0, _ := args.Get(0).([]*domain.Volume)
	return r0, args.Error(1)
}

func (m *MockVolumeRepo) Update(ctx context.Context, v *domain.Volume) error {
	return m.Called(ctx, v).Error(0)
}

func (m *MockVolumeRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

// MockVolumeService
type MockVolumeService struct {
	mock.Mock
}

func (m *MockVolumeService) CreateVolume(ctx context.Context, name string, sizeGB int) (*domain.Volume, error) {
	args := m.Called(ctx, name, sizeGB)
	r0, _ := args.Get(0).(*domain.Volume)
	return r0, args.Error(1)
}
func (m *MockVolumeService) ListVolumes(ctx context.Context) ([]*domain.Volume, error) {
	args := m.Called(ctx)
	r0, _ := args.Get(0).([]*domain.Volume)
	return r0, args.Error(1)
}
func (m *MockVolumeService) GetVolume(ctx context.Context, idOrName string) (*domain.Volume, error) {
	args := m.Called(ctx, idOrName)
	r0, _ := args.Get(0).(*domain.Volume)
	return r0, args.Error(1)
}
func (m *MockVolumeService) DeleteVolume(ctx context.Context, idOrName string) error {
	args := m.Called(ctx, idOrName)
	return args.Error(0)
}
func (m *MockVolumeService) ResizeVolume(ctx context.Context, id string, newSizeGB int) error {
	args := m.Called(ctx, id, newSizeGB)
	return args.Error(0)
}
func (m *MockVolumeService) ReleaseVolumesForInstance(ctx context.Context, instanceID uuid.UUID) error {
	args := m.Called(ctx, instanceID)
	return args.Error(0)
}

func (m *MockVolumeService) AttachVolume(ctx context.Context, volumeID string, instanceID string, mountPath string) (string, error) {
	args := m.Called(ctx, volumeID, instanceID, mountPath)
	return args.String(0), args.Error(1)
}

func (m *MockVolumeService) DetachVolume(ctx context.Context, volumeID string) error {
	args := m.Called(ctx, volumeID)
	return args.Error(0)
}

// MockSnapshotRepo
type MockSnapshotRepo struct {
	mock.Mock
}

func (m *MockSnapshotRepo) Create(ctx context.Context, s *domain.Snapshot) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}

func (m *MockSnapshotRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Snapshot, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Snapshot), args.Error(1)
}

func (m *MockSnapshotRepo) ListByVolumeID(ctx context.Context, volumeID uuid.UUID) ([]*domain.Snapshot, error) {
	args := m.Called(ctx, volumeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Snapshot), args.Error(1)
}

func (m *MockSnapshotRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Snapshot, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Snapshot), args.Error(1)
}
func (m *MockSnapshotRepo) Update(ctx context.Context, s *domain.Snapshot) error {
	return m.Called(ctx, s).Error(0)
}
func (m *MockSnapshotRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

// MockSnapshotService
type MockSnapshotService struct {
	mock.Mock
}

func (m *MockSnapshotService) CreateSnapshot(ctx context.Context, volumeID uuid.UUID, description string) (*domain.Snapshot, error) {
	args := m.Called(ctx, volumeID, description)
	r0, _ := args.Get(0).(*domain.Snapshot)
	return r0, args.Error(1)
}
func (m *MockSnapshotService) ListSnapshots(ctx context.Context) ([]*domain.Snapshot, error) {
	args := m.Called(ctx)
	r0, _ := args.Get(0).([]*domain.Snapshot)
	return r0, args.Error(1)
}
func (m *MockSnapshotService) GetSnapshot(ctx context.Context, id uuid.UUID) (*domain.Snapshot, error) {
	args := m.Called(ctx, id)
	r0, _ := args.Get(0).(*domain.Snapshot)
	return r0, args.Error(1)
}
func (m *MockSnapshotService) DeleteSnapshot(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockSnapshotService) RestoreSnapshot(ctx context.Context, snapshotID uuid.UUID, newVolumeName string) (*domain.Volume, error) {
	args := m.Called(ctx, snapshotID, newVolumeName)
	r0, _ := args.Get(0).(*domain.Volume)
	return r0, args.Error(1)
}

// MockLifecycleRepository
type MockLifecycleRepository struct {
	mock.Mock
}

func (m *MockLifecycleRepository) Create(ctx context.Context, rule *domain.LifecycleRule) error {
	return m.Called(ctx, rule).Error(0)
}
func (m *MockLifecycleRepository) Get(ctx context.Context, id uuid.UUID) (*domain.LifecycleRule, error) {
	args := m.Called(ctx, id)
	r0, _ := args.Get(0).(*domain.LifecycleRule)
	return r0, args.Error(1)
}
func (m *MockLifecycleRepository) List(ctx context.Context, bucketName string) ([]*domain.LifecycleRule, error) {
	args := m.Called(ctx, bucketName)
	r0, _ := args.Get(0).([]*domain.LifecycleRule)
	return r0, args.Error(1)
}
func (m *MockLifecycleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockLifecycleRepository) GetEnabledRules(ctx context.Context) ([]*domain.LifecycleRule, error) {
	args := m.Called(ctx)
	r0, _ := args.Get(0).([]*domain.LifecycleRule)
	return r0, args.Error(1)
}

// MockEncryptionRepository
type MockEncryptionRepository struct {
	mock.Mock
}

func (m *MockEncryptionRepository) SaveKey(ctx context.Context, key ports.EncryptionKey) error {
	return m.Called(ctx, key).Error(0)
}

func (m *MockEncryptionRepository) GetKey(ctx context.Context, bucketName string) (*ports.EncryptionKey, error) {
	args := m.Called(ctx, bucketName)
	r0, _ := args.Get(0).(*ports.EncryptionKey)
	return r0, args.Error(1)
}

// MockEncryptionService
type MockEncryptionService struct {
	mock.Mock
}

func (m *MockEncryptionService) Encrypt(ctx context.Context, bucket string, r io.Reader) (io.Reader, error) {
	args := m.Called(ctx, bucket, r)
	r0, _ := args.Get(0).(io.Reader)
	return r0, args.Error(1)
}

func (m *MockEncryptionService) Decrypt(ctx context.Context, bucket string, r io.Reader) (io.Reader, error) {
	args := m.Called(ctx, bucket, r)
	r0, _ := args.Get(0).(io.Reader)
	return r0, args.Error(1)
}

func (m *MockEncryptionService) CreateKey(ctx context.Context, bucket string) (string, error) {
	args := m.Called(ctx, bucket)
	return args.String(0), args.Error(1)
}

func (m *MockEncryptionService) RotateKey(ctx context.Context, bucket string) (string, error) {
	args := m.Called(ctx, bucket)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	r0, _ := args.Get(0).(string)
	return r0, args.Error(1)
}

// MockStorageBackend
type MockStorageBackend struct{ mock.Mock }

func (m *MockStorageBackend) CreateVolume(ctx context.Context, name string, sizeGB int) (string, error) {
	args := m.Called(ctx, name, sizeGB)
	return args.String(0), args.Error(1)
}
func (m *MockStorageBackend) DeleteVolume(ctx context.Context, name string) error {
	return m.Called(ctx, name).Error(0)
}
func (m *MockStorageBackend) ResizeVolume(ctx context.Context, name string, newSizeGB int) error {
	return m.Called(ctx, name, newSizeGB).Error(0)
}
func (m *MockStorageBackend) AttachVolume(ctx context.Context, volumeName, instanceID string) (string, error) {
	args := m.Called(ctx, volumeName, instanceID)
	return args.String(0), args.Error(1)
}
func (m *MockStorageBackend) DetachVolume(ctx context.Context, volumeName, instanceID string) error {
	return m.Called(ctx, volumeName, instanceID).Error(0)
}
func (m *MockStorageBackend) CreateSnapshot(ctx context.Context, volumeName, snapshotName string) error {
	return m.Called(ctx, volumeName, snapshotName).Error(0)
}
func (m *MockStorageBackend) RestoreSnapshot(ctx context.Context, volumeName, snapshotName string) error {
	return m.Called(ctx, volumeName, snapshotName).Error(0)
}
func (m *MockStorageBackend) DeleteSnapshot(ctx context.Context, snapshotName string) error {
	return m.Called(ctx, snapshotName).Error(0)
}
func (m *MockStorageBackend) Ping(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}
func (m *MockStorageBackend) Type() string {
	return m.Called().String(0)
}
