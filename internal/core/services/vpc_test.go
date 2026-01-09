package services_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const vpcTestCIDR = "10.0.0.0/16"

func setupVpcServiceTest(t *testing.T, cidr string) (*MockVpcRepo, *MockNetworkBackend, *MockAuditService, ports.VpcService) {
	vpcRepo := new(MockVpcRepo)
	network := new(MockNetworkBackend)
	auditSvc := new(MockAuditService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := services.NewVpcService(vpcRepo, network, auditSvc, logger, cidr)
	return vpcRepo, network, auditSvc, svc
}

func TestVpcServiceCreateSuccess(t *testing.T) {
	vpcRepo, network, auditSvc, svc := setupVpcServiceTest(t, vpcTestCIDR)
	defer vpcRepo.AssertExpectations(t)
	defer network.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	ctx := appcontext.WithUserID(context.Background(), uuid.New())
	name := "test-vpc"
	cidr := vpcTestCIDR

	network.On("CreateBridge", ctx, mock.MatchedBy(func(n string) bool {
		return len(n) > 0 // Dynamic name
	}), mock.Anything).Return(nil)
	vpcRepo.On("Create", ctx, mock.MatchedBy(func(vpc *domain.VPC) bool {
		return vpc.Name == name && vpc.CIDRBlock == cidr
	})).Return(nil)
	auditSvc.On("Log", ctx, mock.Anything, "vpc.create", "vpc", mock.Anything, mock.Anything).Return(nil)

	vpc, err := svc.CreateVPC(ctx, name, cidr)

	assert.NoError(t, err)
	assert.NotNil(t, vpc)
	assert.Equal(t, name, vpc.Name)
	assert.Contains(t, vpc.NetworkID, "br-vpc-")
}

func TestVpcServiceCreateDBFailureRollsBackBridge(t *testing.T) {
	vpcRepo, network, _, svc := setupVpcServiceTest(t, vpcTestCIDR)
	defer vpcRepo.AssertExpectations(t)
	defer network.AssertExpectations(t)

	ctx := context.Background()
	name := "fail-vpc"

	network.On("CreateBridge", ctx, mock.Anything, mock.Anything).Return(nil)
	vpcRepo.On("Create", ctx, mock.Anything).Return(assert.AnError)
	network.On("DeleteBridge", ctx, mock.Anything).Return(nil) // Rollback

	vpc, err := svc.CreateVPC(ctx, name, "")

	assert.Error(t, err)
	assert.Nil(t, vpc)
	network.AssertCalled(t, "DeleteBridge", ctx, mock.Anything)
}

func TestVpcServiceDeleteSuccess(t *testing.T) {
	vpcRepo, network, auditSvc, svc := setupVpcServiceTest(t, vpcTestCIDR)
	defer vpcRepo.AssertExpectations(t)
	defer network.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	ctx := context.Background()
	vpcID := uuid.New()
	vpc := &domain.VPC{
		ID:        vpcID,
		Name:      "to-delete",
		NetworkID: "br-vpc-123",
	}

	vpcRepo.On("GetByID", ctx, vpcID).Return(vpc, nil)
	network.On("DeleteBridge", ctx, "br-vpc-123").Return(nil)
	vpcRepo.On("Delete", ctx, vpcID).Return(nil)
	auditSvc.On("Log", ctx, mock.Anything, "vpc.delete", "vpc", mock.Anything, mock.Anything).Return(nil)

	err := svc.DeleteVPC(ctx, vpcID.String())

	assert.NoError(t, err)
}

func TestVpcServiceListSuccess(t *testing.T) {
	vpcRepo, _, _, svc := setupVpcServiceTest(t, vpcTestCIDR)
	defer vpcRepo.AssertExpectations(t)

	ctx := context.Background()

	vpcs := []*domain.VPC{{Name: "vpc1"}, {Name: "vpc2"}}
	vpcRepo.On("List", ctx).Return(vpcs, nil)

	result, err := svc.ListVPCs(ctx)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestVpcServiceGetByName(t *testing.T) {
	vpcRepo, _, _, svc := setupVpcServiceTest(t, vpcTestCIDR)
	defer vpcRepo.AssertExpectations(t)

	ctx := context.Background()
	name := "my-vpc"
	vpc := &domain.VPC{ID: uuid.New(), Name: name}

	vpcRepo.On("GetByName", ctx, name).Return(vpc, nil)

	result, err := svc.GetVPC(ctx, name)

	assert.NoError(t, err)
	assert.Equal(t, name, result.Name)
}
