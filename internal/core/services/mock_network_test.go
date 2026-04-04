package services_test

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/stretchr/testify/mock"
)

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
	return args.Get(0).([]*domain.VPC), args.Error(1)
}
func (m *MockVpcRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

// MockVpcService
type MockVpcService struct{ mock.Mock }

func (m *MockVpcService) CreateVPC(ctx context.Context, name, cidrBlock string) (*domain.VPC, error) {
	args := m.Called(ctx, name, cidrBlock)
	r0, _ := args.Get(0).(*domain.VPC)
	return r0, args.Error(1)
}
func (m *MockVpcService) GetVPC(ctx context.Context, idOrName string) (*domain.VPC, error) {
	args := m.Called(ctx, idOrName)
	r0, _ := args.Get(0).(*domain.VPC)
	return r0, args.Error(1)
}
func (m *MockVpcService) ListVPCs(ctx context.Context) ([]*domain.VPC, error) {
	args := m.Called(ctx)
	r0, _ := args.Get(0).([]*domain.VPC)
	return r0, args.Error(1)
}
func (m *MockVpcService) DeleteVPC(ctx context.Context, idOrName string) error {
	return m.Called(ctx, idOrName).Error(0)
}

// MockSubnetRepo
type MockSubnetRepo struct{ mock.Mock }

func (m *MockSubnetRepo) Create(ctx context.Context, subnet *domain.Subnet) error {
	return m.Called(ctx, subnet).Error(0)
}
func (m *MockSubnetRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subnet, error) {
	args := m.Called(ctx, id)
	r0, _ := args.Get(0).(*domain.Subnet)
	return r0, args.Error(1)
}
func (m *MockSubnetRepo) GetByName(ctx context.Context, vpcID uuid.UUID, name string) (*domain.Subnet, error) {
	args := m.Called(ctx, vpcID, name)
	r0, _ := args.Get(0).(*domain.Subnet)
	return r0, args.Error(1)
}
func (m *MockSubnetRepo) ListByVPC(ctx context.Context, vpcID uuid.UUID) ([]*domain.Subnet, error) {
	args := m.Called(ctx, vpcID)
	r0, _ := args.Get(0).([]*domain.Subnet)
	return r0, args.Error(1)
}
func (m *MockSubnetRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

// MockVPCPeeringRepo
type MockVPCPeeringRepo struct{ mock.Mock }

func (m *MockVPCPeeringRepo) Create(ctx context.Context, peering *domain.VPCPeering) error {
	return m.Called(ctx, peering).Error(0)
}
func (m *MockVPCPeeringRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.VPCPeering, error) {
	args := m.Called(ctx, id)
	r0, _ := args.Get(0).(*domain.VPCPeering)
	return r0, args.Error(1)
}
func (m *MockVPCPeeringRepo) List(ctx context.Context, tenantID uuid.UUID) ([]*domain.VPCPeering, error) {
	args := m.Called(ctx, tenantID)
	r0, _ := args.Get(0).([]*domain.VPCPeering)
	return r0, args.Error(1)
}
func (m *MockVPCPeeringRepo) ListByVPC(ctx context.Context, vpcID uuid.UUID) ([]*domain.VPCPeering, error) {
	args := m.Called(ctx, vpcID)
	r0, _ := args.Get(0).([]*domain.VPCPeering)
	return r0, args.Error(1)
}
func (m *MockVPCPeeringRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return m.Called(ctx, id, status).Error(0)
}
func (m *MockVPCPeeringRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockVPCPeeringRepo) GetActiveByVPCPair(ctx context.Context, v1, v2 uuid.UUID) (*domain.VPCPeering, error) {
	args := m.Called(ctx, v1, v2)
	r0, _ := args.Get(0).(*domain.VPCPeering)
	return r0, args.Error(1)
}

// MockLBRepo
type MockLBRepo struct{ mock.Mock }

func (m *MockLBRepo) Create(ctx context.Context, lb *domain.LoadBalancer) error {
	return m.Called(ctx, lb).Error(0)
}
func (m *MockLBRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.LoadBalancer, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.LoadBalancer), args.Error(1)
}
func (m *MockLBRepo) GetByName(ctx context.Context, name string) (*domain.LoadBalancer, error) {
	args := m.Called(ctx, name)
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
	return m.Called(ctx, id).Error(0)
}
func (m *MockLBRepo) AddTarget(ctx context.Context, target *domain.LBTarget) error {
	return m.Called(ctx, target).Error(0)
}
func (m *MockLBRepo) RemoveTarget(ctx context.Context, lbID, instanceID uuid.UUID) error {
	return m.Called(ctx, lbID, instanceID).Error(0)
}
func (m *MockLBRepo) ListTargets(ctx context.Context, lbID uuid.UUID) ([]*domain.LBTarget, error) {
	args := m.Called(ctx, lbID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.LBTarget), args.Error(1)
}
func (m *MockLBRepo) UpdateTargetHealth(ctx context.Context, lbID, instanceID uuid.UUID, health string) error {
	return m.Called(ctx, lbID, instanceID, health).Error(0)
}
func (m *MockLBRepo) GetTargetsForInstance(ctx context.Context, instanceID uuid.UUID) ([]*domain.LBTarget, error) {
	args := m.Called(ctx, instanceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.LBTarget), args.Error(1)
}

// MockLBService
type MockLBService struct{ mock.Mock }

func (m *MockLBService) Create(ctx context.Context, name string, vpcID uuid.UUID, port int, algo string, idempotencyKey string) (*domain.LoadBalancer, error) {
	args := m.Called(ctx, name, vpcID, port, algo, idempotencyKey)
	r0, _ := args.Get(0).(*domain.LoadBalancer)
	return r0, args.Error(1)
}
func (m *MockLBService) Get(ctx context.Context, idOrName string) (*domain.LoadBalancer, error) {
	args := m.Called(ctx, idOrName)
	r0, _ := args.Get(0).(*domain.LoadBalancer)
	return r0, args.Error(1)
}
func (m *MockLBService) List(ctx context.Context) ([]*domain.LoadBalancer, error) {
	args := m.Called(ctx)
	r0, _ := args.Get(0).([]*domain.LoadBalancer)
	return r0, args.Error(1)
}
func (m *MockLBService) Delete(ctx context.Context, idOrName string) error {
	args := m.Called(ctx, idOrName)
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
	r0, _ := args.Get(0).([]*domain.LBTarget)
	return r0, args.Error(1)
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
	args := m.Called(ctx, instanceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.SecurityGroup), args.Error(1)
}

// MockNetworkBackend
type MockNetworkBackend struct{ mock.Mock }

func (m *MockNetworkBackend) CreateBridge(ctx context.Context, name string, vxlanID int) error {
	return m.Called(ctx, name, vxlanID).Error(0)
}
func (m *MockNetworkBackend) DeleteBridge(ctx context.Context, name string) error {
	return m.Called(ctx, name).Error(0)
}
func (m *MockNetworkBackend) ListBridges(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}
func (m *MockNetworkBackend) AddPort(ctx context.Context, bridge, portName string) error {
	return m.Called(ctx, bridge, portName).Error(0)
}
func (m *MockNetworkBackend) DeletePort(ctx context.Context, bridge, portName string) error {
	return m.Called(ctx, bridge, portName).Error(0)
}
func (m *MockNetworkBackend) CreateVXLANTunnel(ctx context.Context, bridge string, vni int, remoteIP string) error {
	return m.Called(ctx, bridge, vni, remoteIP).Error(0)
}
func (m *MockNetworkBackend) DeleteVXLANTunnel(ctx context.Context, bridge string, remoteIP string) error {
	return m.Called(ctx, bridge, remoteIP).Error(0)
}
func (m *MockNetworkBackend) AddFlowRule(ctx context.Context, bridge string, rule ports.FlowRule) error {
	return m.Called(ctx, bridge, rule).Error(0)
}
func (m *MockNetworkBackend) DeleteFlowRule(ctx context.Context, bridge string, match string) error {
	return m.Called(ctx, bridge, match).Error(0)
}
func (m *MockNetworkBackend) ListFlowRules(ctx context.Context, bridge string) ([]ports.FlowRule, error) {
	args := m.Called(ctx, bridge)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ports.FlowRule), args.Error(1)
}
func (m *MockNetworkBackend) CreateVethPair(ctx context.Context, hostEnd, containerEnd string) error {
	return m.Called(ctx, hostEnd, containerEnd).Error(0)
}
func (m *MockNetworkBackend) AttachVethToBridge(ctx context.Context, bridge, vethEnd string) error {
	return m.Called(ctx, bridge, vethEnd).Error(0)
}
func (m *MockNetworkBackend) DeleteVethPair(ctx context.Context, hostEnd string) error {
	return m.Called(ctx, hostEnd).Error(0)
}
func (m *MockNetworkBackend) SetVethIP(ctx context.Context, vethEnd, ip, cidr string) error {
	return m.Called(ctx, vethEnd, ip, cidr).Error(0)
}
func (m *MockNetworkBackend) Ping(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}
func (m *MockNetworkBackend) Type() string {
	return m.Called().String(0)
}
