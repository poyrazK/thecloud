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

	// Get the migration files to know how many Execs to expect
	entries, err := migrationsFS.ReadDir("migrations")
	require.NoError(t, err)

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}
		// Each migration file in the loop will trigger an Exec call
		mock.ExpectExec(".*").WillReturnResult(pgxmock.NewResult("EXECUTE", 1))
	}

	err = RunMigrations(context.Background(), mock, logger)
	require.NoError(t, err)

	// Ensure all expectations were met
	require.NoError(t, mock.ExpectationsWereMet())
}
