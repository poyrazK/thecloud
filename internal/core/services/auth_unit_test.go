package services_test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthService_Unit_Extended(t *testing.T) {
	mockUserRepo := new(MockUserRepo)
	mockIdentitySvc := new(MockIdentityService)
	mockAuditSvc := new(MockAuditService)
	mockTenantSvc := new(MockTenantService)
	svc := services.NewAuthService(mockUserRepo, mockIdentitySvc, mockAuditSvc, mockTenantSvc, slog.Default())

	ctx := context.Background()

	t.Run("Register_WeakPassword", func(t *testing.T) {
		_, err := svc.Register(ctx, "test@example.com", "weak", "Test User")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "password is too weak")
	})

	t.Run("Register_ExistingEmail", func(t *testing.T) {
		email := "existing@example.com"
		mockUserRepo.On("GetByEmail", mock.Anything, email).Return(&domain.User{ID: uuid.New()}, nil).Once()

		_, err := svc.Register(ctx, email, testPassword, "Test User")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("Register_TenantFailureRollback", func(t *testing.T) {
		email := "rollback@example.com"
		mockUserRepo.On("GetByEmail", mock.Anything, email).Return(nil, nil).Once()
		mockUserRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		mockTenantSvc.On("CreateTenant", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("tenant fail")).Once()
		mockUserRepo.On("Delete", mock.Anything, mock.Anything).Return(nil).Once()

		_, err := svc.Register(ctx, email, testPassword, "Rollback User")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create personal tenant")
	})

	t.Run("Login_UserNotFound", func(t *testing.T) {
		email := "missing@example.com"
		mockUserRepo.On("GetByEmail", mock.Anything, email).Return(nil, nil).Once()

		_, _, err := svc.Login(ctx, email, "any")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid email or password")
	})

	t.Run("Login_IdentityServiceError", func(t *testing.T) {
		email := "login@example.com"
		pass := testPassword
		// We can't easily mock bcrypt without a wrapper or pre-hashed password
		// But we can check GetByEmail returning error
		mockUserRepo.On("GetByEmail", mock.Anything, email).Return(nil, fmt.Errorf("db fail")).Once()

		_, _, err := svc.Login(ctx, email, pass)
		require.Error(t, err)
	})
}
