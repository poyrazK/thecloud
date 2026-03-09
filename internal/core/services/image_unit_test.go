package services_test

import (
	"bytes"
	"context"
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

func TestImageService_Unit(t *testing.T) {
	repo := new(MockImageRepo)
	fileStore := new(MockFileStore)
	rbacSvc := new(MockRBACService)
	rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	svc := services.NewImageService(services.ImageServiceParams{
		Repo:      repo,
		RBACSvc:   rbacSvc,
		FileStore: fileStore,
		Logger:    slog.Default(),
	})

	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)

	t.Run("RegisterImage", func(t *testing.T) {
		repo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		img, err := svc.RegisterImage(ctx, "ubuntu-custom", "desc", "linux", "22.04", false)
		require.NoError(t, err)
		assert.NotNil(t, img)
		assert.Equal(t, "ubuntu-custom", img.Name)
	})

	t.Run("UploadImage", func(t *testing.T) {
		imgID := uuid.New()
		img := &domain.Image{ID: imgID, UserID: userID}
		repo.On("GetByID", mock.Anything, imgID).Return(img, nil).Once()
		fileStore.On("Write", mock.Anything, "images", mock.Anything, mock.Anything).Return(int64(1024), nil).Once()
		repo.On("Update", mock.Anything, mock.MatchedBy(func(i *domain.Image) bool {
			return i.Status == domain.ImageStatusActive && i.SizeGB == 1
		})).Return(nil).Once()

		err := svc.UploadImage(ctx, imgID, bytes.NewReader([]byte("dummy content")))
		require.NoError(t, err)
	})
}
