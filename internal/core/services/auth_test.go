package services_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAuthServiceTest(t *testing.T) (*pgxpool.Pool, *services.AuthService, *postgres.UserRepo, *services.IdentityService, *services.AuditService, *services.TenantService) {
	db := setupDB(t)
	cleanDB(t, db)

	userRepo := postgres.NewUserRepo(db)
	auditRepo := postgres.NewAuditRepository(db)
	identityRepo := postgres.NewIdentityRepository(db)
	tenantRepo := postgres.NewTenantRepo(db)

	auditSvc := services.NewAuditService(auditRepo)
	identitySvc := services.NewIdentityService(identityRepo, auditSvc, slog.Default())
	tenantSvc := services.NewTenantService(tenantRepo, userRepo, slog.Default())
	svc := services.NewAuthService(userRepo, identitySvc, auditSvc, tenantSvc)

	return db, svc, userRepo, identitySvc, auditSvc, tenantSvc
}

func TestAuthServiceRegisterSuccess(t *testing.T) {
	_, svc, userRepo, _, _, _ := setupAuthServiceTest(t)
	ctx := context.Background()

	email := "test-" + uuid.New().String() + "@example.com"
	password := testutil.TestPasswordStrong
	name := "Test User"

	user, err := svc.Register(ctx, email, password, name)
	require.NoError(t, err)
	require.NotNil(t, user)
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
	_, svc, _, _, _, _ := setupAuthServiceTest(t)
	ctx := context.Background()

	user, err := svc.Register(ctx, "test-"+uuid.New().String()+"@example.com", testutil.TestPasswordWeak, "User")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "password is too weak")
}

func TestAuthServiceRegisterDuplicateEmail(t *testing.T) {
	_, svc, _, _, _, _ := setupAuthServiceTest(t)
	ctx := context.Background()

	email := "existing-" + uuid.New().String() + "@example.com"
	_, err := svc.Register(ctx, email, testutil.TestPasswordStrong, "name")
	require.NoError(t, err)

	user, err := svc.Register(ctx, email, testutil.TestPasswordStrong, "name")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "already exists")
}

func TestAuthServiceLoginSuccess(t *testing.T) {
	_, svc, _, identitySvc, _, _ := setupAuthServiceTest(t)
	ctx := context.Background()

	email := "login-" + uuid.New().String() + "@example.com"
	password := testutil.TestPasswordStrong
	name := "Login User"

	user, err := svc.Register(ctx, email, password, name)
	require.NoError(t, err)

	resultUser, apiKey, err := svc.Login(ctx, email, password)
	require.NoError(t, err)
	require.NotNil(t, resultUser)
	assert.Equal(t, user.ID, resultUser.ID)
	assert.NotEmpty(t, apiKey)

	// Verify API key is valid
	key, err := identitySvc.ValidateAPIKey(ctx, apiKey)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, key.UserID)
}

func TestAuthServiceLoginWrongPassword(t *testing.T) {
	_, svc, _, _, _, _ := setupAuthServiceTest(t)
	ctx := context.Background()

	email := "wrong-" + uuid.New().String() + "@example.com"
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
	_, svc, _, _, _, _ := setupAuthServiceTest(t)
	ctx := context.Background()

	user, err := svc.Register(ctx, "val-"+uuid.New().String()+"@example.com", testutil.TestPasswordStrong, "User")
	require.NoError(t, err)

	result, err := svc.ValidateUser(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, result.ID)
}

func TestAuthServiceLoginUserNotFound(t *testing.T) {
	_, svc, _, _, _, _ := setupAuthServiceTest(t)
	ctx := context.Background()

	resultUser, apiKey, err := svc.Login(ctx, "notfound-"+uuid.New().String()+"@example.com", "anypassword")

	assert.Error(t, err)
	assert.Nil(t, resultUser)
	assert.Empty(t, apiKey)
	assert.Contains(t, err.Error(), "invalid")
}

func TestAuthServiceLoginAccountLockout(t *testing.T) {
	_, svc, _, _, _, _ := setupAuthServiceTest(t)
	ctx := context.Background()

	email := "lockout-" + uuid.New().String() + "@example.com"
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

func TestAuthService_TokenExpiry(t *testing.T) {
	db, svc, _, identitySvc, _, _ := setupAuthServiceTest(t)
	ctx := context.Background()

	email := "expiry-" + uuid.New().String() + "@example.com"
	password := testutil.TestPasswordStrong
	user, err := svc.Register(ctx, email, password, "Expiry User")
	require.NoError(t, err)

	_, apiKey, err := svc.Login(ctx, email, password)
	require.NoError(t, err)

	// Verify Valid
	_, err = identitySvc.ValidateAPIKey(ctx, apiKey)
	require.NoError(t, err)

	// Expire Token Manually
	// We update using user_id which is simpler and robust for this test
	_, err = db.Exec(ctx, "UPDATE api_keys SET expires_at = NOW() - INTERVAL '1 minute' WHERE user_id = $1", user.ID)
	require.NoError(t, err)

	// Verify Expired
	_, err = identitySvc.ValidateAPIKey(ctx, apiKey)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired") // Assuming error message
}
