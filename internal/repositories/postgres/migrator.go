// Package postgres provides PostgreSQL-backed repository implementations.
package postgres

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.up.sql migrations/*.down.sql
var MigrationFiles embed.FS

// RunMigrations executes SQL migrations in order, tracking applied versions in schema_migrations.
// Each migration runs exactly once — on subsequent startups, only new migrations run.
// Existing -- +goose Up/Down markers in SQL files are stripped (SQL comments, ignored by postgres).
func RunMigrations(ctx context.Context, db any, logger *slog.Logger) error {
	entries, err := MigrationFiles.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	// Acquire a single connection for the entire migration run to ensure consistency
	var conn interface {
		Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
		Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	}

	switch d := db.(type) {
	case *pgxpool.Pool:
		c, err := d.Acquire(ctx)
		if err != nil {
			return fmt.Errorf("failed to acquire connection for migrations: %w", err)
		}
		defer c.Release()
		conn = c
	case DB:
		conn = d
	default:
		return fmt.Errorf("unsupported db type")
	}

	// Ensure schema_migrations table exists
	if err := ensureSchemaMigrationsTable(ctx, conn); err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// Get set of already-applied versions
	appliedSet, err := getAppliedVersions(ctx, conn)
	if err != nil {
		return fmt.Errorf("failed to read applied versions: %w", err)
	}

	// Run only migrations that are not yet applied
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}

		version := extractVersion(entry.Name())
		if version == 0 {
			continue
		}

		if appliedSet[version] {
			logger.Debug("migration already applied, skipping", "migration", entry.Name())
			continue
		}

		content, err := MigrationFiles.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", entry.Name(), err)
		}

		sql := string(content)
		if parts := strings.Split(sql, "-- +goose Down"); len(parts) > 1 {
			sql = parts[0]
		}
		sql = strings.TrimPrefix(sql, "-- +goose Up")
		sql = strings.TrimSpace(sql)

		if sql == "" {
			continue
		}

		if _, err := conn.Exec(ctx, sql); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", entry.Name(), err)
		}

		if err := recordVersion(ctx, conn, version); err != nil {
			return fmt.Errorf("failed to record version %d: %w", version, err)
		}

		logger.Info("applied migration", "migration", entry.Name())
	}

	return nil
}

// ensureSchemaMigrationsTable creates the schema_migrations table if it doesn't exist.
func ensureSchemaMigrationsTable(ctx context.Context, conn interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
}) error {
	_, err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version bigint not null primary key,
			dirty boolean not null default false,
			created_at timestamptz not null default now()
		)
	`)
	return err
}

// getAppliedVersions returns the set of migration versions already recorded in schema_migrations.
func getAppliedVersions(ctx context.Context, conn interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
}) (map[int64]bool, error) {
	rows, err := conn.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		// Table might not exist yet (shouldn't happen after ensureSchemaMigrationsTable)
		// but treat as no versions applied
		return make(map[int64]bool), nil
	}
	defer rows.Close()

	applied := make(map[int64]bool)
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		applied[v] = true
	}
	return applied, rows.Err()
}

// recordVersion marks a migration version as applied in schema_migrations.
func recordVersion(ctx context.Context, conn interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
}, version int64) error {
	_, err := conn.Exec(ctx, `
		INSERT INTO schema_migrations (version, dirty)
		VALUES ($1, false)
		ON CONFLICT (version) DO NOTHING
	`, version)
	return err
}

// extractVersion extracts the numeric prefix from a migration filename.
// e.g., "072_migrate_to_tenants.up.sql" -> 72
func extractVersion(name string) int64 {
	var v int64
	for _, c := range name {
		if c >= '0' && c <= '9' {
			v = v*10 + int64(c-'0')
		} else if v > 0 {
			break
		}
	}
	return v
}
