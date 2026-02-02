package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAutoScalingServiceIntegrationTest(t *testing.T) (ports.AutoScalingService, ports.VpcRepository, context.Context) {
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewAutoScalingRepo(db)
	vpcRepo := postgres.NewVpcRepository(db)
	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(auditRepo)

	svc := services.NewAutoScalingService(repo, vpcRepo, auditSvc)

	return svc, vpcRepo, ctx
}

func TestAutoScalingService_Integration(t *testing.T) {
	svc, vpcRepo, ctx := setupAutoScalingServiceIntegrationTest(t)
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	// Setup VPC
	vpc := &domain.VPC{
		ID:        uuid.New(),
		UserID:    userID,
		TenantID:  tenantID,
		Name:      "asg-vpc",
		NetworkID: "br-int",
		Status:    "ACTIVE",
		CreatedAt: time.Now(),
	}
	err := vpcRepo.Create(ctx, vpc)
	require.NoError(t, err)

	t.Run("CreateGroup_Success", func(t *testing.T) {
		name := "my-asg"
		group, err := svc.CreateGroup(ctx, ports.CreateScalingGroupParams{
			Name:         name,
			VpcID:        vpc.ID,
			Image:        "nginx",
			Ports:        "80:80",
			MinInstances: 1,
			MaxInstances: 5,
			DesiredCount: 2,
		})

		assert.NoError(t, err)
		assert.NotNil(t, group)
		assert.Equal(t, name, group.Name)
		assert.Equal(t, 2, group.DesiredCount)

		// List
		groups, err := svc.ListGroups(ctx)
		assert.NoError(t, err)
		assert.Len(t, groups, 1)
	})

	t.Run("PolicyLifecycle", func(t *testing.T) {
		group, _ := svc.CreateGroup(ctx, ports.CreateScalingGroupParams{
			Name:         "policy-asg",
			VpcID:        vpc.ID,
			Image:        "nginx",
			MinInstances: 1,
			MaxInstances: 5,
			DesiredCount: 1,
		})

		// Create
		policy, err := svc.CreatePolicy(ctx, ports.CreateScalingPolicyParams{
			GroupID:     group.ID,
			Name:        "cpu-high",
			MetricType:  "cpu",
			TargetValue: 70.0,
			ScaleOut:    1,
			ScaleIn:     1,
			CooldownSec: 300,
		})
		assert.NoError(t, err)
		assert.NotNil(t, policy)

		// Delete
		err = svc.DeletePolicy(ctx, policy.ID)
		assert.NoError(t, err)
	})

	t.Run("CapacityUpdate", func(t *testing.T) {
		group, _ := svc.CreateGroup(ctx, ports.CreateScalingGroupParams{
			Name:         "cap-asg",
			VpcID:        vpc.ID,
			Image:        "nginx",
			MinInstances: 1,
			MaxInstances: 10,
			DesiredCount: 1,
		})

		err := svc.SetDesiredCapacity(ctx, group.ID, 5)
		assert.NoError(t, err)

		updated, _ := svc.GetGroup(ctx, group.ID)
		assert.Equal(t, 5, updated.DesiredCount)
	})

	t.Run("DeleteGroup", func(t *testing.T) {
		group, _ := svc.CreateGroup(ctx, ports.CreateScalingGroupParams{
			Name:         "del-asg",
			VpcID:        vpc.ID,
			Image:        "nginx",
			MinInstances: 1,
			MaxInstances: 5,
			DesiredCount: 1,
		})

		err := svc.DeleteGroup(ctx, group.ID)
		assert.NoError(t, err)

		updated, _ := svc.GetGroup(ctx, group.ID)
		assert.Equal(t, domain.ScalingGroupStatusDeleting, updated.Status)
		assert.Equal(t, 0, updated.DesiredCount)
	})

	t.Run("Validation", func(t *testing.T) {
		// Min > Max
		_, err := svc.CreateGroup(ctx, ports.CreateScalingGroupParams{
			Name:         "fail-asg",
			VpcID:        vpc.ID,
			Image:        "nginx",
			MinInstances: 5,
			MaxInstances: 2,
			DesiredCount: 3,
		})
		assert.Error(t, err)

		// Desired < Min
		_, err = svc.CreateGroup(ctx, ports.CreateScalingGroupParams{
			Name:         "fail-asg-2",
			VpcID:        vpc.ID,
			Image:        "nginx",
			MinInstances: 2,
			MaxInstances: 5,
			DesiredCount: 1,
		})
		assert.Error(t, err)
	})
}
