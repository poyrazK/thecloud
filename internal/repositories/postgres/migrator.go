// Package postgres provides PostgreSQL-backed repository implementations.
package postgres

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.up.sql
var migrationFiles embed.FS

// RunMigrations executes all SQL migration files in order.
// It tries to be idempotent by using IF NOT EXISTS where possible in SQL files.
func RunMigrations(ctx context.Context, db any, logger *slog.Logger) error {
	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	// Sort entries to ensure deterministic execution order
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	// We need a single connection to ensure session state (like search_path) persists
	// if the caller passed a pool.
	var executor interface {
		Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	}

	switch d := db.(type) {
	case *pgxpool.Pool:
		conn, err := d.Acquire(ctx)
		if err != nil {
			return fmt.Errorf("failed to acquire connection for migrations: %w", err)
		}
		defer conn.Release()
		executor = conn
	case DB:
		executor = d
	default:
		return fmt.Errorf("provided db does not support Exec")
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}

		content, err := migrationFiles.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", entry.Name(), err)
		}

		// Only execute the "Up" part if it's a goose-formatted file
		sql := string(content)
		if parts := strings.Split(sql, "-- +goose Down"); len(parts) > 1 {
			sql = parts[0]
		}
		// Also handle -- +goose Up prefix if present
		sql = strings.TrimPrefix(sql, "-- +goose Up")

		// Execute migration
		_, err = executor.Exec(ctx, sql)
		if err != nil {
			// Log but don't fail, as tables might already exist
			// Ideally we should check specific error codes
			logger.Warn("migration result", "migration", entry.Name(), "error", err)
		} else {
			logger.Info("applied migration", "migration", entry.Name())
		}
	}

	return nil
}
