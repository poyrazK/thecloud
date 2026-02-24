package services_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testPassword = "password123ABC!@#123"

func setupAuthServiceTest(t *testing.T) (*pgxpool.Pool, *services.AuthService, *postgres.UserRepo, *services.IdentityService) {
	t.Helper()
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

	return db, svc, userRepo, identitySvc
}

func TestAuthServiceRegister(t *testing.T) {
	_, svc, userRepo, _ := setupAuthServiceTest(t)
	ctx := context.Background()

	email := "new@example.com"
	pass := testPassword
	name := "New User"

	user, err := svc.Register(ctx, email, pass, name)
	require.NoError(t, err)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, name, user.Name)

	// Verify persistence
	dbUser, err := userRepo.GetByEmail(ctx, email)
	require.NoError(t, err)
	assert.Equal(t, user.ID, dbUser.ID)
}

func TestAuthServiceLogin(t *testing.T) {
	_, svc, _, _ := setupAuthServiceTest(t)
	ctx := context.Background()

	email := "login@example.com"
	pass := testPassword
	name := "Login User"

	_, err := svc.Register(ctx, email, pass, name)
	require.NoError(t, err)

	user, token, err := svc.Login(ctx, email, pass)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, email, user.Email)
}

func TestAuthServiceLoginInvalidCredentials(t *testing.T) {
	_, svc, _, _ := setupAuthServiceTest(t)
	ctx := context.Background()

	email := "wrong@example.com"
	_, err := svc.Register(ctx, email, testPassword, "User")
	require.NoError(t, err)

	_, _, err = svc.Login(ctx, email, "wrongpass")
	require.Error(t, err)
}

func TestAuthServiceLoginUserNotFound(t *testing.T) {
	_, svc, _, _ := setupAuthServiceTest(t)

	_, _, err := svc.Login(context.Background(), "user@example.com", "wrong")
	require.Error(t, err)
}

func TestAuthServiceValidateToken(t *testing.T) {
	_, svc, _, identitySvc := setupAuthServiceTest(t)
	ctx := context.Background()

	email := "session@example.com"
	user, err := svc.Register(ctx, email, testPassword, "User")
	require.NoError(t, err)

	apiKey, err := identitySvc.CreateKey(ctx, user.ID, "session")
	require.NoError(t, err)

	validatedKey, err := identitySvc.ValidateAPIKey(ctx, apiKey.Key)
	require.NoError(t, err)
	assert.Equal(t, user.ID, validatedKey.UserID)
}

func TestAuthServiceRevokeToken(t *testing.T) {
	_, svc, _, identitySvc := setupAuthServiceTest(t)
	ctx := context.Background()

	email := "revoke@example.com"
	user, err := svc.Register(ctx, email, testPassword, "User")
	require.NoError(t, err)

	apiKey, err := identitySvc.CreateKey(ctx, user.ID, "session")
	require.NoError(t, err)

	err = identitySvc.RevokeKey(ctx, user.ID, apiKey.ID)
	require.NoError(t, err)

	_, err = identitySvc.ValidateAPIKey(ctx, apiKey.Key)
	require.Error(t, err)
}

func TestAuthServiceRotateToken(t *testing.T) {
	_, svc, _, identitySvc := setupAuthServiceTest(t)
	ctx := context.Background()

	email := "rotate@example.com"
	user, err := svc.Register(ctx, email, testPassword, "User")
	require.NoError(t, err)

	apiKey, err := identitySvc.CreateKey(ctx, user.ID, "session")
	require.NoError(t, err)

	newToken, err := identitySvc.RotateKey(ctx, user.ID, apiKey.ID)
	require.NoError(t, err)
	assert.NotEqual(t, apiKey.Key, newToken.Key)

	// Old token should be invalid
	_, err = identitySvc.ValidateAPIKey(ctx, apiKey.Key)
	require.Error(t, err)

	// New token should be valid
	validatedKey, err := identitySvc.ValidateAPIKey(ctx, newToken.Key)
	require.NoError(t, err)
	assert.Equal(t, user.ID, validatedKey.UserID)
}

func TestAuthServiceLogout(t *testing.T) {
	_, svc, _, identitySvc := setupAuthServiceTest(t)
	ctx := context.Background()

	email := "logout@example.com"
	pass := testPassword
	user, err := svc.Register(ctx, email, pass, "User")
	require.NoError(t, err)

	_, token, err := svc.Login(ctx, email, pass)
	require.NoError(t, err)

	// In current implementation, login creates a key. We need to find it to revoke it.
	keys, err := identitySvc.ListKeys(ctx, user.ID)
	require.NoError(t, err)
	require.NotEmpty(t, keys)

	err = identitySvc.RevokeKey(ctx, user.ID, keys[0].ID)
	require.NoError(t, err)

	_, err = identitySvc.ValidateAPIKey(ctx, token)
	require.Error(t, err)
}

func TestAuthServiceTokenRotationIntegration(t *testing.T) {
	db, svc, _, identitySvc := setupAuthServiceTest(t)
	defer db.Close()

	ctx := context.Background()
	user, err := svc.Register(ctx, "rotate-int@example.com", testPassword, "User")
	require.NoError(t, err)

	// Initial token
	token1, err := identitySvc.CreateKey(ctx, user.ID, "session")
	require.NoError(t, err)

	// Rotate
	token2, err := identitySvc.RotateKey(ctx, user.ID, token1.ID)
	require.NoError(t, err)

	// Verify
	_, err = identitySvc.ValidateAPIKey(ctx, token1.Key)
	require.Error(t, err)

	vKey, err := identitySvc.ValidateAPIKey(ctx, token2.Key)
	require.NoError(t, err)
	assert.Equal(t, user.ID, vKey.UserID)
}
