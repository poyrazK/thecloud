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
	
	defer mockRepo.AssertExpectations(t)
	defer mockStorage.AssertExpectations(t)
	defer mockEventSvc.AssertExpectations(t)
	defer mockAuditSvc.AssertExpectations(t)

	svc := services.NewVolumeService(mockRepo, mockStorage, mockEventSvc, mockAuditSvc, slog.Default())

	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)

	t.Run("CreateVolume", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			mockStorage.On("CreateVolume", mock.Anything, mock.Anything, 10).Return("/dev/vdb", nil).Once()
			mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
			mockEventSvc.On("RecordEvent", mock.Anything, "VOLUME_CREATE", mock.Anything, "VOLUME", mock.Anything).Return(nil).Once()
			mockAuditSvc.On("Log", mock.Anything, userID, "volume.create", "volume", mock.Anything, mock.Anything).Return(nil).Once()

			vol, err := svc.CreateVolume(ctx, "test-vol", 10)
			require.NoError(t, err)
			assert.NotNil(t, vol)
			assert.Equal(t, "/dev/vdb", vol.BackendPath)
		})

		t.Run("Storage Error", func(t *testing.T) {
			mockStorage.On("CreateVolume", mock.Anything, mock.Anything, 10).Return("", errors.New("storage fail")).Once()
			_, err := svc.CreateVolume(ctx, "test-vol", 10)
			require.Error(t, err)
		})

		t.Run("Repo Error Rollback", func(t *testing.T) {
			mockStorage.On("CreateVolume", mock.Anything, mock.Anything, 10).Return("/dev/vdb", nil).Once()
			mockRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db fail")).Once()
			mockStorage.On("DeleteVolume", mock.Anything, mock.Anything).Return(nil).Once()

			_, err := svc.CreateVolume(ctx, "test-vol", 10)
			require.Error(t, err)
		})
	})

	t.Run("GetAndList", func(t *testing.T) {
		t.Run("ListVolumes", func(t *testing.T) {
			mockRepo.On("List", mock.Anything).Return([]*domain.Volume{{ID: uuid.New()}}, nil).Once()
			vols, err := svc.ListVolumes(ctx)
			require.NoError(t, err)
			assert.Len(t, vols, 1)
		})

		t.Run("GetVolume by ID", func(t *testing.T) {
			id := uuid.New()
			mockRepo.On("GetByID", mock.Anything, id).Return(&domain.Volume{ID: id}, nil).Once()
			vol, err := svc.GetVolume(ctx, id.String())
			require.NoError(t, err)
			assert.Equal(t, id, vol.ID)
		})

		t.Run("GetVolume by Name", func(t *testing.T) {
			mockRepo.On("GetByName", mock.Anything, "myvol").Return(&domain.Volume{Name: "myvol"}, nil).Once()
			vol, err := svc.GetVolume(ctx, "myvol")
			require.NoError(t, err)
			assert.Equal(t, "myvol", vol.Name)
		})
	})

	t.Run("DeleteVolume", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			volID := uuid.New()
			vol := &domain.Volume{ID: volID, Status: domain.VolumeStatusAvailable, UserID: userID, SizeGB: 10}
			mockRepo.On("GetByID", mock.Anything, volID).Return(vol, nil).Once()
			mockStorage.On("DeleteVolume", mock.Anything, mock.Anything).Return(nil).Once()
			mockRepo.On("Delete", mock.Anything, volID).Return(nil).Once()
			mockEventSvc.On("RecordEvent", mock.Anything, "VOLUME_DELETE", volID.String(), "VOLUME", mock.Anything).Return(nil).Once()
			mockAuditSvc.On("Log", mock.Anything, userID, "volume.delete", "volume", volID.String(), mock.Anything).Return(nil).Once()

			err := svc.DeleteVolume(ctx, volID.String())
			require.NoError(t, err)
		})

		t.Run("In Use Error", func(t *testing.T) {
			volID := uuid.New()
			vol := &domain.Volume{ID: volID, Status: domain.VolumeStatusInUse}
			mockRepo.On("GetByID", mock.Anything, volID).Return(vol, nil).Once()
			err := svc.DeleteVolume(ctx, volID.String())
			require.Error(t, err)
		})
	})

	t.Run("AttachDetach", func(t *testing.T) {
		t.Run("Attach Success", func(t *testing.T) {
			volID := uuid.New()
			instID := uuid.New()
			vol := &domain.Volume{ID: volID, Status: domain.VolumeStatusAvailable, UserID: userID}

			mockRepo.On("GetByID", mock.Anything, volID).Return(vol, nil).Once()
			mockStorage.On("AttachVolume", mock.Anything, mock.Anything, instID.String()).Return("/dev/vdb", nil).Once()
			mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(v *domain.Volume) bool {
				return v.Status == domain.VolumeStatusInUse && v.InstanceID != nil && *v.InstanceID == instID
			})).Return(nil).Once()
			mockAuditSvc.On("Log", mock.Anything, userID, "volume.attach", "volume", volID.String(), mock.Anything).Return(nil).Once()

			path, err := svc.AttachVolume(ctx, volID.String(), instID.String(), "/mnt/data")
			require.NoError(t, err)
			assert.Equal(t, "/dev/vdb", path)
		})

		t.Run("Attach Repo Update Fail with Rollback", func(t *testing.T) {
			volID := uuid.New()
			instID := uuid.New()
			vol := &domain.Volume{ID: volID, Status: domain.VolumeStatusAvailable, UserID: userID}

			mockRepo.On("GetByID", mock.Anything, volID).Return(vol, nil).Once()
			mockStorage.On("AttachVolume", mock.Anything, mock.Anything, instID.String()).Return("/dev/vdb", nil).Once()
			mockRepo.On("Update", mock.Anything, mock.Anything).Return(errors.New("db update fail")).Once()
			mockStorage.On("DetachVolume", mock.Anything, mock.Anything, instID.String()).Return(nil).Once()

			_, err := svc.AttachVolume(ctx, volID.String(), instID.String(), "/mnt/data")
			require.Error(t, err)
		})

		t.Run("Detach Success", func(t *testing.T) {
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
}
