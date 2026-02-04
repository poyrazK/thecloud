package postgres

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
)

func TestRunMigrations(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Get the migration files to know how many Execs to expect
	entries, err := migrationsFS.ReadDir("migrations")
	assert.NoError(t, err)

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}
		// Each migration file in the loop will trigger an Exec call
		mock.ExpectExec(".*").WillReturnResult(pgxmock.NewResult("EXECUTE", 1))
	}

	err = RunMigrations(context.Background(), mock, logger)
	assert.NoError(t, err)
	
	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}
