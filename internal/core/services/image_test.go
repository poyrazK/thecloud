package services_test

import (
	"context"
	"log/slog"
	"strings"
	"testing"

	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupImageServiceTest(t *testing.T) (ports.ImageService, ports.ImageRepository, *MockFileStore, context.Context) {
	t.Helper()
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewImageRepository(db)
	fileStore := new(MockFileStore)
	rbacSvc := new(MockRBACService)
	rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	svc := services.NewImageService(services.ImageServiceParams{
		Repo:      repo,
		RBACSvc:   rbacSvc,
		FileStore: fileStore,
		Logger:    slog.Default(),
	})

	return svc, repo, fileStore, ctx
}

func TestImageService_Lifecycle(t *testing.T) {
	svc, repo, fileStore, ctx := setupImageServiceTest(t)
	userID := appcontext.UserIDFromContext(ctx)

	t.Run("Create and List Images", func(t *testing.T) {
		name := "ubuntu-22.04"
		img, err := svc.RegisterImage(ctx, name, "Ubuntu 22.04 LTS", "linux", "amd64", false)
		require.NoError(t, err)
		assert.NotNil(t, img)
		assert.Equal(t, name, img.Name)
		assert.Equal(t, domain.ImageStatusPending, img.Status)

		// List
		images, err := svc.ListImages(ctx, userID, true)
		require.NoError(t, err)
		assert.NotEmpty(t, images)
	})

	t.Run("Upload and Delete Image", func(t *testing.T) {
		img, _ := svc.RegisterImage(ctx, "delete-me", "desc", "linux", "amd64", false)

		// Mock FileStore Write
		fileStore.On("Write", mock.Anything, "images", mock.Anything, mock.Anything).Return(int64(1024), nil).Once()

		content := "fake image binary data"
		err := svc.UploadImage(ctx, img.ID, strings.NewReader(content))
		require.NoError(t, err)

		// Verify Active
		updated, _ := repo.GetByID(ctx, img.ID)
		assert.Equal(t, domain.ImageStatusActive, updated.Status)

		// Delete
		fileStore.On("Delete", mock.Anything, "images", mock.Anything).Return(nil).Once()

		err = svc.DeleteImage(ctx, img.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, img.ID)
		require.Error(t, err)
	})
}
