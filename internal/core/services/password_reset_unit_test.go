package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockPasswordResetRepo struct{ mock.Mock }

func (m *MockPasswordResetRepo) Create(ctx context.Context, t *domain.PasswordResetToken) error {
	return m.Called(ctx, t).Error(0)
}
func (m *MockPasswordResetRepo) GetByTokenHash(ctx context.Context, hash string) (*domain.PasswordResetToken, error) {
	args := m.Called(ctx, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PasswordResetToken), args.Error(1)
}
func (m *MockPasswordResetRepo) MarkAsUsed(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockPasswordResetRepo) DeleteExpired(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func TestPasswordResetService_Unit(t *testing.T) {
	mockRepo := new(MockPasswordResetRepo)
	mockUserRepo := new(MockUserRepo)
	svc := services.NewPasswordResetService(mockRepo, mockUserRepo, nil)

	ctx := context.Background()

	t.Run("RequestReset_Success", func(t *testing.T) {
		email := "test@example.com"
		user := &domain.User{ID: uuid.New(), Email: email}
		mockUserRepo.On("GetByEmail", mock.Anything, email).Return(user, nil).Once()
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()

		err := svc.RequestReset(ctx, email)
		require.NoError(t, err)
	})

	t.Run("RequestReset_UserNotFound", func(t *testing.T) {
		email := "unknown@example.com"
		mockUserRepo.On("GetByEmail", mock.Anything, email).Return(nil, errors.New(errors.NotFound, "not found")).Once()

		err := svc.RequestReset(ctx, email)
		require.NoError(t, err) // Should be masked
	})

	t.Run("ResetPassword_InvalidToken", func(t *testing.T) {
		mockRepo.On("GetByTokenHash", mock.Anything, mock.Anything).Return(nil, assert.AnError).Once()
		err := svc.ResetPassword(ctx, "invalid", "new-pass")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid")
	})

	t.Run("ResetPassword_UsedToken", func(t *testing.T) {
		token := &domain.PasswordResetToken{Used: true}
		mockRepo.On("GetByTokenHash", mock.Anything, mock.Anything).Return(token, nil).Once()
		err := svc.ResetPassword(ctx, "used", "new-pass")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already used")
	})

	t.Run("ResetPassword_ExpiredToken", func(t *testing.T) {
		token := &domain.PasswordResetToken{
			ExpiresAt: time.Now().Add(-time.Hour),
			Used:      false,
		}
		mockRepo.On("GetByTokenHash", mock.Anything, mock.Anything).Return(token, nil).Once()
		err := svc.ResetPassword(ctx, "expired", "new-pass")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})

	t.Run("ResetPassword_Success", func(t *testing.T) {
		userID := uuid.New()
		token := &domain.PasswordResetToken{
			ID:        uuid.New(),
			UserID:    userID,
			ExpiresAt: time.Now().Add(time.Hour),
			Used:      false,
		}
		user := &domain.User{ID: userID}

		mockRepo.On("GetByTokenHash", mock.Anything, mock.Anything).Return(token, nil).Once()
		mockUserRepo.On("GetByID", mock.Anything, userID).Return(user, nil).Once()
		mockUserRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
		mockRepo.On("MarkAsUsed", mock.Anything, token.ID.String()).Return(nil).Once()

		err := svc.ResetPassword(ctx, "valid-token", "new-password")
		require.NoError(t, err)
	})
}
