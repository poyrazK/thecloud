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
	name := "test-vol-create"
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
	vol, err := svc.CreateVolume(ctx, "to-delete", 5)
	require.NoError(t, err)

	err = svc.DeleteVolume(ctx, vol.ID.String())
	assert.NoError(t, err)

	// Verify Deleted from DB
	_, err = repo.GetByID(ctx, vol.ID)
	assert.Error(t, err)
}

func TestVolumeServiceDeleteVolumeInUseFails(t *testing.T) {
	svc, repo, ctx := setupVolumeServiceTest(t)
	vol, err := svc.CreateVolume(ctx, "in-use-vol", 5)
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
	_, _ = svc.CreateVolume(ctx, "v1", 1)
	_, _ = svc.CreateVolume(ctx, "v2", 2)

	result, err := svc.ListVolumes(ctx)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestVolumeServiceGetVolume(t *testing.T) {
	svc, _, ctx := setupVolumeServiceTest(t)
	vol, _ := svc.CreateVolume(ctx, "find-me", 5)

	t.Run("get by id", func(t *testing.T) {
		res, err := svc.GetVolume(ctx, vol.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, vol.ID, res.ID)
	})

	t.Run("get by name", func(t *testing.T) {
		res, err := svc.GetVolume(ctx, "find-me")
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

	vol, err := svc.CreateVolume(cancelledCtx, "fail-vol", 5)
	assert.Error(t, err)
	assert.Nil(t, vol)
}
