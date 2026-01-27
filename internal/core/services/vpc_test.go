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
	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupVpcServiceTest(cidr string) (*MockVpcRepo, *MockNetworkBackend, *MockAuditService, ports.VpcService) {
	vpcRepo := new(MockVpcRepo)
	network := new(MockNetworkBackend)
	auditSvc := new(MockAuditService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := services.NewVpcService(vpcRepo, network, auditSvc, logger, cidr)
	return vpcRepo, network, auditSvc, svc
}

func TestVpcServiceCreateSuccess(t *testing.T) {
	vpcRepo, network, auditSvc, svc := setupVpcServiceTest(testutil.TestCIDR)
	defer vpcRepo.AssertExpectations(t)
	defer network.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	ctx := appcontext.WithUserID(context.Background(), uuid.New())
	name := "test-vpc"
	cidr := testutil.TestCIDR

	network.On("CreateBridge", mock.Anything, mock.MatchedBy(func(n string) bool {
		return len(n) > 0 // Dynamic name
	}), mock.Anything).Return(nil)
	vpcRepo.On("Create", mock.Anything, mock.MatchedBy(func(vpc *domain.VPC) bool {
		return vpc.Name == name && vpc.CIDRBlock == cidr
	})).Return(nil)
	auditSvc.On("Log", mock.Anything, mock.Anything, "vpc.create", "vpc", mock.Anything, mock.Anything).Return(nil)

	vpc, err := svc.CreateVPC(ctx, name, cidr)

	assert.NoError(t, err)
	assert.NotNil(t, vpc)
	assert.Equal(t, name, vpc.Name)
	assert.Contains(t, vpc.NetworkID, "br-vpc-")
}

func TestVpcServiceCreateDBFailureRollsBackBridge(t *testing.T) {
	vpcRepo, network, _, svc := setupVpcServiceTest(testutil.TestCIDR)
	defer vpcRepo.AssertExpectations(t)
	defer network.AssertExpectations(t)

	name := "fail-vpc"

	network.On("CreateBridge", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	vpcRepo.On("Create", mock.Anything, mock.Anything).Return(assert.AnError)
	network.On("DeleteBridge", mock.Anything, mock.Anything).Return(nil) // Rollback

	vpc, err := svc.CreateVPC(context.Background(), name, "")

	assert.Error(t, err)
	assert.Nil(t, vpc)
	network.AssertCalled(t, "DeleteBridge", mock.Anything, mock.Anything)
}

func TestVpcServiceCreateDBFailureRollbackError(t *testing.T) {
	vpcRepo, network, _, svc := setupVpcServiceTest(testutil.TestCIDR)
	defer vpcRepo.AssertExpectations(t)
	defer network.AssertExpectations(t)

	name := "fail-vpc"

	network.On("CreateBridge", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	vpcRepo.On("Create", mock.Anything, mock.Anything).Return(assert.AnError)
	network.On("DeleteBridge", mock.Anything, mock.Anything).Return(assert.AnError)

	vpc, err := svc.CreateVPC(context.Background(), name, "")

	assert.Error(t, err)
	assert.Nil(t, vpc)
}

func TestVpcServiceCreateBridgeError(t *testing.T) {
	vpcRepo, network, _, svc := setupVpcServiceTest(testutil.TestCIDR)
	defer vpcRepo.AssertExpectations(t)
	defer network.AssertExpectations(t)

	network.On("CreateBridge", mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError)

	vpc, err := svc.CreateVPC(context.Background(), "bad-bridge", "")

	assert.Error(t, err)
	assert.Nil(t, vpc)
	vpcRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestVpcServiceCreateDefaultCIDR(t *testing.T) {
	vpcRepo, network, auditSvc, svc := setupVpcServiceTest("")
	defer vpcRepo.AssertExpectations(t)
	defer network.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	ctx := appcontext.WithUserID(context.Background(), uuid.New())

	network.On("CreateBridge", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	vpcRepo.On("Create", mock.Anything, mock.MatchedBy(func(vpc *domain.VPC) bool {
		return vpc.CIDRBlock == "10.0.0.0/16"
	})).Return(nil)
	auditSvc.On("Log", mock.Anything, mock.Anything, "vpc.create", "vpc", mock.Anything, mock.Anything).Return(nil)

	vpc, err := svc.CreateVPC(ctx, "default-cidr", "")

	assert.NoError(t, err)
	assert.NotNil(t, vpc)
}

func TestVpcServiceDeleteSuccess(t *testing.T) {
	vpcRepo, network, auditSvc, svc := setupVpcServiceTest(testutil.TestCIDR)
	defer vpcRepo.AssertExpectations(t)
	defer network.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	vpcID := uuid.New()
	vpc := &domain.VPC{
		ID:        vpcID,
		Name:      "to-delete",
		NetworkID: "br-vpc-123",
	}

	vpcRepo.On("GetByID", mock.Anything, vpcID).Return(vpc, nil)
	network.On("DeleteBridge", mock.Anything, "br-vpc-123").Return(nil)
	vpcRepo.On("Delete", mock.Anything, vpcID).Return(nil)
	auditSvc.On("Log", mock.Anything, mock.Anything, "vpc.delete", "vpc", mock.Anything, mock.Anything).Return(nil)

	err := svc.DeleteVPC(context.Background(), vpcID.String())

	assert.NoError(t, err)
}

func TestVpcServiceDeleteBridgeError(t *testing.T) {
	vpcRepo, network, _, svc := setupVpcServiceTest(testutil.TestCIDR)
	defer vpcRepo.AssertExpectations(t)
	defer network.AssertExpectations(t)

	vpcID := uuid.New()
	vpc := &domain.VPC{
		ID:        vpcID,
		Name:      "to-delete",
		NetworkID: "br-vpc-err",
	}

	vpcRepo.On("GetByID", mock.Anything, vpcID).Return(vpc, nil)
	network.On("DeleteBridge", mock.Anything, "br-vpc-err").Return(assert.AnError)

	err := svc.DeleteVPC(context.Background(), vpcID.String())

	assert.Error(t, err)
	vpcRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestVpcServiceDeleteGetError(t *testing.T) {
	vpcRepo, _, _, svc := setupVpcServiceTest(testutil.TestCIDR)
	defer vpcRepo.AssertExpectations(t)

	vpcRepo.On("GetByID", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	err := svc.DeleteVPC(context.Background(), uuid.New().String())
	assert.Error(t, err)
}

func TestVpcServiceDeleteRepoError(t *testing.T) {
	vpcRepo, network, _, svc := setupVpcServiceTest(testutil.TestCIDR)
	defer vpcRepo.AssertExpectations(t)
	defer network.AssertExpectations(t)

	vpcID := uuid.New()
	vpc := &domain.VPC{
		ID:        vpcID,
		Name:      "to-delete",
		NetworkID: "br-vpc-ok",
	}

	vpcRepo.On("GetByID", mock.Anything, vpcID).Return(vpc, nil)
	network.On("DeleteBridge", mock.Anything, "br-vpc-ok").Return(nil)
	vpcRepo.On("Delete", mock.Anything, vpcID).Return(assert.AnError)

	err := svc.DeleteVPC(context.Background(), vpcID.String())

	assert.Error(t, err)
}

func TestVpcServiceDeleteByName(t *testing.T) {
	vpcRepo, network, auditSvc, svc := setupVpcServiceTest(testutil.TestCIDR)
	defer vpcRepo.AssertExpectations(t)
	defer network.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	name := "by-name"
	vpcID := uuid.New()
	vpc := &domain.VPC{ID: vpcID, Name: name, NetworkID: "br-vpc-name"}

	vpcRepo.On("GetByName", mock.Anything, name).Return(vpc, nil)
	network.On("DeleteBridge", mock.Anything, "br-vpc-name").Return(nil)
	vpcRepo.On("Delete", mock.Anything, vpcID).Return(nil)
	auditSvc.On("Log", mock.Anything, mock.Anything, "vpc.delete", "vpc", mock.Anything, mock.Anything).Return(nil)

	err := svc.DeleteVPC(context.Background(), name)
	assert.NoError(t, err)
}

func TestVpcServiceListSuccess(t *testing.T) {
	vpcRepo, _, _, svc := setupVpcServiceTest(testutil.TestCIDR)
	defer vpcRepo.AssertExpectations(t)

	vpcs := []*domain.VPC{{Name: "vpc1"}, {Name: "vpc2"}}
	vpcRepo.On("List", mock.Anything).Return(vpcs, nil)

	result, err := svc.ListVPCs(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestVpcServiceGetByName(t *testing.T) {
	vpcRepo, _, _, svc := setupVpcServiceTest(testutil.TestCIDR)
	defer vpcRepo.AssertExpectations(t)

	name := "my-vpc"
	vpc := &domain.VPC{ID: uuid.New(), Name: name}

	vpcRepo.On("GetByName", mock.Anything, name).Return(vpc, nil)

	result, err := svc.GetVPC(context.Background(), name)

	assert.NoError(t, err)
	assert.Equal(t, name, result.Name)
}
