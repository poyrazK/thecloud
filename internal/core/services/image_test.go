package services_test

import (
	"context"
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
	ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), uuid.New()), uuid.New())

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

	t.Run("Success", func(t *testing.T) {
		name := "test-image"
		os := "linux"
		version := "1.0"

		repo.On("Create", mock.Anything, mock.Anything).Return(nil)

		img, err := svc.RegisterImage(ctx, name, "desc", os, version, false)
		assert.NoError(t, err)
		assert.NotNil(t, img)
		assert.Equal(t, name, img.Name)
		assert.Equal(t, os, img.OS)
	})
}

func TestImageService_UploadImage(t *testing.T) {
	svc, repo, store, ctx := setupImageServiceTest(t)
	uID := appcontext.UserIDFromContext(ctx)

	t.Run("Success", func(t *testing.T) {
		id := uuid.New()
		img := &domain.Image{ID: id, UserID: uID}
		repo.On("GetByID", mock.Anything, id).Return(img, nil)
		store.On("Write", mock.Anything, "images", mock.Anything, mock.Anything).Return(int64(1024), nil)
		repo.On("Update", mock.Anything, mock.Anything).Return(nil)

		err := svc.UploadImage(ctx, id, strings.NewReader("data"))
		assert.NoError(t, err)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		id := uuid.New()
		img := &domain.Image{ID: id, UserID: uuid.New()} // different owner
		repo.On("GetByID", mock.Anything, id).Return(img, nil)

		err := svc.UploadImage(ctx, id, strings.NewReader("data"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot upload")
	})
}

func TestImageService_DeleteImage(t *testing.T) {
	svc, repo, store, ctx := setupImageServiceTest(t)
	uID := appcontext.UserIDFromContext(ctx)

	t.Run("Success", func(t *testing.T) {
		id := uuid.New()
		img := &domain.Image{ID: id, UserID: uID, FilePath: "path"}
		repo.On("GetByID", mock.Anything, id).Return(img, nil)
		store.On("Delete", mock.Anything, "images", "path").Return(nil)
		repo.On("Delete", mock.Anything, id).Return(nil)

		err := svc.DeleteImage(ctx, id)
		assert.NoError(t, err)
	})

	t.Run("NonOwner", func(t *testing.T) {
		id := uuid.New()
		img := &domain.Image{ID: id, UserID: uuid.New()} // different owner
		repo.On("GetByID", mock.Anything, id).Return(img, nil)

		err := svc.DeleteImage(ctx, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot delete someone else's image")
	})
}
