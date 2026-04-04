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
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSnapshotService_Unit(t *testing.T) {
	mockRepo := new(MockSnapshotRepo)
	mockVolRepo := new(MockVolumeRepo)
	mockStorage := new(MockStorageBackend)
	mockEventSvc := new(MockEventService)
	mockAuditSvc := new(MockAuditService)
	rbacSvc := new(MockRBACService)
	rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	svc := services.NewSnapshotService(
		mockRepo,
		rbacSvc,
		mockVolRepo,
		mockStorage,
		mockEventSvc,
		mockAuditSvc,
		slog.Default(),
	)

	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)

	t.Run("CreateSnapshot", func(t *testing.T) {
		volID := uuid.New()
		vol := &domain.Volume{ID: volID, Name: "test-vol", SizeGB: 10}

		mockVolRepo.On("GetByID", mock.Anything, volID).Return(vol, nil).Once()
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Maybe()
		mockStorage.On("CreateSnapshot", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
		mockEventSvc.On("RecordEvent", mock.Anything, "SNAPSHOT_CREATE", mock.Anything, "SNAPSHOT", mock.Anything).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "snapshot.create", "snapshot", mock.Anything, mock.Anything).Return(nil).Once()

		snap, err := svc.CreateSnapshot(ctx, volID, "my backup")
		require.NoError(t, err)
		assert.NotNil(t, snap)
		assert.Equal(t, volID, snap.VolumeID)

		time.Sleep(10 * time.Millisecond) // Wait for async part
	})

	t.Run("RestoreSnapshot", func(t *testing.T) {
		snapID := uuid.New()
		snap := &domain.Snapshot{ID: snapID, UserID: userID, SizeGB: 10, Status: domain.SnapshotStatusAvailable}

		mockRepo.On("GetByID", mock.Anything, snapID).Return(snap, nil).Once()
		mockStorage.On("CreateVolume", mock.Anything, mock.Anything, 10).Return("/dev/vdc", nil).Once()
		mockStorage.On("RestoreSnapshot", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		mockVolRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		mockEventSvc.On("RecordEvent", mock.Anything, "VOLUME_RESTORE", mock.Anything, "VOLUME", mock.Anything).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "volume.restore", "volume", mock.Anything, mock.Anything).Return(nil).Once()

		vol, err := svc.RestoreSnapshot(ctx, snapID, "restored-vol")
		require.NoError(t, err)
		assert.NotNil(t, vol)
		assert.Equal(t, "restored-vol", vol.Name)
	})

	t.Run("RestoreSnapshot_NotAvailable", func(t *testing.T) {
		snapID := uuid.New()
		snap := &domain.Snapshot{ID: snapID, Status: domain.SnapshotStatusCreating}
		mockRepo.On("GetByID", mock.Anything, snapID).Return(snap, nil).Once()

		_, err := svc.RestoreSnapshot(ctx, snapID, "fail")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not available")
	})

	t.Run("RestoreSnapshot_StorageFailure", func(t *testing.T) {
		snapID := uuid.New()
		snap := &domain.Snapshot{ID: snapID, SizeGB: 10, Status: domain.SnapshotStatusAvailable}
		mockRepo.On("GetByID", mock.Anything, snapID).Return(snap, nil).Once()
		mockStorage.On("CreateVolume", mock.Anything, mock.Anything, 10).Return("/dev/vdc", nil).Once()
		mockStorage.On("RestoreSnapshot", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("storage fail")).Once()
		mockStorage.On("DeleteVolume", mock.Anything, mock.Anything).Return(nil).Once()

		_, err := svc.RestoreSnapshot(ctx, snapID, "fail")
		require.Error(t, err)
	})

	t.Run("DeleteSnapshot_Success", func(t *testing.T) {
		id := uuid.New()
		snap := &domain.Snapshot{ID: id, UserID: userID}
		mockRepo.On("GetByID", mock.Anything, id).Return(snap, nil).Once()
		mockStorage.On("DeleteSnapshot", mock.Anything, mock.Anything).Return(nil).Once()
		mockRepo.On("Delete", mock.Anything, id).Return(nil).Once()
		mockEventSvc.On("RecordEvent", mock.Anything, "SNAPSHOT_DELETE", id.String(), "SNAPSHOT", mock.Anything).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "snapshot.delete", "snapshot", id.String(), mock.Anything).Return(nil).Once()

		err := svc.DeleteSnapshot(ctx, id)
		require.NoError(t, err)
	})
}
