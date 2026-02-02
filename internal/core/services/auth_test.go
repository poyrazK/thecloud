package services_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAuthServiceTest(t *testing.T) (*services.AuthService, *postgres.UserRepo, *services.IdentityService, *services.AuditService, *services.TenantService) {
	db := setupDB(t)
	cleanDB(t, db)

	userRepo := postgres.NewUserRepo(db)
	auditRepo := postgres.NewAuditRepository(db)
	identityRepo := postgres.NewIdentityRepository(db)
	tenantRepo := postgres.NewTenantRepo(db)

	auditSvc := services.NewAuditService(auditRepo)
	identitySvc := services.NewIdentityService(identityRepo, auditSvc)
	tenantSvc := services.NewTenantService(tenantRepo, userRepo, slog.Default())
	svc := services.NewAuthService(userRepo, identitySvc, auditSvc, tenantSvc)

	return svc, userRepo, identitySvc, auditSvc, tenantSvc
}

func TestAuthServiceRegisterSuccess(t *testing.T) {
	svc, userRepo, _, _, _ := setupAuthServiceTest(t)
	ctx := context.Background()

	email := "test@example.com"
	password := testutil.TestPasswordStrong
	name := "Test User"

	user, err := svc.Register(ctx, email, password, name)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, name, user.Name)

	// Verify in DB
	fetched, err := userRepo.GetByEmail(ctx, email)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, fetched.ID)

	// Verify default tenant was created (as indicated by DefaultTenantID being set)
	assert.NotNil(t, fetched.DefaultTenantID)
}

func TestAuthServiceRegisterWeakPassword(t *testing.T) {
	svc, _, _, _, _ := setupAuthServiceTest(t)
	ctx := context.Background()

	user, err := svc.Register(ctx, "test@example.com", testutil.TestPasswordWeak, "User")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "password is too weak")
}

func TestAuthServiceRegisterDuplicateEmail(t *testing.T) {
	svc, _, _, _, _ := setupAuthServiceTest(t)
	ctx := context.Background()

	email := "existing@example.com"
	_, err := svc.Register(ctx, email, testutil.TestPasswordStrong, "name")
	require.NoError(t, err)

	user, err := svc.Register(ctx, email, testutil.TestPasswordStrong, "name")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "already exists")
}

func TestAuthServiceLoginSuccess(t *testing.T) {
	svc, _, identitySvc, _, _ := setupAuthServiceTest(t)
	ctx := context.Background()

	email := "login@example.com"
	password := testutil.TestPasswordStrong
	name := "Login User"

	user, err := svc.Register(ctx, email, password, name)
	require.NoError(t, err)

	resultUser, apiKey, err := svc.Login(ctx, email, password)

	assert.NoError(t, err)
	assert.NotNil(t, resultUser)
	assert.Equal(t, user.ID, resultUser.ID)
	assert.NotEmpty(t, apiKey)

	// Verify API key is valid
	key, err := identitySvc.ValidateAPIKey(ctx, apiKey)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, key.UserID)
}

func TestAuthServiceLoginWrongPassword(t *testing.T) {
	svc, _, _, _, _ := setupAuthServiceTest(t)
	ctx := context.Background()

	email := "wrong@example.com"
	password := testutil.TestPasswordStrong
	_, err := svc.Register(ctx, email, password, "name")
	require.NoError(t, err)

	resultUser, apiKey, err := svc.Login(ctx, email, "wrong-password")

	assert.Error(t, err)
	assert.Nil(t, resultUser)
	assert.Empty(t, apiKey)
	assert.Contains(t, err.Error(), "invalid")
}

func TestAuthServiceValidateUser(t *testing.T) {
	svc, _, _, _, _ := setupAuthServiceTest(t)
	ctx := context.Background()

	user, err := svc.Register(ctx, "val@example.com", testutil.TestPasswordStrong, "User")
	require.NoError(t, err)

	result, err := svc.ValidateUser(ctx, user.ID)

	assert.NoError(t, err)
	assert.Equal(t, user.ID, result.ID)
}

func TestAuthServiceLoginUserNotFound(t *testing.T) {
	svc, _, _, _, _ := setupAuthServiceTest(t)
	ctx := context.Background()

	resultUser, apiKey, err := svc.Login(ctx, "notfound@example.com", "anypassword")

	assert.Error(t, err)
	assert.Nil(t, resultUser)
	assert.Empty(t, apiKey)
	assert.Contains(t, err.Error(), "invalid")
}

func TestAuthServiceLoginAccountLockout(t *testing.T) {
	svc, _, _, _, _ := setupAuthServiceTest(t)
	ctx := context.Background()

	email := "lockout@example.com"
	password := testutil.TestPasswordStrong
	_, err := svc.Register(ctx, email, password, "User")
	require.NoError(t, err)

	// Trigger 5 failed login attempts
	for i := 0; i < 5; i++ {
		_, _, err := svc.Login(ctx, email, "wrong-password")
		assert.Error(t, err)
	}

	// The 6th attempt should be locked out
	resultUser, apiKey, err := svc.Login(ctx, email, password)

	assert.Error(t, err)
	assert.Nil(t, resultUser)
	assert.Empty(t, apiKey)
	assert.Contains(t, err.Error(), "locked")
}
