package services_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSubnetServiceTest(t *testing.T) (*services.SubnetService, *postgres.SubnetRepository, *postgres.VpcRepository, context.Context) {
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewSubnetRepository(db)
	vpcRepo := postgres.NewVpcRepository(db)

	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(auditRepo)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := services.NewSubnetService(repo, vpcRepo, auditSvc, logger)
	return svc, repo, vpcRepo, ctx
}

func createTestVPC(t *testing.T, ctx context.Context, vpcRepo *postgres.VpcRepository) *domain.VPC {
	tenantID := appcontext.TenantIDFromContext(ctx)
	userID := appcontext.UserIDFromContext(ctx)
	vpc := &domain.VPC{
		ID:        uuid.New(),
		UserID:    userID,
		TenantID:  tenantID,
		Name:      "test-vpc-" + uuid.New().String(),
		CIDRBlock: testutil.TestCIDR,
		Status:    "available",
	}
	err := vpcRepo.Create(ctx, vpc)
	require.NoError(t, err)
	return vpc
}

func TestSubnetServiceCreateSubnetSuccess(t *testing.T) {
	svc, repo, vpcRepo, ctx := setupSubnetServiceTest(t)
	vpc := createTestVPC(t, ctx, vpcRepo)

	subnet, err := svc.CreateSubnet(ctx, vpc.ID, "test-subnet", testutil.TestSubnetCIDR, "us-east-1a")

	assert.NoError(t, err)
	assert.NotNil(t, subnet)
	assert.Equal(t, testutil.TestGatewayIP, subnet.GatewayIP)

	// Verify in DB
	fetched, err := repo.GetByID(ctx, subnet.ID)
	assert.NoError(t, err)
	assert.Equal(t, subnet.ID, fetched.ID)
}

func TestSubnetServiceCreateSubnetInvalidCIDR(t *testing.T) {
	svc, _, vpcRepo, ctx := setupSubnetServiceTest(t)
	vpc := createTestVPC(t, ctx, vpcRepo)

	// Outside VPC range (e.g. TestOtherCIDR)
	subnet, err := svc.CreateSubnet(ctx, vpc.ID, "bad-subnet", testutil.TestOtherCIDR, "us-east-1a")

	assert.Error(t, err)
	assert.Nil(t, subnet)
	assert.Contains(t, err.Error(), "within VPC CIDR range")
}

func TestSubnetServiceDeleteSubnetSuccess(t *testing.T) {
	svc, repo, vpcRepo, ctx := setupSubnetServiceTest(t)
	vpc := createTestVPC(t, ctx, vpcRepo)

	subnet, err := svc.CreateSubnet(ctx, vpc.ID, "to-delete", testutil.TestSubnetCIDR, "az")
	require.NoError(t, err)
	require.NotNil(t, subnet)

	err = svc.DeleteSubnet(ctx, subnet.ID)
	assert.NoError(t, err)

	// Verify Deleted from DB
	_, err = repo.GetByID(ctx, subnet.ID)
	assert.Error(t, err)
}

func TestSubnetServiceGetSubnet(t *testing.T) {
	svc, _, vpcRepo, ctx := setupSubnetServiceTest(t)
	vpc := createTestVPC(t, ctx, vpcRepo)

	subnet, err := svc.CreateSubnet(ctx, vpc.ID, "find-me", testutil.TestSubnetCIDR, "az")
	require.NoError(t, err)

	t.Run("get by id", func(t *testing.T) {
		res, err := svc.GetSubnet(ctx, subnet.ID.String(), uuid.Nil)
		assert.NoError(t, err)
		assert.Equal(t, subnet.ID, res.ID)
	})

	t.Run("get by name", func(t *testing.T) {
		res, err := svc.GetSubnet(ctx, "find-me", vpc.ID)
		assert.NoError(t, err)
		assert.Equal(t, subnet.ID, res.ID)
	})
}

func TestSubnetServiceListSubnets(t *testing.T) {
	svc, _, vpcRepo, ctx := setupSubnetServiceTest(t)
	vpc := createTestVPC(t, ctx, vpcRepo)

	_, err := svc.CreateSubnet(ctx, vpc.ID, "s1", "10.0.1.0/24", "az")
	require.NoError(t, err)
	_, err = svc.CreateSubnet(ctx, vpc.ID, "s2", "10.0.2.0/24", "az")
	require.NoError(t, err)

	subnets, err := svc.ListSubnets(ctx, vpc.ID)

	assert.NoError(t, err)
	assert.Len(t, subnets, 2)
}
