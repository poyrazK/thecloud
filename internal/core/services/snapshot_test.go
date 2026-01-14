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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupSnapshotServiceTest(_ *testing.T) (*MockSnapshotRepo, *MockVolumeRepo, *MockStorageBackend, *MockEventService, *MockAuditService, ports.SnapshotService) {
	repo := new(MockSnapshotRepo)
	volRepo := new(MockVolumeRepo)
	storage := new(MockStorageBackend)
	eventSvc := new(MockEventService)
	auditSvc := new(MockAuditService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := services.NewSnapshotService(repo, volRepo, storage, eventSvc, auditSvc, logger)
	return repo, volRepo, storage, eventSvc, auditSvc, svc
}

func TestCreateSnapshotSuccess(t *testing.T) {
	repo, volRepo, storage, eventSvc, auditSvc, svc := setupSnapshotServiceTest(t)
	defer repo.AssertExpectations(t)
	defer volRepo.AssertExpectations(t)
	defer storage.AssertExpectations(t)
	defer eventSvc.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	ctx := appcontext.WithUserID(context.Background(), uuid.New())
	volID := uuid.New()
	vol := &domain.Volume{
		ID:     volID,
		Name:   "test-vol",
		SizeGB: 10,
	}

	volRepo.On("GetByID", ctx, volID).Return(vol, nil)
	repo.On("Create", ctx, mock.AnythingOfType("*domain.Snapshot")).Return(nil)
	eventSvc.On("RecordEvent", ctx, "SNAPSHOT_CREATE", mock.Anything, "SNAPSHOT", mock.Anything).Return(nil)
	auditSvc.On("Log", ctx, mock.Anything, "snapshot.create", "snapshot", mock.Anything, mock.Anything).Return(nil)

	// Async expectations
	storage.On("CreateSnapshot", mock.Anything, "thecloud-vol-"+volID.String()[:8], mock.Anything).Return(nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Snapshot")).Return(nil)

	snap, err := svc.CreateSnapshot(ctx, volID, "Test snapshot")

	assert.NoError(t, err)
	assert.NotNil(t, snap)
	assert.Equal(t, volID, snap.VolumeID)
	assert.Equal(t, domain.SnapshotStatusCreating, snap.Status)

	// Wait a bit for the goroutine to finish its mocks
	time.Sleep(100 * time.Millisecond)
}

func TestRestoreSnapshotSuccess(t *testing.T) {
	repo, volRepo, storage, eventSvc, auditSvc, svc := setupSnapshotServiceTest(t)
	defer repo.AssertExpectations(t)
	defer volRepo.AssertExpectations(t)
	defer storage.AssertExpectations(t)
	defer eventSvc.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	ctx := appcontext.WithUserID(context.Background(), uuid.New())
	snapID := uuid.New()
	snap := &domain.Snapshot{
		ID:       snapID,
		VolumeID: uuid.New(),
		SizeGB:   10,
		Status:   domain.SnapshotStatusAvailable,
		UserID:   appcontext.UserIDFromContext(ctx),
	}

	repo.On("GetByID", ctx, snapID).Return(snap, nil)
	// CreateVolume returns (string, error)
	storage.On("CreateVolume", ctx, mock.Anything, 10).Return("vol-path", nil)
	// Restore expectations
	// volume id is dynamic in test but we can use mock.Anything or partial match
	storage.On("RestoreSnapshot", ctx, mock.Anything, mock.Anything).Return(nil)
	volRepo.On("Create", ctx, mock.AnythingOfType("*domain.Volume")).Return(nil)

	eventSvc.On("RecordEvent", ctx, "VOLUME_RESTORE", mock.Anything, "VOLUME", mock.Anything).Return(nil)
	auditSvc.On("Log", ctx, mock.Anything, "volume.restore", "volume", mock.Anything, mock.Anything).Return(nil)

	vol, err := svc.RestoreSnapshot(ctx, snapID, "restored-vol")

	assert.NoError(t, err)
	assert.NotNil(t, vol)
	assert.Equal(t, "restored-vol", vol.Name)
	assert.Equal(t, 10, vol.SizeGB)
}

func TestDeleteSnapshotSuccess(t *testing.T) {
	repo, _, storage, eventSvc, auditSvc, svc := setupSnapshotServiceTest(t)
	defer repo.AssertExpectations(t)
	defer storage.AssertExpectations(t)
	defer eventSvc.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	ctx := appcontext.WithUserID(context.Background(), uuid.New())
	snapID := uuid.New()
	snap := &domain.Snapshot{
		ID:     snapID,
		UserID: appcontext.UserIDFromContext(ctx),
	}

	repo.On("GetByID", ctx, snapID).Return(snap, nil)

	backendSnapName := "thecloud-snap-" + snapID.String()[:8]
	storage.On("DeleteSnapshot", ctx, backendSnapName).Return(nil)

	repo.On("Delete", ctx, snapID).Return(nil)
	eventSvc.On("RecordEvent", ctx, "SNAPSHOT_DELETE", snapID.String(), "SNAPSHOT", mock.Anything).Return(nil)
	auditSvc.On("Log", ctx, mock.Anything, "snapshot.delete", "snapshot", snapID.String(), mock.Anything).Return(nil)

	err := svc.DeleteSnapshot(ctx, snapID)

	assert.NoError(t, err)
}

func TestListSnapshots(t *testing.T) {
	repo, _, _, _, _, svc := setupSnapshotServiceTest(t)
	defer repo.AssertExpectations(t)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)

	items := []*domain.Snapshot{{ID: uuid.New(), UserID: userID}}
	repo.On("ListByUserID", ctx, userID).Return(items, nil).Once()

	res, err := svc.ListSnapshots(ctx)
	assert.NoError(t, err)
	assert.Len(t, res, 1)
}

func TestGetSnapshot(t *testing.T) {
	repo, _, _, _, _, svc := setupSnapshotServiceTest(t)
	defer repo.AssertExpectations(t)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)

	snapID := uuid.New()
	snap := &domain.Snapshot{ID: snapID, UserID: userID}
	repo.On("GetByID", ctx, snapID).Return(snap, nil).Once()

	res, err := svc.GetSnapshot(ctx, snapID)
	assert.NoError(t, err)
	assert.Equal(t, snapID, res.ID)
}
