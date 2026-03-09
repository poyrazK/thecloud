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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestVolumeService_EdgeCases(t *testing.T) {
	mockRepo := new(MockVolumeRepo)
	mockStorage := new(MockStorageBackend)
	mockEventSvc := new(MockEventService)
	mockAuditSvc := new(MockAuditService)
	rbacSvc := new(MockRBACService)
	rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	svc := services.NewVolumeService(services.VolumeServiceParams{
		Repo:     mockRepo,
		RBACSvc:  rbacSvc,
		Storage:  mockStorage,
		EventSvc: mockEventSvc,
		AuditSvc: mockAuditSvc,
		Logger:   slog.Default(),
	})

	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)

	type testCase struct {
		name          string
		op            func() error
		setupMock     func()
		expectedError string
	}

	volID := uuid.New()
	instID := uuid.New()

	testCases := []testCase{
		{
			name: "AttachVolume Already In Use",
			op: func() error {
				_, err := svc.AttachVolume(ctx, volID.String(), instID.String(), "/mnt/data")
				return err
			},
			setupMock: func() {
				vol := &domain.Volume{ID: volID, Status: domain.VolumeStatusInUse}
				mockRepo.On("GetByID", mock.Anything, volID).Return(vol, nil).Once()
			},
			expectedError: "already attached",
		},
		{
			name: "AttachVolume Backend Error",
			op: func() error {
				_, err := svc.AttachVolume(ctx, volID.String(), instID.String(), "/mnt/data")
				return err
			},
			setupMock: func() {
				vol := &domain.Volume{ID: volID, Status: domain.VolumeStatusAvailable}
				mockRepo.On("GetByID", mock.Anything, volID).Return(vol, nil).Once()
				mockStorage.On("AttachVolume", mock.Anything, mock.Anything, instID.String()).Return("", errors.New("backend failure")).Once()
			},
			expectedError: "backend failure",
		},
		{
			name: "DetachVolume Not Attached",
			op: func() error {
				return svc.DetachVolume(ctx, volID.String())
			},
			setupMock: func() {
				vol := &domain.Volume{ID: volID, Status: domain.VolumeStatusAvailable}
				mockRepo.On("GetByID", mock.Anything, volID).Return(vol, nil).Once()
			},
			expectedError: "not attached",
		},
		{
			name: "DeleteVolume In Use Fails",
			op: func() error {
				return svc.DeleteVolume(ctx, volID.String())
			},
			setupMock: func() {
				vol := &domain.Volume{ID: volID, Status: domain.VolumeStatusInUse}
				mockRepo.On("GetByID", mock.Anything, volID).Return(vol, nil).Once()
			},
			expectedError: "cannot delete volume that is in use",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()
			err := tc.op()
			require.Error(t, err)
			require.ErrorContains(t, err, tc.expectedError)
			mockRepo.AssertExpectations(t)
			mockStorage.AssertExpectations(t)
		})
	}
}
