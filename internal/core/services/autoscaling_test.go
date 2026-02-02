package services_test

import (
	"context"
	"log/slog"
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

type RealClock struct{}

func (c *RealClock) Now() time.Time { return time.Now() }

type NoopLBService struct{}

func (s *NoopLBService) Create(ctx context.Context, name string, vpcID uuid.UUID, port int, protocol, algo string) (*domain.LoadBalancer, error) {
	return nil, nil
}
func (s *NoopLBService) Get(ctx context.Context, id uuid.UUID) (*domain.LoadBalancer, error) {
	return nil, nil
}
func (s *NoopLBService) AddTarget(ctx context.Context, lbID, instanceID uuid.UUID, port, weight int) error {
	return nil
}
func (s *NoopLBService) RemoveTarget(ctx context.Context, lbID, instanceID uuid.UUID) error {
	return nil
}
func (s *NoopLBService) ListTargets(ctx context.Context, lbID uuid.UUID) ([]*domain.LBTarget, error) {
	return nil, nil
}
func (s *NoopLBService) List(ctx context.Context) ([]*domain.LoadBalancer, error) { return nil, nil }
func (s *NoopLBService) Delete(ctx context.Context, id uuid.UUID) error           { return nil }
func (s *NoopLBService) CreateListener(ctx context.Context, lbID uuid.UUID, port int, protocol string) error {
	return nil
}

func TestAutoScaling_TriggerScaleUp(t *testing.T) {
	// 1. Setup Infra using InstanceService helper (provides Docker + Postgres)
	db, instSvc, compute, instRepo, vpcRepo, _, ctx := setupInstanceServiceTest(t)
	// db := setupDB(t) // Reuse db from helper

	asgRepo := postgres.NewAutoScalingRepo(db)
	auditSvc := services.NewAuditService(postgres.NewAuditRepository(db))
	eventSvc := services.NewEventService(postgres.NewEventRepository(db), nil, slog.Default())

	worker := services.NewAutoScalingWorker(asgRepo, instSvc, &NoopLBService{}, eventSvc, &RealClock{})

	// 2. Setup Resources (VPC, Image)
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	vpc := &domain.VPC{
		ID:        uuid.New(),
		UserID:    userID,
		TenantID:  tenantID,
		Name:      "asg-scale-vpc",
		NetworkID: "br-test", // Use existing if needed, but docker adapter might ignore
		Status:    "ACTIVE",
		CreatedAt: time.Now(),
	}
	require.NoError(t, vpcRepo.Create(ctx, vpc))

	// 3. Create Scaling Group
	groupName := "scale-out-test"
	asgSvc := services.NewAutoScalingService(asgRepo, vpcRepo, auditSvc)
	group, err := asgSvc.CreateGroup(ctx, ports.CreateScalingGroupParams{
		Name:         groupName,
		VpcID:        vpc.ID,
		Image:        "alpine:latest",
		MinInstances: 1,
		MaxInstances: 5,
		DesiredCount: 1,
	})
	require.NoError(t, err)

	// 4. Initial Reconcile (Should create 1 instance)
	worker.Evaluate(ctx)

	// Verify 1 instance exists
	instances, err := instRepo.List(ctx)
	require.NoError(t, err)
	require.Len(t, instances, 1)
	instID := instances[0].ID

	// Wait for instance to be running (provision job)
	// We manually provision it since we don't have the task worker running
	err = instSvc.Provision(ctx, instID, nil)
	require.NoError(t, err)

	// 5. Create Scaling Policy (CPU > 50%)
	_, err = asgSvc.CreatePolicy(ctx, ports.CreateScalingPolicyParams{
		GroupID:     group.ID,
		Name:        "cpu-high",
		MetricType:  "cpu",
		TargetValue: 50.0,
		ScaleOut:    1,
		ScaleIn:     1,
		CooldownSec: 60,
	})
	require.NoError(t, err)

	// 6. Inject Fake Metrics
	// We directly insert into metrics_history to simulate high load
	_, err = db.Exec(ctx, `
		INSERT INTO metrics_history (id, instance_id, cpu_percent, memory_bytes, recorded_at)
		VALUES ($1, $2, $3, $4, $5)
	`, uuid.New(), instID, 85.0, 1024, time.Now())
	require.NoError(t, err)

	// 7. Evaluate Again (Should Trigger Scale Out)
	worker.Evaluate(ctx)

	// 8. Verify Desired Count Increased
	updatedGroup, err := asgSvc.GetGroup(ctx, group.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, updatedGroup.DesiredCount, "Desired count should update to 2")

	// 9. Reconcile Again (Should create 2nd instance)
	worker.Evaluate(ctx)

	instances, err = instRepo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, instances, 2, "Should have 2 instances")

	// Cleanup
	for _, inst := range instances {
		if inst.ContainerID != "" {
			_ = compute.DeleteInstance(ctx, inst.ContainerID)
		}
	}
}
