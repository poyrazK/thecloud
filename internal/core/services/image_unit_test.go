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

type MockImageRepo struct {
	mock.Mock
}

func (m *MockImageRepo) Create(ctx context.Context, img *domain.Image) error {
	return m.Called(ctx, img).Error(0)
}
func (m *MockImageRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Image, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).(*domain.Image)
	return r0, args.Error(1)
}
func (m *MockImageRepo) List(ctx context.Context, userID uuid.UUID, includePublic bool) ([]*domain.Image, error) {
	args := m.Called(ctx, userID, includePublic)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).([]*domain.Image)
	return r0, args.Error(1)
}
func (m *MockImageRepo) Update(ctx context.Context, img *domain.Image) error {
	return m.Called(ctx, img).Error(0)
}
func (m *MockImageRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func TestImageService_Unit(t *testing.T) {
	repo := new(MockImageRepo)
	fileStore := new(MockFileStore)
	svc := services.NewImageService(repo, fileStore, slog.Default())

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
