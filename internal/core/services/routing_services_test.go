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
	mockNAT := new(MockNATGatewayRepo)
	mockSubnet := new(MockSubnetRepo)
	mockVPC := new(MockVpcRepo)
	mockEIP := new(MockEIPRepo)
	mockNetwork := new(MockNetworkBackend)
	mockRBAC := new(MockRBACService)
	mockAudit := new(MockAuditService)

	svc := services.NewNATGatewayService(services.NATGatewayServiceParams{
		Repo:       mockNAT,
		SubnetRepo: mockSubnet,
		VpcRepo:    mockVPC,
		EIPRepo:    mockEIP,
		Network:    mockNetwork,
		RBACSvc:    mockRBAC,
		AuditSvc:   mockAudit,
		Logger:     slog.Default(),
	})

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	subnetID := uuid.New()
	eipID := uuid.New()
	vpcID := uuid.New()
	subnet := &domain.Subnet{ID: subnetID, VPCID: vpcID, GatewayIP: "10.0.0.1", CIDRBlock: "10.0.0.0/24"}
	vpc := &domain.VPC{ID: vpcID, Name: "vpc", NetworkID: "br-vpc"}
	eip := &domain.ElasticIP{ID: eipID, UserID: userID, TenantID: tenantID, PublicIP: "203.0.113.10", Status: domain.EIPStatusAllocated}

	mockRBAC.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockSubnet.On("GetByID", mock.Anything, subnetID).Return(subnet, nil)
	mockVPC.On("GetByID", mock.Anything, vpcID).Return(vpc, nil)
	mockEIP.On("GetByID", mock.Anything, eipID).Return(eip, nil)
	mockNAT.On("Create", mock.Anything, mock.AnythingOfType("*domain.NATGateway")).Return(nil)
	mockNetwork.On("SetupNATForSubnet", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockNAT.On("Update", mock.Anything, mock.AnythingOfType("*domain.NATGateway")).Return(nil)
	mockAudit.On("Log", mock.Anything, mock.Anything, "nat_gateway.create", "nat_gateway", mock.Anything, mock.Anything).Return(nil)

	nat, err := svc.CreateNATGateway(ctx, subnetID, eipID)

	require.NoError(t, err)
	assert.NotNil(t, nat)
	assert.Equal(t, subnetID, nat.SubnetID)
	assert.Equal(t, eipID, nat.ElasticIPID)
	assert.Equal(t, domain.NATGatewayStatusActive, nat.Status)
	mockNAT.AssertExpectations(t)
}

func TestNATGatewayService_CreateNATGateway_EIPNotAllocated(t *testing.T) {
	mockNAT := new(MockNATGatewayRepo)
	mockSubnet := new(MockSubnetRepo)
	mockVPC := new(MockVpcRepo)
	mockEIP := new(MockEIPRepo)
	mockRBAC := new(MockRBACService)

	svc := services.NewNATGatewayService(services.NATGatewayServiceParams{
		Repo:       mockNAT,
		SubnetRepo: mockSubnet,
		VpcRepo:    mockVPC,
		EIPRepo:    mockEIP,
		RBACSvc:    mockRBAC,
		Logger:     slog.Default(),
	})

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	subnetID := uuid.New()
	eipID := uuid.New()
	vpcID := uuid.New()
	subnet := &domain.Subnet{ID: subnetID, VPCID: vpcID, GatewayIP: "10.0.0.1", CIDRBlock: "10.0.0.0/24"}
	eip := &domain.ElasticIP{ID: eipID, Status: domain.EIPStatusAssociated} // wrong status

	mockRBAC.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockSubnet.On("GetByID", mock.Anything, subnetID).Return(subnet, nil)
	mockVPC.On("GetByID", mock.Anything, vpcID).Return(&domain.VPC{ID: vpcID}, nil)
	mockEIP.On("GetByID", mock.Anything, eipID).Return(eip, nil)

	_, err := svc.CreateNATGateway(ctx, subnetID, eipID)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "allocated")
}

func TestNATGatewayService_DeleteNATGateway(t *testing.T) {
	mockNAT := new(MockNATGatewayRepo)
	mockSubnet := new(MockSubnetRepo)
	mockVPC := new(MockVpcRepo)
	mockEIP := new(MockEIPRepo)
	mockNetwork := new(MockNetworkBackend)
	mockRBAC := new(MockRBACService)
	mockAudit := new(MockAuditService)

	svc := services.NewNATGatewayService(services.NATGatewayServiceParams{
		Repo:       mockNAT,
		SubnetRepo: mockSubnet,
		VpcRepo:    mockVPC,
		EIPRepo:    mockEIP,
		Network:    mockNetwork,
		RBACSvc:    mockRBAC,
		AuditSvc:   mockAudit,
		Logger:     slog.Default(),
	})

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	natID := uuid.New()
	subnetID := uuid.New()
	eipID := uuid.New()
	vpcID := uuid.New()
	nat := &domain.NATGateway{ID: natID, SubnetID: subnetID, ElasticIPID: eipID, Status: domain.NATGatewayStatusActive, VPCID: vpcID, UserID: userID, TenantID: tenantID}
	subnet := &domain.Subnet{ID: subnetID, CIDRBlock: "10.0.0.0/24"}
	vpc := &domain.VPC{ID: vpcID, NetworkID: "br-vpc"}
	eip := &domain.ElasticIP{ID: eipID, Status: domain.EIPStatusAssociated, PublicIP: "203.0.113.10"}

	mockRBAC.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockNAT.On("GetByID", mock.Anything, natID).Return(nat, nil)
	mockSubnet.On("GetByID", mock.Anything, subnetID).Return(subnet, nil)
	mockVPC.On("GetByID", mock.Anything, vpcID).Return(vpc, nil)
	mockEIP.On("GetByID", mock.Anything, eipID).Return(eip, nil)
	mockEIP.On("Update", mock.Anything, mock.AnythingOfType("*domain.ElasticIP")).Return(nil)
	mockNetwork.On("RemoveNATForSubnet", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockNAT.On("Delete", mock.Anything, natID).Return(nil)
	mockAudit.On("Log", mock.Anything, mock.Anything, "nat_gateway.delete", "nat_gateway", mock.Anything, mock.Anything).Return(nil)

	err := svc.DeleteNATGateway(ctx, natID)

	require.NoError(t, err)
	mockNAT.AssertExpectations(t)
}

func TestNATGatewayService_DeleteNATGateway_NotFound(t *testing.T) {
	mockNAT := new(MockNATGatewayRepo)
	mockRBAC := new(MockRBACService)

	svc := services.NewNATGatewayService(services.NATGatewayServiceParams{
		Repo:     mockNAT,
		RBACSvc:  mockRBAC,
		Logger:   slog.Default(),
	})

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	natID := uuid.New()
	mockRBAC.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockNAT.On("GetByID", mock.Anything, natID).Return(nil, errors.New("not found"))

	err := svc.DeleteNATGateway(ctx, natID)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}