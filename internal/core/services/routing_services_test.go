package services_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ---- Test Helpers ----

func newTestCtx(userID, tenantID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)
	return ctx
}

type natGatewaySvcMocks struct {
	nat     *MockNATGatewayRepo
	subnet  *MockSubnetRepo
	vpc     *MockVpcRepo
	eip     *MockEIPRepo
	network *MockNetworkBackend
	rbac    *MockRBACService
	audit   *MockAuditService
}

func setupNATGatewaySvcMocks() natGatewaySvcMocks {
	return natGatewaySvcMocks{
		nat:     new(MockNATGatewayRepo),
		subnet:  new(MockSubnetRepo),
		vpc:     new(MockVpcRepo),
		eip:     new(MockEIPRepo),
		network: new(MockNetworkBackend),
		rbac:    new(MockRBACService),
		audit:   new(MockAuditService),
	}
}

func (m *natGatewaySvcMocks) service() *services.NATGatewayService {
	return services.NewNATGatewayService(services.NATGatewayServiceParams{
		Repo:       m.nat,
		SubnetRepo: m.subnet,
		VpcRepo:    m.vpc,
		EIPRepo:    m.eip,
		Network:    m.network,
		RBACSvc:    m.rbac,
		AuditSvc:   m.audit,
		Logger:     slog.Default(),
	})
}

func (m *natGatewaySvcMocks) expectAuthSuccess() {
	m.rbac.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
}

// ---- InternetGateway Service Tests ----

func TestInternetGatewayService_CreateIGW(t *testing.T) {
	mockIGW := new(MockIGWRepo)
	mockRT := new(MockRTRepo)
	mockVPC := new(MockVpcRepo)
	mockRBAC := new(MockRBACService)
	mockAudit := new(MockAuditService)
	mockRBAC.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	svc := services.NewInternetGatewayService(services.InternetGatewayServiceParams{
		Repo:     mockIGW,
		RTRepo:   mockRT,
		VpcRepo:  mockVPC,
		RBACSvc:  mockRBAC,
		AuditSvc: mockAudit,
		Logger:   slog.Default(),
	})

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	mockIGW.On("Create", mock.Anything, mock.AnythingOfType("*domain.InternetGateway")).Return(nil)
	mockAudit.On("Log", mock.Anything, userID, "igw.create", "internet_gateway", mock.Anything, mock.Anything).Return(nil)

	igw, err := svc.CreateIGW(ctx)

	require.NoError(t, err)
	assert.NotNil(t, igw)
	assert.Equal(t, domain.IGWStatusDetached, igw.Status)
	assert.Nil(t, igw.VPCID)
	mockIGW.AssertExpectations(t)
}

func TestInternetGatewayService_AttachIGW_Success(t *testing.T) {
	mockIGW := new(MockIGWRepo)
	mockRT := new(MockRTRepo)
	mockVPC := new(MockVpcRepo)
	mockRBAC := new(MockRBACService)
	mockAudit := new(MockAuditService)

	svc := services.NewInternetGatewayService(services.InternetGatewayServiceParams{
		Repo:     mockIGW,
		RTRepo:   mockRT,
		VpcRepo:  mockVPC,
		RBACSvc:  mockRBAC,
		AuditSvc: mockAudit,
		Logger:   slog.Default(),
	})

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	vpcID := uuid.New()
	igwID := uuid.New()
	vpc := &domain.VPC{ID: vpcID, Name: "test-vpc"}
	igw := &domain.InternetGateway{ID: igwID, Status: domain.IGWStatusDetached, VPCID: nil, UserID: userID, TenantID: tenantID}

	mockRBAC.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockIGW.On("GetByID", mock.Anything, igwID).Return(igw, nil)
	mockVPC.On("GetByID", mock.Anything, vpcID).Return(vpc, nil)
	mockIGW.On("GetByVPC", mock.Anything, vpcID).Return(nil, errors.New("not found"))
	mockIGW.On("Update", mock.Anything, mock.AnythingOfType("*domain.InternetGateway")).Return(nil)
	mockRT.On("GetMainByVPC", mock.Anything, vpcID).Return((*domain.RouteTable)(nil), errors.New("no main"))
	mockAudit.On("Log", mock.Anything, userID, "igw.attach", "internet_gateway", mock.Anything, map[string]interface{}{"vpc_id": vpcID.String()}).Return(nil)

	err := svc.AttachIGW(ctx, igwID, vpcID)

	require.NoError(t, err)
	assert.Equal(t, domain.IGWStatusAttached, igw.Status)
	assert.Equal(t, &vpcID, igw.VPCID)
	mockIGW.AssertExpectations(t)
}

func TestInternetGatewayService_AttachIGW_AlreadyAttached(t *testing.T) {
	mockIGW := new(MockIGWRepo)
	mockVPC := new(MockVpcRepo)
	mockRBAC := new(MockRBACService)

	svc := services.NewInternetGatewayService(services.InternetGatewayServiceParams{
		Repo:    mockIGW,
		VpcRepo: mockVPC,
		RBACSvc: mockRBAC,
		Logger:  slog.Default(),
	})

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	vpcID := uuid.New()
	igwID := uuid.New()
	igw := &domain.InternetGateway{ID: igwID, Status: domain.IGWStatusAttached, VPCID: &vpcID}

	mockRBAC.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockIGW.On("GetByID", mock.Anything, igwID).Return(igw, nil)

	err := svc.AttachIGW(ctx, igwID, vpcID)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "already attached")
}

func TestInternetGatewayService_Delete_AttachedIGW(t *testing.T) {
	mockIGW := new(MockIGWRepo)
	mockRBAC := new(MockRBACService)
	mockAudit := new(MockAuditService)

	svc := services.NewInternetGatewayService(services.InternetGatewayServiceParams{
		Repo:     mockIGW,
		RBACSvc:  mockRBAC,
		AuditSvc: mockAudit,
		Logger:   slog.Default(),
	})

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	vpcID := uuid.New()
	igwID := uuid.New()
	igw := &domain.InternetGateway{ID: igwID, Status: domain.IGWStatusAttached, VPCID: &vpcID, UserID: userID, TenantID: tenantID}

	mockRBAC.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockIGW.On("GetByID", mock.Anything, igwID).Return(igw, nil)

	err := svc.DeleteIGW(ctx, igwID)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "attached")
	mockIGW.AssertExpectations(t)
}

func TestInternetGatewayService_Delete_DetachedIGW(t *testing.T) {
	mockIGW := new(MockIGWRepo)
	mockRBAC := new(MockRBACService)
	mockAudit := new(MockAuditService)

	svc := services.NewInternetGatewayService(services.InternetGatewayServiceParams{
		Repo:     mockIGW,
		RBACSvc:  mockRBAC,
		AuditSvc: mockAudit,
		Logger:   slog.Default(),
	})

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	igwID := uuid.New()
	igw := &domain.InternetGateway{ID: igwID, Status: domain.IGWStatusDetached, VPCID: nil, UserID: userID, TenantID: tenantID}

	mockRBAC.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockIGW.On("GetByID", mock.Anything, igwID).Return(igw, nil)
	mockIGW.On("Delete", mock.Anything, igwID).Return(nil)
	mockAudit.On("Log", mock.Anything, mock.Anything, "igw.delete", "internet_gateway", mock.Anything, mock.Anything).Return(nil)

	err := svc.DeleteIGW(ctx, igwID)

	require.NoError(t, err)
	mockIGW.AssertExpectations(t)
}

// ---- RouteTable Service Tests ----

func TestRouteTableService_CreateRouteTable(t *testing.T) {
	mockRT := new(MockRTRepo)
	mockVPC := new(MockVpcRepo)
	mockRBAC := new(MockRBACService)
	mockAudit := new(MockAuditService)

	svc := services.NewRouteTableService(services.RouteTableServiceParams{
		Repo:     mockRT,
		VpcRepo:  mockVPC,
		RBACSvc:  mockRBAC,
		AuditSvc: mockAudit,
		Logger:   slog.Default(),
	})

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	vpcID := uuid.New()
	vpc := &domain.VPC{ID: vpcID, Name: "test-vpc"}

	mockRBAC.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockVPC.On("GetByID", mock.Anything, vpcID).Return(vpc, nil)
	mockRT.On("Create", mock.Anything, mock.AnythingOfType("*domain.RouteTable")).Return(nil)
	mockAudit.On("Log", mock.Anything, mock.Anything, "route_table.create", "route_table", mock.Anything, mock.Anything).Return(nil)

	rt, err := svc.CreateRouteTable(ctx, vpcID, "test-rt", false)

	require.NoError(t, err)
	assert.NotNil(t, rt)
	assert.Equal(t, "test-rt", rt.Name)
	assert.Equal(t, vpcID, rt.VPCID)
	assert.False(t, rt.IsMain)
	mockRT.AssertExpectations(t)
}

func TestRouteTableService_CreateRouteTable_VPCNotFound(t *testing.T) {
	mockRT := new(MockRTRepo)
	mockVPC := new(MockVpcRepo)

	svc := services.NewRouteTableService(services.RouteTableServiceParams{
		Repo:    mockRT,
		VpcRepo: mockVPC,
		Logger:  slog.Default(),
	})

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	vpcID := uuid.New()
	mockVPC.On("GetByID", mock.Anything, vpcID).Return(nil, errors.New("not found"))

	_, err := svc.CreateRouteTable(ctx, vpcID, "test-rt", false)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRouteTableService_AddRoute(t *testing.T) {
	mockRT := new(MockRTRepo)
	mockVPC := new(MockVpcRepo)
	mockRBAC := new(MockRBACService)
	mockAudit := new(MockAuditService)
	mockNetwork := new(MockNetworkBackend)

	svc := services.NewRouteTableService(services.RouteTableServiceParams{
		Repo:     mockRT,
		VpcRepo:  mockVPC,
		RBACSvc:  mockRBAC,
		AuditSvc: mockAudit,
		Network:  mockNetwork,
		Logger:   slog.Default(),
	})

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	rtID := uuid.New()
	vpcID := uuid.New()
	rt := &domain.RouteTable{ID: rtID, VPCID: vpcID, Name: "test-rt"}
	vpc := &domain.VPC{ID: vpcID, NetworkID: "br-test"}
	igwID := uuid.New()

	mockRBAC.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockRT.On("GetByID", mock.Anything, rtID).Return(rt, nil)
	mockVPC.On("GetByID", mock.Anything, vpcID).Return(vpc, nil)
	mockRT.On("AddRoute", mock.Anything, rtID, mock.AnythingOfType("*domain.Route")).Return(nil)
	mockNetwork.On("AddFlowRule", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockAudit.On("Log", mock.Anything, mock.Anything, "route_table.add_route", "route_table", mock.Anything, mock.Anything).Return(nil)

	route, err := svc.AddRoute(ctx, rtID, "0.0.0.0/0", domain.RouteTargetIGW, &igwID)

	require.NoError(t, err)
	assert.Equal(t, "0.0.0.0/0", route.DestinationCIDR)
	assert.Equal(t, domain.RouteTargetIGW, route.TargetType)
	mockRT.AssertExpectations(t)
}

func TestRouteTableService_AssociateSubnet(t *testing.T) {
	mockRT := new(MockRTRepo)
	mockRBAC := new(MockRBACService)
	mockAudit := new(MockAuditService)

	svc := services.NewRouteTableService(services.RouteTableServiceParams{
		Repo:     mockRT,
		RBACSvc:  mockRBAC,
		AuditSvc: mockAudit,
		Logger:   slog.Default(),
	})

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	rtID := uuid.New()
	subnetID := uuid.New()

	mockRBAC.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockRT.On("AssociateSubnet", mock.Anything, rtID, subnetID).Return(nil)
	mockAudit.On("Log", mock.Anything, mock.Anything, "route_table.associate", "route_table", mock.Anything, mock.Anything).Return(nil)

	err := svc.AssociateSubnet(ctx, rtID, subnetID)

	require.NoError(t, err)
	mockRT.AssertExpectations(t)
}

func TestRouteTableService_DisassociateSubnet(t *testing.T) {
	mockRT := new(MockRTRepo)
	mockRBAC := new(MockRBACService)
	mockAudit := new(MockAuditService)

	svc := services.NewRouteTableService(services.RouteTableServiceParams{
		Repo:     mockRT,
		RBACSvc:  mockRBAC,
		AuditSvc: mockAudit,
		Logger:   slog.Default(),
	})

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	rtID := uuid.New()
	subnetID := uuid.New()

	mockRBAC.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockRT.On("DisassociateSubnet", mock.Anything, rtID, subnetID).Return(nil)
	mockAudit.On("Log", mock.Anything, mock.Anything, "route_table.disassociate", "route_table", mock.Anything, mock.Anything).Return(nil)

	err := svc.DisassociateSubnet(ctx, rtID, subnetID)

	require.NoError(t, err)
	mockRT.AssertExpectations(t)
}

// ---- NATGateway Service Tests ----

func TestNATGatewayService_CreateNATGateway(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	subnetID := uuid.New()
	eipID := uuid.New()
	vpcID := uuid.New()

	subnet := &domain.Subnet{ID: subnetID, VPCID: vpcID, GatewayIP: "10.0.0.1", CIDRBlock: "10.0.0.0/24"}
	vpc := &domain.VPC{ID: vpcID, Name: "vpc", NetworkID: "br-vpc"}
	eipAllocated := &domain.ElasticIP{ID: eipID, UserID: userID, TenantID: tenantID, PublicIP: "203.0.113.10", Status: domain.EIPStatusAllocated}

	tests := []struct {
		name       string
		subnet     *domain.Subnet
		vpc        *domain.VPC
		eip        *domain.ElasticIP
		networkErr error
		wantErr    bool
		errContains string
	}{
		{
			name:    "success",
			subnet:  subnet,
			vpc:     vpc,
			eip:     eipAllocated,
			wantErr: false,
		},
		{
			name:        "eip not allocated",
			subnet:      subnet,
			vpc:         vpc,
			eip:         &domain.ElasticIP{ID: eipID, Status: domain.EIPStatusAssociated},
			wantErr:     true,
			errContains: "allocated",
		},
		{
			name:        "invalid egress ip format",
			subnet:      subnet,
			vpc:         vpc,
			eip:         &domain.ElasticIP{ID: eipID, Status: domain.EIPStatusAllocated, PublicIP: "invalid-ip"},
			wantErr:     true,
			errContains: "invalid",
		},
		{
			name:        "network setup failed",
			subnet:      subnet,
			vpc:         vpc,
			eip:         eipAllocated,
			networkErr:  errors.New("network error"),
			wantErr:     true,
			errContains: "failed to setup NAT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mocks := setupNATGatewaySvcMocks()
			mocks.expectAuthSuccess()

			if tt.subnet != nil {
				mocks.subnet.On("GetByID", mock.Anything, subnetID).Return(tt.subnet, nil)
			}
			if tt.vpc != nil {
				mocks.vpc.On("GetByID", mock.Anything, tt.subnet.VPCID).Return(tt.vpc, nil)
			}
			if tt.eip != nil {
				mocks.eip.On("GetByID", mock.Anything, eipID).Return(tt.eip, nil)
			}

			// Create is always called
			mocks.nat.On("Create", mock.Anything, mock.AnythingOfType("*domain.NATGateway")).Return(nil).Maybe()

			if tt.networkErr != nil {
				// Network fails - NAT is created then marked as failed
				mocks.network.On("SetupNATForSubnet", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tt.networkErr).Maybe()
				mocks.nat.On("Update", mock.Anything, mock.AnythingOfType("*domain.NATGateway")).Return(nil).Maybe()
			} else if !tt.wantErr {
				// Success case
				mocks.network.On("SetupNATForSubnet", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
				mocks.nat.On("Update", mock.Anything, mock.AnythingOfType("*domain.NATGateway")).Return(nil).Maybe()
				mocks.audit.On("Log", mock.Anything, mock.Anything, "nat_gateway.create", "nat_gateway", mock.Anything, mock.Anything).Return(nil).Maybe()
			}

			ctx := newTestCtx(userID, tenantID)
			nat, err := mocks.service().CreateNATGateway(ctx, subnetID, eipID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, nat)
				assert.Equal(t, subnetID, nat.SubnetID)
				assert.Equal(t, eipID, nat.ElasticIPID)
				assert.Equal(t, domain.NATGatewayStatusActive, nat.Status)
			}
		})
	}
}

func TestNATGatewayService_DeleteNATGateway(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	natID := uuid.New()
	subnetID := uuid.New()
	eipID := uuid.New()
	vpcID := uuid.New()

	nat := &domain.NATGateway{ID: natID, SubnetID: subnetID, ElasticIPID: eipID, Status: domain.NATGatewayStatusActive, VPCID: vpcID, UserID: userID, TenantID: tenantID}
	subnet := &domain.Subnet{ID: subnetID, CIDRBlock: "10.0.0.0/24"}
	vpc := &domain.VPC{ID: vpcID, NetworkID: "br-vpc"}
	eip := &domain.ElasticIP{ID: eipID, Status: domain.EIPStatusAssociated, PublicIP: "203.0.113.10"}

	tests := []struct {
		name        string
		nat         *domain.NATGateway
		subnet      *domain.Subnet
		vpc         *domain.VPC
		eip         *domain.ElasticIP
		natGetErr   error
		subnetGetErr error
		vpcGetErr    error
		eipGetErr    error
		networkErr   error
		wantErr      bool
		errContains  string
	}{
		{
			name:    "success",
			nat:     nat,
			subnet:  subnet,
			vpc:     vpc,
			eip:     eip,
			wantErr: false,
		},
		{
			name:       "nat not found",
			natGetErr:  errors.New("not found"),
			wantErr:    true,
			errContains: "not found",
		},
		{
			name:        "subnet not found",
			nat:         nat,
			subnetGetErr: errors.New("subnet not found"),
			wantErr:     true,
			errContains: "subnet",
		},
		{
			name:     "vpc not found",
			nat:      nat,
			subnet:   subnet,
			vpcGetErr: errors.New("vpc not found"),
			wantErr:  true,
			errContains: "vpc",
		},
		{
			name:      "eip not found",
			nat:       nat,
			subnet:    subnet,
			vpc:       vpc,
			eipGetErr: errors.New("eip not found"),
			wantErr:  true,
			errContains: "eip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mocks := setupNATGatewaySvcMocks()
			mocks.expectAuthSuccess()

			if tt.natGetErr != nil {
				mocks.nat.On("GetByID", mock.Anything, natID).Return(nil, tt.natGetErr)
			} else {
				mocks.nat.On("GetByID", mock.Anything, natID).Return(tt.nat, nil)
				mocks.subnet.On("GetByID", mock.Anything, tt.nat.SubnetID).Return(tt.subnet, tt.subnetGetErr)
			}
			if tt.subnetGetErr == nil && tt.subnet != nil {
				mocks.vpc.On("GetByID", mock.Anything, tt.nat.VPCID).Return(tt.vpc, tt.vpcGetErr)
			}
			if tt.vpcGetErr == nil && tt.vpc != nil {
				mocks.eip.On("GetByID", mock.Anything, tt.nat.ElasticIPID).Return(tt.eip, tt.eipGetErr)
			}

			if !tt.wantErr {
				mocks.eip.On("Update", mock.Anything, mock.AnythingOfType("*domain.ElasticIP")).Return(nil).Maybe()
				mocks.network.On("RemoveNATForSubnet", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tt.networkErr).Maybe()
				if tt.networkErr == nil {
					mocks.nat.On("Delete", mock.Anything, natID).Return(nil).Maybe()
					mocks.audit.On("Log", mock.Anything, mock.Anything, "nat_gateway.delete", "nat_gateway", mock.Anything, mock.Anything).Return(nil).Maybe()
				}
			}

			ctx := newTestCtx(userID, tenantID)
			err := mocks.service().DeleteNATGateway(ctx, natID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}