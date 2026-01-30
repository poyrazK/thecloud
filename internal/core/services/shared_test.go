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

// MockUserRepo
type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) Create(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *MockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepo) Update(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *MockUserRepo) List(ctx context.Context) ([]*domain.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.User), args.Error(1)
}

func (m *MockUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

// MockIdentityService
type MockIdentityService struct {
	mock.Mock
}

func (m *MockIdentityService) CreateKey(ctx context.Context, userID uuid.UUID, name string) (*domain.APIKey, error) {
	args := m.Called(ctx, userID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.APIKey), args.Error(1)
}
func (m *MockIdentityService) ValidateAPIKey(ctx context.Context, key string) (*domain.APIKey, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.APIKey), args.Error(1)
}
func (m *MockIdentityService) ListKeys(ctx context.Context, userID uuid.UUID) ([]*domain.APIKey, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.APIKey), args.Error(1)
}
func (m *MockIdentityService) RevokeKey(ctx context.Context, userID, id uuid.UUID) error {
	args := m.Called(ctx, userID, id)
	return args.Error(0)
}
func (m *MockIdentityService) RotateKey(ctx context.Context, userID, id uuid.UUID) (*domain.APIKey, error) {
	args := m.Called(ctx, userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.APIKey), args.Error(1)
}

// MockAutoScalingRepo
type MockAutoScalingRepo struct{ mock.Mock }

func (m *MockAutoScalingRepo) CreateGroup(ctx context.Context, group *domain.ScalingGroup) error {
	return m.Called(ctx, group).Error(0)
}
func (m *MockAutoScalingRepo) GetGroupByID(ctx context.Context, id uuid.UUID) (*domain.ScalingGroup, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ScalingGroup), args.Error(1)
}
func (m *MockAutoScalingRepo) GetGroupByIdempotencyKey(ctx context.Context, key string) (*domain.ScalingGroup, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ScalingGroup), args.Error(1)
}
func (m *MockAutoScalingRepo) ListGroups(ctx context.Context) ([]*domain.ScalingGroup, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ScalingGroup), args.Error(1)
}
func (m *MockAutoScalingRepo) ListAllGroups(ctx context.Context) ([]*domain.ScalingGroup, error) {
	ret := m.Called(ctx)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).([]*domain.ScalingGroup), ret.Error(1)
}
func (m *MockAutoScalingRepo) CountGroupsByVPC(ctx context.Context, vpcID uuid.UUID) (int, error) {
	args := m.Called(ctx, vpcID)
	return args.Int(0), args.Error(1)
}
func (m *MockAutoScalingRepo) UpdateGroup(ctx context.Context, group *domain.ScalingGroup) error {
	return m.Called(ctx, group).Error(0)
}
func (m *MockAutoScalingRepo) DeleteGroup(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockAutoScalingRepo) CreatePolicy(ctx context.Context, policy *domain.ScalingPolicy) error {
	args := m.Called(ctx, policy)
	return args.Error(0)
}
func (m *MockAutoScalingRepo) GetPoliciesForGroup(ctx context.Context, groupID uuid.UUID) ([]*domain.ScalingPolicy, error) {
	args := m.Called(ctx, groupID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ScalingPolicy), args.Error(1)
}
func (m *MockAutoScalingRepo) GetAllPolicies(ctx context.Context, groupIDs []uuid.UUID) (map[uuid.UUID][]*domain.ScalingPolicy, error) {
	args := m.Called(ctx, groupIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID][]*domain.ScalingPolicy), args.Error(1)
}
func (m *MockAutoScalingRepo) UpdatePolicyLastScaled(ctx context.Context, policyID uuid.UUID, t time.Time) error {
	args := m.Called(ctx, policyID, t)
	return args.Error(0)
}
func (m *MockAutoScalingRepo) DeletePolicy(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockAutoScalingRepo) AddInstanceToGroup(ctx context.Context, groupID, instanceID uuid.UUID) error {
	return m.Called(ctx, groupID, instanceID).Error(0)
}
func (m *MockAutoScalingRepo) RemoveInstanceFromGroup(ctx context.Context, groupID, instanceID uuid.UUID) error {
	return m.Called(ctx, groupID, instanceID).Error(0)
}
func (m *MockAutoScalingRepo) GetInstancesInGroup(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, groupID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uuid.UUID), args.Error(1)
}
func (m *MockAutoScalingRepo) GetAllScalingGroupInstances(ctx context.Context, groupIDs []uuid.UUID) (map[uuid.UUID][]uuid.UUID, error) {
	args := m.Called(ctx, groupIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID][]uuid.UUID), args.Error(1)
}
func (m *MockAutoScalingRepo) GetAverageCPU(ctx context.Context, instanceIDs []uuid.UUID, since time.Time) (float64, error) {
	args := m.Called(ctx, instanceIDs, since)
	return args.Get(0).(float64), args.Error(1)
}

// MockInstanceService
type MockInstanceService struct{ mock.Mock }

func (m *MockInstanceService) LaunchInstance(ctx context.Context, name, image, ports, instanceType string, vpcID, subnetID *uuid.UUID, volumes []domain.VolumeAttachment) (*domain.Instance, error) {
	args := m.Called(ctx, name, image, ports, instanceType, vpcID, subnetID, volumes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Instance), args.Error(1)
}
func (m *MockInstanceService) StopInstance(ctx context.Context, idOrName string) error {
	return m.Called(ctx, idOrName).Error(0)
}
func (m *MockInstanceService) ListInstances(ctx context.Context) ([]*domain.Instance, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Instance), args.Error(1)
}
func (m *MockInstanceService) GetInstance(ctx context.Context, idOrName string) (*domain.Instance, error) {
	args := m.Called(ctx, idOrName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Instance), args.Error(1)
}
func (m *MockInstanceService) GetInstanceLogs(ctx context.Context, idOrName string) (string, error) {
	args := m.Called(ctx, idOrName)
	return args.String(0), args.Error(1)
}
func (m *MockInstanceService) GetInstanceStats(ctx context.Context, idOrName string) (*domain.InstanceStats, error) {
	args := m.Called(ctx, idOrName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.InstanceStats), args.Error(1)
}
func (m *MockInstanceService) GetConsoleURL(ctx context.Context, idOrName string) (string, error) {
	return m.GetInstanceLogs(ctx, idOrName)
}
func (m *MockInstanceService) TerminateInstance(ctx context.Context, idOrName string) error {
	return m.Called(ctx, idOrName).Error(0)
}
func (m *MockInstanceService) Exec(ctx context.Context, idOrName string, cmd []string) (string, error) {
	args := m.Called(ctx, idOrName, cmd)
	return args.String(0), args.Error(1)
}

// MockInstanceTypeRepo
type MockInstanceTypeRepo struct{ mock.Mock }

func (m *MockInstanceTypeRepo) List(ctx context.Context) ([]*domain.InstanceType, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.InstanceType), args.Error(1)
}
func (m *MockInstanceTypeRepo) GetByID(ctx context.Context, id string) (*domain.InstanceType, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.InstanceType), args.Error(1)
}
func (m *MockInstanceTypeRepo) Create(ctx context.Context, it *domain.InstanceType) (*domain.InstanceType, error) {
	args := m.Called(ctx, it)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.InstanceType), args.Error(1)
}
func (m *MockInstanceTypeRepo) Update(ctx context.Context, it *domain.InstanceType) (*domain.InstanceType, error) {
	args := m.Called(ctx, it)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.InstanceType), args.Error(1)
}
func (m *MockInstanceTypeRepo) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

// MockLBService
type MockLBService struct{ mock.Mock }

func (m *MockLBService) Create(ctx context.Context, name string, vpcID uuid.UUID, port int, algo string, idempotencyKey string) (*domain.LoadBalancer, error) {
	args := m.Called(ctx, name, vpcID, port, algo, idempotencyKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.LoadBalancer), args.Error(1)
}
func (m *MockLBService) Get(ctx context.Context, id uuid.UUID) (*domain.LoadBalancer, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.LoadBalancer), args.Error(1)
}
func (m *MockLBService) List(ctx context.Context) ([]*domain.LoadBalancer, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.LoadBalancer), args.Error(1)
}
func (m *MockLBService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockLBService) AddTarget(ctx context.Context, lbID, instanceID uuid.UUID, port, weight int) error {
	args := m.Called(ctx, lbID, instanceID, port, weight)
	return args.Error(0)
}
func (m *MockLBService) RemoveTarget(ctx context.Context, lbID, instanceID uuid.UUID) error {
	args := m.Called(ctx, lbID, instanceID)
	return args.Error(0)
}
func (m *MockLBService) ListTargets(ctx context.Context, lbID uuid.UUID) ([]*domain.LBTarget, error) {
	args := m.Called(ctx, lbID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.LBTarget), args.Error(1)
}

// MockEventService
type MockEventService struct{ mock.Mock }

func (m *MockEventService) RecordEvent(ctx context.Context, eType, resourceID, resourceType string, meta map[string]interface{}) error {
	args := m.Called(ctx, eType, resourceID, resourceType, meta)
	return args.Error(0)
}
func (m *MockEventService) ListEvents(ctx context.Context, limit int) ([]*domain.Event, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Event), args.Error(1)
}

// MockClock
type MockClock struct{ mock.Mock }

func (m *MockClock) Now() time.Time {
	args := m.Called()
	return args.Get(0).(time.Time)
}

// MockVpcRepo
type MockVpcRepo struct{ mock.Mock }

func (m *MockVpcRepo) Create(ctx context.Context, vpc *domain.VPC) error {
	return m.Called(ctx, vpc).Error(0)
}
func (m *MockVpcRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.VPC, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.VPC), args.Error(1)
}
func (m *MockVpcRepo) GetByName(ctx context.Context, name string) (*domain.VPC, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.VPC), args.Error(1)
}
func (m *MockVpcRepo) List(ctx context.Context) ([]*domain.VPC, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.VPC), args.Error(1)
}

func (m *MockVpcRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

// MockSubnetRepo
type MockSubnetRepo struct{ mock.Mock }

func (m *MockSubnetRepo) Create(ctx context.Context, subnet *domain.Subnet) error {
	return m.Called(ctx, subnet).Error(0)
}
func (m *MockSubnetRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subnet, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Subnet), args.Error(1)
}
func (m *MockSubnetRepo) GetByName(ctx context.Context, vpcID uuid.UUID, name string) (*domain.Subnet, error) {
	args := m.Called(ctx, vpcID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Subnet), args.Error(1)
}
func (m *MockSubnetRepo) ListByVPC(ctx context.Context, vpcID uuid.UUID) ([]*domain.Subnet, error) {
	args := m.Called(ctx, vpcID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Subnet), args.Error(1)
}
func (m *MockSubnetRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Object), args.Error(1)
}
func (m *MockStorageRepo) List(ctx context.Context, bucket string) ([]*domain.Object, error) {
	args := m.Called(ctx, bucket)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Object), args.Error(1)
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Object), args.Error(1)
}
func (m *MockStorageRepo) ListVersions(ctx context.Context, bucket, key string) ([]*domain.Object, error) {
	args := m.Called(ctx, bucket, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Object), args.Error(1)
}

func (m *MockStorageRepo) CreateBucket(ctx context.Context, bucket *domain.Bucket) error {
	return m.Called(ctx, bucket).Error(0)
}
func (m *MockStorageRepo) GetBucket(ctx context.Context, name string) (*domain.Bucket, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Bucket), args.Error(1)
}
func (m *MockStorageRepo) DeleteBucket(ctx context.Context, name string) error {
	return m.Called(ctx, name).Error(0)
}
func (m *MockStorageRepo) ListBuckets(ctx context.Context, userID string) ([]*domain.Bucket, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Bucket), args.Error(1)
}
func (m *MockStorageRepo) SetBucketVersioning(ctx context.Context, name string, enabled bool) error {
	return m.Called(ctx, name, enabled).Error(0)
}

func (m *MockStorageRepo) SaveMultipartUpload(ctx context.Context, upload *domain.MultipartUpload) error {
	return m.Called(ctx, upload).Error(0)
}
func (m *MockStorageRepo) GetMultipartUpload(ctx context.Context, uploadID uuid.UUID) (*domain.MultipartUpload, error) {
	args := m.Called(ctx, uploadID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MultipartUpload), args.Error(1)
}
func (m *MockStorageRepo) DeleteMultipartUpload(ctx context.Context, uploadID uuid.UUID) error {
	return m.Called(ctx, uploadID).Error(0)
}
func (m *MockStorageRepo) SavePart(ctx context.Context, part *domain.Part) error {
	return m.Called(ctx, part).Error(0)
}
func (m *MockStorageRepo) ListParts(ctx context.Context, uploadID uuid.UUID) ([]*domain.Part, error) {
	args := m.Called(ctx, uploadID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Part), args.Error(1)
}

// MockFileStore
type MockFileStore struct {
	mock.Mock
}

func (m *MockFileStore) Write(ctx context.Context, bucket, key string, r io.Reader) (int64, error) {
	args := m.Called(ctx, bucket, key, r)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockFileStore) Read(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	args := m.Called(ctx, bucket, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}
func (m *MockFileStore) Delete(ctx context.Context, bucket, key string) error {
	args := m.Called(ctx, bucket, key)
	return args.Error(0)
}
func (m *MockFileStore) GetClusterStatus(ctx context.Context) (*domain.StorageCluster, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.StorageCluster), args.Error(1)
}
func (m *MockFileStore) Assemble(ctx context.Context, bucket, key string, parts []string) (int64, error) {
	args := m.Called(ctx, bucket, key, parts)
	return args.Get(0).(int64), args.Error(1)
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Volume), args.Error(1)
}

func (m *MockVolumeRepo) GetByName(ctx context.Context, name string) (*domain.Volume, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Volume), args.Error(1)
}

func (m *MockVolumeRepo) List(ctx context.Context) ([]*domain.Volume, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Volume), args.Error(1)
}

func (m *MockVolumeRepo) ListByInstanceID(ctx context.Context, id uuid.UUID) ([]*domain.Volume, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Volume), args.Error(1)
}

func (m *MockVolumeRepo) Update(ctx context.Context, v *domain.Volume) error {
	return m.Called(ctx, v).Error(0)
}

func (m *MockVolumeRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

// MockQueueRepo
type MockQueueRepo struct{ mock.Mock }

func (m *MockQueueRepo) Create(ctx context.Context, q *domain.Queue) error {
	args := m.Called(ctx, q)
	return args.Error(0)
}
func (m *MockQueueRepo) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Queue, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Queue), args.Error(1)
}
func (m *MockQueueRepo) GetByName(ctx context.Context, name string, userID uuid.UUID) (*domain.Queue, error) {
	args := m.Called(ctx, name, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Queue), args.Error(1)
}
func (m *MockQueueRepo) List(ctx context.Context, userID uuid.UUID) ([]*domain.Queue, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Queue), args.Error(1)
}
func (m *MockQueueRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockQueueRepo) SendMessage(ctx context.Context, queueID uuid.UUID, body string) (*domain.Message, error) {
	args := m.Called(ctx, queueID, body)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Message), args.Error(1)
}
func (m *MockQueueRepo) ReceiveMessages(ctx context.Context, queueID uuid.UUID, maxMessages, visibilityTimeout int) ([]*domain.Message, error) {
	args := m.Called(ctx, queueID, maxMessages, visibilityTimeout)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Message), args.Error(1)
}
func (m *MockQueueRepo) DeleteMessage(ctx context.Context, queueID uuid.UUID, receiptHandle string) error {
	args := m.Called(ctx, queueID, receiptHandle)
	return args.Error(0)
}
func (m *MockQueueRepo) PurgeMessages(ctx context.Context, queueID uuid.UUID) (int64, error) {
	args := m.Called(ctx, queueID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockQueueRepo) GetQueueStats(ctx context.Context, queueID uuid.UUID) (int, int, error) {
	args := m.Called(ctx, queueID)
	return args.Int(0), args.Int(1), args.Error(2)
}

// MockQueueService
type MockQueueService struct{ mock.Mock }

func (m *MockQueueService) CreateQueue(ctx context.Context, name string, opts *ports.CreateQueueOptions) (*domain.Queue, error) {
	args := m.Called(ctx, name, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Queue), args.Error(1)
}
func (m *MockQueueService) GetQueue(ctx context.Context, id uuid.UUID) (*domain.Queue, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Queue), args.Error(1)
}
func (m *MockQueueService) ListQueues(ctx context.Context) ([]*domain.Queue, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Queue), args.Error(1)
}
func (m *MockQueueService) DeleteQueue(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockQueueService) SendMessage(ctx context.Context, queueID uuid.UUID, body string) (*domain.Message, error) {
	args := m.Called(ctx, queueID, body)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Message), args.Error(1)
}
func (m *MockQueueService) ReceiveMessages(ctx context.Context, queueID uuid.UUID, maxMessages int) ([]*domain.Message, error) {
	args := m.Called(ctx, queueID, maxMessages)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Message), args.Error(1)
}
func (m *MockQueueService) DeleteMessage(ctx context.Context, queueID uuid.UUID, receiptHandle string) error {
	args := m.Called(ctx, queueID, receiptHandle)
	return args.Error(0)
}
func (m *MockQueueService) PurgeQueue(ctx context.Context, queueID uuid.UUID) error {
	args := m.Called(ctx, queueID)
	return args.Error(0)
}

// MockNotifyRepo
type MockNotifyRepo struct{ mock.Mock }

func (m *MockNotifyRepo) CreateTopic(ctx context.Context, topic *domain.Topic) error {
	args := m.Called(ctx, topic)
	return args.Error(0)
}
func (m *MockNotifyRepo) GetTopicByID(ctx context.Context, id, userID uuid.UUID) (*domain.Topic, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Topic), args.Error(1)
}
func (m *MockNotifyRepo) GetTopicByName(ctx context.Context, name string, userID uuid.UUID) (*domain.Topic, error) {
	args := m.Called(ctx, name, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Topic), args.Error(1)
}
func (m *MockNotifyRepo) ListTopics(ctx context.Context, userID uuid.UUID) ([]*domain.Topic, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Topic), args.Error(1)
}
func (m *MockNotifyRepo) DeleteTopic(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockNotifyRepo) CreateSubscription(ctx context.Context, sub *domain.Subscription) error {
	return m.Called(ctx, sub).Error(0)
}
func (m *MockNotifyRepo) GetSubscriptionByID(ctx context.Context, id, userID uuid.UUID) (*domain.Subscription, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Subscription), args.Error(1)
}
func (m *MockNotifyRepo) ListSubscriptions(ctx context.Context, topicID uuid.UUID) ([]*domain.Subscription, error) {
	args := m.Called(ctx, topicID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Subscription), args.Error(1)
}
func (m *MockNotifyRepo) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockNotifyRepo) SaveMessage(ctx context.Context, msg *domain.NotifyMessage) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

// MockCronRepo
type MockCronRepo struct{ mock.Mock }

func (m *MockCronRepo) CreateJob(ctx context.Context, job *domain.CronJob) error {
	return m.Called(ctx, job).Error(0)
}
func (m *MockCronRepo) GetJobByID(ctx context.Context, id, userID uuid.UUID) (*domain.CronJob, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CronJob), args.Error(1)
}
func (m *MockCronRepo) ListJobs(ctx context.Context, userID uuid.UUID) ([]*domain.CronJob, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.CronJob), args.Error(1)
}
func (m *MockCronRepo) UpdateJob(ctx context.Context, job *domain.CronJob) error {
	return m.Called(ctx, job).Error(0)
}
func (m *MockCronRepo) DeleteJob(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockCronRepo) GetNextJobsToRun(ctx context.Context) ([]*domain.CronJob, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.CronJob), args.Error(1)
}
func (m *MockCronRepo) SaveJobRun(ctx context.Context, run *domain.CronJobRun) error {
	return m.Called(ctx, run).Error(0)
}

// MockGatewayRepo
type MockGatewayRepo struct{ mock.Mock }

func (m *MockGatewayRepo) CreateRoute(ctx context.Context, route *domain.GatewayRoute) error {
	return m.Called(ctx, route).Error(0)
}
func (m *MockGatewayRepo) GetRouteByID(ctx context.Context, id, userID uuid.UUID) (*domain.GatewayRoute, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.GatewayRoute), args.Error(1)
}
func (m *MockGatewayRepo) ListRoutes(ctx context.Context, userID uuid.UUID) ([]*domain.GatewayRoute, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.GatewayRoute), args.Error(1)
}
func (m *MockGatewayRepo) DeleteRoute(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockGatewayRepo) GetAllActiveRoutes(ctx context.Context) ([]*domain.GatewayRoute, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.GatewayRoute), args.Error(1)
}

// MockContainerRepo
type MockContainerRepo struct{ mock.Mock }

func (m *MockContainerRepo) CreateDeployment(ctx context.Context, d *domain.Deployment) error {
	args := m.Called(ctx, d)
	return args.Error(0)
}
func (m *MockContainerRepo) GetDeploymentByID(ctx context.Context, id, userID uuid.UUID) (*domain.Deployment, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Deployment), args.Error(1)
}
func (m *MockContainerRepo) ListDeployments(ctx context.Context, userID uuid.UUID) ([]*domain.Deployment, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Deployment), args.Error(1)
}
func (m *MockContainerRepo) UpdateDeployment(ctx context.Context, d *domain.Deployment) error {
	return m.Called(ctx, d).Error(0)
}
func (m *MockContainerRepo) DeleteDeployment(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockContainerRepo) AddContainer(ctx context.Context, deploymentID, instanceID uuid.UUID) error {
	return m.Called(ctx, deploymentID, instanceID).Error(0)
}
func (m *MockContainerRepo) RemoveContainer(ctx context.Context, deploymentID, instanceID uuid.UUID) error {
	return m.Called(ctx, deploymentID, instanceID).Error(0)
}
func (m *MockContainerRepo) GetContainers(ctx context.Context, deploymentID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, deploymentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uuid.UUID), args.Error(1)
}
func (m *MockContainerRepo) ListAllDeployments(ctx context.Context) ([]*domain.Deployment, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Deployment), args.Error(1)
}

// MockAuditRepo
type MockAuditRepo struct{ mock.Mock }

func (m *MockAuditRepo) Create(ctx context.Context, log *domain.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}
func (m *MockAuditRepo) ListByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.AuditLog, error) {
	args := m.Called(ctx, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.AuditLog), args.Error(1)
}

// MockAuditService
type MockAuditService struct{ mock.Mock }

func (m *MockAuditService) Log(ctx context.Context, userID uuid.UUID, action, resourceType, resourceID string, details map[string]interface{}) error {
	args := m.Called(ctx, userID, action, resourceType, resourceID, details)
	return args.Error(0)
}
func (m *MockAuditService) ListLogs(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.AuditLog, error) {
	args := m.Called(ctx, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.AuditLog), args.Error(1)
}

// MockInstanceRepo
type MockInstanceRepo struct{ mock.Mock }

func (m *MockInstanceRepo) Create(ctx context.Context, inst *domain.Instance) error {
	return m.Called(ctx, inst).Error(0)
}
func (m *MockInstanceRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Instance, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Instance), args.Error(1)
}
func (m *MockInstanceRepo) GetByName(ctx context.Context, name string) (*domain.Instance, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Instance), args.Error(1)
}
func (m *MockInstanceRepo) List(ctx context.Context) ([]*domain.Instance, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Instance), args.Error(1)
}
func (m *MockInstanceRepo) ListAll(ctx context.Context) ([]*domain.Instance, error) {
	return m.List(ctx)
}
func (m *MockInstanceRepo) ListBySubnet(ctx context.Context, subnetID uuid.UUID) ([]*domain.Instance, error) {
	args := m.Called(ctx, subnetID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Instance), args.Error(1)
}
func (m *MockInstanceRepo) Update(ctx context.Context, inst *domain.Instance) error {
	return m.Called(ctx, inst).Error(0)
}
func (m *MockInstanceRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

// MockClusterRepo
type MockClusterRepo struct{ mock.Mock }

func (m *MockClusterRepo) Create(ctx context.Context, cluster *domain.Cluster) error {
	return m.Called(ctx, cluster).Error(0)
}
func (m *MockClusterRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Cluster, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Cluster), args.Error(1)
}
func (m *MockClusterRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Cluster, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Cluster), args.Error(1)
}
func (m *MockClusterRepo) ListAll(ctx context.Context) ([]*domain.Cluster, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Cluster), args.Error(1)
}
func (m *MockClusterRepo) Update(ctx context.Context, cluster *domain.Cluster) error {
	return m.Called(ctx, cluster).Error(0)
}
func (m *MockClusterRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockClusterRepo) AddNode(ctx context.Context, node *domain.ClusterNode) error {
	return m.Called(ctx, node).Error(0)
}
func (m *MockClusterRepo) GetNodes(ctx context.Context, clusterID uuid.UUID) ([]*domain.ClusterNode, error) {
	args := m.Called(ctx, clusterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ClusterNode), args.Error(1)
}
func (m *MockClusterRepo) DeleteNode(ctx context.Context, nodeID uuid.UUID) error {
	return m.Called(ctx, nodeID).Error(0)
}
func (m *MockClusterRepo) UpdateNode(ctx context.Context, node *domain.ClusterNode) error {
	return m.Called(ctx, node).Error(0)
}

// MockClusterProvisioner
type MockClusterProvisioner struct{ mock.Mock }

func (m *MockClusterProvisioner) Provision(ctx context.Context, cluster *domain.Cluster) error {
	return m.Called(ctx, cluster).Error(0)
}
func (m *MockClusterProvisioner) Deprovision(ctx context.Context, cluster *domain.Cluster) error {
	return m.Called(ctx, cluster).Error(0)
}
func (m *MockClusterProvisioner) GetStatus(ctx context.Context, cluster *domain.Cluster) (domain.ClusterStatus, error) {
	args := m.Called(ctx, cluster)
	return args.Get(0).(domain.ClusterStatus), args.Error(1)
}
func (m *MockClusterProvisioner) Repair(ctx context.Context, cluster *domain.Cluster) error {
	return m.Called(ctx, cluster).Error(0)
}
func (m *MockClusterProvisioner) Scale(ctx context.Context, cluster *domain.Cluster) error {
	return m.Called(ctx, cluster).Error(0)
}
func (m *MockClusterProvisioner) GetKubeconfig(ctx context.Context, cluster *domain.Cluster, role string) (string, error) {
	args := m.Called(ctx, cluster, role)
	return args.String(0), args.Error(1)
}
func (m *MockClusterProvisioner) GetHealth(ctx context.Context, cluster *domain.Cluster) (*ports.ClusterHealth, error) {
	args := m.Called(ctx, cluster)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ports.ClusterHealth), args.Error(1)
}
func (m *MockClusterProvisioner) Upgrade(ctx context.Context, cluster *domain.Cluster, version string) error {
	return m.Called(ctx, cluster, version).Error(0)
}
func (m *MockClusterProvisioner) RotateSecrets(ctx context.Context, cluster *domain.Cluster) error {
	return m.Called(ctx, cluster).Error(0)
}
func (m *MockClusterProvisioner) CreateBackup(ctx context.Context, cluster *domain.Cluster) error {
	return m.Called(ctx, cluster).Error(0)
}
func (m *MockClusterProvisioner) Restore(ctx context.Context, cluster *domain.Cluster, backupPath string) error {
	return m.Called(ctx, cluster, backupPath).Error(0)
}
func (m *MockInstanceRepo) ListByVPC(ctx context.Context, vpcID uuid.UUID) ([]*domain.Instance, error) {
	args := m.Called(ctx, vpcID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Instance), args.Error(1)
}

// MockDockerClient
type MockComputeBackend struct{ mock.Mock }

func (m *MockComputeBackend) CreateInstance(ctx context.Context, opts ports.CreateInstanceOptions) (string, error) {
	args := m.Called(ctx, opts)
	return args.String(0), args.Error(1)
}
func (m *MockComputeBackend) StopInstance(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockComputeBackend) DeleteInstance(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockComputeBackend) GetInstanceLogs(ctx context.Context, id string) (io.ReadCloser, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}
func (m *MockComputeBackend) GetInstanceStats(ctx context.Context, id string) (io.ReadCloser, error) {
	args := m.Called(ctx, id)
	// Return the result directly if valid, checking nil first
	res := args.Get(0)
	if res == nil {
		return nil, args.Error(1)
	}
	return res.(io.ReadCloser), args.Error(1)
}
func (m *MockComputeBackend) GetInstancePort(ctx context.Context, id string, port string) (int, error) {
	args := m.Called(ctx, id, port)
	return args.Int(0), args.Error(1)
}
func (m *MockComputeBackend) CreateNetwork(ctx context.Context, name string) (string, error) {
	args := m.Called(ctx, name)
	return args.String(0), args.Error(1)
}
func (m *MockComputeBackend) DeleteNetwork(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockComputeBackend) AttachVolume(ctx context.Context, id string, volumePath string) error {
	return m.Called(ctx, id, volumePath).Error(0)
}
func (m *MockComputeBackend) DetachVolume(ctx context.Context, id string, volumePath string) error {
	args := m.Called(ctx, id, volumePath)
	return args.Error(0)
}
func (m *MockComputeBackend) RunTask(ctx context.Context, opts ports.RunTaskOptions) (string, error) {
	args := m.Called(ctx, opts)
	return args.String(0), args.Error(1)
}
func (m *MockComputeBackend) WaitTask(ctx context.Context, id string) (int64, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockComputeBackend) Exec(ctx context.Context, containerID string, cmd []string) (string, error) {
	args := m.Called(ctx, containerID, cmd)
	return args.String(0), args.Error(1)
}

func (m *MockComputeBackend) GetInstanceIP(ctx context.Context, id string) (string, error) {
	args := m.Called(ctx, id)
	return args.String(0), args.Error(1)
}
func (m *MockComputeBackend) GetConsoleURL(ctx context.Context, id string) (string, error) {
	return m.GetInstanceIP(ctx, id)
}
func (m *MockComputeBackend) CreateVolumeSnapshot(ctx context.Context, volumeID string, destinationPath string) error {
	return m.Called(ctx, volumeID, destinationPath).Error(0)
}

func (m *MockComputeBackend) RestoreVolumeSnapshot(ctx context.Context, volumeID string, sourcePath string) error {
	args := m.Called(ctx, volumeID, sourcePath)
	return args.Error(0)
}

func (m *MockComputeBackend) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockComputeBackend) Type() string {
	args := m.Called()
	return args.String(0)
}

// MockNetworkBackend
type MockNetworkBackend struct{ mock.Mock }

func (m *MockNetworkBackend) CreateBridge(ctx context.Context, name string, vxlanID int) error {
	args := m.Called(ctx, name, vxlanID)
	return args.Error(0)
}
func (m *MockNetworkBackend) DeleteBridge(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}
func (m *MockNetworkBackend) ListBridges(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}
func (m *MockNetworkBackend) AddPort(ctx context.Context, bridge, portName string) error {
	return m.Called(ctx, bridge, portName).Error(0)
}
func (m *MockNetworkBackend) DeletePort(ctx context.Context, bridge, portName string) error {
	return m.AddPort(ctx, bridge, portName)
}
func (m *MockNetworkBackend) CreateVXLANTunnel(ctx context.Context, bridge string, vni int, remoteIP string) error {
	args := m.Called(ctx, bridge, vni, remoteIP)
	return args.Error(0)
}
func (m *MockNetworkBackend) DeleteVXLANTunnel(ctx context.Context, bridge string, remoteIP string) error {
	args := m.Called(ctx, bridge, remoteIP)
	return args.Error(0)
}
func (m *MockNetworkBackend) AddFlowRule(ctx context.Context, bridge string, rule ports.FlowRule) error {
	args := m.Called(ctx, bridge, rule)
	return args.Error(0)
}
func (m *MockNetworkBackend) DeleteFlowRule(ctx context.Context, bridge string, match string) error {
	args := m.Called(ctx, bridge, match)
	return args.Error(0)
}
func (m *MockNetworkBackend) ListFlowRules(ctx context.Context, bridge string) ([]ports.FlowRule, error) {
	args := m.Called(ctx, bridge)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ports.FlowRule), args.Error(1)
}

func (m *MockNetworkBackend) CreateVethPair(ctx context.Context, hostEnd, containerEnd string) error {
	args := m.Called(ctx, hostEnd, containerEnd)
	return args.Error(0)
}

func (m *MockNetworkBackend) AttachVethToBridge(ctx context.Context, bridge, vethEnd string) error {
	args := m.Called(ctx, bridge, vethEnd)
	return args.Error(0)
}

func (m *MockNetworkBackend) DeleteVethPair(ctx context.Context, hostEnd string) error {
	args := m.Called(ctx, hostEnd)
	return args.Error(0)
}

func (m *MockNetworkBackend) SetVethIP(ctx context.Context, vethEnd, ip, cidr string) error {
	args := m.Called(ctx, vethEnd, ip, cidr)
	return args.Error(0)
}
func (m *MockNetworkBackend) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
func (m *MockNetworkBackend) Type() string {
	args := m.Called()
	return args.String(0)
}

// MockPasswordResetService
type MockPasswordResetService struct {
	mock.Mock
}

func (m *MockPasswordResetService) RequestReset(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *MockPasswordResetService) ResetPassword(ctx context.Context, token, newPassword string) error {
	args := m.Called(ctx, token, newPassword)
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
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockRoleRepo
type MockRoleRepo struct {
	mock.Mock
}

func (m *MockRoleRepo) CreateRole(ctx context.Context, role *domain.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}
func (m *MockRoleRepo) GetRoleByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Role), args.Error(1)
}
func (m *MockRoleRepo) GetRoleByName(ctx context.Context, name string) (*domain.Role, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Role), args.Error(1)
}
func (m *MockRoleRepo) ListRoles(ctx context.Context) ([]*domain.Role, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Role), args.Error(1)
}
func (m *MockRoleRepo) UpdateRole(ctx context.Context, role *domain.Role) error {
	return m.Called(ctx, role).Error(0)
}
func (m *MockRoleRepo) DeleteRole(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockRoleRepo) AddPermissionToRole(ctx context.Context, roleID uuid.UUID, permission domain.Permission) error {
	return m.Called(ctx, roleID, permission).Error(0)
}
func (m *MockRoleRepo) RemovePermissionFromRole(ctx context.Context, roleID uuid.UUID, permission domain.Permission) error {
	return m.Called(ctx, roleID, permission).Error(0)
}
func (m *MockRoleRepo) GetPermissionsForRole(ctx context.Context, roleID uuid.UUID) ([]domain.Permission, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Permission), args.Error(1)
}

// MockStackRepo
type MockStackRepo struct {
	mock.Mock
}

func (m *MockStackRepo) Create(ctx context.Context, s *domain.Stack) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}
func (m *MockStackRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Stack, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Stack), args.Error(1)
}
func (m *MockStackRepo) GetByName(ctx context.Context, userID uuid.UUID, name string) (*domain.Stack, error) {
	args := m.Called(ctx, userID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Stack), args.Error(1)
}
func (m *MockStackRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Stack, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Stack), args.Error(1)
}
func (m *MockStackRepo) Update(ctx context.Context, s *domain.Stack) error {
	return m.Called(ctx, s).Error(0)
}
func (m *MockStackRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockStackRepo) AddResource(ctx context.Context, r *domain.StackResource) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}
func (m *MockStackRepo) ListResources(ctx context.Context, stackID uuid.UUID) ([]domain.StackResource, error) {
	args := m.Called(ctx, stackID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.StackResource), args.Error(1)
}
func (m *MockStackRepo) DeleteResources(ctx context.Context, stackID uuid.UUID) error {
	args := m.Called(ctx, stackID)
	return args.Error(0)
}

// MockVpcService
type MockVpcService struct {
	mock.Mock
}

func (m *MockVpcService) CreateVPC(ctx context.Context, name, cidrBlock string) (*domain.VPC, error) {
	args := m.Called(ctx, name, cidrBlock)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.VPC), args.Error(1)
}
func (m *MockVpcService) GetVPC(ctx context.Context, idOrName string) (*domain.VPC, error) {
	args := m.Called(ctx, idOrName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.VPC), args.Error(1)
}
func (m *MockVpcService) ListVPCs(ctx context.Context) ([]*domain.VPC, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.VPC), args.Error(1)
}
func (m *MockVpcService) DeleteVPC(ctx context.Context, idOrName string) error {
	args := m.Called(ctx, idOrName)
	return args.Error(0)
}

// MockVolumeService
type MockVolumeService struct {
	mock.Mock
}

func (m *MockVolumeService) CreateVolume(ctx context.Context, name string, sizeGB int) (*domain.Volume, error) {
	args := m.Called(ctx, name, sizeGB)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Volume), args.Error(1)
}
func (m *MockVolumeService) ListVolumes(ctx context.Context) ([]*domain.Volume, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Volume), args.Error(1)
}
func (m *MockVolumeService) GetVolume(ctx context.Context, idOrName string) (*domain.Volume, error) {
	args := m.Called(ctx, idOrName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Volume), args.Error(1)
}
func (m *MockVolumeService) DeleteVolume(ctx context.Context, idOrName string) error {
	args := m.Called(ctx, idOrName)
	return args.Error(0)
}
func (m *MockVolumeService) ReleaseVolumesForInstance(ctx context.Context, instanceID uuid.UUID) error {
	args := m.Called(ctx, instanceID)
	return args.Error(0)
}

// MockSnapshotService
type MockSnapshotService struct {
	mock.Mock
}

func (m *MockSnapshotService) CreateSnapshot(ctx context.Context, volumeID uuid.UUID, description string) (*domain.Snapshot, error) {
	args := m.Called(ctx, volumeID, description)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Snapshot), args.Error(1)
}
func (m *MockSnapshotService) ListSnapshots(ctx context.Context) ([]*domain.Snapshot, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Snapshot), args.Error(1)
}
func (m *MockSnapshotService) GetSnapshot(ctx context.Context, id uuid.UUID) (*domain.Snapshot, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Snapshot), args.Error(1)
}
func (m *MockSnapshotService) DeleteSnapshot(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockSnapshotService) RestoreSnapshot(ctx context.Context, snapshotID uuid.UUID, newVolumeName string) (*domain.Volume, error) {
	args := m.Called(ctx, snapshotID, newVolumeName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Volume), args.Error(1)
}

// MockFunctionRepository
type MockFunctionRepository struct {
	mock.Mock
}

func (m *MockFunctionRepository) Create(ctx context.Context, function *domain.Function) error {
	args := m.Called(ctx, function)
	return args.Error(0)
}
func (m *MockFunctionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Function, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Function), args.Error(1)
}
func (m *MockFunctionRepository) GetByName(ctx context.Context, userID uuid.UUID, name string) (*domain.Function, error) {
	args := m.Called(ctx, userID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Function), args.Error(1)
}
func (m *MockFunctionRepository) List(ctx context.Context, userID uuid.UUID) ([]*domain.Function, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Function), args.Error(1)
}
func (m *MockFunctionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockFunctionRepository) CreateInvocation(ctx context.Context, i *domain.Invocation) error {
	args := m.Called(ctx, i)
	return args.Error(0)
}
func (m *MockFunctionRepository) GetInvocations(ctx context.Context, functionID uuid.UUID, limit int) ([]*domain.Invocation, error) {
	args := m.Called(ctx, functionID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Invocation), args.Error(1)
}

// MockContainerRepository
type MockContainerRepository struct {
	mock.Mock
}

func (m *MockContainerRepository) CreateDeployment(ctx context.Context, dep *domain.Deployment) error {
	args := m.Called(ctx, dep)
	return args.Error(0)
}
func (m *MockContainerRepository) GetDeploymentByID(ctx context.Context, id, userID uuid.UUID) (*domain.Deployment, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Deployment), args.Error(1)
}
func (m *MockContainerRepository) ListDeployments(ctx context.Context, userID uuid.UUID) ([]*domain.Deployment, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Deployment), args.Error(1)
}
func (m *MockContainerRepository) UpdateDeployment(ctx context.Context, dep *domain.Deployment) error {
	return m.Called(ctx, dep).Error(0)
}
func (m *MockContainerRepository) DeleteDeployment(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Replication management
func (m *MockContainerRepository) AddContainer(ctx context.Context, deploymentID, instanceID uuid.UUID) error {
	return m.Called(ctx, deploymentID, instanceID).Error(0)
}
func (m *MockContainerRepository) RemoveContainer(ctx context.Context, deploymentID, instanceID uuid.UUID) error {
	return m.AddContainer(ctx, deploymentID, instanceID)
}
func (m *MockContainerRepository) GetContainers(ctx context.Context, deploymentID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, deploymentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

// Worker
func (m *MockContainerRepository) ListAllDeployments(ctx context.Context) ([]*domain.Deployment, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Deployment), args.Error(1)
}

// MockCronRepository
type MockCronRepository struct {
	mock.Mock
}

func (m *MockCronRepository) CreateJob(ctx context.Context, job *domain.CronJob) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}
func (m *MockCronRepository) GetJobByID(ctx context.Context, id, userID uuid.UUID) (*domain.CronJob, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CronJob), args.Error(1)
}
func (m *MockCronRepository) ListJobs(ctx context.Context, userID uuid.UUID) ([]*domain.CronJob, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.CronJob), args.Error(1)
}
func (m *MockCronRepository) UpdateJob(ctx context.Context, job *domain.CronJob) error {
	return m.Called(ctx, job).Error(0)
}
func (m *MockCronRepository) DeleteJob(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockCronRepository) GetNextJobsToRun(ctx context.Context) ([]*domain.CronJob, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.CronJob), args.Error(1)
}
func (m *MockCronRepository) SaveJobRun(ctx context.Context, run *domain.CronJobRun) error {
	args := m.Called(ctx, run)
	return args.Error(0)
}

// MockIdentityRepo
type MockIdentityRepo struct{ mock.Mock }

func (m *MockIdentityRepo) CreateAPIKey(ctx context.Context, apiKey *domain.APIKey) error {
	args := m.Called(ctx, apiKey)
	return args.Error(0)
}
func (m *MockIdentityRepo) GetAPIKeyByKey(ctx context.Context, key string) (*domain.APIKey, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.APIKey), args.Error(1)
}
func (m *MockIdentityRepo) GetAPIKeyByID(ctx context.Context, id uuid.UUID) (*domain.APIKey, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.APIKey), args.Error(1)
}
func (m *MockIdentityRepo) ListAPIKeysByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.APIKey, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.APIKey), args.Error(1)
}
func (m *MockIdentityRepo) DeleteAPIKey(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockLBRepo
type MockLBRepo struct{ mock.Mock }

func (m *MockLBRepo) Create(ctx context.Context, lb *domain.LoadBalancer) error {
	args := m.Called(ctx, lb)
	return args.Error(0)
}
func (m *MockLBRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.LoadBalancer, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.LoadBalancer), args.Error(1)
}
func (m *MockLBRepo) GetByIdempotencyKey(ctx context.Context, key string) (*domain.LoadBalancer, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.LoadBalancer), args.Error(1)
}
func (m *MockLBRepo) List(ctx context.Context) ([]*domain.LoadBalancer, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.LoadBalancer), args.Error(1)
}
func (m *MockLBRepo) ListAll(ctx context.Context) ([]*domain.LoadBalancer, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.LoadBalancer), args.Error(1)
}
func (m *MockLBRepo) Update(ctx context.Context, lb *domain.LoadBalancer) error {
	return m.Called(ctx, lb).Error(0)
}
func (m *MockLBRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockLBRepo) AddTarget(ctx context.Context, target *domain.LBTarget) error {
	args := m.Called(ctx, target)
	return args.Error(0)
}
func (m *MockLBRepo) RemoveTarget(ctx context.Context, lbID, instanceID uuid.UUID) error {
	args := m.Called(ctx, lbID, instanceID)
	return args.Error(0)
}
func (m *MockLBRepo) ListTargets(ctx context.Context, lbID uuid.UUID) ([]*domain.LBTarget, error) {
	args := m.Called(ctx, lbID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.LBTarget), args.Error(1)
}
func (m *MockLBRepo) UpdateTargetHealth(ctx context.Context, lbID, instanceID uuid.UUID, health string) error {
	args := m.Called(ctx, lbID, instanceID, health)
	return args.Error(0)
}
func (m *MockLBRepo) GetTargetsForInstance(ctx context.Context, instanceID uuid.UUID) ([]*domain.LBTarget, error) {
	args := m.Called(ctx, instanceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.LBTarget), args.Error(1)
}

// MockSecurityGroupRepo
type MockSecurityGroupRepo struct{ mock.Mock }

func (m *MockSecurityGroupRepo) Create(ctx context.Context, sg *domain.SecurityGroup) error {
	return m.Called(ctx, sg).Error(0)
}
func (m *MockSecurityGroupRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.SecurityGroup, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SecurityGroup), args.Error(1)
}
func (m *MockSecurityGroupRepo) GetByName(ctx context.Context, vpcID uuid.UUID, name string) (*domain.SecurityGroup, error) {
	args := m.Called(ctx, vpcID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SecurityGroup), args.Error(1)
}
func (m *MockSecurityGroupRepo) ListByVPC(ctx context.Context, vpcID uuid.UUID) ([]*domain.SecurityGroup, error) {
	args := m.Called(ctx, vpcID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.SecurityGroup), args.Error(1)
}
func (m *MockSecurityGroupRepo) AddRule(ctx context.Context, rule *domain.SecurityRule) error {
	return m.Called(ctx, rule).Error(0)
}
func (m *MockSecurityGroupRepo) GetRuleByID(ctx context.Context, ruleID uuid.UUID) (*domain.SecurityRule, error) {
	args := m.Called(ctx, ruleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SecurityRule), args.Error(1)
}
func (m *MockSecurityGroupRepo) DeleteRule(ctx context.Context, ruleID uuid.UUID) error {
	return m.Called(ctx, ruleID).Error(0)
}
func (m *MockSecurityGroupRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockSecurityGroupRepo) AddInstanceToGroup(ctx context.Context, instanceID, groupID uuid.UUID) error {
	return m.Called(ctx, instanceID, groupID).Error(0)
}
func (m *MockSecurityGroupRepo) RemoveInstanceFromGroup(ctx context.Context, instanceID, groupID uuid.UUID) error {
	return m.Called(ctx, instanceID, groupID).Error(0)
}
func (m *MockSecurityGroupRepo) ListInstanceGroups(ctx context.Context, instanceID uuid.UUID) ([]*domain.SecurityGroup, error) {
	return nil, nil
}

// MockQueueRepository
type MockQueueRepository struct{ mock.Mock }

func (m *MockQueueRepository) Create(ctx context.Context, queue *domain.Queue) error {
	args := m.Called(ctx, queue)
	return args.Error(0)
}
func (m *MockQueueRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Queue, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Queue), args.Error(1)
}
func (m *MockQueueRepository) GetByName(ctx context.Context, name string, userID uuid.UUID) (*domain.Queue, error) {
	args := m.Called(ctx, name, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Queue), args.Error(1)
}
func (m *MockQueueRepository) List(ctx context.Context, userID uuid.UUID) ([]*domain.Queue, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*domain.Queue), args.Error(1)
}
func (m *MockQueueRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockQueueRepository) SendMessage(ctx context.Context, queueID uuid.UUID, body string) (*domain.Message, error) {
	args := m.Called(ctx, queueID, body)
	return args.Get(0).(*domain.Message), args.Error(1)
}
func (m *MockQueueRepository) ReceiveMessages(ctx context.Context, queueID uuid.UUID, maxMessages, visibilityTimeout int) ([]*domain.Message, error) {
	args := m.Called(ctx, queueID, maxMessages, visibilityTimeout)
	return args.Get(0).([]*domain.Message), args.Error(1)
}
func (m *MockQueueRepository) DeleteMessage(ctx context.Context, queueID uuid.UUID, receiptHandle string) error {
	args := m.Called(ctx, queueID, receiptHandle)
	return args.Error(0)
}
func (m *MockQueueRepository) PurgeMessages(ctx context.Context, queueID uuid.UUID) (int64, error) {
	args := m.Called(ctx, queueID)
	return int64(args.Int(0)), args.Error(1)
}

// MockStorageBackend
type MockStorageBackend struct {
	mock.Mock
}

func (m *MockStorageBackend) CreateVolume(ctx context.Context, name string, sizeGB int) (string, error) {
	args := m.Called(ctx, name, sizeGB)
	return args.String(0), args.Error(1)
}
func (m *MockStorageBackend) DeleteVolume(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}
func (m *MockStorageBackend) AttachVolume(ctx context.Context, volumeName, instanceID string) error {
	args := m.Called(ctx, volumeName, instanceID)
	return args.Error(0)
}
func (m *MockStorageBackend) DetachVolume(ctx context.Context, volumeName, instanceID string) error {
	return m.Called(ctx, volumeName, instanceID).Error(0)
}
func (m *MockStorageBackend) CreateSnapshot(ctx context.Context, volumeName, snapshotName string) error {
	args := m.Called(ctx, volumeName, snapshotName)
	return args.Error(0)
}
func (m *MockStorageBackend) RestoreSnapshot(ctx context.Context, volumeName, snapshotName string) error {
	return m.Called(ctx, volumeName, snapshotName).Error(0)
}
func (m *MockStorageBackend) DeleteSnapshot(ctx context.Context, snapshotName string) error {
	args := m.Called(ctx, snapshotName)
	return args.Error(0)
}
func (m *MockStorageBackend) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
func (m *MockStorageBackend) Type() string {
	args := m.Called()
	return args.String(0)
}

// MockSecretService
type MockSecretService struct {
	mock.Mock
}

func (m *MockSecretService) CreateSecret(ctx context.Context, name, value, description string) (*domain.Secret, error) {
	args := m.Called(ctx, name, value, description)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Secret), args.Error(1)
}
func (m *MockSecretService) GetSecret(ctx context.Context, id uuid.UUID) (*domain.Secret, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Secret), args.Error(1)
}
func (m *MockSecretService) GetSecretByName(ctx context.Context, name string) (*domain.Secret, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Secret), args.Error(1)
}
func (m *MockSecretService) ListSecrets(ctx context.Context) ([]*domain.Secret, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Secret), args.Error(1)
}
func (m *MockSecretService) DeleteSecret(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockSecretService) Encrypt(ctx context.Context, userID uuid.UUID, plainText string) (string, error) {
	args := m.Called(ctx, userID, plainText)
	return args.String(0), args.Error(1)
}
func (m *MockSecretService) Decrypt(ctx context.Context, userID uuid.UUID, cipherText string) (string, error) {
	args := m.Called(ctx, userID, cipherText)
	return args.String(0), args.Error(1)
}

// MockTaskQueue
type MockTaskQueue struct {
	mock.Mock
}

func (m *MockTaskQueue) Enqueue(ctx context.Context, queue string, payload interface{}) error {
	args := m.Called(ctx, queue, payload)
	return args.Error(0)
}

func (m *MockTaskQueue) Dequeue(ctx context.Context, queue string) (string, error) {
	args := m.Called(ctx, queue)
	return args.String(0), args.Error(1)
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.LifecycleRule), args.Error(1)
}
func (m *MockLifecycleRepository) List(ctx context.Context, bucketName string) ([]*domain.LifecycleRule, error) {
	args := m.Called(ctx, bucketName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.LifecycleRule), args.Error(1)
}
func (m *MockLifecycleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockLifecycleRepository) GetEnabledRules(ctx context.Context) ([]*domain.LifecycleRule, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.LifecycleRule), args.Error(1)
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ports.EncryptionKey), args.Error(1)
}

// MockEncryptionService
type MockEncryptionService struct {
	mock.Mock
}

func (m *MockEncryptionService) Encrypt(ctx context.Context, bucket string, data []byte) ([]byte, error) {
	args := m.Called(ctx, bucket, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEncryptionService) Decrypt(ctx context.Context, bucket string, encryptedData []byte) ([]byte, error) {
	args := m.Called(ctx, bucket, encryptedData)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
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
	return args.Get(0).(string), args.Error(1)
}

// MockTenantRepo
type MockTenantRepo struct {
	mock.Mock
}

func (m *MockTenantRepo) Create(ctx context.Context, tenant *domain.Tenant) error {
	return m.Called(ctx, tenant).Error(0)
}
func (m *MockTenantRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Tenant), args.Error(1)
}
func (m *MockTenantRepo) GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Tenant), args.Error(1)
}
func (m *MockTenantRepo) Update(ctx context.Context, tenant *domain.Tenant) error {
	return m.Called(ctx, tenant).Error(0)
}
func (m *MockTenantRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockTenantRepo) AddMember(ctx context.Context, tenantID, userID uuid.UUID, role string) error {
	return m.Called(ctx, tenantID, userID, role).Error(0)
}
func (m *MockTenantRepo) RemoveMember(ctx context.Context, tenantID, userID uuid.UUID) error {
	return m.Called(ctx, tenantID, userID).Error(0)
}
func (m *MockTenantRepo) ListMembers(ctx context.Context, tenantID uuid.UUID) ([]domain.TenantMember, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TenantMember), args.Error(1)
}
func (m *MockTenantRepo) GetMembership(ctx context.Context, tenantID, userID uuid.UUID) (*domain.TenantMember, error) {
	args := m.Called(ctx, tenantID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TenantMember), args.Error(1)
}
func (m *MockTenantRepo) ListUserTenants(ctx context.Context, userID uuid.UUID) ([]domain.Tenant, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Tenant), args.Error(1)
}
func (m *MockTenantRepo) GetQuota(ctx context.Context, tenantID uuid.UUID) (*domain.TenantQuota, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TenantQuota), args.Error(1)
}
func (m *MockTenantRepo) UpdateQuota(ctx context.Context, quota *domain.TenantQuota) error {
	return m.Called(ctx, quota).Error(0)
}

// MockTenantService
type MockTenantService struct {
	mock.Mock
}

func (m *MockTenantService) CreateTenant(ctx context.Context, name, slug string, ownerID uuid.UUID) (*domain.Tenant, error) {
	args := m.Called(ctx, name, slug, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Tenant), args.Error(1)
}
func (m *MockTenantService) GetTenant(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Tenant), args.Error(1)
}
func (m *MockTenantService) InviteMember(ctx context.Context, tenantID uuid.UUID, email, role string) error {
	return m.Called(ctx, tenantID, email, role).Error(0)
}
func (m *MockTenantService) RemoveMember(ctx context.Context, tenantID, userID uuid.UUID) error {
	return m.Called(ctx, tenantID, userID).Error(0)
}
func (m *MockTenantService) SwitchTenant(ctx context.Context, userID, tenantID uuid.UUID) error {
	return m.Called(ctx, userID, tenantID).Error(0)
}
func (m *MockTenantService) CheckQuota(ctx context.Context, tenantID uuid.UUID, resource string, requested int) error {
	return m.Called(ctx, tenantID, resource, requested).Error(0)
}
func (m *MockTenantService) GetMembership(ctx context.Context, tenantID, userID uuid.UUID) (*domain.TenantMember, error) {
	args := m.Called(ctx, tenantID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TenantMember), args.Error(1)
}
