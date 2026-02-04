package services

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const testVolumeName = "test-vol"

// mockVolumeRepo is already defined in dashboard_test.go (package services)

func TestInstanceServiceInternalGetVolumeByIDOrName(t *testing.T) {
	t.Parallel()
	repo := new(mockVolumeRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := &InstanceService{volumeRepo: repo, logger: logger}
	ctx := context.Background()
	volID := uuid.New()

	t.Run("ByID", func(t *testing.T) {
		repo.On("GetByID", ctx, volID).Return(&domain.Volume{ID: volID}, nil).Once()
		res, err := svc.getVolumeByIDOrName(ctx, volID.String())
		assert.NoError(t, err)
		assert.Equal(t, volID, res.ID)
	})

	t.Run("ByName", func(t *testing.T) {
		repo.On("GetByName", ctx, testVolumeName).Return(&domain.Volume{Name: testVolumeName}, nil).Once()
		res, err := svc.getVolumeByIDOrName(ctx, testVolumeName)
		assert.NoError(t, err)
		assert.Equal(t, testVolumeName, res.Name)
	})
}

func TestInstanceServiceInternalResolveVolumes(t *testing.T) {
	t.Parallel()
	repo := new(mockVolumeRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := &InstanceService{volumeRepo: repo, logger: logger}
	ctx := context.Background()
	volID := uuid.New()

	repo.On("GetByID", ctx, volID).Return(&domain.Volume{ID: volID, Name: "vol1", Status: domain.VolumeStatusAvailable}, nil).Once()

	binds, vols, err := svc.resolveVolumes(ctx, []domain.VolumeAttachment{{VolumeIDOrName: volID.String(), MountPath: "/data"}})
	assert.NoError(t, err)
	assert.Len(t, binds, 1)
	assert.Len(t, vols, 1)
}

func TestInstanceServiceInternalResolveVolumesUnavailable(t *testing.T) {
	t.Parallel()
	repo := new(mockVolumeRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := &InstanceService{volumeRepo: repo, logger: logger}
	ctx := context.Background()
	volID := uuid.New()

	repo.On("GetByID", ctx, volID).Return(&domain.Volume{ID: volID, Name: "vol1", Status: domain.VolumeStatusInUse}, nil).Once()

	_, _, err := svc.resolveVolumes(ctx, []domain.VolumeAttachment{{VolumeIDOrName: volID.String(), MountPath: "/data"}})
	assert.Error(t, err)
}

func TestInstanceServiceInternalUpdateVolumesAfterLaunch(t *testing.T) {
	t.Parallel()
	repo := new(mockVolumeRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := &InstanceService{volumeRepo: repo, logger: logger}
	ctx := context.Background()
	instID := uuid.New()
	vol := &domain.Volume{ID: uuid.New(), Status: domain.VolumeStatusAvailable}

	repo.On("Update", ctx, mock.MatchedBy(func(v *domain.Volume) bool {
		return v.Status == domain.VolumeStatusInUse && v.InstanceID != nil && *v.InstanceID == instID
	})).Return(nil).Once()

	svc.updateVolumesAfterLaunch(ctx, []*domain.Volume{vol}, instID)
	repo.AssertExpectations(t)
}
