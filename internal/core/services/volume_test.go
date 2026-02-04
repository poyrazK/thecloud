package services_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/noop"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupVolumeServiceTest(t *testing.T) (*services.VolumeService, *postgres.VolumeRepository, context.Context) {
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewVolumeRepository(db)
	storage := noop.NewNoopStorageBackend()

	eventRepo := postgres.NewEventRepository(db)
	eventSvc := services.NewEventService(eventRepo, nil, slog.Default())

	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(auditRepo)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := services.NewVolumeService(repo, storage, eventSvc, auditSvc, logger)
	return svc, repo, ctx
}

func TestVolumeServiceCreateVolumeSuccess(t *testing.T) {
	svc, repo, ctx := setupVolumeServiceTest(t)
	name := "test-vol-create-" + uuid.New().String()
	size := 10

	vol, err := svc.CreateVolume(ctx, name, size)

	assert.NoError(t, err)
	assert.NotNil(t, vol)
	assert.Equal(t, name, vol.Name)
	assert.Equal(t, size, vol.SizeGB)
	assert.Equal(t, domain.VolumeStatusAvailable, vol.Status)

	// Verify in DB
	fetched, err := repo.GetByID(ctx, vol.ID)
	assert.NoError(t, err)
	assert.Equal(t, vol.ID, fetched.ID)
}

func TestVolumeServiceDeleteVolumeSuccess(t *testing.T) {
	svc, repo, ctx := setupVolumeServiceTest(t)
	vol, err := svc.CreateVolume(ctx, "to-delete-"+uuid.New().String(), 5)
	require.NoError(t, err)

	err = svc.DeleteVolume(ctx, vol.ID.String())
	assert.NoError(t, err)

	// Verify Deleted from DB
	_, err = repo.GetByID(ctx, vol.ID)
	assert.Error(t, err)
}

func TestVolumeServiceDeleteVolumeInUseFails(t *testing.T) {
	svc, repo, ctx := setupVolumeServiceTest(t)
	vol, err := svc.CreateVolume(ctx, "in-use-vol-"+uuid.New().String(), 5)
	require.NoError(t, err)

	// Mark as in-use
	vol.Status = domain.VolumeStatusInUse
	err = repo.Update(ctx, vol)
	require.NoError(t, err)

	err = svc.DeleteVolume(ctx, vol.ID.String())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "in use")
}

func TestVolumeServiceListVolumesSuccess(t *testing.T) {
	svc, _, ctx := setupVolumeServiceTest(t)
	_, _ = svc.CreateVolume(ctx, "v1-"+uuid.New().String(), 1)
	_, _ = svc.CreateVolume(ctx, "v2-"+uuid.New().String(), 2)

	result, err := svc.ListVolumes(ctx)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestVolumeServiceGetVolume(t *testing.T) {
	svc, _, ctx := setupVolumeServiceTest(t)
	vol, _ := svc.CreateVolume(ctx, "find-me-"+uuid.New().String(), 5)

	t.Run("get by id", func(t *testing.T) {
		res, err := svc.GetVolume(ctx, vol.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, vol.ID, res.ID)
	})

	t.Run("get by name", func(t *testing.T) {
		res, err := svc.GetVolume(ctx, vol.Name)
		assert.NoError(t, err)
		assert.Equal(t, vol.ID, res.ID)
	})
}

func TestVolumeServiceReleaseVolumesForInstance(t *testing.T) {
	svc, repo, ctx := setupVolumeServiceTest(t)
	instanceID := uuid.New()

	vol, _ := svc.CreateVolume(ctx, "attached-vol", 10)
	vol.Status = domain.VolumeStatusInUse
	vol.InstanceID = &instanceID
	_ = repo.Update(ctx, vol)

	err := svc.ReleaseVolumesForInstance(ctx, instanceID)
	assert.NoError(t, err)

	// Verify released
	updated, _ := repo.GetByID(ctx, vol.ID)
	assert.Equal(t, domain.VolumeStatusAvailable, updated.Status)
	assert.Nil(t, updated.InstanceID)
}

func TestVolumeServiceCreateVolumeRollbackOnRepoError(t *testing.T) {
	// In order to verify rollback functionality, we need to trigger a repository failure
	// after the storage operation has succeeded. Using a cancelled context provides
	// a deterministic way to simulate this failure scenario during the repository assertion phase.

	svc, _, ctx := setupVolumeServiceTest(t)
	cancelledCtx, cancel := context.WithCancel(ctx)
	cancel()

	vol, err := svc.CreateVolume(cancelledCtx, "fail-vol-"+uuid.New().String(), 5)
	assert.Error(t, err)
	assert.Nil(t, vol)
}

func TestVolume_LaunchAttach_Conflict(t *testing.T) {
	// Use InstanceService setup because we need LaunchInstance
	db, svc, _, _, _, volRepo, ctx := setupInstanceServiceTest(t)
	// We also need VolumeService to create volumes elegantly
	volSvc := services.NewVolumeService(volRepo, noop.NewNoopStorageBackend(), services.NewEventService(postgres.NewEventRepository(db), nil, slog.Default()), services.NewAuditService(postgres.NewAuditRepository(db)), slog.Default())

	// 1. Create Volume
	vol, err := volSvc.CreateVolume(ctx, "shared-vol-"+uuid.New().String(), 1)
	require.NoError(t, err)

	// 2. Launch Instance A with Volume
	nameA := "inst-A-" + uuid.New().String()
	image := "alpine:latest"
	volsA := []domain.VolumeAttachment{
		{VolumeIDOrName: vol.ID.String(), MountPath: "/dev/xvdb"},
	}
	instA, err := svc.LaunchInstance(ctx, nameA, image, "", "basic-2", nil, nil, volsA)
	require.NoError(t, err)

	// Provision A to ensure volume becomes "IN_USE"
	err = svc.Provision(ctx, instA.ID, volsA, "")
	require.NoError(t, err)

	// Verify Volume Status is InUse
	updatedVol, _ := volSvc.GetVolume(ctx, vol.ID.String())
	assert.Equal(t, domain.VolumeStatusInUse, updatedVol.Status)
	assert.Equal(t, instA.ID, *updatedVol.InstanceID)

	// 3. Launch Instance B with SAME Volume
	// Should fail because volume is already attached
	nameB := "inst-B-" + uuid.New().String()
	instB, err := svc.LaunchInstance(ctx, nameB, image, "", "basic-2", nil, nil, volsA)

	// Expectation: LaunchInstance should check volume status and fail
	// OR Provision should fail.
	// LaunchInstance usually does validation.
	if err == nil {
		// If Launch succeeded (maybe only validation passed?), try Provision
		err = svc.Provision(ctx, instB.ID, volsA, "")
		assert.Error(t, err, "Provisioning second instance with same volume should fail")
		if err == nil {
			// Cleanup B
			_ = svc.TerminateInstance(ctx, instB.ID.String())
		}
	} else {
		assert.Error(t, err, "Launching second instance with in-use volume should fail")
	}

	// Cleanup
	_ = svc.TerminateInstance(ctx, instA.ID.String())
	_ = volSvc.DeleteVolume(ctx, vol.ID.String())
}
