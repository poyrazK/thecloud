//go:build integration
// +build integration

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
	"github.com/poyrazk/thecloud/internal/repositories/docker"
	"github.com/poyrazk/thecloud/internal/repositories/noop"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testDBName = "test-db"
)

func setupDatabaseServiceTest(t *testing.T) (ports.DatabaseService, ports.DatabaseRepository, *docker.DockerAdapter, ports.VpcRepository, context.Context) {
	t.Helper()
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewDatabaseRepository(db)
	vpcRepo := postgres.NewVpcRepository(db)

	compute, err := docker.NewDockerAdapter(slog.Default())
	require.NoError(t, err)

	eventRepo := postgres.NewEventRepository(db)
	eventSvc := services.NewEventService(eventRepo, nil, slog.Default())

	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(auditRepo)

	volRepo := postgres.NewVolumeRepository(db)
	storage := noop.NewNoopStorageBackendAdapter()
	volumeSvc := services.NewVolumeService(volRepo, storage, eventSvc, auditSvc, slog.Default())

	logger := slog.Default()

	svc := services.NewDatabaseService(services.DatabaseServiceParams{
		Repo:         repo,
		Compute:      compute,
		VpcRepo:      vpcRepo,
		VolumeSvc:    volumeSvc,
		SnapshotSvc:  nil,
		SnapshotRepo: nil,
		EventSvc:     eventSvc,
		AuditSvc:     auditSvc,
		Logger:       logger,
	})

	return svc, repo, compute, vpcRepo, ctx
}

func TestCreateDatabaseSuccess(t *testing.T) {
	svc, repo, compute, _, ctx := setupDatabaseServiceTest(t)

	db, err := svc.CreateDatabase(ctx, ports.CreateDatabaseRequest{
		Name:             testDBName,
		Engine:           "postgres",
		Version:          "16",
		AllocatedStorage: 20,
	})

	require.NoError(t, err)
	assert.NotNil(t, db)
	assert.Equal(t, testDBName, db.Name)
	assert.Equal(t, domain.EnginePostgres, db.Engine)
	assert.Equal(t, 20, db.AllocatedStorage)
	assert.NotEmpty(t, db.ContainerID)

	// Verify instance creation by checking connectivity
	ip, err := compute.GetInstanceIP(ctx, db.ContainerID)
	require.NoError(t, err)
	assert.NotEmpty(t, ip)

	// Clean up created container
	err = compute.DeleteInstance(ctx, db.ContainerID)
	require.NoError(t, err)

	// Verify in DB
	fetched, err := repo.GetByID(ctx, db.ID)
	require.NoError(t, err)
	assert.Equal(t, db.ID, fetched.ID)
}

func TestCreateDatabaseWithVpc(t *testing.T) {
	svc, _, compute, vpcRepo, ctx := setupDatabaseServiceTest(t)

	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	vpcID := uuid.New()
	networkName := "net-" + vpcID.String()
	netID, err := compute.CreateNetwork(ctx, networkName)
	require.NoError(t, err)
	defer func() {
		err := compute.DeleteNetwork(ctx, netID)
		assert.NoError(t, err)
	}()

	vpc := &domain.VPC{
		ID:        vpcID,
		UserID:    userID,
		TenantID:  tenantID,
		Name:      "test-vpc",
		CIDRBlock: "10.0.0.0/16",
		NetworkID: netID,
		Status:    "ACTIVE",
		CreatedAt: time.Now(),
	}
	err = vpcRepo.Create(ctx, vpc)
	require.NoError(t, err)

	db, err := svc.CreateDatabase(ctx, ports.CreateDatabaseRequest{
		Name:             testDBName,
		Engine:           "postgres",
		Version:          "16",
		VpcID:            &vpcID,
		AllocatedStorage: 10,
	})
	require.NoError(t, err)
	require.NotNil(t, db)
	assert.Equal(t, &vpcID, db.VpcID)

	// Cleanup
	err = compute.DeleteInstance(ctx, db.ContainerID)
	require.NoError(t, err)
}

func TestCreateReplica(t *testing.T) {
	svc, repo, compute, _, ctx := setupDatabaseServiceTest(t)

	// 1. Create primary
	primary, err := svc.CreateDatabase(ctx, ports.CreateDatabaseRequest{
		Name:             "primary-db",
		Engine:           "postgres",
		Version:          "16",
		AllocatedStorage: 20,
	})
	require.NoError(t, err)
	defer func() {
		err := compute.DeleteInstance(ctx, primary.ContainerID)
		assert.NoError(t, err)
	}()

	// 2. Create replica
	replica, err := svc.CreateReplica(ctx, primary.ID, "replica-db")
	require.NoError(t, err)
	assert.NotNil(t, replica)
	assert.Equal(t, domain.RoleReplica, replica.Role)
	assert.Equal(t, &primary.ID, replica.PrimaryID)
	assert.NotEmpty(t, replica.ContainerID)

	defer func() {
		err := compute.DeleteInstance(ctx, replica.ContainerID)
		assert.NoError(t, err)
	}()

	// 3. Verify in repo
	fetched, err := repo.GetByID(ctx, replica.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.RoleReplica, fetched.Role)
}

func TestModifyDatabaseVolumeResize(t *testing.T) {
	svc, repo, compute, _, ctx := setupDatabaseServiceTest(t)

	// 1. Create a database with 10GB
	db, err := svc.CreateDatabase(ctx, ports.CreateDatabaseRequest{
		Name:             "resize-db",
		Engine:           "postgres",
		Version:          "16",
		AllocatedStorage: 10,
	})
	require.NoError(t, err)
	defer func() {
		_ = compute.DeleteInstance(ctx, db.ContainerID)
	}()

	// 2. Resize to 20GB
	newSize := 20
	updated, err := svc.ModifyDatabase(ctx, ports.ModifyDatabaseRequest{
		ID:               db.ID,
		AllocatedStorage: &newSize,
	})
	require.NoError(t, err)
	assert.Equal(t, 20, updated.AllocatedStorage)

	// 3. Verify in repository
	fetched, err := repo.GetByID(ctx, db.ID)
	require.NoError(t, err)
	assert.Equal(t, 20, fetched.AllocatedStorage)
}
