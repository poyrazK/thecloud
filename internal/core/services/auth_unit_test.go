package services_test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockDB struct {
	mock.Mock
}

func (m *mockDB) Begin(ctx context.Context) (services.Transaction, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(services.Transaction), args.Error(1)
}

type mockTransaction struct {
	mock.Mock
}

func (m *mockTransaction) Commit(ctx context.Context) error    { return m.Called(ctx).Error(0) }
func (m *mockTransaction) Rollback(ctx context.Context) error    { return m.Called(ctx).Error(0) }

func TestAuthService_Unit(t *testing.T) {
	t.Run("Extended", testAuthServiceUnitExtended)
}

func testAuthServiceUnitExtended(t *testing.T) {
	mockUserRepo := new(MockUserRepo)
	mockIdentitySvc := new(MockIdentityService)
	mockAuditSvc := new(MockAuditService)
	mockTenantSvc := new(MockTenantService)
	mockDB := new(mockDB)
	svc := services.NewAuthService(mockUserRepo, mockIdentitySvc, mockAuditSvc, mockTenantSvc, mockDB, slog.Default())

	ctx := context.Background()

	t.Run("Register_WeakPassword", func(t *testing.T) {
		_, err := svc.Register(ctx, "test@example.com", "weak", "Test User")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "password is too weak")
	})

	t.Run("Register_ExistingEmail", func(t *testing.T) {
		email := "existing@example.com"
		mockUserRepo.On("GetByEmail", mock.Anything, email).Return(&domain.User{ID: uuid.New()}, nil).Once()

		_, err := svc.Register(ctx, email, testutil.TestPasswordStrong, "Test User")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("Register_TenantFailureRollback", func(t *testing.T) {
		email := "rollback@example.com"
		mockUserRepo.On("GetByEmail", mock.Anything, email).Return(nil, nil).Once()
		mockDB.On("Begin", mock.Anything).Return(nil, fmt.Errorf("begin failed")).Once()

		_, err := svc.Register(ctx, email, testutil.TestPasswordStrong, "Rollback User")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "begin failed")
		mockUserRepo.AssertExpectations(t)
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
		pass := testutil.TestPasswordStrong
		mockUserRepo.On("GetByEmail", mock.Anything, email).Return(nil, fmt.Errorf("db fail")).Once()

		_, _, err := svc.Login(ctx, email, pass)
		require.Error(t, err)
	})
}

func TestAuthService_RegisterWithTransactionRollback(t *testing.T) {
	mockUserRepo := new(MockUserRepo)
	mockIdentitySvc := new(MockIdentityService)
	mockAuditSvc := new(MockAuditService)
	mockTenantSvc := new(MockTenantService)

	mockTx := new(mockTransaction)
	mockDB := new(mockDB)
	mockDB.On("Begin", mock.Anything).Return(mockTx, nil)

	svc := services.NewAuthService(mockUserRepo, mockIdentitySvc, mockAuditSvc, mockTenantSvc, mockDB, slog.Default())

	ctx := context.Background()
	email := "trans@example.com"

	mockUserRepo.On("GetByEmail", mock.Anything, email).Return(nil, nil).Once()
	mockUserRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
	mockTenantSvc.On("CreateTenant", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("tenant fail")).Once()
	mockTx.On("Rollback", mock.Anything).Return(nil).Once()

	_, err := svc.Register(ctx, email, testutil.TestPasswordStrong, "Trans User")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tenant fail")
	mockUserRepo.AssertExpectations(t)
	mockTenantSvc.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}
