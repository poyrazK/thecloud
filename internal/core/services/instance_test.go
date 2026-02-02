package services_test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/docker"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type InMemoryTaskQueue struct {
	jobs []string
}

func (q *InMemoryTaskQueue) Enqueue(ctx context.Context, queueName string, payload interface{}) error {
	q.jobs = append(q.jobs, fmt.Sprintf("%v", payload))
	return nil
}

func (q *InMemoryTaskQueue) Dequeue(ctx context.Context, queueName string) (string, error) {
	if len(q.jobs) > 0 {
		job := q.jobs[0]
		q.jobs = q.jobs[1:]
		return job, nil
	}
	return "", nil
}

func setupInstanceServiceTest(t *testing.T) (*services.InstanceService, *docker.DockerAdapter, ports.InstanceRepository, ports.VpcRepository, ports.VolumeRepository, context.Context) {
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewInstanceRepository(db)
	vpcRepo := postgres.NewVpcRepository(db)
	subnetRepo := postgres.NewSubnetRepository(db)
	volumeRepo := postgres.NewVolumeRepository(db)
	itRepo := postgres.NewInstanceTypeRepository(db)

	compute, err := docker.NewDockerAdapter()
	require.NoError(t, err)

	// Ensure default instance type exists
	defaultType := &domain.InstanceType{
		ID:       "basic-2",
		Name:     "Basic 2",
		VCPUs:    1,
		MemoryMB: 128,
		DiskGB:   1,
	}
	_, _ = itRepo.Create(ctx, defaultType)

	eventRepo := postgres.NewEventRepository(db)
	eventSvc := services.NewEventService(eventRepo, nil, slog.Default())

	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(auditRepo)

	taskQueue := &InMemoryTaskQueue{}

	svc := services.NewInstanceService(services.InstanceServiceParams{
		Repo:             repo,
		VpcRepo:          vpcRepo,
		SubnetRepo:       subnetRepo,
		VolumeRepo:       volumeRepo,
		InstanceTypeRepo: itRepo,
		Compute:          compute,
		Network:          nil,
		EventSvc:         eventSvc,
		AuditSvc:         auditSvc,
		TaskQueue:        taskQueue,
		Logger:           slog.Default(),
	})

	return svc, compute, repo, vpcRepo, volumeRepo, ctx
}

func TestLaunchInstanceSuccess(t *testing.T) {
	svc, compute, repo, _, _, ctx := setupInstanceServiceTest(t)
	name := "test-inst-launch"
	image := "alpine:latest"
	ports := "8080:80"

	// 1. Launch (Enqueue)
	inst, err := svc.LaunchInstance(ctx, name, image, ports, "basic-2", nil, nil, nil)
	require.NoError(t, err)
	assert.NotNil(t, inst)
	assert.Equal(t, domain.StatusStarting, inst.Status)

	// 2. Provision (Simulate Worker)
	err = svc.Provision(ctx, inst.ID, nil)
	require.NoError(t, err)

	// 3. Verify Running
	updatedInst, err := repo.GetByID(ctx, inst.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusRunning, updatedInst.Status)
	assert.NotEmpty(t, updatedInst.ContainerID)

	// 4. Verify connectivity
	ip, err := compute.GetInstanceIP(ctx, updatedInst.ContainerID)
	assert.NoError(t, err)
	assert.NotEmpty(t, ip)

	// Cleanup
	_ = compute.DeleteInstance(ctx, updatedInst.ContainerID)
}

func TestTerminateInstanceSuccess(t *testing.T) {
	svc, compute, repo, _, _, ctx := setupInstanceServiceTest(t)
	name := "test-inst-term"
	image := "alpine:latest"

	// Setup: Launch & Provision
	inst, err := svc.LaunchInstance(ctx, name, image, "", "basic-2", nil, nil, nil)
	require.NoError(t, err)
	err = svc.Provision(ctx, inst.ID, nil)
	require.NoError(t, err)

	updatedInst, _ := repo.GetByID(ctx, inst.ID)
	require.NotEmpty(t, updatedInst.ContainerID)

	// Execute Terminate
	err = svc.TerminateInstance(ctx, updatedInst.ID.String())
	assert.NoError(t, err)

	// Verify Deleted from DB
	_, err = repo.GetByID(ctx, updatedInst.ID)
	assert.Error(t, err)

	// Verify container is gone
	_, err = compute.GetInstanceIP(ctx, updatedInst.ContainerID)
	assert.Error(t, err)
}

func TestInstanceNetworking(t *testing.T) {
	svc, compute, _, vpcRepo, _, ctx := setupInstanceServiceTest(t)

	vpcID := uuid.New()
	networkName := "net-" + vpcID.String()
	netID, err := compute.CreateNetwork(ctx, networkName)
	require.NoError(t, err)
	defer compute.DeleteNetwork(ctx, netID)

	vpc := &domain.VPC{
		ID:        vpcID,
		UserID:    appcontext.UserIDFromContext(ctx),
		TenantID:  appcontext.TenantIDFromContext(ctx),
		Name:      "test-vpc",
		CIDRBlock: "10.0.0.0/16",
		NetworkID: netID,
		Status:    "ACTIVE",
		CreatedAt: time.Now(),
	}
	err = vpcRepo.Create(ctx, vpc)
	require.NoError(t, err)

	// Use Launch with VPC
	inst, err := svc.LaunchInstance(ctx, "test-vpc-inst", "alpine:latest", "", "basic-2", &vpcID, nil, nil)
	require.NoError(t, err)

	err = svc.Provision(ctx, inst.ID, nil)
	require.NoError(t, err)

	// Check if attached to network
	updatedInst, err := svc.GetInstance(ctx, inst.ID.String())
	require.NoError(t, err)
	assert.NotNil(t, updatedInst.VpcID)
	assert.Equal(t, vpcID, *updatedInst.VpcID)

	// Cleanup
	if updatedInst.ContainerID != "" {
		_ = compute.DeleteInstance(ctx, updatedInst.ContainerID)
	}
}
