package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/poyraz/cloud/internal/core/domain"
	"github.com/poyraz/cloud/internal/core/services"
)

// Mocks
type MockAutoScalingRepo struct{ mock.Mock }

func (m *MockAutoScalingRepo) CreateGroup(ctx context.Context, group *domain.ScalingGroup) error {
	args := m.Called(ctx, group)
	return args.Error(0)
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
	return args.Get(0).([]*domain.ScalingGroup), args.Error(1)
}
func (m *MockAutoScalingRepo) CountGroupsByVPC(ctx context.Context, vpcID uuid.UUID) (int, error) {
	args := m.Called(ctx, vpcID)
	return args.Int(0), args.Error(1)
}
func (m *MockAutoScalingRepo) UpdateGroup(ctx context.Context, group *domain.ScalingGroup) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}
func (m *MockAutoScalingRepo) DeleteGroup(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockAutoScalingRepo) CreatePolicy(ctx context.Context, policy *domain.ScalingPolicy) error {
	args := m.Called(ctx, policy)
	return args.Error(0)
}
func (m *MockAutoScalingRepo) GetPoliciesForGroup(ctx context.Context, groupID uuid.UUID) ([]*domain.ScalingPolicy, error) {
	args := m.Called(ctx, groupID)
	return args.Get(0).([]*domain.ScalingPolicy), args.Error(1)
}
func (m *MockAutoScalingRepo) GetAllPolicies(ctx context.Context, groupIDs []uuid.UUID) (map[uuid.UUID][]*domain.ScalingPolicy, error) {
	args := m.Called(ctx, groupIDs)
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
	args := m.Called(ctx, groupID, instanceID)
	return args.Error(0)
}
func (m *MockAutoScalingRepo) RemoveInstanceFromGroup(ctx context.Context, groupID, instanceID uuid.UUID) error {
	args := m.Called(ctx, groupID, instanceID)
	return args.Error(0)
}
func (m *MockAutoScalingRepo) GetInstancesInGroup(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, groupID)
	return args.Get(0).([]uuid.UUID), args.Error(1)
}
func (m *MockAutoScalingRepo) GetAllScalingGroupInstances(ctx context.Context, groupIDs []uuid.UUID) (map[uuid.UUID][]uuid.UUID, error) {
	args := m.Called(ctx, groupIDs)
	return args.Get(0).(map[uuid.UUID][]uuid.UUID), args.Error(1)
}
func (m *MockAutoScalingRepo) GetAverageCPU(ctx context.Context, instanceIDs []uuid.UUID, since time.Time) (float64, error) {
	args := m.Called(ctx, instanceIDs, since)
	return args.Get(0).(float64), args.Error(1)
}

type MockVpcRepo struct{ mock.Mock }

func (m *MockVpcRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.VPC, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.VPC), args.Error(1)
}
func (m *MockVpcRepo) GetByName(ctx context.Context, name string) (*domain.VPC, error) {
	return nil, nil
}
func (m *MockVpcRepo) Create(ctx context.Context, vpc *domain.VPC) error { return nil }
func (m *MockVpcRepo) List(ctx context.Context) ([]*domain.VPC, error)   { return nil, nil }
func (m *MockVpcRepo) Delete(ctx context.Context, id uuid.UUID) error    { return nil }

type MockInstanceService struct{ mock.Mock }

func (m *MockInstanceService) LaunchInstance(ctx context.Context, name, image, ports string, vpcID *uuid.UUID, volumes []domain.VolumeAttachment) (*domain.Instance, error) {
	args := m.Called(ctx, name, image, ports, vpcID, volumes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Instance), args.Error(1)
}
func (m *MockInstanceService) StopInstance(ctx context.Context, idOrName string) error { return nil }
func (m *MockInstanceService) ListInstances(ctx context.Context) ([]*domain.Instance, error) {
	return nil, nil
}
func (m *MockInstanceService) GetInstance(ctx context.Context, idOrName string) (*domain.Instance, error) {
	return nil, nil
}
func (m *MockInstanceService) GetInstanceLogs(ctx context.Context, idOrName string) (string, error) {
	return "", nil
}
func (m *MockInstanceService) GetInstanceStats(ctx context.Context, idOrName string) (*domain.InstanceStats, error) {
	return nil, nil
}
func (m *MockInstanceService) TerminateInstance(ctx context.Context, idOrName string) error {
	args := m.Called(ctx, idOrName)
	return args.Error(0)
}

func TestCreateGroup_SecurityLimits(t *testing.T) {
	mockRepo := new(MockAutoScalingRepo)
	mockVpcRepo := new(MockVpcRepo)
	mockInstSvc := new(MockInstanceService)
	svc := services.NewAutoScalingService(mockRepo, mockVpcRepo, mockInstSvc)
	ctx := context.Background()
	vpcID := uuid.New()

	mockVpcRepo.On("GetByID", ctx, vpcID).Return(&domain.VPC{ID: vpcID}, nil)

	t.Run("ExceedsMaxInstances", func(t *testing.T) {
		_, err := svc.CreateGroup(ctx, "test", vpcID, "img", "80:80", 1, 1000, 1, nil, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max_instances cannot exceed")
	})

	t.Run("ExceedsVPCLimit", func(t *testing.T) {
		mockRepo.On("CountGroupsByVPC", ctx, vpcID).Return(10, nil)
		_, err := svc.CreateGroup(ctx, "test", vpcID, "img", "80:80", 1, 5, 1, nil, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "VPC already has")
	})
}
