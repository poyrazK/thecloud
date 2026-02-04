package services

import (
	"context"
	"log/slog"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockIdentityService struct {
	mock.Mock
}

func (m *mockIdentityService) CreateKey(ctx context.Context, userID uuid.UUID, name string) (*domain.APIKey, error) {
	args := m.Called(ctx, userID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.APIKey), args.Error(1)
}

func (m *mockIdentityService) ValidateAPIKey(ctx context.Context, key string) (*domain.APIKey, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.APIKey), args.Error(1)
}

func (m *mockIdentityService) ListKeys(ctx context.Context, userID uuid.UUID) ([]*domain.APIKey, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*domain.APIKey), args.Error(1)
}

func (m *mockIdentityService) RevokeKey(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	return m.Called(ctx, userID, id).Error(0)
}

func (m *mockIdentityService) RotateKey(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*domain.APIKey, error) {
	args := m.Called(ctx, userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.APIKey), args.Error(1)
}

func setupCachedIdentityTest(t *testing.T) (*mockIdentityService, *redis.Client, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	base := new(mockIdentityService)
	return base, client, mr
}

func TestCachedIdentityServiceValidateAPIKey(t *testing.T) {
	t.Parallel()
	base, client, _ := setupCachedIdentityTest(t)
	svc := NewCachedIdentityService(base, client, slog.Default())
	ctx := context.Background()
	key := "test-key"
	apiKey := &domain.APIKey{ID: uuid.New(), Key: key, Name: "token1"}

	t.Run("cache miss", func(t *testing.T) {
		base.On("ValidateAPIKey", mock.Anything, key).Return(apiKey, nil).Once()

		res, err := svc.ValidateAPIKey(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, apiKey.ID, res.ID)
		base.AssertExpectations(t)
	})

	t.Run("cache hit", func(t *testing.T) {
		// Should not call base service
		res, err := svc.ValidateAPIKey(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, apiKey.ID, res.ID)
		base.AssertExpectations(t)
	})

	t.Run("invalid cache data", func(t *testing.T) {
		// Corrupt cache
		client.Set(ctx, "apikey:invalid", "corrupt", 0)
		base.On("ValidateAPIKey", mock.Anything, "invalid").Return(apiKey, nil).Once()

		res, err := svc.ValidateAPIKey(ctx, "invalid")
		assert.NoError(t, err)
		assert.NotNil(t, res)
		base.AssertExpectations(t)
	})
}

func TestCachedIdentityServicePassthrough(t *testing.T) {
	t.Parallel()
	base, client, _ := setupCachedIdentityTest(t)
	svc := NewCachedIdentityService(base, client, slog.Default())
	ctx := context.Background()
	userID := uuid.New()
	keyID := uuid.New()

	t.Run("CreateKey", func(t *testing.T) {
		base.On("CreateKey", mock.Anything, userID, "name").Return(&domain.APIKey{}, nil)
		_, err := svc.CreateKey(ctx, userID, "name")
		assert.NoError(t, err)
		base.AssertExpectations(t)
	})

	t.Run("ListKeys", func(t *testing.T) {
		base.On("ListKeys", mock.Anything, userID).Return([]*domain.APIKey{}, nil)
		_, err := svc.ListKeys(ctx, userID)
		assert.NoError(t, err)
		base.AssertExpectations(t)
	})

	t.Run("RevokeKey", func(t *testing.T) {
		base.On("RevokeKey", mock.Anything, userID, keyID).Return(nil)
		err := svc.RevokeKey(ctx, userID, keyID)
		assert.NoError(t, err)
		base.AssertExpectations(t)
	})

	t.Run("RotateKey", func(t *testing.T) {
		base.On("RotateKey", mock.Anything, userID, keyID).Return(&domain.APIKey{}, nil)
		_, err := svc.RotateKey(ctx, userID, keyID)
		assert.NoError(t, err)
		base.AssertExpectations(t)
	})
}
