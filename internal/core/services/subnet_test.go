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

const subnetTestAZ = "us-east-1a"

func setupSubnetServiceTest(_ *testing.T) (*MockSubnetRepo, *MockVpcRepo, *MockAuditService, ports.SubnetService) {
	repo := new(MockSubnetRepo)
	vpcRepo := new(MockVpcRepo)
	auditSvc := new(MockAuditService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := services.NewSubnetService(repo, vpcRepo, auditSvc, logger)
	return repo, vpcRepo, auditSvc, svc
}

func TestSubnetServiceCreateSubnetSuccess(t *testing.T) {
	repo, vpcRepo, auditSvc, svc := setupSubnetServiceTest(t)
	defer repo.AssertExpectations(t)
	defer vpcRepo.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	ctx := appcontext.WithUserID(context.Background(), uuid.New())
	vpcID := uuid.New()
	vpc := &domain.VPC{
		ID:        vpcID,
		CIDRBlock: testutil.TestCIDR,
	}

	vpcRepo.On("GetByID", ctx, vpcID).Return(vpc, nil)
	repo.On("Create", ctx, mock.MatchedBy(func(s *domain.Subnet) bool {
		return s.VPCID == vpcID && s.CIDRBlock == testutil.TestSubnetCIDR && s.GatewayIP == testutil.TestGatewayIP
	})).Return(nil)
	auditSvc.On("Log", ctx, mock.Anything, "subnet.create", "subnet", mock.Anything, mock.Anything).Return(nil)

	subnet, err := svc.CreateSubnet(ctx, vpcID, "test-subnet", testutil.TestSubnetCIDR, subnetTestAZ)

	assert.NoError(t, err)
	assert.NotNil(t, subnet)
	assert.Equal(t, testutil.TestGatewayIP, subnet.GatewayIP)
}

func TestSubnetServiceCreateSubnetRepoError(t *testing.T) {
	repo, vpcRepo, _, svc := setupSubnetServiceTest(t)
	defer repo.AssertExpectations(t)
	defer vpcRepo.AssertExpectations(t)

	ctx := appcontext.WithUserID(context.Background(), uuid.New())
	vpcID := uuid.New()
	vpc := &domain.VPC{ID: vpcID, CIDRBlock: testutil.TestCIDR}

	vpcRepo.On("GetByID", ctx, vpcID).Return(vpc, nil)
	repo.On("Create", ctx, mock.Anything).Return(assert.AnError)

	subnet, err := svc.CreateSubnet(ctx, vpcID, "test-subnet", testutil.TestSubnetCIDR, subnetTestAZ)
	assert.Error(t, err)
	assert.Nil(t, subnet)
}

func TestSubnetServiceCreateSubnetInvalidCIDR(t *testing.T) {
	repo, vpcRepo, _, svc := setupSubnetServiceTest(t)
	defer repo.AssertExpectations(t)
	defer vpcRepo.AssertExpectations(t)

	ctx := context.Background()
	vpcID := uuid.New()
	vpc := &domain.VPC{
		ID:        vpcID,
		CIDRBlock: testutil.TestCIDR,
	}

	vpcRepo.On("GetByID", ctx, vpcID).Return(vpc, nil)

	// Outside VPC range
	subnet, err := svc.CreateSubnet(ctx, vpcID, "bad-subnet", testutil.TestOtherCIDR, subnetTestAZ)

	assert.Error(t, err)
	assert.Nil(t, subnet)
	assert.Contains(t, err.Error(), "within VPC CIDR range")
}

func TestSubnetServiceCreateSubnetVpcRepoError(t *testing.T) {
	repo, vpcRepo, _, svc := setupSubnetServiceTest(t)
	defer repo.AssertExpectations(t)
	defer vpcRepo.AssertExpectations(t)

	ctx := context.Background()
	vpcID := uuid.New()

	vpcRepo.On("GetByID", ctx, vpcID).Return(nil, assert.AnError)

	subnet, err := svc.CreateSubnet(ctx, vpcID, "bad-vpc", testutil.TestSubnetCIDR, subnetTestAZ)
	assert.Error(t, err)
	assert.Nil(t, subnet)
}

func TestSubnetServiceCreateSubnetInvalidSubnetCIDRFormat(t *testing.T) {
	repo, vpcRepo, _, svc := setupSubnetServiceTest(t)
	defer repo.AssertExpectations(t)
	defer vpcRepo.AssertExpectations(t)

	ctx := context.Background()
	vpcID := uuid.New()
	vpc := &domain.VPC{
		ID:        vpcID,
		CIDRBlock: testutil.TestCIDR,
	}

	vpcRepo.On("GetByID", ctx, vpcID).Return(vpc, nil)

	subnet, err := svc.CreateSubnet(ctx, vpcID, "bad-subnet", "not-a-cidr", subnetTestAZ)

	assert.Error(t, err)
	assert.Nil(t, subnet)
}

func TestSubnetServiceCreateSubnetInvalidVPCCIDRFormat(t *testing.T) {
	repo, vpcRepo, _, svc := setupSubnetServiceTest(t)
	defer repo.AssertExpectations(t)
	defer vpcRepo.AssertExpectations(t)

	ctx := context.Background()
	vpcID := uuid.New()
	vpc := &domain.VPC{
		ID:        vpcID,
		CIDRBlock: "bad-cidr",
	}

	vpcRepo.On("GetByID", ctx, vpcID).Return(vpc, nil)

	subnet, err := svc.CreateSubnet(ctx, vpcID, "bad-vpc", testutil.TestSubnetCIDR, subnetTestAZ)

	assert.Error(t, err)
	assert.Nil(t, subnet)
}

func TestSubnetServiceDeleteSubnetSuccess(t *testing.T) {
	repo, vpcRepo, auditSvc, svc := setupSubnetServiceTest(t)
	defer repo.AssertExpectations(t)
	defer vpcRepo.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	ctx := context.Background()
	subnetID := uuid.New()
	subnet := &domain.Subnet{ID: subnetID, UserID: uuid.New()}

	repo.On("GetByID", ctx, subnetID).Return(subnet, nil)
	repo.On("Delete", ctx, subnetID).Return(nil)
	auditSvc.On("Log", ctx, subnet.UserID, "subnet.delete", "subnet", subnetID.String(), mock.Anything).Return(nil)

	err := svc.DeleteSubnet(ctx, subnetID)

	assert.NoError(t, err)
}

func TestSubnetServiceDeleteSubnetGetError(t *testing.T) {
	repo, _, _, svc := setupSubnetServiceTest(t)
	defer repo.AssertExpectations(t)

	ctx := context.Background()
	subnetID := uuid.New()

	repo.On("GetByID", ctx, subnetID).Return(nil, assert.AnError)

	err := svc.DeleteSubnet(ctx, subnetID)
	assert.Error(t, err)
}

func TestSubnetServiceDeleteSubnetRepoError(t *testing.T) {
	repo, _, _, svc := setupSubnetServiceTest(t)
	defer repo.AssertExpectations(t)

	ctx := context.Background()
	subnetID := uuid.New()
	subnet := &domain.Subnet{ID: subnetID, UserID: uuid.New()}

	repo.On("GetByID", ctx, subnetID).Return(subnet, nil)
	repo.On("Delete", ctx, subnetID).Return(assert.AnError)

	err := svc.DeleteSubnet(ctx, subnetID)
	assert.Error(t, err)
}

func TestSubnetServiceGetSubnet(t *testing.T) {
	repo, _, _, svc := setupSubnetServiceTest(t)
	defer repo.AssertExpectations(t)

	ctx := context.Background()
	id := uuid.New()
	expected := &domain.Subnet{ID: id, Name: "test"}

	repo.On("GetByID", ctx, id).Return(expected, nil)

	subnet, err := svc.GetSubnet(ctx, id.String(), uuid.Nil)

	assert.NoError(t, err)
	assert.Equal(t, expected, subnet)
}

func TestSubnetServiceGetSubnetByName(t *testing.T) {
	repo, _, _, svc := setupSubnetServiceTest(t)
	defer repo.AssertExpectations(t)

	ctx := context.Background()
	vpcID := uuid.New()
	name := "subnet-name"
	expected := &domain.Subnet{ID: uuid.New(), Name: name}

	repo.On("GetByName", ctx, vpcID, name).Return(expected, nil)

	subnet, err := svc.GetSubnet(ctx, name, vpcID)
	assert.NoError(t, err)
	assert.Equal(t, expected, subnet)
}

func TestSubnetServiceListSubnets(t *testing.T) {
	repo, _, _, svc := setupSubnetServiceTest(t)
	defer repo.AssertExpectations(t)

	ctx := context.Background()
	vpcID := uuid.New()
	expected := []*domain.Subnet{{ID: uuid.New(), Name: "s1"}}

	repo.On("ListByVPC", ctx, vpcID).Return(expected, nil)

	subnets, err := svc.ListSubnets(ctx, vpcID)

	assert.NoError(t, err)
	assert.Equal(t, expected, subnets)
}
