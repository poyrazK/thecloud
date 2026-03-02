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
		tests := []struct {
			name      string
			volName   string
			sizeGB    int
			setup     func()
			wantError bool
		}{
			{
				name:    "Success",
				volName: "test-vol",
				sizeGB:  10,
				setup: func() {
					mockStorage.On("CreateVolume", mock.Anything, mock.Anything, 10).Return("/dev/vdb", nil).Once()
					mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
					mockEventSvc.On("RecordEvent", mock.Anything, "VOLUME_CREATE", mock.Anything, "VOLUME", mock.Anything).Return(nil).Once()
					mockAuditSvc.On("Log", mock.Anything, userID, "volume.create", "volume", mock.Anything, mock.Anything).Return(nil).Once()
				},
				wantError: false,
			},
			{
				name:    "Storage Error",
				volName: "test-vol",
				sizeGB:  10,
				setup: func() {
					mockStorage.On("CreateVolume", mock.Anything, mock.Anything, 10).Return("", errors.New("storage fail")).Once()
				},
				wantError: true,
			},
			{
				name:    "Repo Error Rollback",
				volName: "test-vol",
				sizeGB:  10,
				setup: func() {
					mockStorage.On("CreateVolume", mock.Anything, mock.Anything, 10).Return("/dev/vdb", nil).Once()
					mockRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db fail")).Once()
					mockStorage.On("DeleteVolume", mock.Anything, mock.Anything).Return(nil).Once()
				},
				wantError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tt.setup()
				vol, err := svc.CreateVolume(ctx, tt.volName, tt.sizeGB)
				if tt.wantError {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
					assert.NotNil(t, vol)
					assert.Equal(t, "/dev/vdb", vol.BackendPath)
				}
			})
		}
	})

	t.Run("GetAndList", func(t *testing.T) {
		t.Run("ListVolumes", func(t *testing.T) {
			mockRepo.On("List", mock.Anything).Return([]*domain.Volume{{ID: uuid.New()}}, nil).Once()
			vols, err := svc.ListVolumes(ctx)
			require.NoError(t, err)
			assert.Len(t, vols, 1)
		})

		t.Run("GetVolume", func(t *testing.T) {
			id := uuid.New()
			tests := []struct {
				name     string
				idOrName string
				setup    func()
				wantID   uuid.UUID
				wantName string
			}{
				{
					name:     "By ID",
					idOrName: id.String(),
					setup: func() {
						mockRepo.On("GetByID", mock.Anything, id).Return(&domain.Volume{ID: id}, nil).Once()
					},
					wantID: id,
				},
				{
					name:     "By Name",
					idOrName: "myvol",
					setup: func() {
						mockRepo.On("GetByName", mock.Anything, "myvol").Return(&domain.Volume{Name: "myvol"}, nil).Once()
					},
					wantName: "myvol",
				},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					tt.setup()
					vol, err := svc.GetVolume(ctx, tt.idOrName)
					require.NoError(t, err)
					if tt.wantID != uuid.Nil {
						assert.Equal(t, tt.wantID, vol.ID)
					}
					if tt.wantName != "" {
						assert.Equal(t, tt.wantName, vol.Name)
					}
				})
			}
		})
	})

	t.Run("DeleteVolume", func(t *testing.T) {
		volID := uuid.New()
		tests := []struct {
			name      string
			idOrName  string
			setup     func()
			wantError bool
		}{
			{
				name:     "Success",
				idOrName: volID.String(),
				setup: func() {
					vol := &domain.Volume{ID: volID, Status: domain.VolumeStatusAvailable, UserID: userID, SizeGB: 10}
					mockRepo.On("GetByID", mock.Anything, volID).Return(vol, nil).Once()
					mockStorage.On("DeleteVolume", mock.Anything, mock.Anything).Return(nil).Once()
					mockRepo.On("Delete", mock.Anything, volID).Return(nil).Once()
					mockEventSvc.On("RecordEvent", mock.Anything, "VOLUME_DELETE", volID.String(), "VOLUME", mock.Anything).Return(nil).Once()
					mockAuditSvc.On("Log", mock.Anything, userID, "volume.delete", "volume", volID.String(), mock.Anything).Return(nil).Once()
				},
				wantError: false,
			},
			{
				name:     "In Use Error",
				idOrName: volID.String(),
				setup: func() {
					vol := &domain.Volume{ID: volID, Status: domain.VolumeStatusInUse}
					mockRepo.On("GetByID", mock.Anything, volID).Return(vol, nil).Once()
				},
				wantError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tt.setup()
				err := svc.DeleteVolume(ctx, tt.idOrName)
				if tt.wantError {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			})
		}
	})

	t.Run("AttachDetach", func(t *testing.T) {
		volID := uuid.New()
		instID := uuid.New()

		t.Run("Attach", func(t *testing.T) {
			tests := []struct {
				name      string
				setup     func()
				wantError bool
			}{
				{
					name: "Success",
					setup: func() {
						vol := &domain.Volume{ID: volID, Status: domain.VolumeStatusAvailable, UserID: userID}
						mockRepo.On("GetByID", mock.Anything, volID).Return(vol, nil).Once()
						mockStorage.On("AttachVolume", mock.Anything, mock.Anything, instID.String()).Return("/dev/vdb", nil).Once()
						mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(v *domain.Volume) bool {
							return v.Status == domain.VolumeStatusInUse && v.InstanceID != nil && *v.InstanceID == instID
						})).Return(nil).Once()
						mockAuditSvc.On("Log", mock.Anything, userID, "volume.attach", "volume", volID.String(), mock.Anything).Return(nil).Once()
					},
					wantError: false,
				},
				{
					name: "Repo Update Fail with Rollback",
					setup: func() {
						vol := &domain.Volume{ID: volID, Status: domain.VolumeStatusAvailable, UserID: userID}
						mockRepo.On("GetByID", mock.Anything, volID).Return(vol, nil).Once()
						mockStorage.On("AttachVolume", mock.Anything, mock.Anything, instID.String()).Return("/dev/vdb", nil).Once()
						mockRepo.On("Update", mock.Anything, mock.Anything).Return(errors.New("db update fail")).Once()
						mockStorage.On("DetachVolume", mock.Anything, mock.Anything, instID.String()).Return(nil).Once()
					},
					wantError: true,
				},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					tt.setup()
					path, err := svc.AttachVolume(ctx, volID.String(), instID.String(), "/mnt/data")
					if tt.wantError {
						require.Error(t, err)
					} else {
						require.NoError(t, err)
						assert.Equal(t, "/dev/vdb", path)
					}
				})
			}
		})

		t.Run("Detach Success", func(t *testing.T) {
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
