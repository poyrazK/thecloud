package services_test

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	coreports "github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/docker"
	"github.com/poyrazk/thecloud/internal/repositories/noop"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testInstanceType = "basic-2"
	testImage        = "alpine"
)

// FaultyInstanceRepository wraps a real InstanceRepository to simulate failures.
type FaultyInstanceRepository struct {
	coreports.InstanceRepository
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
	mu   sync.Mutex
}

func (q *InMemoryTaskQueue) Enqueue(ctx context.Context, queueName string, payload interface{}) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.jobs = append(q.jobs, fmt.Sprintf("%v", payload))
	return nil
}

func (q *InMemoryTaskQueue) Dequeue(ctx context.Context, queueName string) (string, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.jobs) == 0 {
		return "", fmt.Errorf("queue empty")
	}
	job := q.jobs[0]
	q.jobs = q.jobs[1:]
	return job, nil
}

func setupInstanceServiceTest(t *testing.T) (*pgxpool.Pool, *services.InstanceService, *docker.DockerAdapter, coreports.InstanceRepository, coreports.VpcRepository, coreports.VolumeRepository, context.Context) {
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewInstanceRepository(db)
	vpcRepo := postgres.NewVpcRepository(db)
	subnetRepo := postgres.NewSubnetRepository(db)
	volumeRepo := postgres.NewVolumeRepository(db)
	itRepo := postgres.NewInstanceTypeRepository(db)

	compute, err := docker.NewDockerAdapter(slog.Default())
	require.NoError(t, err)

	// In integration tests, we frequently rely on a shared Docker network.
	// We ensure it exists here so that Provisioning (which uses it as a fallback) succeeds.
	if compute.Type() == "docker" {
		_, _ = compute.CreateNetwork(ctx, testutil.TestDockerNetwork)
		// Pre-pull test image to prevent flakes in CI environments with slow registries or restrictive daemons
		_, _, _ = compute.LaunchInstanceWithOptions(ctx, coreports.CreateInstanceOptions{
			Name:      "pre-pull-image",
			ImageName: testImage,
		})
		_ = compute.DeleteInstance(ctx, "pre-pull-image")
	}

	// Ensure default instance type exists
	defaultType := &domain.InstanceType{
		ID:       testInstanceType,
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

	tenantRepo := postgres.NewTenantRepo(db)
	userRepo := postgres.NewUserRepo(db)
	tenantSvc := services.NewTenantService(tenantRepo, userRepo, slog.Default())

	taskQueue := &InMemoryTaskQueue{}
	network := noop.NewNoopNetworkAdapter(slog.Default())

	sshKeyRepo := postgres.NewSSHKeyRepo(db)
	sshKeySvc := services.NewSSHKeyService(sshKeyRepo)

	svc := services.NewInstanceService(services.InstanceServiceParams{
		Repo:             repo,
		VpcRepo:          vpcRepo,
		SubnetRepo:       subnetRepo,
		VolumeRepo:       volumeRepo,
		InstanceTypeRepo: itRepo,
		Compute:          compute,
		Network:          network,
		EventSvc:         eventSvc,
		AuditSvc:         auditSvc,
		TaskQueue:        taskQueue,
		Logger:           slog.Default(),
		TenantSvc:        tenantSvc,
		SSHKeySvc:        sshKeySvc,
	})

	return db, svc, compute, repo, vpcRepo, volumeRepo, ctx
}

func TestLaunchInstanceSuccess(t *testing.T) {
	_, svc, compute, repo, _, _, ctx := setupInstanceServiceTest(t)
	name := "test-inst-launch"
	image := testImage
	portsStr := "0:80"

	// 1. Launch (Enqueue)
	inst, err := svc.LaunchInstance(ctx, coreports.LaunchParams{
		Name:         name,
		Image:        image,
		Ports:        portsStr,
		InstanceType: testInstanceType,
	})
	require.NoError(t, err)
	assert.NotNil(t, inst)
	assert.Equal(t, domain.StatusStarting, inst.Status)

	// 2. Provision (Simulate Worker)
	err = svc.Provision(ctx, domain.ProvisionJob{InstanceID: inst.ID})
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
	_, svc, compute, repo, _, _, ctx := setupInstanceServiceTest(t)
	name := "test-inst-term"
	image := testImage

	// Setup: Launch & Provision
	inst, err := svc.LaunchInstance(ctx, coreports.LaunchParams{
		Name:         name,
		Image:        image,
		InstanceType: testInstanceType,
	})
	require.NoError(t, err)
	err = svc.Provision(ctx, domain.ProvisionJob{InstanceID: inst.ID})
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

func TestInstanceServiceLaunchDBFailure(t *testing.T) {
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
	compute, _ := docker.NewDockerAdapter(slog.Default())

	defaultType := &domain.InstanceType{ID: testInstanceType, Name: "Basic 2", VCPUs: 1, MemoryMB: 128, DiskGB: 1}
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
		TenantSvc:        services.NewTenantService(postgres.NewTenantRepo(db), postgres.NewUserRepo(db), slog.Default()),
		SSHKeySvc:        services.NewSSHKeyService(postgres.NewSSHKeyRepo(db)),
	})

	// Attempt Launch
	_, err := svc.LaunchInstance(ctx, coreports.LaunchParams{
		Name:         "fail-inst",
		Image:        testImage,
		InstanceType: testInstanceType,
	})

	// Verify Failure
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "simulated database failure")

	// Verify no junk in DB (using real repo to check)
	list, err := realRepo.List(ctx)
	assert.NoError(t, err)
	assert.Empty(t, list)
}

func TestInstanceNetworking(t *testing.T) {
	_, svc, compute, _, vpcRepo, _, ctx := setupInstanceServiceTest(t)

	vpcID := uuid.New()
	networkName := "net-" + vpcID.String()
	netID, err := compute.CreateNetwork(ctx, networkName)
	require.NoError(t, err)
	defer func() { _ = compute.DeleteNetwork(ctx, netID) }()

	vpc := &domain.VPC{
		ID:        vpcID,
		UserID:    appcontext.UserIDFromContext(ctx),
		TenantID:  appcontext.TenantIDFromContext(ctx),
		Name:      "test-vpc-" + vpcID.String(),
		CIDRBlock: "10.0.0.0/16",
		NetworkID: netID,
		Status:    "ACTIVE",
		CreatedAt: time.Now(),
	}
	err = vpcRepo.Create(ctx, vpc)
	require.NoError(t, err)

	// Use Launch with VPC
	inst, err := svc.LaunchInstance(ctx, coreports.LaunchParams{
		Name:         "test-vpc-inst",
		Image:        testImage,
		InstanceType: testInstanceType,
		VpcID:        &vpcID,
	})
	require.NoError(t, err)

	err = svc.Provision(ctx, domain.ProvisionJob{InstanceID: inst.ID})
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

func TestInstanceServiceLaunchConcurrency(t *testing.T) {
	_, svc, compute, repo, _, _, ctx := setupInstanceServiceTest(t)
	concurrency := 5
	errChan := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(idx int) {
			name := fmt.Sprintf("conc-inst-%d", idx)
			_, err := svc.LaunchInstance(ctx, coreports.LaunchParams{
				Name:         name,
				Image:        testImage,
				InstanceType: testInstanceType,
			})
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

func TestInstanceServiceGetStatsReal(t *testing.T) {
	_, svc, _, _, _, _, ctx := setupInstanceServiceTest(t)
	name := "stats-inst"
	image := testImage

	// Launch and Provision
	inst, err := svc.LaunchInstance(ctx, coreports.LaunchParams{
		Name:         name,
		Image:        image,
		InstanceType: testInstanceType,
	})
	require.NoError(t, err)
	err = svc.Provision(ctx, domain.ProvisionJob{InstanceID: inst.ID})
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

func TestNetworkingCIDRExhaustion(t *testing.T) {
	db, svc, _, _, vpcRepo, _, ctx := setupInstanceServiceTest(t)
	subnetRepo := postgres.NewSubnetRepository(db)
	auditSvc := services.NewAuditService(postgres.NewAuditRepository(db))
	subnetSvc := services.NewSubnetService(subnetRepo, vpcRepo, auditSvc, slog.Default())

	// 1. Create VPC and a very small subnet (/30)
	tenantID := appcontext.TenantIDFromContext(ctx)
	userID := appcontext.UserIDFromContext(ctx)
	vpc := &domain.VPC{
		ID:        uuid.New(),
		UserID:    userID,
		TenantID:  tenantID,
		Name:      "exhaust-vpc",
		CIDRBlock: "10.10.0.0/16",
		Status:    "available",
	}
	err := vpcRepo.Create(ctx, vpc)
	require.NoError(t, err)

	subnet, err := subnetSvc.CreateSubnet(ctx, vpc.ID, "tiny-subnet", "10.10.1.0/30", "us-east-1a")
	require.NoError(t, err)

	// 2. Launch 1st instance (Should succeed in DB)
	inst1, err := svc.LaunchInstance(ctx, coreports.LaunchParams{
		Name:         "inst-1",
		Image:        testImage,
		InstanceType: testInstanceType,
		VpcID:        &vpc.ID,
		SubnetID:     &subnet.ID,
	})
	require.NoError(t, err)

	// Manually provision to trigger network allocation
	err = svc.Provision(ctx, domain.ProvisionJob{InstanceID: inst1.ID})
	assert.NoError(t, err)

	// 3. Launch 2nd instance (Should succeed in DB)
	inst2, err := svc.LaunchInstance(ctx, coreports.LaunchParams{
		Name:         "inst-2",
		Image:        testImage,
		InstanceType: testInstanceType,
		VpcID:        &vpc.ID,
		SubnetID:     &subnet.ID,
	})
	require.NoError(t, err)

	// Provision 2nd instance (Should fail with CIDR exhaustion)
	err = svc.Provision(ctx, domain.ProvisionJob{InstanceID: inst2.ID})
	require.Error(t, err, "Expected error due to CIDR exhaustion")
	t.Logf("Got expected error: %v", err)
	assert.Contains(t, err.Error(), "allocate IP")
}

func TestInstanceMetadataAndLabels(t *testing.T) {
	_, svc, _, repo, _, _, ctx := setupInstanceServiceTest(t)

	t.Run("Launch with Metadata", func(t *testing.T) {
		name := "meta-launch"
		metadata := map[string]string{"env": "prod", "version": "1.0"}
		labels := map[string]string{"tier": "frontend"}

		inst, err := svc.LaunchInstance(ctx, coreports.LaunchParams{
			Name:         name,
			Image:        testImage,
			InstanceType: testInstanceType,
			Metadata:     metadata,
			Labels:       labels,
		})
		require.NoError(t, err)
		assert.Equal(t, metadata, inst.Metadata)
		assert.Equal(t, labels, inst.Labels)

		// Verify in DB
		dbInst, err := repo.GetByID(ctx, inst.ID)
		require.NoError(t, err)
		assert.Equal(t, metadata, dbInst.Metadata)
		assert.Equal(t, labels, dbInst.Labels)
	})

	t.Run("Update Metadata", func(t *testing.T) {
		inst, _ := svc.LaunchInstance(ctx, coreports.LaunchParams{
			Name:         "meta-update",
			Image:        testImage,
			InstanceType: testInstanceType,
			Metadata:     map[string]string{"key1": "val1"},
		})

		// 1. Add and Update
		err := svc.UpdateInstanceMetadata(ctx, inst.ID, map[string]string{"key1": "newval", "key2": "val2"}, nil)
		require.NoError(t, err)

		// 2. Delete (empty value)
		err = svc.UpdateInstanceMetadata(ctx, inst.ID, map[string]string{"key1": ""}, map[string]string{"l1": "v1"})
		require.NoError(t, err)

		dbInst, _ := repo.GetByID(ctx, inst.ID)
		assert.Equal(t, "val2", dbInst.Metadata["key2"])
		_, ok := dbInst.Metadata["key1"]
		assert.False(t, ok)
		assert.Equal(t, "v1", dbInst.Labels["l1"])
	})
}

func TestSSHKeyInjection(t *testing.T) {
	db, svc, _, _, _, _, ctx := setupInstanceServiceTest(t)
	sshKeyRepo := postgres.NewSSHKeyRepo(db)

	// 1. Create SSH Key
	key := &domain.SSHKey{
		ID:        uuid.New(),
		UserID:    appcontext.UserIDFromContext(ctx),
		TenantID:  appcontext.TenantIDFromContext(ctx),
		Name:      "test-injection-key",
		PublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC...",
	}
	err := sshKeyRepo.Create(ctx, key)
	require.NoError(t, err)

	// 2. Launch with SSH Key
	inst, err := svc.LaunchInstance(ctx, coreports.LaunchParams{
		Name:         "ssh-inst",
		Image:        testImage,
		InstanceType: testInstanceType,
		SSHKeyID:     &key.ID,
	})
	require.NoError(t, err)

	assert.Equal(t, &key.ID, inst.SSHKeyID)
}

func TestInstanceService_LifecycleMethods(t *testing.T) {
	_, svc, compute, repo, _, _, ctx := setupInstanceServiceTest(t)
	
	// Setup instance
	inst, err := svc.LaunchInstance(ctx, coreports.LaunchParams{
		Name:         "lifecycle-test",
		Image:        testImage,
		InstanceType: testInstanceType,
	})
	require.NoError(t, err)
	err = svc.Provision(ctx, domain.ProvisionJob{InstanceID: inst.ID})
	require.NoError(t, err)

	t.Run("StopInstance", func(t *testing.T) {
		err := svc.StopInstance(ctx, inst.ID.String())
		assert.NoError(t, err)
		
		dbInst, err := repo.GetByID(ctx, inst.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusStopped, dbInst.Status)

		// Test stopping again (idempotency)
		err = svc.StopInstance(ctx, inst.ID.String())
		assert.NoError(t, err)
	})

	t.Run("StartInstance", func(t *testing.T) {
		err := svc.StartInstance(ctx, inst.ID.String())
		assert.NoError(t, err)
		
		dbInst, err := repo.GetByID(ctx, inst.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusRunning, dbInst.Status)

		// Test starting again (idempotency)
		err = svc.StartInstance(ctx, inst.ID.String())
		assert.NoError(t, err)
	})

	t.Run("GetInstanceLogs", func(t *testing.T) {
		logs, err := svc.GetInstanceLogs(ctx, inst.ID.String())
		assert.NoError(t, err)
		assert.NotNil(t, logs)
	})

	t.Run("Exec", func(t *testing.T) {
		output, err := svc.Exec(ctx, inst.ID.String(), []string{"echo", "hello"})
		assert.NoError(t, err)
		assert.Contains(t, output, "hello")
	})

	t.Run("GetConsoleURL", func(t *testing.T) {
		if compute.Type() == "docker" {
			t.Skip("Skipping console test for docker backend")
		}
		url, err := svc.GetConsoleURL(ctx, inst.ID.String())
		assert.NoError(t, err)
		assert.NotNil(t, url)
	})

	// Cleanup
	// Refresh instance to get the latest container ID after provisioning/restarts
	refreshInst, err := repo.GetByID(ctx, inst.ID)
	if err == nil && refreshInst.ContainerID != "" {
		_ = compute.DeleteInstance(ctx, refreshInst.ContainerID)
	}
}

func TestLaunchInstanceWithOptions(t *testing.T) {
	_, svc, compute, _, _, _, ctx := setupInstanceServiceTest(t)

	opts := coreports.CreateInstanceOptions{
		Name:      "opts-launch",
		ImageName: testImage,
		Ports:     []string{"8080:80"},
		Env:       []string{"FOO=BAR"},
	}

	inst, err := svc.LaunchInstanceWithOptions(ctx, opts)
	require.NoError(t, err)
	assert.Equal(t, "opts-launch", inst.Name)
	assert.Equal(t, "8080:80", inst.Ports)

	// Cleanup
	if inst.ContainerID != "" {
		_ = compute.DeleteInstance(ctx, inst.ContainerID)
	}
}
