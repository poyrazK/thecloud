package services_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupVolumeServiceTest(t *testing.T) (*MockVolumeRepo, *MockStorageBackend, *MockEventService, *MockAuditService, ports.VolumeService) {
	repo := new(MockVolumeRepo)
	storage := new(MockStorageBackend)
	eventSvc := new(MockEventService)
	auditSvc := new(MockAuditService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := services.NewVolumeService(repo, storage, eventSvc, auditSvc, logger)
	return repo, storage, eventSvc, auditSvc, svc
}

func TestVolumeServiceCreateVolumeSuccess(t *testing.T) {
	repo, storage, eventSvc, auditSvc, svc := setupVolumeServiceTest(t)
	defer repo.AssertExpectations(t)
	defer storage.AssertExpectations(t)
	defer eventSvc.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	ctx := appcontext.WithUserID(context.Background(), uuid.New())
	name := "test-vol"
	size := 10

	// CreateVolume(ctx, name, size) -> (path, error)
	storage.On("CreateVolume", ctx, mock.MatchedBy(func(n string) bool {
		return len(n) > 0 // Ensure some name is generated
	}), size).Return("vol-path", nil)

	repo.On("Create", ctx, mock.AnythingOfType("*domain.Volume")).Return(nil)
	eventSvc.On("RecordEvent", ctx, "VOLUME_CREATE", mock.Anything, "VOLUME", mock.Anything).Return(nil)
	auditSvc.On("Log", ctx, mock.Anything, "volume.create", "volume", mock.Anything, mock.Anything).Return(nil)

	vol, err := svc.CreateVolume(ctx, name, size)

	assert.NoError(t, err)
	assert.NotNil(t, vol)
	assert.Equal(t, name, vol.Name)
	assert.Equal(t, size, vol.SizeGB)
	assert.Equal(t, domain.VolumeStatusAvailable, vol.Status)
}

func TestVolumeServiceDeleteVolumeSuccess(t *testing.T) {
	repo, storage, eventSvc, auditSvc, svc := setupVolumeServiceTest(t)
	defer repo.AssertExpectations(t)
	defer storage.AssertExpectations(t)
	defer eventSvc.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	ctx := context.Background()
	volID := uuid.New()
	vol := &domain.Volume{
		ID:     volID,
		Name:   "test-vol",
		Status: domain.VolumeStatusAvailable,
	}

	repo.On("GetByID", ctx, volID).Return(vol, nil)
	dockerName := "thecloud-vol-" + volID.String()[:8]
	storage.On("DeleteVolume", ctx, dockerName).Return(nil)
	repo.On("Delete", ctx, volID).Return(nil)
	eventSvc.On("RecordEvent", ctx, "VOLUME_DELETE", volID.String(), "VOLUME", mock.Anything).Return(nil)
	auditSvc.On("Log", ctx, mock.Anything, "volume.delete", "volume", mock.Anything, mock.Anything).Return(nil)

	err := svc.DeleteVolume(ctx, volID.String())

	assert.NoError(t, err)
}

func TestVolumeServiceDeleteVolumeInUseFails(t *testing.T) {
	repo, storage, _, _, svc := setupVolumeServiceTest(t)
	defer repo.AssertExpectations(t)
	defer storage.AssertExpectations(t)

	ctx := context.Background()
	volID := uuid.New()
	vol := &domain.Volume{
		ID:     volID,
		Status: domain.VolumeStatusInUse,
	}

	repo.On("GetByID", ctx, volID).Return(vol, nil)

	err := svc.DeleteVolume(ctx, volID.String())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "in use")

	storage.AssertNotCalled(t, "DeleteVolume", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}
