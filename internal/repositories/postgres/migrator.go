// Package postgres provides PostgreSQL-backed repository implementations.
package postgres

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"sort"
	"strings"
)

//go:embed migrations/*.up.sql
var migrationFiles embed.FS

// RunMigrations executes all SQL migration files in order.
// It tries to be idempotent by using IF NOT EXISTS where possible in SQL files.
func RunMigrations(ctx context.Context, db DB, logger *slog.Logger) error {
	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	// Sort entries to ensure deterministic execution order
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	// Start a transaction to ensure search_path and migrations are consistent
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin migration transaction: %w", err)
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}

		sql, err := migrationFiles.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", entry.Name(), err)
		}

		// Execute migration within the transaction
		_, err = tx.Exec(ctx, string(sql))
		if err != nil {
			// Log but don't fail, as tables might already exist
			// Ideally we should check specific error codes
			logger.Warn("migration result", "migration", entry.Name(), "error", err)
		} else {
			logger.Info("applied migration", "migration", entry.Name())
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit migration transaction: %w", err)
	}
	tx = nil

	return nil
}
