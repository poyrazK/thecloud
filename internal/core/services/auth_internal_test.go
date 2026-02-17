package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *MockUserRepository) List(ctx context.Context) ([]*domain.User, error) {
	return nil, nil
}
func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func TestAuthService_Internal(t *testing.T) {
	repo := new(MockUserRepository)
	svc := &AuthService{userRepo: repo}
	ctx := context.Background()
	userID := uuid.New()

	t.Run("ValidateUser", func(t *testing.T) {
		user := &domain.User{ID: userID}
		repo.On("GetByID", mock.Anything, userID).Return(user, nil).Once()
		res, err := svc.ValidateUser(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, user, res)
	})
}
