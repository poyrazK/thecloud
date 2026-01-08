//go:build integration

package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/require"
)

func setupDB(t *testing.T) *pgxpool.Pool {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://cloud:cloud@localhost:5433/thecloud"
	}

	ctx := context.Background()
	db, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err)

	err = db.Ping(ctx)
	if err != nil {
		t.Skipf("Skipping integration test: database not available: %v", err)
	}

	return db
}

func setupTestUser(t *testing.T, db *pgxpool.Pool) context.Context {
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

	return appcontext.WithUserID(ctx, userID)
}

func cleanDB(t *testing.T, db *pgxpool.Pool) {
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
		// Users are usually not deleted to keep test user valid if reused,
		// but here we create a new user per test with setupTestUser, so we accumulate users.
		// We can leave users or delete them if we track the ID.
		// For now, let's just clean resources.
	}

	for _, q := range queries {
		_, err := db.Exec(ctx, q)
		// Ignore errors if table doesn't exist (42P01 error code)
		// This allows tests to run even if not all migrations have been applied
		if err != nil {
			t.Logf("Cleanup query failed (ignoring): %s - %v", q, err)
		}
	}
}
