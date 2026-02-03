//go:build integration

package postgres

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
	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/require"
)

// SetupDB initializes the database connection for integration tests.
// It prioritizes the DATABASE_URL environment variable, falling back to a
// temporary Docker container via Testcontainers if the variable is not set.
func SetupDB(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()
	dbURL := os.Getenv("DATABASE_URL")

	if dbURL == "" {
		// Initialize a temporary PostgreSQL container if a local instance is not available.
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
	err = RunMigrations(ctx, db, slog.Default())
	require.NoError(t, err, "Failed to run migrations")

	return db
}

// SetupTestUser creates a dedicated test user and tenant context for a test run.
// It returns a context containing both the UserID and TenantID for multi-tenant isolation.
func SetupTestUser(t *testing.T, db *pgxpool.Pool) context.Context {
	ctx := context.Background()
	userRepo := NewUserRepo(db)

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

// CleanDB removes test data from the database to ensure test isolation.
// This function executes a series of DELETE operations on standard resource tables.
func CleanDB(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	queries := []string{
		"DELETE FROM cron_job_runs",
		"DELETE FROM cron_jobs",
		"DELETE FROM gateway_routes",
		"DELETE FROM deployment_containers",
		"DELETE FROM deployments",
		"DELETE FROM subscriptions",
		"DELETE FROM topics",
		"DELETE FROM queue_messages",
		"DELETE FROM queues",
		"DELETE FROM lb_targets",
		"DELETE FROM scaling_group_instances",
		"DELETE FROM scaling_policies",
		"DELETE FROM scaling_groups",
		"DELETE FROM load_balancers",
		"DELETE FROM volumes",
		"DELETE FROM instances",
		"DELETE FROM vpcs",
		"DELETE FROM tenant_members",
		"DELETE FROM tenant_quotas",
		"DELETE FROM tenants",
		// Users are intentionally preserved to ensure the established identity context
		// remains valid for subsequent test executions within the same container.
	}

	for _, q := range queries {
		_, err := db.Exec(ctx, q)
		// Suppress errors if a table does not exist (PostgreSQL error 42P01).
		// This ensures stability when migrations are partially applied.
		if err != nil {
			t.Logf("Cleanup query failed (ignoring): %s - %v", q, err)
		}
	}
}
