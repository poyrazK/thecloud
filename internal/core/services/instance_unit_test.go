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

type MockSSHKeyService struct{ mock.Mock } 
func (m *MockSSHKeyService) CreateKey(ctx context.Context, name, publicKey string) (*domain.SSHKey, error) { return nil, nil } 
func (m *MockSSHKeyService) GetKey(ctx context.Context, id uuid.UUID) (*domain.SSHKey, error) { 
	args := m.Called(ctx, id) 
	r0, _ := args.Get(0).(*domain.SSHKey) 
	return r0, args.Error(1) 
} 
func (m *MockSSHKeyService) ListKeys(ctx context.Context) ([]*domain.SSHKey, error) { return nil, nil } 
func (m *MockSSHKeyService) DeleteKey(ctx context.Context, id uuid.UUID) error { return nil }
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

func TestInstanceService_LaunchInstance_Unit(t *testing.T) {
	repo := new(MockInstanceRepo)
	vpcRepo := new(MockVpcRepo)
	subnetRepo := new(MockSubnetRepo)
	volRepo := new(MockVolumeRepo)
	typeRepo := new(MockInstanceTypeRepo)
	compute := new(MockComputeBackend)
	network := new(MockNetworkBackend)
	rbacSvc := new(MockRBACService)
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
		RBAC:             rbacSvc,
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
			Name:         "test-inst",
			Image:        "alpine",
			InstanceType: "t2.micro",
		}

		rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
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
		rbacSvc.AssertExpectations(t)
	})

	t.Run("QuotaExceeded", func(t *testing.T) {
		params := ports.LaunchParams{
			Name:         "no-quota",
			Image:        "alpine",
			InstanceType: "t2.micro",
		}

		rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		typeRepo.On("GetByID", mock.Anything, "t2.micro").Return(&domain.InstanceType{
			ID: "t2.micro", VCPUs: 1, MemoryMB: 1024,
		}, nil).Once()
		tenantSvc.On("CheckQuota", mock.Anything, tenantID, "instances", 1).Return(fmt.Errorf("quota exceeded")).Once()

		_, err := svc.LaunchInstance(ctx, params)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "quota exceeded")
		rbacSvc.AssertExpectations(t)
	})
}
