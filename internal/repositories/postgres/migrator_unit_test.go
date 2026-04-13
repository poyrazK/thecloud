package postgres

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/require"
)

func TestRunMigrations(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	entries, err := MigrationFiles.ReadDir("migrations")
	require.NoError(t, err)

	var upFiles int
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}
		upFiles++
	}

	// CREATE TABLE IF NOT EXISTS schema_migrations
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS schema_migrations").
		WillReturnResult(pgxmock.NewResult("CREATE TABLE", 0))

	// SELECT from schema_migrations (empty - first run)
	rows := pgxmock.NewRows([]string{"version"})
	mock.ExpectQuery("SELECT version FROM schema_migrations").WillReturnRows(rows)

	// For each migration file, expect 2 Execs:
	// 1. Migration SQL (any content from the file)
	// 2. recordVersion INSERT INTO schema_migrations VALUES ($1, $2) — 2 args
	for i := 0; i < upFiles; i++ {
		// Run migration SQL — any DDL/DML from .up.sql file
		mock.ExpectExec("CREATE|ALTER|DROP|INSERT").
			WillReturnResult(pgxmock.NewResult("EXECUTE", 1))
		// Record version in schema_migrations — 2 args: version (int64) and dirty (bool)
		mock.ExpectExec("INSERT INTO schema_migrations").
			WithArgs(pgxmock.AnyArg(), false).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))
	}

	err = RunMigrations(context.Background(), mock, logger)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}