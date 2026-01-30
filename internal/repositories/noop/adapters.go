// Package noop provides no-op infrastructure adapters for testing and local runs.
package noop

import (
	"context"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

// NoopInstanceRepository is a no-op instance repository.
type NoopInstanceRepository struct{}

func (r *NoopInstanceRepository) Create(ctx context.Context, i *domain.Instance) error { return nil }
func (r *NoopInstanceRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Instance, error) {
	return &domain.Instance{ID: id}, nil
}
func (r *NoopInstanceRepository) GetByName(ctx context.Context, name string) (*domain.Instance, error) {
	return &domain.Instance{ID: uuid.New(), Name: name}, nil
}
func (r *NoopInstanceRepository) List(ctx context.Context) ([]*domain.Instance, error) {
	return []*domain.Instance{}, nil
}
func (r *NoopInstanceRepository) ListAll(ctx context.Context) ([]*domain.Instance, error) {
	return []*domain.Instance{}, nil
}
func (r *NoopInstanceRepository) Update(ctx context.Context, i *domain.Instance) error { return nil }
func (r *NoopInstanceRepository) Delete(ctx context.Context, id uuid.UUID) error       { return nil }
func (r *NoopInstanceRepository) ListBySubnet(ctx context.Context, subnetID uuid.UUID) ([]*domain.Instance, error) {
	return []*domain.Instance{}, nil
}
func (r *NoopInstanceRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Instance, error) {
	return []*domain.Instance{}, nil
}

// NoopInstanceTypeRepository is a no-op instance type repository.
type NoopInstanceTypeRepository struct{}

func (r *NoopInstanceTypeRepository) List(ctx context.Context) ([]*domain.InstanceType, error) {
	return []*domain.InstanceType{
		{ID: "basic-1", Name: "Basic 1", VCPUs: 1, MemoryMB: 512, DiskGB: 8},
		{ID: "basic-2", Name: "Basic 2", VCPUs: 1, MemoryMB: 1024, DiskGB: 10},
	}, nil
}

func (r *NoopInstanceTypeRepository) GetByID(ctx context.Context, id string) (*domain.InstanceType, error) {
	return &domain.InstanceType{ID: id, Name: id, VCPUs: 1, MemoryMB: 1024, DiskGB: 10}, nil
}

func (r *NoopInstanceTypeRepository) Create(ctx context.Context, it *domain.InstanceType) (*domain.InstanceType, error) {
	return it, nil
}

func (r *NoopInstanceTypeRepository) Update(ctx context.Context, it *domain.InstanceType) (*domain.InstanceType, error) {
	return it, nil
}

func (r *NoopInstanceTypeRepository) Delete(ctx context.Context, id string) error {
	return nil
}

// NoopVpcRepository is a no-op VPC repository.
type NoopVpcRepository struct{}

func (r *NoopVpcRepository) Create(ctx context.Context, v *domain.VPC) error { return nil }
func (r *NoopVpcRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.VPC, error) {
	return &domain.VPC{ID: id}, nil
}
func (r *NoopVpcRepository) GetByName(ctx context.Context, name string) (*domain.VPC, error) {
	return &domain.VPC{ID: uuid.New(), Name: name}, nil
}
func (r *NoopVpcRepository) List(ctx context.Context) ([]*domain.VPC, error) {
	return []*domain.VPC{}, nil
}
func (r *NoopVpcRepository) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (r *NoopVpcRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.VPC, error) {
	return []*domain.VPC{}, nil
}

// NoopSubnetRepository is a no-op subnet repository.
type NoopSubnetRepository struct{}

func (r *NoopSubnetRepository) Create(ctx context.Context, s *domain.Subnet) error { return nil }

var staticSubnet = &domain.Subnet{
	ID:        uuid.New(),
	CIDRBlock: "10.0.0.0/24",
	GatewayIP: "10.0.0.1",
	Status:    "available",
}

func (r *NoopSubnetRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subnet, error) {
	return staticSubnet, nil
}
func (r *NoopSubnetRepository) GetByName(ctx context.Context, vpcID uuid.UUID, name string) (*domain.Subnet, error) {
	return &domain.Subnet{ID: uuid.New(), Name: name, VPCID: vpcID}, nil
}
func (r *NoopSubnetRepository) ListByVPC(ctx context.Context, vpcID uuid.UUID) ([]*domain.Subnet, error) {
	return []*domain.Subnet{}, nil
}
func (r *NoopSubnetRepository) Delete(ctx context.Context, id uuid.UUID) error { return nil }

// NoopVolumeRepository is a no-op volume repository.
type NoopVolumeRepository struct{}

func (r *NoopVolumeRepository) Create(ctx context.Context, v *domain.Volume) error { return nil }
func (r *NoopVolumeRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Volume, error) {
	return &domain.Volume{ID: id}, nil
}
func (r *NoopVolumeRepository) GetByName(ctx context.Context, name string) (*domain.Volume, error) {
	return &domain.Volume{ID: uuid.New(), Name: name}, nil
}
func (r *NoopVolumeRepository) List(ctx context.Context) ([]*domain.Volume, error) {
	return []*domain.Volume{}, nil
}
func (r *NoopVolumeRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Volume, error) {
	return []*domain.Volume{}, nil
}
func (r *NoopVolumeRepository) ListByInstanceID(ctx context.Context, instanceID uuid.UUID) ([]*domain.Volume, error) {
	return []*domain.Volume{}, nil
}
func (r *NoopVolumeRepository) Update(ctx context.Context, v *domain.Volume) error { return nil }
func (r *NoopVolumeRepository) Delete(ctx context.Context, id uuid.UUID) error     { return nil }

// NewNoopComputeBackend returns a no-op compute backend.
func NewNoopComputeBackend() ports.ComputeBackend {
	return &NoopComputeBackend{}
}

// NoopComputeBackend is a no-op compute backend implementation.
type NoopComputeBackend struct{}

func (c *NoopComputeBackend) CreateInstance(ctx context.Context, opts ports.CreateInstanceOptions) (string, error) {
	return "noop-id", nil
}
func (c *NoopComputeBackend) StopInstance(ctx context.Context, id string) error   { return nil }
func (c *NoopComputeBackend) DeleteInstance(ctx context.Context, id string) error { return nil }
func (c *NoopComputeBackend) GetInstanceLogs(ctx context.Context, id string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("")), nil
}
func (c *NoopComputeBackend) GetInstanceStats(ctx context.Context, id string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("{}")), nil
}
func (c *NoopComputeBackend) GetInstancePort(ctx context.Context, id string, internalPort string) (int, error) {
	return 0, nil
}
func (c *NoopComputeBackend) GetInstanceIP(ctx context.Context, id string) (string, error) {
	return "127.0.0.1", nil
}
func (c *NoopComputeBackend) GetConsoleURL(ctx context.Context, id string) (string, error) {
	return "vnc://127.0.0.1:5900", nil
}
func (c *NoopComputeBackend) Exec(ctx context.Context, id string, cmd []string) (string, error) {
	return "", nil
}
func (c *NoopComputeBackend) RunTask(ctx context.Context, opts ports.RunTaskOptions) (string, error) {
	return "task-id", nil
}
func (c *NoopComputeBackend) WaitTask(ctx context.Context, id string) (int64, error) { return 0, nil }
func (c *NoopComputeBackend) CreateNetwork(ctx context.Context, name string) (string, error) {
	return "net-id", nil
}
func (c *NoopComputeBackend) DeleteNetwork(ctx context.Context, id string) error { return nil }
func (c *NoopComputeBackend) AttachVolume(ctx context.Context, id string, volumePath string) error {
	return nil
}
func (c *NoopComputeBackend) DetachVolume(ctx context.Context, id string, volumePath string) error {
	return nil
}
func (c *NoopComputeBackend) Ping(ctx context.Context) error { return nil }
func (c *NoopComputeBackend) Type() string                   { return "noop" }

// NoopEventService is a no-op event service.
type NoopEventService struct{}

func (e *NoopEventService) RecordEvent(ctx context.Context, eventType, resourceID, resourceType string, data map[string]interface{}) error {
	return nil
}
func (e *NoopEventService) ListEvents(ctx context.Context, limit int) ([]*domain.Event, error) {
	return []*domain.Event{}, nil
}

// NoopAuditService is a no-op audit service.
type NoopAuditService struct{}

func (a *NoopAuditService) Log(ctx context.Context, userID uuid.UUID, action, resourceType, resourceID string, metadata map[string]interface{}) error {
	return nil
}
func (a *NoopAuditService) ListLogs(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.AuditLog, error) {
	return []*domain.AuditLog{}, nil
}

// NoopDatabaseRepository is a no-op database repository.
type NoopDatabaseRepository struct{}

func (r *NoopDatabaseRepository) Create(ctx context.Context, db *domain.Database) error { return nil }
func (r *NoopDatabaseRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Database, error) {
	return &domain.Database{ID: id}, nil
}
func (r *NoopDatabaseRepository) List(ctx context.Context) ([]*domain.Database, error) {
	return []*domain.Database{}, nil
}
func (r *NoopDatabaseRepository) Update(ctx context.Context, db *domain.Database) error { return nil }
func (r *NoopDatabaseRepository) Delete(ctx context.Context, id uuid.UUID) error        { return nil }

// NoopCacheRepository is a no-op cache repository.
type NoopCacheRepository struct{}

func (r *NoopCacheRepository) Create(ctx context.Context, cache *domain.Cache) error { return nil }
func (r *NoopCacheRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Cache, error) {
	return &domain.Cache{ID: id}, nil
}
func (r *NoopCacheRepository) GetByName(ctx context.Context, userID uuid.UUID, name string) (*domain.Cache, error) {
	return &domain.Cache{ID: uuid.New(), Name: name, UserID: userID}, nil
}
func (r *NoopCacheRepository) List(ctx context.Context, userID uuid.UUID) ([]*domain.Cache, error) {
	return []*domain.Cache{}, nil
}
func (r *NoopCacheRepository) Update(ctx context.Context, cache *domain.Cache) error { return nil }
func (r *NoopCacheRepository) Delete(ctx context.Context, id uuid.UUID) error        { return nil }

// NoopStorageRepository is a no-op storage repository.
type NoopStorageRepository struct{}

func (r *NoopStorageRepository) SaveMeta(ctx context.Context, obj *domain.Object) error { return nil }
func (r *NoopStorageRepository) GetMeta(ctx context.Context, bucket, key string) (*domain.Object, error) {
	return &domain.Object{Bucket: bucket, Key: key}, nil
}
func (r *NoopStorageRepository) List(ctx context.Context, bucket string) ([]*domain.Object, error) {
	return []*domain.Object{}, nil
}
func (r *NoopStorageRepository) SoftDelete(ctx context.Context, bucket, key string) error { return nil }
func (r *NoopStorageRepository) DeleteVersion(ctx context.Context, bucket, key, versionID string) error {
	return nil
}
func (r *NoopStorageRepository) GetMetaByVersion(ctx context.Context, bucket, key, versionID string) (*domain.Object, error) {
	return &domain.Object{Bucket: bucket, Key: key, VersionID: versionID}, nil
}
func (r *NoopStorageRepository) ListVersions(ctx context.Context, bucket, key string) ([]*domain.Object, error) {
	return []*domain.Object{}, nil
}

// Bucket operations
func (r *NoopStorageRepository) CreateBucket(ctx context.Context, bucket *domain.Bucket) error {
	return nil
}
func (r *NoopStorageRepository) GetBucket(ctx context.Context, name string) (*domain.Bucket, error) {
	return &domain.Bucket{Name: name}, nil
}
func (r *NoopStorageRepository) DeleteBucket(ctx context.Context, name string) error {
	return nil
}
func (r *NoopStorageRepository) ListBuckets(ctx context.Context, userID string) ([]*domain.Bucket, error) {
	return []*domain.Bucket{}, nil
}
func (r *NoopStorageRepository) SetBucketVersioning(ctx context.Context, name string, enabled bool) error {
	return nil
}

// Multipart operations
func (r *NoopStorageRepository) SaveMultipartUpload(ctx context.Context, upload *domain.MultipartUpload) error {
	return nil
}
func (r *NoopStorageRepository) GetMultipartUpload(ctx context.Context, uploadID uuid.UUID) (*domain.MultipartUpload, error) {
	return &domain.MultipartUpload{ID: uploadID}, nil
}
func (r *NoopStorageRepository) DeleteMultipartUpload(ctx context.Context, uploadID uuid.UUID) error {
	return nil
}
func (r *NoopStorageRepository) SavePart(ctx context.Context, part *domain.Part) error {
	return nil
}
func (r *NoopStorageRepository) ListParts(ctx context.Context, uploadID uuid.UUID) ([]*domain.Part, error) {
	return []*domain.Part{}, nil
}

// NoopFileStore is a no-op file store.
type NoopFileStore struct{}

func (r *NoopFileStore) Write(ctx context.Context, bucket, key string, reader io.Reader) (int64, error) {
	return 0, nil
}
func (r *NoopFileStore) Read(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("")), nil
}
func (r *NoopFileStore) Delete(ctx context.Context, bucket, key string) error { return nil }
func (r *NoopFileStore) GetClusterStatus(ctx context.Context) (*domain.StorageCluster, error) {
	return &domain.StorageCluster{Nodes: []domain.StorageNode{}}, nil
}

func (r *NoopFileStore) Assemble(ctx context.Context, bucket, key string, parts []string) (int64, error) {
	return 0, nil
}

// NoopFunctionRepository is a no-op function repository.
type NoopFunctionRepository struct{}

func (r *NoopFunctionRepository) Create(ctx context.Context, f *domain.Function) error { return nil }
func (r *NoopFunctionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Function, error) {
	return &domain.Function{ID: id}, nil
}
func (r *NoopFunctionRepository) GetByName(ctx context.Context, userID uuid.UUID, name string) (*domain.Function, error) {
	return &domain.Function{ID: uuid.New(), Name: name, UserID: userID}, nil
}
func (r *NoopFunctionRepository) List(ctx context.Context, userID uuid.UUID) ([]*domain.Function, error) {
	return []*domain.Function{}, nil
}
func (r *NoopFunctionRepository) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (r *NoopFunctionRepository) CreateInvocation(ctx context.Context, i *domain.Invocation) error {
	return nil
}
func (r *NoopFunctionRepository) GetInvocations(ctx context.Context, functionID uuid.UUID, limit int) ([]*domain.Invocation, error) {
	return []*domain.Invocation{}, nil
}

// NoopUserRepository is a no-op user repository.
type NoopUserRepository struct{}

func (r *NoopUserRepository) Create(ctx context.Context, user *domain.User) error { return nil }
func (r *NoopUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, nil // Return nil so Register doesn't fail
}
func (r *NoopUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return &domain.User{ID: id}, nil
}
func (r *NoopUserRepository) Update(ctx context.Context, user *domain.User) error { return nil }
func (r *NoopUserRepository) List(ctx context.Context) ([]*domain.User, error) {
	return []*domain.User{}, nil
}
func (r *NoopUserRepository) Delete(ctx context.Context, id uuid.UUID) error { return nil }

// NoopIdentityService is a no-op identity service.
type NoopIdentityService struct{}

func (s *NoopIdentityService) CreateKey(ctx context.Context, userID uuid.UUID, name string) (*domain.APIKey, error) {
	return &domain.APIKey{ID: uuid.New(), UserID: userID, Name: name, Key: "noop-key"}, nil
}
func (s *NoopIdentityService) ValidateAPIKey(ctx context.Context, key string) (*domain.APIKey, error) {
	return &domain.APIKey{ID: uuid.New(), Key: key}, nil
}
func (s *NoopIdentityService) ListKeys(ctx context.Context, userID uuid.UUID) ([]*domain.APIKey, error) {
	return []*domain.APIKey{}, nil
}
func (s *NoopIdentityService) RevokeKey(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	return nil
}
func (s *NoopIdentityService) RotateKey(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*domain.APIKey, error) {
	return &domain.APIKey{ID: id, UserID: userID, Key: "rotated-key"}, nil
}

// NoopIdentityRepository is a no-op identity repository.
type NoopIdentityRepository struct{}

func (r *NoopIdentityRepository) CreateAPIKey(ctx context.Context, apiKey *domain.APIKey) error {
	return nil
}
func (r *NoopIdentityRepository) GetAPIKeyByKey(ctx context.Context, key string) (*domain.APIKey, error) {
	return &domain.APIKey{Key: key}, nil
}
func (r *NoopIdentityRepository) GetAPIKeyByID(ctx context.Context, id uuid.UUID) (*domain.APIKey, error) {
	return &domain.APIKey{ID: id}, nil
}
func (r *NoopIdentityRepository) ListAPIKeysByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.APIKey, error) {
	return []*domain.APIKey{}, nil
}
func (r *NoopIdentityRepository) DeleteAPIKey(ctx context.Context, id uuid.UUID) error { return nil }

// NewNoopStorageBackend returns a no-op storage backend.
func NewNoopStorageBackend() ports.StorageBackend {
	return &NoopStorageBackend{}
}

// NoopStorageBackend is a no-op storage backend implementation.
type NoopStorageBackend struct{}

func (b *NoopStorageBackend) CreateVolume(ctx context.Context, name string, sizeGB int) (string, error) {
	return "/tmp/" + name, nil
}
func (b *NoopStorageBackend) DeleteVolume(ctx context.Context, name string) error { return nil }
func (b *NoopStorageBackend) AttachVolume(ctx context.Context, vol, inst string) error {
	return nil
}
func (b *NoopStorageBackend) DetachVolume(ctx context.Context, vol, inst string) error {
	return nil
}
func (b *NoopStorageBackend) CreateSnapshot(ctx context.Context, vol, snap string) error  { return nil }
func (b *NoopStorageBackend) RestoreSnapshot(ctx context.Context, vol, snap string) error { return nil }
func (b *NoopStorageBackend) DeleteSnapshot(ctx context.Context, snap string) error       { return nil }
func (b *NoopStorageBackend) Ping(ctx context.Context) error                              { return nil }
func (b *NoopStorageBackend) Type() string                                                { return "noop" }

// NoopLBRepository is a no-op load balancer repository.
type NoopLBRepository struct{}

func (r *NoopLBRepository) Create(ctx context.Context, lb *domain.LoadBalancer) error { return nil }
func (r *NoopLBRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.LoadBalancer, error) {
	return &domain.LoadBalancer{ID: id}, nil
}
func (r *NoopLBRepository) GetByIdempotencyKey(ctx context.Context, key string) (*domain.LoadBalancer, error) {
	return &domain.LoadBalancer{ID: uuid.New(), IdempotencyKey: key}, nil
}
func (r *NoopLBRepository) List(ctx context.Context) ([]*domain.LoadBalancer, error) {
	return []*domain.LoadBalancer{}, nil
}
func (r *NoopLBRepository) ListAll(ctx context.Context) ([]*domain.LoadBalancer, error) {
	return []*domain.LoadBalancer{}, nil
}
func (r *NoopLBRepository) Update(ctx context.Context, lb *domain.LoadBalancer) error { return nil }
func (r *NoopLBRepository) Delete(ctx context.Context, id uuid.UUID) error            { return nil }
func (r *NoopLBRepository) AddTarget(ctx context.Context, target *domain.LBTarget) error {
	return nil
}
func (r *NoopLBRepository) RemoveTarget(ctx context.Context, lbID, instanceID uuid.UUID) error {
	return nil
}
func (r *NoopLBRepository) ListTargets(ctx context.Context, lbID uuid.UUID) ([]*domain.LBTarget, error) {
	return []*domain.LBTarget{}, nil
}
func (r *NoopLBRepository) UpdateTargetHealth(ctx context.Context, lbID, instanceID uuid.UUID, health string) error {
	return nil
}
func (r *NoopLBRepository) GetTargetsForInstance(ctx context.Context, instanceID uuid.UUID) ([]*domain.LBTarget, error) {
	return []*domain.LBTarget{}, nil
}
