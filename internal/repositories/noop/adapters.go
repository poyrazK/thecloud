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
func (r *NoopSubnetRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subnet, error) {
	return &domain.Subnet{ID: id}, nil
}
func (r *NoopSubnetRepository) GetByName(ctx context.Context, vpcID uuid.UUID, name string) (*domain.Subnet, error) {
	return &domain.Subnet{ID: uuid.New(), Name: name, VPCID: vpcID}, nil
}
func (r *NoopSubnetRepository) ListByVPC(ctx context.Context, vpcID uuid.UUID) ([]*domain.Subnet, error) {
	return []*domain.Subnet{}, nil
}
func (r *NoopSubnetRepository) Delete(ctx context.Context, id uuid.UUID) error { return nil }

// NoopComputeBackend is a no-op compute backend.
type NoopComputeBackend struct{}

func (c *NoopComputeBackend) LaunchInstanceWithOptions(ctx context.Context, opts ports.CreateInstanceOptions) (string, error) {
	return "cid", nil
}
func (c *NoopComputeBackend) StartInstance(ctx context.Context, id string) error  { return nil }
func (c *NoopComputeBackend) StopInstance(ctx context.Context, id string) error   { return nil }
func (c *NoopComputeBackend) DeleteInstance(ctx context.Context, id string) error { return nil }
func (c *NoopComputeBackend) GetInstanceLogs(ctx context.Context, id string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("logs")), nil
}
func (c *NoopComputeBackend) GetInstanceStats(ctx context.Context, id string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("{}")), nil
}
func (c *NoopComputeBackend) GetInstancePort(ctx context.Context, id string, internalPort string) (int, error) {
	return 80, nil
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

// NoopEventService implements ports.EventService.
type NoopEventService struct{}

func (e *NoopEventService) RecordEvent(ctx context.Context, action, resourceID, resourceType string, metadata map[string]interface{}) error {
	return nil
}
func (e *NoopEventService) ListEvents(ctx context.Context, limit int) ([]*domain.Event, error) {
	return []*domain.Event{}, nil
}

// NoopAuditService implements ports.AuditService.
type NoopAuditService struct{}

func (a *NoopAuditService) Log(ctx context.Context, userID uuid.UUID, action, resourceType, resourceID string, details map[string]interface{}) error {
	return nil
}
func (a *NoopAuditService) ListLogs(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.AuditLog, error) {
	return []*domain.AuditLog{}, nil
}

// NoopClusterRepository is a no-op cluster repository.
type NoopClusterRepository struct{}

func (r *NoopClusterRepository) Create(ctx context.Context, c *domain.Cluster) error { return nil }
func (r *NoopClusterRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Cluster, error) {
	return nil, nil
}
func (r *NoopClusterRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Cluster, error) {
	return nil, nil
}
func (r *NoopClusterRepository) ListAll(ctx context.Context) ([]*domain.Cluster, error) {
	return nil, nil
}
func (r *NoopClusterRepository) Update(ctx context.Context, c *domain.Cluster) error      { return nil }
func (r *NoopClusterRepository) Delete(ctx context.Context, id uuid.UUID) error           { return nil }
func (r *NoopClusterRepository) AddNode(ctx context.Context, n *domain.ClusterNode) error { return nil }
func (r *NoopClusterRepository) GetNodes(ctx context.Context, clusterID uuid.UUID) ([]*domain.ClusterNode, error) {
	return []*domain.ClusterNode{}, nil
}
func (r *NoopClusterRepository) UpdateNode(ctx context.Context, n *domain.ClusterNode) error {
	return nil
}
func (r *NoopClusterRepository) DeleteNode(ctx context.Context, id uuid.UUID) error { return nil }

// NoopClusterService
type NoopClusterService struct{}

func (s *NoopClusterService) CreateCluster(ctx context.Context, name string, vpcID uuid.UUID) (*domain.Cluster, error) {
	return nil, nil
}
func (s *NoopClusterService) DeleteCluster(ctx context.Context, id uuid.UUID) error { return nil }
func (s *NoopClusterService) GetCluster(ctx context.Context, id uuid.UUID) (*domain.Cluster, error) {
	return nil, nil
}
func (s *NoopClusterService) ListClusters(ctx context.Context) ([]*domain.Cluster, error) {
	return nil, nil
}
func (s *NoopClusterService) AddNode(ctx context.Context, clusterID uuid.UUID, role string) (*domain.ClusterNode, error) {
	return nil, nil
}
func (s *NoopClusterService) RemoveNode(ctx context.Context, clusterID, nodeID uuid.UUID) error {
	return nil
}
func (s *NoopClusterService) GetKubeconfig(ctx context.Context, clusterID uuid.UUID) (string, error) {
	return "", nil
}

// NoopSecretService is a no-op secret service.
type NoopSecretService struct{}

func (s *NoopSecretService) CreateSecret(ctx context.Context, name, value, desc string) (*domain.Secret, error) {
	return &domain.Secret{ID: uuid.New(), Name: name}, nil
}
func (s *NoopSecretService) GetSecret(ctx context.Context, id uuid.UUID) (*domain.Secret, error) {
	return nil, nil
}
func (s *NoopSecretService) GetSecretByName(ctx context.Context, name string) (*domain.Secret, error) {
	return nil, nil
}
func (s *NoopSecretService) ListSecrets(ctx context.Context) ([]*domain.Secret, error) {
	return []*domain.Secret{}, nil
}
func (s *NoopSecretService) DeleteSecret(ctx context.Context, id uuid.UUID) error { return nil }
func (s *NoopSecretService) Encrypt(ctx context.Context, userID uuid.UUID, plain string) (string, error) {
	return plain, nil
}
func (s *NoopSecretService) Decrypt(ctx context.Context, userID uuid.UUID, cipher string) (string, error) {
	return cipher, nil
}

// NoopSecurityGroupService is a no-op security group service.
type NoopSecurityGroupService struct{}

func (s *NoopSecurityGroupService) CreateGroup(ctx context.Context, vpcID uuid.UUID, name, desc string) (*domain.SecurityGroup, error) {
	return &domain.SecurityGroup{ID: uuid.New(), Name: name}, nil
}
func (s *NoopSecurityGroupService) GetGroup(ctx context.Context, idOrName string, vpcID uuid.UUID) (*domain.SecurityGroup, error) {
	return &domain.SecurityGroup{ID: uuid.New(), Name: idOrName}, nil
}
func (s *NoopSecurityGroupService) ListGroups(ctx context.Context, vpcID uuid.UUID) ([]*domain.SecurityGroup, error) {
	return []*domain.SecurityGroup{}, nil
}
func (s *NoopSecurityGroupService) DeleteGroup(ctx context.Context, id uuid.UUID) error { return nil }
func (s *NoopSecurityGroupService) AddRule(ctx context.Context, groupID uuid.UUID, rule domain.SecurityRule) (*domain.SecurityRule, error) {
	return &rule, nil
}
func (s *NoopSecurityGroupService) RemoveRule(ctx context.Context, ruleID uuid.UUID) error {
	return nil
}
func (s *NoopSecurityGroupService) AttachToInstance(ctx context.Context, instID, groupID uuid.UUID) error {
	return nil
}
func (s *NoopSecurityGroupService) DetachFromInstance(ctx context.Context, instID, groupID uuid.UUID) error {
	return nil
}

// NoopStorageService is a no-op storage service.
type NoopStorageService struct{}

func (s *NoopStorageService) Upload(ctx context.Context, bucket, key string, r io.Reader) (*domain.Object, error) {
	return &domain.Object{Key: key}, nil
}
func (s *NoopStorageService) Download(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("data")), nil
}
func (s *NoopStorageService) Delete(ctx context.Context, bucket, key string) error { return nil }
func (s *NoopStorageService) CreateBucket(ctx context.Context, name string) error  { return nil }
func (s *NoopStorageService) DeleteBucket(ctx context.Context, name string) error  { return nil }

// NoopLBService is a no-op LB service.
type NoopLBService struct{}

func (s *NoopLBService) Create(ctx context.Context, name string, vpcID uuid.UUID, port int, algo string, idempotencyKey string) (*domain.LoadBalancer, error) {
	return &domain.LoadBalancer{ID: uuid.New(), Name: name}, nil
}
func (s *NoopLBService) Get(ctx context.Context, id uuid.UUID) (*domain.LoadBalancer, error) {
	return &domain.LoadBalancer{ID: id}, nil
}
func (s *NoopLBService) List(ctx context.Context) ([]*domain.LoadBalancer, error) {
	return []*domain.LoadBalancer{}, nil
}
func (s *NoopLBService) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (s *NoopLBService) AddTarget(ctx context.Context, lbID, instanceID uuid.UUID, port int, weight int) error {
	return nil
}
func (s *NoopLBService) RemoveTarget(ctx context.Context, lbID, instanceID uuid.UUID) error {
	return nil
}
func (s *NoopLBService) ListTargets(ctx context.Context, lbID uuid.UUID) ([]*domain.LBTarget, error) {
	return []*domain.LBTarget{}, nil
}

// NoopTaskQueue is a no-op task queue.
type NoopTaskQueue struct{}

func (q *NoopTaskQueue) Enqueue(ctx context.Context, queue string, payload interface{}) error {
	return nil
}
func (q *NoopTaskQueue) Dequeue(ctx context.Context, queue string) (string, error) { return "", nil }

// --- New No-Ops (for benchmarks and system tests) ---

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
func (r *NoopVolumeRepository) ListByInstanceID(ctx context.Context, id uuid.UUID) ([]*domain.Volume, error) {
	return []*domain.Volume{}, nil
}
func (r *NoopVolumeRepository) Update(ctx context.Context, v *domain.Volume) error { return nil }
func (r *NoopVolumeRepository) Delete(ctx context.Context, id uuid.UUID) error     { return nil }

type NoopFunctionRepository struct{}

func (r *NoopFunctionRepository) Create(ctx context.Context, fn *domain.Function) error { return nil }
func (r *NoopFunctionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Function, error) {
	return &domain.Function{ID: id}, nil
}
func (r *NoopFunctionRepository) GetByName(ctx context.Context, userID uuid.UUID, name string) (*domain.Function, error) {
	return &domain.Function{ID: uuid.New(), Name: name}, nil
}
func (r *NoopFunctionRepository) List(ctx context.Context, userID uuid.UUID) ([]*domain.Function, error) {
	return []*domain.Function{}, nil
}
func (r *NoopFunctionRepository) Update(ctx context.Context, fn *domain.Function) error { return nil }
func (r *NoopFunctionRepository) Delete(ctx context.Context, id uuid.UUID) error        { return nil }
func (r *NoopFunctionRepository) GetInvocations(ctx context.Context, fnID uuid.UUID, limit int) ([]*domain.Invocation, error) {
	return []*domain.Invocation{}, nil
}
func (r *NoopFunctionRepository) CreateInvocation(ctx context.Context, inv *domain.Invocation) error {
	return nil
}

type NoopFileStore struct{}

func (s *NoopFileStore) Write(ctx context.Context, bucket, key string, r io.Reader) (int64, error) {
	return 0, nil
}
func (s *NoopFileStore) Read(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("")), nil
}
func (s *NoopFileStore) Delete(ctx context.Context, bucket, key string) error { return nil }
func (s *NoopFileStore) GetClusterStatus(ctx context.Context) (*domain.StorageCluster, error) {
	return &domain.StorageCluster{}, nil
}
func (s *NoopFileStore) Assemble(ctx context.Context, bucket, key string, parts []string) (int64, error) {
	return 0, nil
}

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

type NoopUserRepository struct{}

func (r *NoopUserRepository) Create(ctx context.Context, u *domain.User) error { return nil }
func (r *NoopUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return &domain.User{ID: id}, nil
}
func (r *NoopUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return &domain.User{ID: uuid.New(), Email: email}, nil
}
func (r *NoopUserRepository) Update(ctx context.Context, u *domain.User) error { return nil }
func (r *NoopUserRepository) List(ctx context.Context) ([]*domain.User, error) {
	return []*domain.User{}, nil
}
func (r *NoopUserRepository) Delete(ctx context.Context, id uuid.UUID) error { return nil }

type NoopIdentityService struct{}

func (s *NoopIdentityService) CreateKey(ctx context.Context, userID uuid.UUID, name string) (*domain.APIKey, error) {
	return &domain.APIKey{ID: uuid.New()}, nil
}
func (s *NoopIdentityService) ValidateAPIKey(ctx context.Context, key string) (*domain.APIKey, error) {
	return &domain.APIKey{ID: uuid.New()}, nil
}
func (s *NoopIdentityService) ListKeys(ctx context.Context, userID uuid.UUID) ([]*domain.APIKey, error) {
	return []*domain.APIKey{}, nil
}
func (s *NoopIdentityService) RevokeKey(ctx context.Context, userID, id uuid.UUID) error { return nil }
func (s *NoopIdentityService) RotateKey(ctx context.Context, userID, id uuid.UUID) (*domain.APIKey, error) {
	return &domain.APIKey{ID: id}, nil
}

type NoopCacheRepository struct{}

func (r *NoopCacheRepository) Create(ctx context.Context, c *domain.Cache) error { return nil }
func (r *NoopCacheRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Cache, error) {
	return &domain.Cache{ID: id}, nil
}
func (r *NoopCacheRepository) List(ctx context.Context) ([]*domain.Cache, error) {
	return []*domain.Cache{}, nil
}
func (r *NoopCacheRepository) Update(ctx context.Context, c *domain.Cache) error { return nil }
func (r *NoopCacheRepository) Delete(ctx context.Context, id uuid.UUID) error    { return nil }

type NoopStorageRepository struct{}

func (r *NoopStorageRepository) SaveMeta(ctx context.Context, obj *domain.Object) error { return nil }
func (r *NoopStorageRepository) GetMeta(ctx context.Context, bucket, key string) (*domain.Object, error) {
	return &domain.Object{Bucket: bucket, Key: key}, nil
}
func (r *NoopStorageRepository) List(ctx context.Context, bucket string) ([]*domain.Object, error) {
	return []*domain.Object{}, nil
}
func (r *NoopStorageRepository) SoftDelete(ctx context.Context, bucket, key string) error { return nil }
func (r *NoopStorageRepository) DeleteVersion(ctx context.Context, bucket, key, ver string) error {
	return nil
}
func (r *NoopStorageRepository) GetMetaByVersion(ctx context.Context, bucket, key, ver string) (*domain.Object, error) {
	return &domain.Object{}, nil
}
func (r *NoopStorageRepository) ListVersions(ctx context.Context, bucket, key string) ([]*domain.Object, error) {
	return []*domain.Object{}, nil
}
func (r *NoopStorageRepository) CreateBucket(ctx context.Context, b *domain.Bucket) error { return nil }
func (r *NoopStorageRepository) GetBucket(ctx context.Context, name string) (*domain.Bucket, error) {
	return &domain.Bucket{Name: name}, nil
}
func (r *NoopStorageRepository) DeleteBucket(ctx context.Context, name string) error { return nil }
func (r *NoopStorageRepository) ListBuckets(ctx context.Context, uid string) ([]*domain.Bucket, error) {
	return []*domain.Bucket{}, nil
}
func (r *NoopStorageRepository) SetBucketVersioning(ctx context.Context, name string, enabled bool) error {
	return nil
}
func (r *NoopStorageRepository) SaveMultipartUpload(ctx context.Context, u *domain.MultipartUpload) error {
	return nil
}
func (r *NoopStorageRepository) GetMultipartUpload(ctx context.Context, id uuid.UUID) (*domain.MultipartUpload, error) {
	return &domain.MultipartUpload{}, nil
}
func (r *NoopStorageRepository) DeleteMultipartUpload(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (r *NoopStorageRepository) SavePart(ctx context.Context, p *domain.Part) error { return nil }
func (r *NoopStorageRepository) ListParts(ctx context.Context, uid uuid.UUID) ([]*domain.Part, error) {
	return []*domain.Part{}, nil
}

// NoopStorageBackend
type NoopStorageBackend struct{}

func NewNoopStorageBackend() *NoopStorageBackend { return &NoopStorageBackend{} }
func (s *NoopStorageBackend) CreateSnapshot(ctx context.Context, volumeID, name string) error {
	return nil
}
func (s *NoopStorageBackend) DeleteSnapshot(ctx context.Context, snapshotID string) error { return nil }
func (s *NoopStorageBackend) RestoreSnapshot(ctx context.Context, snapshotID string) (string, error) {
	return "vol-restored", nil
}
func (s *NoopStorageBackend) CreateVolume(ctx context.Context, size int64) (string, error) {
	return "vol-1", nil
}
func (s *NoopStorageBackend) DeleteVolume(ctx context.Context, id string) error { return nil }
func (s *NoopStorageBackend) ResizeVolume(ctx context.Context, id string, newSize int64) error {
	return nil
}
func (s *NoopStorageBackend) AttachVolume(ctx context.Context, volID, instanceID string) error {
	return nil
}
func (s *NoopStorageBackend) DetachVolume(ctx context.Context, volID, instanceID string) error {
	return nil
}
func (s *NoopStorageBackend) Ping(ctx context.Context) error { return nil }
func (s *NoopStorageBackend) Type() string                   { return "noop" }
