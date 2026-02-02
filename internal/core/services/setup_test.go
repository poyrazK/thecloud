package services_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/require"
)

func setupDB(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()
	dbURL := os.Getenv("DATABASE_URL")

	if dbURL == "" {
		container, cleanup := testutil.SetupPostgresContainer(t)
		t.Cleanup(cleanup)
		dbURL = container.ConnString
	}

	db, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err)

	err = db.Ping(ctx)
	if err != nil {
		t.Skipf("Skipping integration test: database not available: %v", err)
	}

	// Run migrations
	err = postgres.RunMigrations(ctx, db, slog.Default())
	require.NoError(t, err, "Failed to run migrations")

	return db
}

func setupTestUser(t *testing.T, db *pgxpool.Pool) context.Context {
	ctx := context.Background()
	userRepo := postgres.NewUserRepo(db)

	userID := uuid.New()
	user := &domain.User{
		ID:           userID,
		Email:        "testuser_" + userID.String() + "@test.com",
		PasswordHash: "hash",
		Name:         "Test User",
		Role:         "user",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	tenantID := uuid.New()
	slug := "test-tenant-" + tenantID.String()
	_, err = db.Exec(ctx, `
		INSERT INTO tenants (id, name, slug, owner_id, plan, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'free', 'active', NOW(), NOW())
	`, tenantID, "Test Tenant", slug, userID)
	require.NoError(t, err)

	_, err = db.Exec(ctx, `
		INSERT INTO tenant_members (tenant_id, user_id, role)
		VALUES ($1, $2, 'owner')
	`, tenantID, userID)
	require.NoError(t, err)

	_, err = db.Exec(ctx, `
		UPDATE users SET default_tenant_id = $1 WHERE id = $2
	`, tenantID, userID)
	require.NoError(t, err)

	return appcontext.WithTenantID(appcontext.WithUserID(ctx, userID), tenantID)
}

func cleanDB(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	tables := []string{
		"invocations",
		"functions",
		"instance_types",
		"instances",
		"subnets",
		"vpcs",
		"volumes",
		"load_balancers",
		"lb_targets",
		"audit_logs",
		"gateway_routes",
		"events",
		"usage_records",
		"role_permissions",
		"roles",
		"encryption_keys",
		"api_keys",
		"users",
		"tenants",
		"scaling_group_instances",
		"scaling_policies",
		"scaling_groups",
		"metrics_history",
		"deployment_containers",
		"deployments",
	}

	for _, table := range tables {
		_, _ = db.Exec(ctx, "DELETE FROM "+table+" CASCADE")
	}
}
