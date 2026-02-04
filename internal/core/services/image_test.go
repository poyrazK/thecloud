package services_test

import (
	"context"
	"strings"
	"testing"

	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/filesystem"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupImageServiceTest(t *testing.T) (ports.ImageService, ports.ImageRepository, context.Context) {
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewImageRepository(db)

	tmpDir := t.TempDir()
	store, err := filesystem.NewLocalFileStore(tmpDir)
	require.NoError(t, err)

	svc := services.NewImageService(repo, store, nil)
	return svc, repo, ctx
}

func TestImageService(t *testing.T) {
	svc, repo, ctx := setupImageServiceTest(t)
	userID := appcontext.UserIDFromContext(ctx)

	t.Run("RegisterImage", func(t *testing.T) {
		img, err := svc.RegisterImage(ctx, "ubuntu", "Ubuntu 22.04", "linux", "22.04", true)
		assert.NoError(t, err)
		assert.NotNil(t, img)
		assert.Equal(t, "ubuntu", img.Name)

		// Verify in DB
		fetched, err := repo.GetByID(ctx, img.ID)
		assert.NoError(t, err)
		assert.Equal(t, img.ID, fetched.ID)
	})

	t.Run("UploadImage", func(t *testing.T) {
		img, _ := svc.RegisterImage(ctx, "upload-test", "desc", "linux", "v1", false)
		require.NotNil(t, img)

		err := svc.UploadImage(ctx, img.ID, strings.NewReader("fake content"))
		assert.NoError(t, err)

		fetched, _ := repo.GetByID(ctx, img.ID)
		assert.Equal(t, domain.ImageStatusActive, fetched.Status)
	})

	t.Run("GetImage", func(t *testing.T) {
		img, _ := svc.RegisterImage(ctx, "get-test", "desc", "linux", "v1", false)

		res, err := svc.GetImage(ctx, img.ID)
		assert.NoError(t, err)
		assert.Equal(t, img.ID, res.ID)
	})

	t.Run("ListImages", func(t *testing.T) {
		_, _ = svc.RegisterImage(ctx, "list1", "desc", "linux", "v1", true)
		_, _ = svc.RegisterImage(ctx, "list2", "desc", "linux", "v1", false)

		imgs, err := svc.ListImages(ctx, userID, true)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(imgs), 2)
	})

	t.Run("DeleteImage", func(t *testing.T) {
		img, _ := svc.RegisterImage(ctx, "del-test", "desc", "linux", "v1", false)

		err := svc.DeleteImage(ctx, img.ID)
		assert.NoError(t, err)

		_, err = repo.GetByID(ctx, img.ID)
		assert.Error(t, err)
	})
}
