package services_test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockDNSService struct{ mock.Mock }

func (m *MockDNSService) CreateZone(ctx context.Context, vpcID uuid.UUID, name, description string) (*domain.DNSZone, error) {
	return nil, nil
}
func (m *MockDNSService) ListZones(ctx context.Context) ([]*domain.DNSZone, error) {
	return nil, nil
}
func (m *MockDNSService) GetZone(ctx context.Context, idOrName string) (*domain.DNSZone, error) {
	return nil, nil
}
func (m *MockDNSService) GetZoneByVPC(ctx context.Context, vpcID uuid.UUID) (*domain.DNSZone, error) {
	return nil, nil
}
func (m *MockDNSService) DeleteZone(ctx context.Context, idOrName string) error {
	return nil
}
func (m *MockDNSService) CreateRecord(ctx context.Context, zoneID uuid.UUID, name string, recordType domain.RecordType, content string, ttl int, priority *int) (*domain.DNSRecord, error) {
	return nil, nil
}
func (m *MockDNSService) ListRecords(ctx context.Context, zoneID uuid.UUID) ([]*domain.DNSRecord, error) {
	return nil, nil
}
func (m *MockDNSService) GetRecord(ctx context.Context, id uuid.UUID) (*domain.DNSRecord, error) {
	return nil, nil
}
func (m *MockDNSService) UpdateRecord(ctx context.Context, id uuid.UUID, content string, ttl int, priority *int) (*domain.DNSRecord, error) {
	return nil, nil
}
func (m *MockDNSService) DeleteRecord(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *MockDNSService) RegisterInstance(ctx context.Context, instance *domain.Instance, ip string) error {
	args := m.Called(ctx, instance, ip)
	return args.Error(0)
}
func (m *MockDNSService) UnregisterInstance(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockSSHKeyService struct{ mock.Mock }

func (m *MockSSHKeyService) CreateKey(ctx context.Context, name, publicKey string) (*domain.SSHKey, error) {
	return nil, nil
}
func (m *MockSSHKeyService) GetKey(ctx context.Context, id uuid.UUID) (*domain.SSHKey, error) {
	args := m.Called(ctx, id)
	r0, _ := args.Get(0).(*domain.SSHKey)
	return r0, args.Error(1)
}
func (m *MockSSHKeyService) ListKeys(ctx context.Context) ([]*domain.SSHKey, error) {
	return nil, nil
}
func (m *MockSSHKeyService) DeleteKey(ctx context.Context, id uuid.UUID) error {
	return nil
}

func TestInstanceService_LaunchInstance_Unit(t *testing.T) {
	repo := new(MockInstanceRepo)
	vpcRepo := new(MockVpcRepo)
	subnetRepo := new(MockSubnetRepo)
	volRepo := new(MockVolumeRepo)
	typeRepo := new(MockInstanceTypeRepo)
	compute := new(MockComputeBackend)
	network := new(MockNetworkBackend)
	eventSvc := new(MockEventService)
	auditSvc := new(MockAuditService)
	dnsSvc := new(MockDNSService)
	taskQueue := new(MockTaskQueue)
	tenantSvc := new(MockTenantService)
	sshKeySvc := new(MockSSHKeyService)

	svc := services.NewInstanceService(services.InstanceServiceParams{
		Repo:             repo,
		VpcRepo:          vpcRepo,
		SubnetRepo:       subnetRepo,
		VolumeRepo:       volRepo,
		InstanceTypeRepo: typeRepo,
		Compute:          compute,
		Network:          network,
		EventSvc:         eventSvc,
		AuditSvc:         auditSvc,
		DNSSvc:           dnsSvc,
		TaskQueue:        taskQueue,
		TenantSvc:        tenantSvc,
		SSHKeySvc:        sshKeySvc,
		DockerNetwork:    "bridge",
		Logger:           slog.Default(),
	})

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	t.Run("Success_Basic", func(t *testing.T) {
		params := ports.LaunchParams{
			Name:         "test-instance",
			Image:        "ubuntu",
			InstanceType: "t2.micro",
		}

		typeRepo.On("GetByID", mock.Anything, "t2.micro").Return(&domain.InstanceType{
			ID: "t2.micro", VCPUs: 1, MemoryMB: 1024,
		}, nil).Once()
		tenantSvc.On("CheckQuota", mock.Anything, tenantID, "instances", 1).Return(nil).Once()
		tenantSvc.On("CheckQuota", mock.Anything, tenantID, "vcpus", 1).Return(nil).Once()
		tenantSvc.On("CheckQuota", mock.Anything, tenantID, "memory", 1).Return(nil).Once()
		tenantSvc.On("IncrementUsage", mock.Anything, tenantID, "vcpus", 1).Return(nil).Once()
		tenantSvc.On("IncrementUsage", mock.Anything, tenantID, "memory", 1).Return(nil).Once()

		repo.On("Create", mock.Anything, mock.MatchedBy(func(i *domain.Instance) bool {
			return i.Name == params.Name && i.UserID == userID
		})).Return(nil).Once()

		taskQueue.On("Enqueue", mock.Anything, "provision_queue", mock.Anything).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "instance.launch", "instance", mock.Anything, mock.Anything).Return(nil).Once()

		inst, err := svc.LaunchInstance(ctx, params)

		require.NoError(t, err)
		assert.NotNil(t, inst)
		assert.Equal(t, params.Name, inst.Name)
		repo.AssertExpectations(t)
		tenantSvc.AssertExpectations(t)
	})

	t.Run("QuotaExceeded", func(t *testing.T) {
		params := ports.LaunchParams{
			Name:         "no-quota",
			Image:        "ubuntu",
			InstanceType: "t2.large",
		}

		typeRepo.On("GetByID", mock.Anything, "t2.large").Return(&domain.InstanceType{
			ID: "t2.large", VCPUs: 4, MemoryMB: 4096,
		}, nil).Once()
		tenantSvc.On("CheckQuota", mock.Anything, tenantID, "instances", 1).Return(fmt.Errorf("quota exceeded")).Once()

		inst, err := svc.LaunchInstance(ctx, params)

		require.Error(t, err)
		assert.Nil(t, inst)
	})
}

func TestInstanceService_Lifecycle_Unit(t *testing.T) {
	repo := new(MockInstanceRepo)
	vpcRepo := new(MockVpcRepo)
	subnetRepo := new(MockSubnetRepo)
	volRepo := new(MockVolumeRepo)
	typeRepo := new(MockInstanceTypeRepo)
	compute := new(MockComputeBackend)
	network := new(MockNetworkBackend)
	eventSvc := new(MockEventService)
	auditSvc := new(MockAuditService)
	dnsSvc := new(MockDNSService)
	taskQueue := new(MockTaskQueue)
	tenantSvc := new(MockTenantService)
	sshKeySvc := new(MockSSHKeyService)

	svc := services.NewInstanceService(services.InstanceServiceParams{
		Repo:             repo,
		VpcRepo:          vpcRepo,
		SubnetRepo:       subnetRepo,
		VolumeRepo:       volRepo,
		InstanceTypeRepo: typeRepo,
		Compute:          compute,
		Network:          network,
		EventSvc:         eventSvc,
		AuditSvc:         auditSvc,
		DNSSvc:           dnsSvc,
		TaskQueue:        taskQueue,
		TenantSvc:        tenantSvc,
		SSHKeySvc:        sshKeySvc,
		DockerNetwork:    "bridge",
		Logger:           slog.Default(),
	})

	ctx := context.Background()
	instanceID := uuid.New()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	t.Run("StartInstance", func(t *testing.T) {
		inst := &domain.Instance{ID: instanceID, UserID: userID, Status: domain.StatusStopped, ContainerID: "cid-1"}
		repo.On("GetByID", mock.Anything, instanceID).Return(inst, nil).Once()
		compute.On("StartInstance", mock.Anything, "cid-1").Return(nil).Once()
		compute.On("Type").Return("docker").Maybe()
		repo.On("Update", mock.Anything, mock.MatchedBy(func(i *domain.Instance) bool {
			return i.Status == domain.StatusRunning
		})).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "instance.start", "instance", instanceID.String(), mock.Anything).Return(nil).Once()

		err := svc.StartInstance(ctx, instanceID.String())
		require.NoError(t, err)
	})

	t.Run("StopInstance", func(t *testing.T) {
		inst := &domain.Instance{ID: instanceID, UserID: userID, Status: domain.StatusRunning, ContainerID: "cid-1"}
		repo.On("GetByID", mock.Anything, instanceID).Return(inst, nil).Once()
		compute.On("StopInstance", mock.Anything, "cid-1").Return(nil).Once()
		compute.On("Type").Return("docker").Maybe()
		repo.On("Update", mock.Anything, mock.MatchedBy(func(i *domain.Instance) bool {
			return i.Status == domain.StatusStopped
		})).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "instance.stop", "instance", instanceID.String(), mock.Anything).Return(nil).Once()

		err := svc.StopInstance(ctx, instanceID.String())
		require.NoError(t, err)
	})

	t.Run("TerminateInstance", func(t *testing.T) {
		inst := &domain.Instance{
			ID: instanceID, UserID: userID, TenantID: tenantID, Status: domain.StatusRunning, ContainerID: "cid-1",
			InstanceType: "t2.micro",
		}
		typeRepo.On("GetByID", mock.Anything, "t2.micro").Return(&domain.InstanceType{VCPUs: 1, MemoryMB: 1024}, nil).Maybe()
		repo.On("GetByID", mock.Anything, instanceID).Return(inst, nil).Once()
		compute.On("DeleteInstance", mock.Anything, "cid-1").Return(nil).Once()
		compute.On("Type").Return("docker").Maybe()
		volRepo.On("ListByInstanceID", mock.Anything, instanceID).Return([]*domain.Volume{}, nil).Once()
		repo.On("Delete", mock.Anything, instanceID).Return(nil).Once()

		tenantSvc.On("DecrementUsage", mock.Anything, tenantID, "instances", 1).Return(nil).Once()
		tenantSvc.On("DecrementUsage", mock.Anything, tenantID, "vcpus", 1).Return(nil).Once()
		tenantSvc.On("DecrementUsage", mock.Anything, tenantID, "memory", 1).Return(nil).Once()

		dnsSvc.On("UnregisterInstance", mock.Anything, instanceID).Return(nil).Maybe()
		eventSvc.On("RecordEvent", mock.Anything, "INSTANCE_TERMINATE", instanceID.String(), "INSTANCE", mock.Anything).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "instance.terminate", "instance", instanceID.String(), mock.Anything).Return(nil).Once()

		err := svc.TerminateInstance(ctx, instanceID.String())
		require.NoError(t, err)
	})
}
