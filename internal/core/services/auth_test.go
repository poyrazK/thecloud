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

func TestAuthService_Register(t *testing.T) {
	t.Parallel()
	_, svc, userRepo, _ := setupAuthServiceTest(t)
	ctx := context.Background()

	email := "new@example.com"
	pass := "password123"
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

func TestAuthService_Login(t *testing.T) {
	t.Parallel()
	_, svc, _, _ := setupAuthServiceTest(t)
	ctx := context.Background()

	email := "login@example.com"
	pass := "password123"
	name := "Login User"

	_, err := svc.Register(ctx, email, pass, name)
	require.NoError(t, err)

	token, user, err := svc.Login(ctx, email, pass)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, email, user.Email)
}

func TestAuthService_LoginInvalidCredentials(t *testing.T) {
	t.Parallel()
	_, svc, _, _ := setupAuthServiceTest(t)
	ctx := context.Background()

	email := "wrong@example.com"
	_, err := svc.Register(ctx, email, "password123", "User")
	require.NoError(t, err)

	_, _, err = svc.Login(ctx, email, "wrongpass")
	assert.Error(t, err)
}

func TestAuthService_LoginUserNotFound(t *testing.T) {
	t.Parallel()
	_, svc, _, _ := setupAuthServiceTest(t)
	ctx := context.Background()

	_, _, err := svc.Login(ctx, "nonexistent@example.com", "pass")
	assert.Error(t, err)
}

func TestAuthService_ValidateToken(t *testing.T) {
	t.Parallel()
	_, svc, _, identitySvc := setupAuthServiceTest(t)
	ctx := context.Background()

	email := "token@example.com"
	user, err := svc.Register(ctx, email, "pass", "User")
	require.NoError(t, err)

	apiKey, err := identitySvc.CreateKey(ctx, user.ID, "session")
	require.NoError(t, err)

	validatedUser, err := svc.ValidateToken(ctx, apiKey.Key)
	require.NoError(t, err)
	assert.Equal(t, user.ID, validatedUser.ID)
}

func TestAuthService_RevokeToken(t *testing.T) {
	t.Parallel()
	_, svc, _, identitySvc := setupAuthServiceTest(t)
	ctx := context.Background()

	email := "revoke@example.com"
	user, err := svc.Register(ctx, email, "pass", "User")
	require.NoError(t, err)

	apiKey, err := identitySvc.CreateKey(ctx, user.ID, "session")
	require.NoError(t, err)

	err = svc.RevokeToken(ctx, user.ID, apiKey.ID)
	assert.NoError(t, err)

	_, err = svc.ValidateToken(ctx, apiKey.Key)
	assert.Error(t, err)
}

func TestAuthService_RotateToken(t *testing.T) {
	t.Parallel()
	_, svc, _, identitySvc := setupAuthServiceTest(t)
	ctx := context.Background()

	email := "rotate@example.com"
	user, err := svc.Register(ctx, email, "pass", "User")
	require.NoError(t, err)

	apiKey, err := identitySvc.CreateKey(ctx, user.ID, "session")
	require.NoError(t, err)

	newToken, err := svc.RotateToken(ctx, user.ID, apiKey.ID)
	require.NoError(t, err)
	assert.NotEqual(t, apiKey.Key, newToken.Key)

	// Old token should be invalid
	_, err = svc.ValidateToken(ctx, apiKey.Key)
	assert.Error(t, err)

	// New token should be valid
	validatedUser, err := svc.ValidateToken(ctx, newToken.Key)
	require.NoError(t, err)
	assert.Equal(t, user.ID, validatedUser.ID)
}

func TestAuthService_Logout(t *testing.T) {
	t.Parallel()
	_, svc, _, _ := setupAuthServiceTest(t)
	ctx := context.Background()

	email := "logout@example.com"
	_, err := svc.Register(ctx, email, "pass", "User")
	require.NoError(t, err)

	token, _, err := svc.Login(ctx, email, "pass")
	require.NoError(t, err)

	err = svc.Logout(ctx, token)
	assert.NoError(t, err)

	_, err = svc.ValidateToken(ctx, token)
	assert.Error(t, err)
}

func TestAuthService_TokenRotationIntegration(t *testing.T) {
	t.Parallel()
	db, svc, _, identitySvc := setupAuthServiceTest(t)
	defer db.Close()

	ctx := context.Background()
	user, err := svc.Register(ctx, "rotate-int@example.com", "pass", "User")
	require.NoError(t, err)

	// Initial token
	token1, err := identitySvc.CreateKey(ctx, user.ID, "session")
	require.NoError(t, err)

	// Rotate
	token2, err := svc.RotateToken(ctx, user.ID, token1.ID)
	require.NoError(t, err)

	// Verify
	_, err = svc.ValidateToken(ctx, token1.Key)
	assert.Error(t, err)

	vUser, err := svc.ValidateToken(ctx, token2.Key)
	require.NoError(t, err)
	assert.Equal(t, user.ID, vUser.ID)
}
