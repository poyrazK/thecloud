package services_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockImageRepo struct {
	mock.Mock
}

func (m *mockImageRepo) Create(ctx context.Context, img *domain.Image) error {
	return m.Called(ctx, img).Error(0)
}

func (m *mockImageRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Image, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Image), args.Error(1)
}

func (m *mockImageRepo) List(ctx context.Context, userID uuid.UUID, includePublic bool) ([]*domain.Image, error) {
	args := m.Called(ctx, userID, includePublic)
	return args.Get(0).([]*domain.Image), args.Error(1)
}

func (m *mockImageRepo) Update(ctx context.Context, img *domain.Image) error {
	return m.Called(ctx, img).Error(0)
}

func (m *mockImageRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

type mockFileStore struct {
	mock.Mock
}

func (m *mockFileStore) Write(ctx context.Context, bucket, key string, r io.Reader) (int64, error) {
	args := m.Called(ctx, bucket, key, r)
	val := args.Get(0)
	if v, ok := val.(int); ok {
		return int64(v), args.Error(1)
	}
	return val.(int64), args.Error(1)
}

func (m *mockFileStore) Read(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	args := m.Called(ctx, bucket, key)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *mockFileStore) Delete(ctx context.Context, bucket, key string) error {
	return m.Called(ctx, bucket, key).Error(0)
}

func (m *mockFileStore) GetClusterStatus(ctx context.Context) (*domain.StorageCluster, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.StorageCluster), args.Error(1)
}

func (m *mockFileStore) Assemble(ctx context.Context, bucket, key string, parts []string) (int64, error) {
	args := m.Called(ctx, bucket, key, parts)
	return args.Get(0).(int64), args.Error(1)
}

func TestImageService(t *testing.T) {
	repo := new(mockImageRepo)
	store := new(mockFileStore)
	svc := services.NewImageService(repo, store, nil)
	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)

	t.Run("RegisterImage", func(t *testing.T) {
		repo.On("Create", mock.Anything, mock.Anything).Return(nil)
		img, err := svc.RegisterImage(ctx, "ubuntu", "Ubuntu 22.04", "linux", "22.04", true)
		assert.NoError(t, err)
		assert.NotNil(t, img)
		assert.Equal(t, "ubuntu", img.Name)
		repo.AssertExpectations(t)
	})

	t.Run("UploadImage", func(t *testing.T) {
		id := uuid.New()
		img := &domain.Image{ID: id, UserID: userID}
		repo.On("GetByID", mock.Anything, id).Return(img, nil)
		store.On("Write", mock.Anything, "images", mock.Anything, mock.Anything).Return(1024, nil)
		repo.On("Update", mock.Anything, img).Return(nil)

		err := svc.UploadImage(ctx, id, strings.NewReader("fake content"))
		assert.NoError(t, err)
		assert.Equal(t, domain.ImageStatusActive, img.Status)
	})

	t.Run("GetImage", func(t *testing.T) {
		id := uuid.New()
		expected := &domain.Image{ID: id}
		repo.On("GetByID", mock.Anything, id).Return(expected, nil)

		img, err := svc.GetImage(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, id, img.ID)
	})

	t.Run("ListImages", func(t *testing.T) {
		repo.On("List", mock.Anything, userID, true).Return([]*domain.Image{{ID: uuid.New()}}, nil)
		imgs, err := svc.ListImages(ctx, userID, true)
		assert.NoError(t, err)
		assert.Len(t, imgs, 1)
	})

	t.Run("DeleteImage", func(t *testing.T) {
		id := uuid.New()
		img := &domain.Image{ID: id, UserID: userID, FilePath: "key1"}
		repo.On("GetByID", mock.Anything, id).Return(img, nil)
		store.On("Delete", mock.Anything, "images", "key1").Return(nil)
		repo.On("Delete", mock.Anything, id).Return(nil)

		err := svc.DeleteImage(ctx, id)
		assert.NoError(t, err)
	})
}
