package services_test

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupImageServiceTest(t *testing.T) (ports.ImageService, *MockImageRepo, *MockFileStore, context.Context) {
	repo := new(MockImageRepo)
	rbacSvc := new(MockRBACService)
	rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	store := new(MockFileStore)
	uID := uuid.New()
	tID := uuid.New()
	ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), uID), tID)

	svc := services.NewImageService(services.ImageServiceParams{
		Repo:      repo,
		RBACSvc:   rbacSvc,
		FileStore: store,
		Logger:    slog.Default(),
	})

	return svc, repo, store, ctx
}

func TestImageService_RegisterImage(t *testing.T) {
	svc, repo, _, ctx := setupImageServiceTest(t)

	tests := []struct {
		name      string
		imgName   string
		os        string
		version   string
		isPublic  bool
		mockSetup func()
		wantErr   bool
	}{
		{
			name:    "Success",
			imgName: "test-image",
			os:      "linux",
			version: "1.0",
			mockSetup: func() {
				repo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			img, err := svc.RegisterImage(ctx, tt.imgName, "desc", tt.os, tt.version, tt.isPublic)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, img)
				assert.Equal(t, tt.imgName, img.Name)
				assert.Equal(t, tt.os, img.OS)
			}
		})
	}
}

func TestImageService_UploadImage(t *testing.T) {
	svc, repo, store, ctx := setupImageServiceTest(t)
	uID := appcontext.UserIDFromContext(ctx)
	tID := appcontext.TenantIDFromContext(ctx)

	tests := []struct {
		name      string
		id        uuid.UUID
		mockSetup func()
		wantErr   bool
		errMsg    string
	}{
		{
			name: "Success",
			id:   uuid.New(),
			mockSetup: func() {
				// Empty mockSetup, handled in loop below
			},
			wantErr: false,
		},
		{
			name: "Unauthorized - Different User",
			id:   uuid.New(),
			mockSetup: func() {
			},
			wantErr: true,
			errMsg:  "cannot upload",
		},
		{
			name: "Tenant Mismatch",
			id:   uuid.New(),
			mockSetup: func() {
			},
			wantErr: true,
			errMsg:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Success" {
				img := &domain.Image{ID: tt.id, UserID: uID, TenantID: &tID}
				repo.On("GetByID", mock.Anything, tt.id).Return(img, nil).Once()
				store.On("Write", mock.Anything, "images", mock.Anything, mock.Anything).Return(int64(1024), nil).Once()
				repo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
			} else if tt.name == "Unauthorized - Different User" {
				img := &domain.Image{ID: tt.id, UserID: uuid.New(), TenantID: &tID}
				repo.On("GetByID", mock.Anything, tt.id).Return(img, nil).Once()
			} else if tt.name == "Tenant Mismatch" {
				otherTID := uuid.New()
				img := &domain.Image{ID: tt.id, UserID: uID, TenantID: &otherTID}
				repo.On("GetByID", mock.Anything, tt.id).Return(img, nil).Once()
			}

			err := svc.UploadImage(ctx, tt.id, strings.NewReader("data"))
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestImageService_DeleteImage(t *testing.T) {
	svc, repo, store, ctx := setupImageServiceTest(t)
	uID := appcontext.UserIDFromContext(ctx)
	tID := appcontext.TenantIDFromContext(ctx)

	tests := []struct {
		name      string
		id        uuid.UUID
		mockSetup func(id uuid.UUID)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "Success",
			id:   uuid.New(),
			mockSetup: func(id uuid.UUID) {
				img := &domain.Image{ID: id, UserID: uID, TenantID: &tID, FilePath: "path"}
				repo.On("GetByID", mock.Anything, id).Return(img, nil).Once()
				store.On("Delete", mock.Anything, "images", "path").Return(nil).Once()
				repo.On("Delete", mock.Anything, id).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "Non-Owner Denial",
			id:   uuid.New(),
			mockSetup: func(id uuid.UUID) {
				img := &domain.Image{ID: id, UserID: uuid.New(), TenantID: &tID}
				repo.On("GetByID", mock.Anything, id).Return(img, nil).Once()
			},
			wantErr: true,
			errMsg:  "cannot delete someone else's image",
		},
		{
			name: "Tenant Isolation Violation",
			id:   uuid.New(),
			mockSetup: func(id uuid.UUID) {
				otherTID := uuid.New()
				img := &domain.Image{ID: id, UserID: uID, TenantID: &otherTID}
				repo.On("GetByID", mock.Anything, id).Return(img, nil).Once()
			},
			wantErr: true,
			errMsg:  "not found",
		},
		{
			name: "Storage Delete Failure",
			id:   uuid.New(),
			mockSetup: func(id uuid.UUID) {
				img := &domain.Image{ID: id, UserID: uID, TenantID: &tID, FilePath: "path"}
				repo.On("GetByID", mock.Anything, id).Return(img, nil).Once()
				store.On("Delete", mock.Anything, "images", "path").Return(io.ErrUnexpectedEOF).Once()
			},
			wantErr: true,
			errMsg:  "failed to delete image file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(tt.id)
			err := svc.DeleteImage(ctx, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
