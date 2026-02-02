package services_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/noop"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSnapshotServiceIntegrationTest(t *testing.T) (ports.SnapshotService, ports.SnapshotRepository, ports.VolumeRepository, context.Context) {
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewSnapshotRepository(db)
	volRepo := postgres.NewVolumeRepository(db)
	storage := noop.NewNoopStorageBackend()

	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(auditRepo)

	eventRepo := postgres.NewEventRepository(db)
	eventSvc := services.NewEventService(eventRepo, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := services.NewSnapshotService(repo, volRepo, storage, eventSvc, auditSvc, logger)

	return svc, repo, volRepo, ctx
}

func TestSnapshotService_Integration(t *testing.T) {
	svc, repo, volRepo, ctx := setupSnapshotServiceIntegrationTest(t)
	userID := appcontext.UserIDFromContext(ctx)

	// Setup Volume
	vol := &domain.Volume{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      "snap-test-vol",
		SizeGB:    10,
		Status:    domain.VolumeStatusAvailable,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := volRepo.Create(ctx, vol)
	require.NoError(t, err)

	t.Run("CreateSnapshot", func(t *testing.T) {
		snap, err := svc.CreateSnapshot(ctx, vol.ID, "integration snapshot")
		assert.NoError(t, err)
		assert.NotNil(t, snap)
		assert.Equal(t, vol.ID, snap.VolumeID)
		assert.Equal(t, domain.SnapshotStatusCreating, snap.Status)

		// Wait for async graduation
		time.Sleep(150 * time.Millisecond)

		fetched, err := svc.GetSnapshot(ctx, snap.ID)
		assert.NoError(t, err)
		assert.Equal(t, domain.SnapshotStatusAvailable, fetched.Status)
	})

	t.Run("ListAndRestore", func(t *testing.T) {
		snap, _ := svc.CreateSnapshot(ctx, vol.ID, "list-snap")
		time.Sleep(150 * time.Millisecond)

		// List
		snaps, err := svc.ListSnapshots(ctx)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(snaps), 2)

		// Restore
		restoredVol, err := svc.RestoreSnapshot(ctx, snap.ID, "restored-vol-name")
		assert.NoError(t, err)
		assert.NotNil(t, restoredVol)
		assert.Equal(t, "restored-vol-name", restoredVol.Name)
		assert.Equal(t, vol.SizeGB, restoredVol.SizeGB)

		// Verify restored volume in DB
		fetchedVol, err := volRepo.GetByID(ctx, restoredVol.ID)
		assert.NoError(t, err)
		assert.Equal(t, restoredVol.ID, fetchedVol.ID)
	})

	t.Run("Delete", func(t *testing.T) {
		snap, _ := svc.CreateSnapshot(ctx, vol.ID, "to-delete")

		err = svc.DeleteSnapshot(ctx, snap.ID)
		assert.NoError(t, err)

		_, err = repo.GetByID(ctx, snap.ID)
		assert.Error(t, err)
	})
}
