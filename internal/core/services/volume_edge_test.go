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
)

func TestVolumeService_EdgeCases(t *testing.T) {
	mockRepo := new(MockVolumeRepo)
	mockStorage := new(MockStorageBackend)
	mockEventSvc := new(MockEventService)
	mockAuditSvc := new(MockAuditService)
	svc := services.NewVolumeService(mockRepo, mockStorage, mockEventSvc, mockAuditSvc, slog.Default())

	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)

	t.Run("AttachVolume Already In Use", func(t *testing.T) {
		volID := uuid.New()
		instID := uuid.New()
		vol := &domain.Volume{
			ID:     volID,
			Status: domain.VolumeStatusInUse,
		}

		mockRepo.On("GetByID", mock.Anything, volID).Return(vol, nil).Once()

		err := svc.AttachVolume(ctx, volID.String(), instID.String(), "/mnt/data")
		require.Error(t, err)
		require.ErrorContains(t, err, "already attached")
	})

	t.Run("AttachVolume Backend Error Rollback", func(t *testing.T) {
		volID := uuid.New()
		instID := uuid.New()
		vol := &domain.Volume{
			ID:     volID,
			Status: domain.VolumeStatusAvailable,
		}

		mockRepo.On("GetByID", mock.Anything, volID).Return(vol, nil).Once()
		mockStorage.On("AttachVolume", mock.Anything, mock.Anything, instID.String()).Return(errors.New("backend failure")).Once()

		err := svc.AttachVolume(ctx, volID.String(), instID.String(), "/mnt/data")
		require.Error(t, err)
		require.ErrorContains(t, err, "backend failure")
	})

	t.Run("DetachVolume Not Attached", func(t *testing.T) {
		volID := uuid.New()
		vol := &domain.Volume{
			ID:     volID,
			Status: domain.VolumeStatusAvailable,
		}

		mockRepo.On("GetByID", mock.Anything, volID).Return(vol, nil).Once()

		err := svc.DetachVolume(ctx, volID.String())
		require.Error(t, err)
		require.ErrorContains(t, err, "not attached")
	})

	t.Run("DeleteVolume In Use Fails", func(t *testing.T) {
		volID := uuid.New()
		vol := &domain.Volume{
			ID:     volID,
			Status: domain.VolumeStatusInUse,
		}

		mockRepo.On("GetByID", mock.Anything, volID).Return(vol, nil).Once()

		err := svc.DeleteVolume(ctx, volID.String())
		require.Error(t, err)
		require.ErrorContains(t, err, "cannot delete volume that is in use")
	})
}
