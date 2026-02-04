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
	"github.com/poyrazk/thecloud/internal/repositories/ovs"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSecurityGroupServiceIntegrationTest(t *testing.T) (ports.SecurityGroupService, ports.SecurityGroupRepository, ports.VpcRepository, context.Context) {
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewSecurityGroupRepository(db)
	vpcRepo := postgres.NewVpcRepository(db)
	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(auditRepo)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	network, err := ovs.NewOvsAdapter(logger)
	if err != nil {
		t.Skipf("Skipping OVS integration: %v", err)
	}

	svc := services.NewSecurityGroupService(repo, vpcRepo, network, auditSvc, logger)

	return svc, repo, vpcRepo, ctx
}

func TestSecurityGroupService_Integration(t *testing.T) {
	svc, _, vpcRepo, ctx := setupSecurityGroupServiceIntegrationTest(t)
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	// Setup VPC
	vpc := &domain.VPC{
		ID:        uuid.New(),
		UserID:    userID,
		TenantID:  tenantID,
		Name:      "sg-vpc",
		NetworkID: "br-int", // Use a real or mock-ish bridge name that OVS might have
	}
	// In a full end-to-end integration, the OVS bridge would ideally be pre-provisioned.
	// For this test scope, we focus on verifying database application logic, accepting
	// that OVS operations may fail if the underlying network infrastructure is absent.

	err := vpcRepo.Create(ctx, vpc)
	require.NoError(t, err)

	t.Run("GroupLifecycle", func(t *testing.T) {
		name := "web-sg"
		sg, err := svc.CreateGroup(ctx, vpc.ID, name, "web servers")
		assert.NoError(t, err)
		assert.Equal(t, name, sg.Name)

		// Get
		fetched, err := svc.GetGroup(ctx, sg.ID.String(), vpc.ID)
		assert.NoError(t, err)
		assert.Equal(t, sg.ID, fetched.ID)

		// List
		groups, err := svc.ListGroups(ctx, vpc.ID)
		assert.NoError(t, err)
		assert.Len(t, groups, 1)

		// Delete
		err = svc.DeleteGroup(ctx, sg.ID)
		assert.NoError(t, err)

		_, err = svc.GetGroup(ctx, sg.ID.String(), vpc.ID)
		assert.Error(t, err)
	})

	t.Run("Rules", func(t *testing.T) {
		sg, _ := svc.CreateGroup(ctx, vpc.ID, "rule-sg", "")

		rule := domain.SecurityRule{
			Protocol:  "tcp",
			PortMin:   80,
			PortMax:   80,
			CIDR:      "0.0.0.0/0",
			Direction: domain.RuleIngress,
		}

		// Attempt to add the rule. Note that in environments without the OVS bridge 'br-int',
		// the network layer operation is expected to fail. We handle this typically by
		// asserting specifically on network errors when running in restricted environments.
		res, err := svc.AddRule(ctx, sg.ID, rule)

		if err != nil && (assert.Contains(t, err.Error(), "failed to add flow rule") || assert.Contains(t, err.Error(), "failed to add bridge")) {
			t.Logf("OVS failed as expected in restricted env: %v", err)
			return
		}

		assert.NoError(t, err)
		assert.NotNil(t, res)

		// Remove rule
		err = svc.RemoveRule(ctx, res.ID)
		assert.NoError(t, err)
	})
}
