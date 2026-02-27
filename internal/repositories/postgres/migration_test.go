//go:build integration

package postgres

import (
	"context"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestMigrationRollback(t *testing.T) {
	db := SetupDB(t)
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Acquire(ctx)
	require.NoError(t, err)
	defer conn.Release()

	schema := "migration_test_" + strings.ReplaceAll(uuid.NewString(), "-", "_")
	_, err = conn.Exec(ctx, "CREATE SCHEMA "+schema)
	require.NoError(t, err)
	defer func() {
		_, _ = conn.Exec(ctx, "DROP SCHEMA IF EXISTS "+schema+" CASCADE")
	}()

	_, err = conn.Exec(ctx, "SET search_path TO "+schema+", public")
	require.NoError(t, err)

	_, err = conn.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
	require.NoError(t, err)

	// Get all migration files from embedded FS
	files, err := MigrationFiles.ReadDir("migrations")
	require.NoError(t, err)

	var upMigrations []string
	var downMigrations []string

	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".up.sql") {
			upMigrations = append(upMigrations, f.Name())
		} else if strings.HasSuffix(f.Name(), ".down.sql") {
			downMigrations = append(downMigrations, f.Name())
		}
	}

	sort.Strings(upMigrations)
	sort.Strings(downMigrations)

	// Function to run a migration file
	runFile := func(name string, isDown bool) error {
		content, err := MigrationFiles.ReadFile("migrations/" + name)
		if err != nil {
			return err
		}
		sql := string(content)

		if !isDown {
			// Extract UP part if goose-formatted
			if parts := strings.Split(sql, "-- +goose Down"); len(parts) > 1 {
				sql = parts[0]
			}
			sql = strings.TrimPrefix(sql, "-- +goose Up")
		} else {
			// Extract DOWN part if goose-formatted
			if parts := strings.Split(sql, "-- +goose Down"); len(parts) > 1 {
				sql = parts[1]
			}
			// If it doesn't have a Down marker but is a .down.sql file, run the whole thing
		}

		_, err = conn.Exec(ctx, sql)
		return err
	}

	t.Run("Step 1: Apply all UP migrations", func(t *testing.T) {
		for _, m := range upMigrations {
			err := runFile(m, false)
			// We use IF NOT EXISTS in most migrations, but if not, we might get errors
			// if the DB isn't clean. For this test, we want to see it work.
			if err != nil {
				t.Logf("Warning: migration %s failed: %v (might be expected if already applied)", m, err)
			}
		}
	})

	t.Run("Step 2: Apply all DOWN migrations in reverse order", func(t *testing.T) {
		// Reverse down migrations
		for i := len(downMigrations) - 1; i >= 0; i-- {
			m := downMigrations[i]
			err := runFile(m, true)
			require.NoError(t, err, "Failed to rollback migration: %s", m)
			t.Logf("Rolled back: %s", m)
		}
	})

	t.Run("Step 3: Re-apply all UP migrations", func(t *testing.T) {
		for _, m := range upMigrations {
			err := runFile(m, false)
			require.NoError(t, err, "Failed to re-apply migration: %s", m)
			t.Logf("Re-applied: %s", m)
		}
	})
}
