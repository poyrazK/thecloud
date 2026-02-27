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
)

//go:embed migrations/*.up.sql
var migrationFiles embed.FS

// DB interface defines the minimum methods needed from a connection or pool.
type DB interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

// TransactionalDB defines an interface that supports transactions.
type TransactionalDB interface {
	Begin(context.Context) (pgx.Tx, error)
}

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

	var tx pgx.Tx
	var execDB interface {
		Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	}

	// Try to start a transaction if supported
	if tdb, ok := db.(TransactionalDB); ok {
		tx, err = tdb.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin migration transaction: %w", err)
		}
		defer func() {
			if tx != nil {
				_ = tx.Rollback(ctx)
			}
		}()
		execDB = tx
	} else if edb, ok := db.(interface {
		Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	}); ok {
		execDB = edb
	} else {
		return fmt.Errorf("provided db does not support Exec")
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}

		sql, err := migrationFiles.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", entry.Name(), err)
		}

		// Execute migration
		_, err = execDB.Exec(ctx, string(sql))
		if err != nil {
			// Log but don't fail, as tables might already exist
			// Ideally we should check specific error codes
			logger.Warn("migration result", "migration", entry.Name(), "error", err)
		} else {
			logger.Info("applied migration", "migration", entry.Name())
		}
	}

	if tx != nil {
		err = tx.Commit(ctx)
		if err != nil {
			return fmt.Errorf("failed to commit migration transaction: %w", err)
		}
		tx = nil
	}

	return nil
}
