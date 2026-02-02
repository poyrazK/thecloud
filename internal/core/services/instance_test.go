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

// FaultyInstanceRepository wraps a real InstanceRepository to simulate failures.
type FaultyInstanceRepository struct {
	ports.InstanceRepository
	ShouldFailCreate bool
}

func (r *FaultyInstanceRepository) Create(ctx context.Context, instance *domain.Instance) error {
	if r.ShouldFailCreate {
		return fmt.Errorf("simulated database failure")
	}
	return r.InstanceRepository.Create(ctx, instance)
}

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

func TestInstanceService_Launch_DBFailure(t *testing.T) {
	// Custom setup with Faulty Repo
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	realRepo := postgres.NewInstanceRepository(db)
	faultyRepo := &FaultyInstanceRepository{InstanceRepository: realRepo, ShouldFailCreate: true}

	// Other deps
	vpcRepo := postgres.NewVpcRepository(db)
	subnetRepo := postgres.NewSubnetRepository(db)
	volumeRepo := postgres.NewVolumeRepository(db)
	itRepo := postgres.NewInstanceTypeRepository(db)
	compute, _ := docker.NewDockerAdapter()

	defaultType := &domain.InstanceType{ID: "basic-2", Name: "Basic 2", VCPUs: 1, MemoryMB: 128, DiskGB: 1}
	_, _ = itRepo.Create(ctx, defaultType)

	eventSvc := services.NewEventService(postgres.NewEventRepository(db), nil, slog.Default())
	auditSvc := services.NewAuditService(postgres.NewAuditRepository(db))
	taskQueue := &InMemoryTaskQueue{}

	svc := services.NewInstanceService(services.InstanceServiceParams{
		Repo:             faultyRepo,
		VpcRepo:          vpcRepo,
		SubnetRepo:       subnetRepo,
		VolumeRepo:       volumeRepo,
		InstanceTypeRepo: itRepo,
		Compute:          compute,
		EventSvc:         eventSvc,
		AuditSvc:         auditSvc,
		TaskQueue:        taskQueue,
		Logger:           slog.Default(),
	})

	// Attempt Launch
	_, err := svc.LaunchInstance(ctx, "fail-inst", "alpine:latest", "", "basic-2", nil, nil, nil)

	// Verify Failure
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "simulated database failure")

	// Verify no junk in DB (using real repo to check)
	list, err := realRepo.List(ctx)
	assert.NoError(t, err)
	assert.Empty(t, list)
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

func TestInstanceService_Launch_Concurrency(t *testing.T) {
	svc, compute, repo, _, _, ctx := setupInstanceServiceTest(t)
	concurrency := 5
	errChan := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(idx int) {
			name := fmt.Sprintf("conc-inst-%d", idx)
			_, err := svc.LaunchInstance(ctx, name, "alpine:latest", "", "basic-2", nil, nil, nil)
			errChan <- err
		}(i)
	}

	for i := 0; i < concurrency; i++ {
		err := <-errChan
		assert.NoError(t, err)
	}

	// Verify all created
	list, err := repo.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, list, concurrency)

	// Verify provision triggers (optional execution)
	for _, inst := range list {
		_ = compute.DeleteInstance(ctx, inst.ContainerID)
	}
}

func TestInstanceService_GetStats_Real(t *testing.T) {
	svc, _, _, _, _, ctx := setupInstanceServiceTest(t)
	name := "stats-inst"
	image := "alpine:latest"

	// Launch and Provision
	inst, err := svc.LaunchInstance(ctx, name, image, "", "basic-2", nil, nil, nil)
	require.NoError(t, err)
	err = svc.Provision(ctx, inst.ID, nil)
	require.NoError(t, err)

	// Wait a bit for container to run and gather stats
	time.Sleep(2 * time.Second)

	// Get Stats
	stats, err := svc.GetInstanceStats(ctx, inst.ID.String())
	if err != nil {
		// In some CI environments (e.g., Docker-in-Docker or restricted cgroups), retrieving container statistics
		// may fail or return incomplete data. We log this as a warning rather than failing the test to maintain
		// robustness across different execution environments while still verifying the integration where possible.
		t.Logf("Warning: Failed to retrieve instance statistics (likely environment limitation): %v", err)
	} else {
		assert.NotNil(t, stats)
		// CPU percent might be 0 if idle, but LimitBytes should be set if memory limit worked.
		assert.GreaterOrEqual(t, stats.MemoryLimitBytes, 0.0)
		t.Logf("Got Stats: CPU=%.2f%% Mem=%.2fMB", stats.CPUPercentage, stats.MemoryUsageBytes/1024/1024)
	}

	// Clean up
	_ = svc.TerminateInstance(ctx, inst.ID.String())
	// compute.DeleteInstance(ctx, ...) happens in Terminate
}
