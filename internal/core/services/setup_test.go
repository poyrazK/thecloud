package services_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
)

func setupDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	// Use helper from postgres package
	db, _ := postgres.SetupDB(t)
	return db
}

func setupTestUser(t *testing.T, db *pgxpool.Pool) context.Context {
	t.Helper()
	return postgres.SetupTestUser(t, db)
}

func cleanDB(t *testing.T, db *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	tables := []string{
		"instances", "vpcs", "subnets", "volumes", "instance_types",
		"roles", "role_permissions", "users", "tenants", "tenant_members",
		"audit_logs", "events", "snapshots", "stacks", "api_keys",
		"ssh_keys", "elastic_ips", "iam_policies", "iam_user_policies",
	}

	// Double check we are in test environment before TRUNCATE
	if os.Getenv("POSTGRES_DB") == "" && os.Getenv("DATABASE_URL") == "" {
		return
	}

	truncateQuery := "TRUNCATE " + strings.Join(tables, ", ") + " RESTART IDENTITY CASCADE"
	_, err := db.Exec(ctx, truncateQuery)
	if err != nil {
		t.Logf("Warning: failed to truncate tables: %v", err)
	}
}
