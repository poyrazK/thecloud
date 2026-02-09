//go:build integration

package k8s_test

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/platform"
	"github.com/poyrazk/thecloud/internal/repositories/docker"
	"github.com/poyrazk/thecloud/internal/repositories/k8s"
	"github.com/poyrazk/thecloud/internal/repositories/noop"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestK8sProvisionerLifecycle provides a realistic integration test for the K8s provisioner
// using a real Docker compute backend and real Postgres storage.
func TestK8sProvisionerLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Infrastructure Setup: Real Postgres
	db := postgres.SetupDB(t)
	defer db.Close()

	// Inject real identity context
	ctx = postgres.SetupTestUser(t, db)

	// Infrastructure Setup: Real Docker
	compute, err := docker.NewDockerAdapter(logger)
	require.NoError(t, err)

	// Repositories: Using real Postgres implementations
	clusterRepo := postgres.NewClusterRepository(db)
	instanceRepo := postgres.NewInstanceRepository(db)
	vpcRepo := postgres.NewVpcRepository(db)
	subnetRepo := postgres.NewSubnetRepository(db)
	secretRepo := postgres.NewSecretRepository(db)
	sgRepo := postgres.NewSecurityGroupRepository(db)
	storageRepo := postgres.NewStorageRepository(db)
	lbRepo := postgres.NewLBRepository(db)

	// Secondary Infrastructure (No-op services for side-effects)
	eventSvc := &noop.NoopEventService{}
	auditSvc := &noop.NoopAuditService{}

	// Real SecretService
	secretSvc := services.NewSecretService(secretRepo, eventSvc, auditSvc, logger, "test-master-key-32-chars-long-!!!", "default")

	// Network: No-op for OVS logic (requires local system support)
	netBackend := &noopNetworkBackend{}

	// Core Services
	sgSvc := services.NewSecurityGroupService(sgRepo, vpcRepo, netBackend, auditSvc, logger)
	storageSvc := services.NewStorageService(storageRepo, nil, auditSvc, nil, &platform.Config{})
	lbSvc := services.NewLBService(lbRepo, vpcRepo, instanceRepo, auditSvc)

	// InstanceService: The real one!
	// We use a SyncTaskQueue to make the provisioning synchronous in the test.
	taskQueue := &syncTaskQueue{}

	instSvc := services.NewInstanceService(services.InstanceServiceParams{
		Repo:             instanceRepo,
		VpcRepo:          vpcRepo,
		SubnetRepo:       subnetRepo,
		VolumeRepo:       nil,
		InstanceTypeRepo: &noop.NoopInstanceTypeRepository{},
		Compute:          compute,
		Network:          netBackend,
		EventSvc:         eventSvc,
		AuditSvc:         auditSvc,
		TaskQueue:        taskQueue,
		Logger:           logger,
	})

	// Inject the service into the task queue so it can call Provision
	taskQueue.svc = instSvc

	provisioner := k8s.NewKubeadmProvisioner(instSvc, clusterRepo, secretSvc, sgSvc, storageSvc, lbSvc, logger)

	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	// Define a functional test VPC and Subnet
	vpc := &domain.VPC{
		ID:        uuid.New(),
		UserID:    userID,
		TenantID:  tenantID,
		Name:      "k8s-vpc",
		NetworkID: "bridge",
	}
	err = vpcRepo.Create(ctx, vpc)
	require.NoError(t, err)

	cluster := &domain.Cluster{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        "integration-cluster-real-" + time.Now().Format("20060102150405"),
		VpcID:       vpc.ID,
		Version:     "v1.29.0",
		PodCIDR:     "10.244.0.0/16",
		ServiceCIDR: "10.96.0.0/12",
		WorkerCount: 1,
		Status:      domain.ClusterStatusProvisioning,
		CreatedAt:   time.Now(),
	}
	err = clusterRepo.Create(ctx, cluster)
	require.NoError(t, err)

	// EXECUTION: Start the real provisioning flow
	// We use a short timeout (30s) because we know full K8s install (apt-get, kubeadm) takes minutes.
	// We only want to verify that the container launches and the provisioner starts waiting.
	t.Log("Starting Kubeadm provisioning (real Docker containers)...")
	provisionCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err = provisioner.Provision(provisionCtx, cluster)

	// We expect a timeout because we cut it short.
	// If it fails immediately with something else (e.g. image pull error), that's a real fail.
	if err != nil {
		if strings.Contains(err.Error(), "context deadline exceeded") ||
			strings.Contains(err.Error(), "timed out") {
			t.Logf("Provisioning timed out as expected (full install takes too long): %v", err)
		} else {
			require.NoError(t, err, "Provisioning failed with unexpected error")
		}
	}

	// Verify outcome
	// instSvc.ListInstances uses repo.List under the hood
	instances, _ := instSvc.ListInstances(ctx)
	found := false
	for _, inst := range instances {
		if strings.Contains(inst.Name, cluster.Name) {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected at least one Docker container to be created for the cluster")
}

// noopNetworkBackend stub for environments without OVS
type noopNetworkBackend struct {
	ports.NetworkBackend
}

func (n *noopNetworkBackend) CreateBridge(ctx context.Context, name string, vxlanID int) error {
	return nil
}
func (n *noopNetworkBackend) DeleteBridge(ctx context.Context, name string) error        { return nil }
func (n *noopNetworkBackend) ListBridges(ctx context.Context) ([]string, error)          { return nil, nil }
func (n *noopNetworkBackend) AddPort(ctx context.Context, bridge, portName string) error { return nil }
func (n *noopNetworkBackend) DeletePort(ctx context.Context, bridge, portName string) error {
	return nil
}
func (n *noopNetworkBackend) CreateVXLANTunnel(ctx context.Context, b string, v int, r string) error {
	return nil
}
func (n *noopNetworkBackend) DeleteVXLANTunnel(ctx context.Context, b, r string) error { return nil }
func (n *noopNetworkBackend) AddFlowRule(ctx context.Context, bridge string, rule ports.FlowRule) error {
	return nil
}
func (n *noopNetworkBackend) DeleteFlowRule(ctx context.Context, bridge, match string) error {
	return nil
}
func (n *noopNetworkBackend) ListFlowRules(ctx context.Context, bridge string) ([]ports.FlowRule, error) {
	return nil, nil
}
func (n *noopNetworkBackend) CreateVethPair(ctx context.Context, h, c string) error { return nil }
func (n *noopNetworkBackend) AttachVethToBridge(ctx context.Context, bridge, vethEnd string) error {
	return nil
}
func (n *noopNetworkBackend) DeleteVethPair(ctx context.Context, hostEnd string) error { return nil }
func (n *noopNetworkBackend) SetVethIP(ctx context.Context, v, ip, c string) error     { return nil }
func (n *noopNetworkBackend) Type() string                                             { return "noop" }
func (n *noopNetworkBackend) Ping(ctx context.Context) error                           { return nil }

type syncTaskQueue struct {
	ports.TaskQueue
	svc *services.InstanceService
}

func (q *syncTaskQueue) Enqueue(ctx context.Context, queue string, payload interface{}) error {
	if job, ok := payload.(domain.ProvisionJob); ok {
		// Run synchronously for integration test
		return q.svc.Provision(ctx, job)
	}
	return nil
}
