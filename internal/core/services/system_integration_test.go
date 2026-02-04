//go:build integration

package services_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/docker"
	"github.com/poyrazk/thecloud/internal/repositories/noop"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SyncTaskQueue for system tests to execute jobs immediately or locally
type SyncTaskQueue struct {
	mu   sync.Mutex
	jobs []struct {
		Queue   string
		Payload interface{}
	}
}

func (q *SyncTaskQueue) Enqueue(ctx context.Context, queueName string, payload interface{}) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.jobs = append(q.jobs, struct {
		Queue   string
		Payload interface{}
	}{Queue: queueName, Payload: payload})
	return nil
}

func (q *SyncTaskQueue) Dequeue(ctx context.Context, queueName string) (string, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.jobs) == 0 {
		return "", fmt.Errorf("queue empty")
	}
	job := q.jobs[0]
	q.jobs = q.jobs[1:]
	// Return payload as string (mock serialization)
	return fmt.Sprintf("%v", job.Payload), nil
}

func TestSystem_ComputeLifecycle_Full(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running system integration test")
	}

	// 1. Setup Environment
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	// 2. Real Infrastructure Adapters
	logger := slog.Default()
	compute, err := docker.NewDockerAdapter(logger)
	require.NoError(t, err)

	// Ensure network exists for Docker backend
	if compute.Type() == "docker" {
		_, _ = compute.CreateNetwork(ctx, "cloud-network")
	}

	// 3. Repositories
	repo := postgres.NewInstanceRepository(db)
	vpcRepo := postgres.NewVpcRepository(db)
	subnetRepo := postgres.NewSubnetRepository(db)
	volumeRepo := postgres.NewVolumeRepository(db)
	itRepo := postgres.NewInstanceTypeRepository(db)
	eventRepo := postgres.NewEventRepository(db)
	auditRepo := postgres.NewAuditRepository(db)

	// 4. Services
	eventSvc := services.NewEventService(eventRepo, nil, logger)
	auditSvc := services.NewAuditService(auditRepo)
	taskQueue := &SyncTaskQueue{}
	network := noop.NewNoopNetworkAdapter(logger)

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
		Logger:           logger,
	})

	// 5. Prerequisites: Instance Type
	// We need to create "custom" instance type in DB because LaunchInstanceWithOptions defaults to it?
	// No, Wait. LaunchInstanceWithOptions in instance.go sets `InstanceType: "custom"`.
	// But it doesn't validate it against DB in `LaunchInstanceWithOptions`.
	// HOWEVER, `Provision` method DOES validate it: `it, itErr := s.instanceTypeRepo.GetByID(ctx, inst.InstanceType)` in instance.go:214.
	// So we MUST create "custom" instance type.
	customType := &domain.InstanceType{
		ID:       "custom",
		Name:     "Custom Type",
		VCPUs:    1,
		MemoryMB: 256,
		DiskGB:   1,
	}
	_, _ = itRepo.Create(ctx, customType)

	// Create "sys-1" too just in case we switch test params
	sys1 := &domain.InstanceType{
		ID:       "sys-1",
		Name:     "System 1",
		VCPUs:    1,
		MemoryMB: 256,
		DiskGB:   1,
	}
	_, _ = itRepo.Create(ctx, sys1)

	// 6. User Data for Cloud-Init Simulation
	// We verify that the DockerAdapter parses this and creates the file.
	userData := `#cloud-config
write_files:
  - path: /tmp/system_test_proof
    content: "cloud-init-is-working"
runcmd:
  - echo "booted" > /tmp/boot_status
`

	// 7. LAUNCH INSTANCE
	t.Log("Step 7: Launching Instance with Cloud-Init...")
	opts := ports.CreateInstanceOptions{
		Name:      "sys-vm-full",
		ImageName: "ubuntu:22.04", // Use canonical field
		// InstanceType field removed as per lint
		UserData: userData,
	}
	inst, err := svc.LaunchInstanceWithOptions(ctx, opts)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusStarting, inst.Status)

	// 8. PROVISION
	// In a real system, a worker would pick this up. Here we call it directly.
	t.Log("Step 8: Provisioning Instance...")
	// Provision signature: func (s *InstanceService) Provision(ctx context.Context, instanceID uuid.UUID, volumes []domain.VolumeAttachment, userData string) error
	err = svc.Provision(ctx, inst.ID, nil, userData)
	require.NoError(t, err)

	// 9. VERIFY RUNNING
	t.Log("Step 9: Verifying Running State...")
	updatedInst, err := repo.GetByID(ctx, inst.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusRunning, updatedInst.Status)
	assert.NotEmpty(t, updatedInst.ContainerID)

	// 10. VERIFY CLOUD-INIT EXECUTION (File Existence)
	// We check if the file /tmp/system_test_proof exists and contains "cloud-init-is-working"
	t.Log("Step 10: Verifying Cloud-Init execution inside container...")
	executed := false
	for i := 0; i < 20; i++ { // Retry for up to 10s
		out, err := compute.Exec(ctx, updatedInst.ContainerID, []string{"cat", "/tmp/system_test_proof"})
		if err == nil && strings.Contains(out, "cloud-init-is-working") {
			executed = true
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if !executed {
		// Attempt to get logs to debug
		logs, _ := compute.GetInstanceLogs(ctx, updatedInst.ContainerID)
		if logs != nil {
			logContent, _ := io.ReadAll(logs)
			t.Logf("Container Logs: %s", string(logContent))
		}
	}
	assert.True(t, executed, "Cloud-Init failed to create proof file")

	// 11. STOP INSTANCE
	t.Log("Step 11: Stopping Instance...")
	err = svc.StopInstance(ctx, inst.ID.String())
	require.NoError(t, err)

	updatedInst, _ = repo.GetByID(ctx, inst.ID)
	assert.Equal(t, domain.StatusStopped, updatedInst.Status)

	// 12. START INSTANCE
	// Now this should work!
	t.Log("Step 12: Starting Instance...")
	err = svc.StartInstance(ctx, inst.ID.String())
	require.NoError(t, err)

	updatedInst, _ = repo.GetByID(ctx, inst.ID)
	assert.Equal(t, domain.StatusRunning, updatedInst.Status)

	// 13. TERMINATE
	t.Log("Step 13: Terminating Instance...")
	err = svc.TerminateInstance(ctx, inst.ID.String())
	require.NoError(t, err)

	// Verify Cleanup
	_, err = repo.GetByID(ctx, inst.ID)
	assert.Error(t, err) // Should be deleted or not found

	// Verify Container Removal
	// Using a new context for verification as previous might be canceled/timeout depending on test runner
	_, err = compute.GetInstanceIP(context.Background(), updatedInst.ContainerID)
	assert.Error(t, err) // Container should be gone
}
