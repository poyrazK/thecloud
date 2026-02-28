package services_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestVolumeServiceUnit(t *testing.T) {
	mockRepo := new(MockVolumeRepo)
	mockStorage := new(MockStorageBackend)
	mockEventSvc := new(MockEventService)
	mockAuditSvc := new(MockAuditService)
	svc := services.NewVolumeService(mockRepo, mockStorage, mockEventSvc, mockAuditSvc, slog.Default())

	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)

	t.Run("CreateVolume Success", func(t *testing.T) {
		mockStorage.On("CreateVolume", mock.Anything, mock.Anything, 10).Return("/dev/vdb", nil).Once()
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		mockEventSvc.On("RecordEvent", mock.Anything, "VOLUME_CREATE", mock.Anything, "VOLUME", mock.Anything).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "volume.create", "volume", mock.Anything, mock.Anything).Return(nil).Once()

		vol, err := svc.CreateVolume(ctx, "test-vol", 10)
		require.NoError(t, err)
		assert.NotNil(t, vol)
		assert.Equal(t, "/dev/vdb", vol.BackendPath)
	})

	t.Run("CreateVolume Storage Error", func(t *testing.T) {
		mockStorage.On("CreateVolume", mock.Anything, mock.Anything, 10).Return("", errors.New("storage fail")).Once()
		_, err := svc.CreateVolume(ctx, "test-vol", 10)
		assert.Error(t, err)
	})

	t.Run("CreateVolume Repo Error Rollback", func(t *testing.T) {
		mockStorage.On("CreateVolume", mock.Anything, mock.Anything, 10).Return("/dev/vdb", nil).Once()
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db fail")).Once()
		mockStorage.On("DeleteVolume", mock.Anything, mock.Anything).Return(nil).Once()

		_, err := svc.CreateVolume(ctx, "test-vol", 10)
		assert.Error(t, err)
	})

	t.Run("ListVolumes", func(t *testing.T) {
		mockRepo.On("List", mock.Anything).Return([]*domain.Volume{{ID: uuid.New()}}, nil).Once()
		vols, err := svc.ListVolumes(ctx)
		assert.NoError(t, err)
		assert.Len(t, vols, 1)
	})

	t.Run("GetVolume by ID", func(t *testing.T) {
		id := uuid.New()
		mockRepo.On("GetByID", mock.Anything, id).Return(&domain.Volume{ID: id}, nil).Once()
		vol, err := svc.GetVolume(ctx, id.String())
		assert.NoError(t, err)
		assert.Equal(t, id, vol.ID)
	})

	t.Run("GetVolume by Name", func(t *testing.T) {
		mockRepo.On("GetByName", mock.Anything, "myvol").Return(&domain.Volume{Name: "myvol"}, nil).Once()
		vol, err := svc.GetVolume(ctx, "myvol")
		assert.NoError(t, err)
		assert.Equal(t, "myvol", vol.Name)
	})

	t.Run("DeleteVolume Success", func(t *testing.T) {
		volID := uuid.New()
		vol := &domain.Volume{ID: volID, Status: domain.VolumeStatusAvailable, UserID: userID, SizeGB: 10}
		mockRepo.On("GetByID", mock.Anything, volID).Return(vol, nil).Once()
		mockStorage.On("DeleteVolume", mock.Anything, mock.Anything).Return(nil).Once()
		mockRepo.On("Delete", mock.Anything, volID).Return(nil).Once()
		mockEventSvc.On("RecordEvent", mock.Anything, "VOLUME_DELETE", volID.String(), "VOLUME", mock.Anything).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "volume.delete", "volume", volID.String(), mock.Anything).Return(nil).Once()

		err := svc.DeleteVolume(ctx, volID.String())
		assert.NoError(t, err)
	})

	t.Run("DeleteVolume Storage Success Repo Fail", func(t *testing.T) {
		volID := uuid.New()
		vol := &domain.Volume{ID: volID, Status: domain.VolumeStatusAvailable, UserID: userID, SizeGB: 10}
		mockRepo.On("GetByID", mock.Anything, volID).Return(vol, nil).Once()
		mockStorage.On("DeleteVolume", mock.Anything, mock.Anything).Return(nil).Once()
		mockRepo.On("Delete", mock.Anything, volID).Return(errors.New("repo fail")).Once()

		err := svc.DeleteVolume(ctx, volID.String())
		assert.Error(t, err)
	})

	t.Run("DeleteVolume In Use Error", func(t *testing.T) {
		volID := uuid.New()
		vol := &domain.Volume{ID: volID, Status: domain.VolumeStatusInUse}
		mockRepo.On("GetByID", mock.Anything, volID).Return(vol, nil).Once()
		err := svc.DeleteVolume(ctx, volID.String())
		assert.Error(t, err)
	})

	t.Run("AttachVolume Success", func(t *testing.T) {
		volID := uuid.New()
		instID := uuid.New()
		vol := &domain.Volume{ID: volID, Status: domain.VolumeStatusAvailable, UserID: userID}

		mockRepo.On("GetByID", mock.Anything, volID).Return(vol, nil).Once()
		mockStorage.On("AttachVolume", mock.Anything, mock.Anything, instID.String()).Return(nil).Once()
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(v *domain.Volume) bool {
			return v.Status == domain.VolumeStatusInUse && v.InstanceID != nil && *v.InstanceID == instID
		})).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "volume.attach", "volume", volID.String(), mock.Anything).Return(nil).Once()

		err := svc.AttachVolume(ctx, volID.String(), instID.String(), "/mnt/data")
		require.NoError(t, err)
	})

	t.Run("AttachVolume Repo Update Fail", func(t *testing.T) {
		volID := uuid.New()
		instID := uuid.New()
		vol := &domain.Volume{ID: volID, Status: domain.VolumeStatusAvailable, UserID: userID}

		mockRepo.On("GetByID", mock.Anything, volID).Return(vol, nil).Once()
		mockStorage.On("AttachVolume", mock.Anything, mock.Anything, instID.String()).Return(nil).Once()
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(errors.New("db update fail")).Once()
		// Verify rollback
		mockStorage.On("DetachVolume", mock.Anything, mock.Anything, instID.String()).Return(nil).Once()

		err := svc.AttachVolume(ctx, volID.String(), instID.String(), "/mnt/data")
		assert.Error(t, err)
	})

	t.Run("AttachVolume Storage Fail", func(t *testing.T) {
		volID := uuid.New()
		instID := uuid.New()
		vol := &domain.Volume{ID: volID, Status: domain.VolumeStatusAvailable, UserID: userID}

		mockRepo.On("GetByID", mock.Anything, volID).Return(vol, nil).Once()
		mockStorage.On("AttachVolume", mock.Anything, mock.Anything, instID.String()).Return(errors.New("storage fail")).Once()

		err := svc.AttachVolume(ctx, volID.String(), instID.String(), "/mnt/data")
		assert.Error(t, err)
	})

	t.Run("AttachVolume Already In Use", func(t *testing.T) {
		volID := uuid.New()
		vol := &domain.Volume{ID: volID, Status: domain.VolumeStatusInUse}
		mockRepo.On("GetByID", mock.Anything, volID).Return(vol, nil).Once()
		err := svc.AttachVolume(ctx, volID.String(), uuid.New().String(), "/m")
		assert.Error(t, err)
	})

	t.Run("DetachVolume Success", func(t *testing.T) {
		volID := uuid.New()
		instID := uuid.New()
		vol := &domain.Volume{ID: volID, Status: domain.VolumeStatusInUse, InstanceID: &instID, UserID: userID}

		mockRepo.On("GetByID", mock.Anything, volID).Return(vol, nil).Once()
		mockStorage.On("DetachVolume", mock.Anything, mock.Anything, instID.String()).Return(nil).Once()
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(v *domain.Volume) bool {
			return v.Status == domain.VolumeStatusAvailable && v.InstanceID == nil
		})).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "volume.detach", "volume", volID.String(), mock.Anything).Return(nil).Once()

		err := svc.DetachVolume(ctx, volID.String())
		require.NoError(t, err)
	})

	t.Run("DetachVolume Repo Update Fail", func(t *testing.T) {
		volID := uuid.New()
		instID := uuid.New()
		vol := &domain.Volume{ID: volID, Status: domain.VolumeStatusInUse, InstanceID: &instID, UserID: userID}

		mockRepo.On("GetByID", mock.Anything, volID).Return(vol, nil).Once()
		mockStorage.On("DetachVolume", mock.Anything, mock.Anything, instID.String()).Return(nil).Once()
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(errors.New("db update fail")).Once()

		err := svc.DetachVolume(ctx, volID.String())
		assert.Error(t, err)
	})

	t.Run("DetachVolume Storage Fail", func(t *testing.T) {
		volID := uuid.New()
		instID := uuid.New()
		vol := &domain.Volume{ID: volID, Status: domain.VolumeStatusInUse, InstanceID: &instID, UserID: userID}

		mockRepo.On("GetByID", mock.Anything, volID).Return(vol, nil).Once()
		mockStorage.On("DetachVolume", mock.Anything, mock.Anything, instID.String()).Return(errors.New("storage fail")).Once()

		err := svc.DetachVolume(ctx, volID.String())
		assert.Error(t, err)
	})

	t.Run("DetachVolume Not Attached", func(t *testing.T) {
		volID := uuid.New()
		vol := &domain.Volume{ID: volID, Status: domain.VolumeStatusAvailable}
		mockRepo.On("GetByID", mock.Anything, volID).Return(vol, nil).Once()
		err := svc.DetachVolume(ctx, volID.String())
		assert.Error(t, err)
	})

	t.Run("ReleaseVolumesForInstance", func(t *testing.T) {
		instID := uuid.New()
		vol := &domain.Volume{ID: uuid.New(), Status: domain.VolumeStatusInUse, InstanceID: &instID}
		mockRepo.On("ListByInstanceID", mock.Anything, instID).Return([]*domain.Volume{vol}, nil).Once()
		mockStorage.On("DetachVolume", mock.Anything, mock.Anything, instID.String()).Return(nil).Once()
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(v *domain.Volume) bool {
			return v.Status == domain.VolumeStatusAvailable && v.InstanceID == nil
		})).Return(nil).Once()

		err := svc.ReleaseVolumesForInstance(ctx, instID)
		require.NoError(t, err)
	})

	t.Run("ReleaseVolumesForInstance List Fail", func(t *testing.T) {
		instID := uuid.New()
		mockRepo.On("ListByInstanceID", mock.Anything, instID).Return(nil, errors.New("list fail")).Once()
		err := svc.ReleaseVolumesForInstance(ctx, instID)
		assert.Error(t, err)
	})

	t.Run("ReleaseVolumesForInstance Partial Update Fail", func(t *testing.T) {
		instID := uuid.New()
		vol1 := &domain.Volume{ID: uuid.New(), Status: domain.VolumeStatusInUse, InstanceID: &instID}
		vol2 := &domain.Volume{ID: uuid.New(), Status: domain.VolumeStatusInUse, InstanceID: &instID}
		
		mockRepo.On("ListByInstanceID", mock.Anything, instID).Return([]*domain.Volume{vol1, vol2}, nil).Once()
		mockStorage.On("DetachVolume", mock.Anything, mock.Anything, instID.String()).Return(nil).Twice()
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(errors.New("update fail")).Once() // Fail first
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once() // Succeed second

		err := svc.ReleaseVolumesForInstance(ctx, instID)
		assert.NoError(t, err) // It continues on partial failure
	})
}
