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

const (
	testVolName = "test-vol"
)

func setupVolumeServiceTest(_ *testing.T) (*MockVolumeRepo, *MockStorageBackend, *MockEventService, *MockAuditService, ports.VolumeService) {
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
	name := testVolName
	size := 10

	// CreateVolume(ctx, name, size) -> (path, error)
	storage.On("CreateVolume", mock.Anything, mock.MatchedBy(func(n string) bool {
		return len(n) > 0 // Ensure some name is generated
	}), size).Return("vol-path", nil)

	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Volume")).Return(nil)
	eventSvc.On("RecordEvent", mock.Anything, "VOLUME_CREATE", mock.Anything, "VOLUME", mock.Anything).Return(nil)
	auditSvc.On("Log", mock.Anything, mock.Anything, "volume.create", "volume", mock.Anything, mock.Anything).Return(nil)

	vol, err := svc.CreateVolume(ctx, name, size)

	assert.NoError(t, err)
	assert.NotNil(t, vol)
	assert.Equal(t, name, vol.Name)
	assert.Equal(t, size, vol.SizeGB)
	assert.Equal(t, domain.VolumeStatusAvailable, vol.Status)
}

func TestVolumeServiceCreateVolumeStorageError(t *testing.T) {
	repo, storage, _, _, svc := setupVolumeServiceTest(t)
	defer storage.AssertExpectations(t)
	defer repo.AssertExpectations(t)

	ctx := appcontext.WithUserID(context.Background(), uuid.New())

	storage.On("CreateVolume", mock.Anything, mock.Anything, 5).Return("", assert.AnError)

	vol, err := svc.CreateVolume(ctx, testVolName, 5)
	assert.Error(t, err)
	assert.Nil(t, vol)
	// Should not create record when backend fails
	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestVolumeServiceCreateVolumeRepoErrorRollsBack(t *testing.T) {
	repo, storage, _, _, svc := setupVolumeServiceTest(t)
	defer storage.AssertExpectations(t)
	defer repo.AssertExpectations(t)

	ctx := appcontext.WithUserID(context.Background(), uuid.New())

	storage.On("CreateVolume", mock.Anything, mock.Anything, 5).Return("vol-path", nil)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Volume")).Return(assert.AnError)
	storage.On("DeleteVolume", mock.Anything, mock.Anything).Return(nil)

	vol, err := svc.CreateVolume(ctx, testVolName, 5)
	assert.Error(t, err)
	assert.Nil(t, vol)
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
		Name:   testVolName,
		Status: domain.VolumeStatusAvailable,
	}

	repo.On("GetByID", mock.Anything, volID).Return(vol, nil)
	dockerName := "thecloud-vol-" + volID.String()[:8]
	storage.On("DeleteVolume", mock.Anything, dockerName).Return(nil)
	repo.On("Delete", mock.Anything, volID).Return(nil)
	eventSvc.On("RecordEvent", mock.Anything, "VOLUME_DELETE", volID.String(), "VOLUME", mock.Anything).Return(nil)
	auditSvc.On("Log", mock.Anything, mock.Anything, "volume.delete", "volume", mock.Anything, mock.Anything).Return(nil)

	err := svc.DeleteVolume(ctx, volID.String())

	assert.NoError(t, err)
}

func TestVolumeServiceDeleteVolumeStorageErrorContinues(t *testing.T) {
	repo, storage, eventSvc, auditSvc, svc := setupVolumeServiceTest(t)
	defer repo.AssertExpectations(t)
	defer storage.AssertExpectations(t)
	defer eventSvc.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	ctx := context.Background()
	volID := uuid.New()
	vol := &domain.Volume{ID: volID, Name: testVolName, Status: domain.VolumeStatusAvailable}

	repo.On("GetByID", mock.Anything, volID).Return(vol, nil)
	dockerName := "thecloud-vol-" + volID.String()[:8]
	storage.On("DeleteVolume", mock.Anything, dockerName).Return(assert.AnError)
	repo.On("Delete", mock.Anything, volID).Return(nil)
	eventSvc.On("RecordEvent", mock.Anything, "VOLUME_DELETE", volID.String(), "VOLUME", mock.Anything).Return(nil)
	auditSvc.On("Log", mock.Anything, mock.Anything, "volume.delete", "volume", mock.Anything, mock.Anything).Return(nil)

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

	repo.On("GetByID", mock.Anything, volID).Return(vol, nil)

	err := svc.DeleteVolume(ctx, volID.String())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "in use")

	storage.AssertNotCalled(t, "DeleteVolume", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestVolumeServiceDeleteVolumeRepoError(t *testing.T) {
	repo, storage, _, _, svc := setupVolumeServiceTest(t)
	defer repo.AssertExpectations(t)
	defer storage.AssertExpectations(t)

	ctx := context.Background()
	volID := uuid.New()
	vol := &domain.Volume{ID: volID, Name: testVolName, Status: domain.VolumeStatusAvailable}

	repo.On("GetByID", mock.Anything, volID).Return(vol, nil)
	storage.On("DeleteVolume", mock.Anything, mock.Anything).Return(nil)
	repo.On("Delete", mock.Anything, volID).Return(assert.AnError)

	err := svc.DeleteVolume(ctx, volID.String())

	assert.Error(t, err)
}

func TestVolumeServiceDeleteVolumeGetError(t *testing.T) {
	repo, _, _, _, svc := setupVolumeServiceTest(t)
	defer repo.AssertExpectations(t)

	ctx := context.Background()
	volID := uuid.New()

	repo.On("GetByID", mock.Anything, volID).Return(nil, assert.AnError)

	err := svc.DeleteVolume(ctx, volID.String())
	assert.Error(t, err)
}

func TestVolumeServiceDeleteVolumeByName(t *testing.T) {
	repo, storage, eventSvc, auditSvc, svc := setupVolumeServiceTest(t)
	defer repo.AssertExpectations(t)
	defer storage.AssertExpectations(t)
	defer eventSvc.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	ctx := context.Background()
	volID := uuid.New()
	vol := &domain.Volume{ID: volID, Name: testVolName, Status: domain.VolumeStatusAvailable}

	repo.On("GetByName", mock.Anything, testVolName).Return(vol, nil)
	storage.On("DeleteVolume", mock.Anything, mock.Anything).Return(nil)
	repo.On("Delete", mock.Anything, volID).Return(nil)
	eventSvc.On("RecordEvent", mock.Anything, "VOLUME_DELETE", volID.String(), "VOLUME", mock.Anything).Return(nil)
	auditSvc.On("Log", mock.Anything, mock.Anything, "volume.delete", "volume", mock.Anything, mock.Anything).Return(nil)

	err := svc.DeleteVolume(ctx, testVolName)
	assert.NoError(t, err)
}
func TestVolumeServiceListVolumesSuccess(t *testing.T) {
	repo, _, _, _, svc := setupVolumeServiceTest(t)
	defer repo.AssertExpectations(t)

	ctx := context.Background()
	volumes := []*domain.Volume{{ID: uuid.New(), Name: "v1"}, {ID: uuid.New(), Name: "v2"}}
	repo.On("List", mock.Anything).Return(volumes, nil)

	result, err := svc.ListVolumes(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
}

func TestVolumeServiceGetVolume(t *testing.T) {
	repo, _, _, _, svc := setupVolumeServiceTest(t)
	defer repo.AssertExpectations(t)

	ctx := context.Background()
	volID := uuid.New()
	vol := &domain.Volume{ID: volID, Name: testVolName}

	t.Run("get by id", func(t *testing.T) {
		repo.On("GetByID", mock.Anything, volID).Return(vol, nil).Once()
		res, err := svc.GetVolume(ctx, volID.String())
		assert.NoError(t, err)
		assert.Equal(t, vol, res)
	})

	t.Run("get by name", func(t *testing.T) {
		repo.On("GetByName", mock.Anything, testVolName).Return(vol, nil).Once()
		res, err := svc.GetVolume(ctx, testVolName)
		assert.NoError(t, err)
		assert.Equal(t, vol, res)
	})
}

func TestVolumeServiceReleaseVolumesForInstance(t *testing.T) {
	repo, _, _, _, svc := setupVolumeServiceTest(t)
	defer repo.AssertExpectations(t)

	ctx := context.Background()
	instanceID := uuid.New()
	volumes := []*domain.Volume{
		{ID: uuid.New(), InstanceID: &instanceID, Status: domain.VolumeStatusInUse},
		{ID: uuid.New(), InstanceID: &instanceID, Status: domain.VolumeStatusInUse},
	}

	repo.On("ListByInstanceID", mock.Anything, instanceID).Return(volumes, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Volume")).Return(nil).Twice()

	err := svc.ReleaseVolumesForInstance(ctx, instanceID)

	assert.NoError(t, err)
	for _, v := range volumes {
		assert.Equal(t, domain.VolumeStatusAvailable, v.Status)
		assert.Nil(t, v.InstanceID)
	}
}

func TestVolumeServiceReleaseVolumesForInstanceListError(t *testing.T) {
	repo, _, _, _, svc := setupVolumeServiceTest(t)
	defer repo.AssertExpectations(t)

	ctx := context.Background()
	instanceID := uuid.New()

	repo.On("ListByInstanceID", mock.Anything, instanceID).Return(nil, assert.AnError)

	err := svc.ReleaseVolumesForInstance(ctx, instanceID)
	assert.Error(t, err)
}

func TestVolumeServiceReleaseVolumesForInstanceUpdateErrorContinues(t *testing.T) {
	repo, _, _, _, svc := setupVolumeServiceTest(t)
	defer repo.AssertExpectations(t)

	ctx := context.Background()
	instanceID := uuid.New()
	volumes := []*domain.Volume{
		{ID: uuid.New(), InstanceID: &instanceID, Status: domain.VolumeStatusInUse},
		{ID: uuid.New(), InstanceID: &instanceID, Status: domain.VolumeStatusInUse},
	}

	repo.On("ListByInstanceID", mock.Anything, instanceID).Return(volumes, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Volume")).Return(assert.AnError).Once()
	repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Volume")).Return(nil).Once()

	err := svc.ReleaseVolumesForInstance(ctx, instanceID)
	assert.NoError(t, err)
}
