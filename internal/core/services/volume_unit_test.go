package services_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestVolumeService_Unit(t *testing.T) {
	mockRepo := new(MockVolumeRepo)
	mockStorage := new(MockStorageBackend)
	mockEventSvc := new(MockEventService)
	mockAuditSvc := new(MockAuditService)
	svc := services.NewVolumeService(mockRepo, mockStorage, mockEventSvc, mockAuditSvc, slog.Default())

	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)

	t.Run("CreateVolume", func(t *testing.T) {
		mockStorage.On("CreateVolume", mock.Anything, mock.Anything, 10).Return("/dev/vdb", nil).Once()
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		mockEventSvc.On("RecordEvent", mock.Anything, "VOLUME_CREATE", mock.Anything, "VOLUME", mock.Anything).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "volume.create", "volume", mock.Anything, mock.Anything).Return(nil).Once()

		vol, err := svc.CreateVolume(ctx, "test-vol", 10)
		assert.NoError(t, err)
		assert.NotNil(t, vol)
		assert.Equal(t, "/dev/vdb", vol.BackendPath)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ReleaseVolumesForInstance", func(t *testing.T) {
		instID := uuid.New()
		vol := &domain.Volume{ID: uuid.New(), Status: domain.VolumeStatusInUse, InstanceID: &instID}
		mockRepo.On("ListByInstanceID", mock.Anything, instID).Return([]*domain.Volume{vol}, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(v *domain.Volume) bool {
			return v.Status == domain.VolumeStatusAvailable && v.InstanceID == nil
		})).Return(nil).Once()

		err := svc.ReleaseVolumesForInstance(ctx, instID)
		assert.NoError(t, err)
	})
}
