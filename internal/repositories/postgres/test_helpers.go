//go:build integration

package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
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
// It returns both the connection pool and the unique schema name created for the test.
func SetupDB(t *testing.T) (*pgxpool.Pool, string) {
	t.Helper()
	ctx := context.Background()
	dbURL := os.Getenv("DATABASE_URL")

	if dbURL == "" {
		// Initialize a temporary PostgreSQL container if a local instance is not available.
		container, cleanup := testutil.SetupPostgresContainer(t)
		t.Cleanup(cleanup)
		dbURL = container.ConnString
	}

	// Use a unique schema for this test run to allow parallel execution in CI
	schema := "test_repo_" + strings.ReplaceAll(uuid.New().String(), "-", "_")
	
	// Create base connection to initialize schema
	baseDB, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err)
	defer baseDB.Close()

	_, err = baseDB.Exec(ctx, fmt.Sprintf("CREATE SCHEMA %s", schema))
	require.NoError(t, err)

	// Configure pool to use the schema for ALL connections
	// We include 'public' for extensions like uuid-ossp
	config, err := pgxpool.ParseConfig(dbURL)
	require.NoError(t, err)
	
	if config.ConnConfig.RuntimeParams == nil {
		config.ConnConfig.RuntimeParams = make(map[string]string)
	}
	config.ConnConfig.RuntimeParams["search_path"] = fmt.Sprintf("%s, public", schema)

	db, err := pgxpool.NewWithConfig(ctx, config)
	require.NoError(t, err)

	err = db.Ping(ctx)
	if err != nil {
		t.Skipf("Skipping integration test: database not available: %v", err)
	}

	// Ensure search_path is set for the current connection too (redundant but safe)
	_, err = db.Exec(ctx, fmt.Sprintf("SET search_path TO %s, public", schema))
	require.NoError(t, err)

	// Run migrations
	err = RunMigrations(ctx, db, slog.Default())
	require.NoError(t, err, "Failed to run migrations")

	t.Cleanup(func() {
		db.Close()
		// Clean up the schema
		cleanupDB, _ := pgxpool.New(ctx, dbURL)
		if cleanupDB != nil {
			_, _ = cleanupDB.Exec(ctx, fmt.Sprintf("DROP SCHEMA %s CASCADE", schema))
			cleanupDB.Close()
		}
	})

	return db, schema
}

// SetupTestUser creates a dedicated test user and tenant context for a test run.
// It returns a context containing both the UserID and TenantID for multi-tenant isolation.
func SetupTestUser(t *testing.T, db *pgxpool.Pool) context.Context {
	t.Helper()
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
	t.Helper()
	ctx := context.Background()

	// Get current schema from search_path
	var schema string
	err := db.QueryRow(ctx, "SHOW search_path").Scan(&schema)
	if err != nil {
		schema = "public"
	}
	schema = strings.Split(schema, ",")[0]
	schema = strings.TrimSpace(schema)

	query := fmt.Sprintf(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = '%s' 
		AND table_type = 'BASE TABLE'
		AND table_name != 'schema_migrations'
	`, schema)
	
	rows, err := db.Query(ctx, query)
	if err != nil {
		t.Logf("Warning: failed to query tables for cleanup: %v", err)
		return
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err == nil {
			tables = append(tables, table)
		}
	}

	if len(tables) == 0 {
		return
	}

	truncateQuery := "TRUNCATE " + strings.Join(tables, ", ") + " RESTART IDENTITY CASCADE"
	_, err = db.Exec(ctx, truncateQuery)
	if err != nil {
		t.Logf("Warning: failed to truncate tables: %v", err)
	}
}
