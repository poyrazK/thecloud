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
	"github.com/poyrazk/thecloud/internal/repositories/noop"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupVpcServiceTest(t *testing.T, cidr string) (*services.VpcService, *postgres.VpcRepository, *postgres.LBRepository, context.Context) {
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	vpcRepo := postgres.NewVpcRepository(db)
	lbRepo := postgres.NewLBRepository(db)

	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(auditRepo)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	network := noop.NewNoopNetworkAdapter(logger)

	svc := services.NewVpcService(vpcRepo, lbRepo, network, auditSvc, logger, cidr)
	return svc, vpcRepo, lbRepo, ctx
}

func TestVpcServiceCreateSuccess(t *testing.T) {
	svc, repo, _, ctx := setupVpcServiceTest(t, testutil.TestCIDR)
	name := "test-vpc-" + uuid.New().String()
	cidr := testutil.TestCIDR

	vpc, err := svc.CreateVPC(ctx, name, cidr)
	require.NoError(t, err)
	require.NotNil(t, vpc)
	assert.Equal(t, name, vpc.Name)
	assert.Equal(t, cidr, vpc.CIDRBlock)
	assert.Contains(t, vpc.NetworkID, "br-vpc-")

	// Verify in DB
	fetched, err := repo.GetByID(ctx, vpc.ID)
	assert.NoError(t, err)
	assert.Equal(t, vpc.ID, fetched.ID)
}

func TestVpcServiceCreateDBFailureRollback(t *testing.T) {
	// Induced failure via context cancellation
	svc, _, _, ctx := setupVpcServiceTest(t, testutil.TestCIDR)
	cancelledCtx, cancel := context.WithCancel(ctx)
	cancel()

	vpc, err := svc.CreateVPC(cancelledCtx, "fail-vpc", "")
	assert.Error(t, err)
	assert.Nil(t, vpc)
}

func TestVpcServiceCreateDefaultCIDR(t *testing.T) {
	svc, _, _, ctx := setupVpcServiceTest(t, "")

	vpc, err := svc.CreateVPC(ctx, "default-cidr-"+uuid.New().String(), "")

	assert.NoError(t, err)
	assert.NotNil(t, vpc)
	assert.Equal(t, "10.0.0.0/16", vpc.CIDRBlock)
}

func TestVpcServiceDeleteSuccess(t *testing.T) {
	svc, repo, _, ctx := setupVpcServiceTest(t, testutil.TestCIDR)
	vpc, err := svc.CreateVPC(ctx, "to-delete-"+uuid.New().String(), testutil.TestCIDR)
	require.NoError(t, err)

	err = svc.DeleteVPC(ctx, vpc.ID.String())
	assert.NoError(t, err)

	// Verify Deleted from DB
	_, err = repo.GetByID(ctx, vpc.ID)
	assert.Error(t, err)
}

func TestVpcServiceDeleteFailureWithLBs(t *testing.T) {
	svc, _, lbRepo, ctx := setupVpcServiceTest(t, testutil.TestCIDR)
	vpc, err := svc.CreateVPC(ctx, "in-use-"+uuid.New().String(), testutil.TestCIDR)
	require.NoError(t, err)

	// Add a Load Balancer to this VPC
	lb := &domain.LoadBalancer{
		ID:     uuid.New(),
		UserID: appcontext.UserIDFromContext(ctx),
		Name:   "test-lb",
		VpcID:  vpc.ID,
		Status: domain.LBStatusActive,
	}
	err = lbRepo.Create(ctx, lb)
	require.NoError(t, err)

	err = svc.DeleteVPC(ctx, vpc.ID.String())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "load balancers still exist")
}

func TestVpcServiceListSuccess(t *testing.T) {
	svc, _, _, ctx := setupVpcServiceTest(t, testutil.TestCIDR)
	_, _ = svc.CreateVPC(ctx, "vpc1-"+uuid.New().String(), testutil.TestCIDR)
	_, _ = svc.CreateVPC(ctx, "vpc2-"+uuid.New().String(), testutil.TestCIDR)

	result, err := svc.ListVPCs(ctx)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestVpcServiceGetByName(t *testing.T) {
	svc, _, _, ctx := setupVpcServiceTest(t, testutil.TestCIDR)
	name := "my-vpc-" + uuid.New().String()
	vpc, _ := svc.CreateVPC(ctx, name, testutil.TestCIDR)

	result, err := svc.GetVPC(ctx, name)

	assert.NoError(t, err)
	assert.Equal(t, vpc.ID, result.ID)
}
