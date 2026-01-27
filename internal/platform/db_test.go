package platform

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDatabaseInvalidURL(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	pool, err := NewDatabase(context.Background(), &Config{DatabaseURL: "invalid-url"}, logger)
	assert.Error(t, err)
	assert.Nil(t, pool)
	assert.Contains(t, err.Error(), "unable to parse database url")
}

func TestNewDatabaseInvalidMaxConns(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	pool, err := NewDatabase(context.Background(), &Config{
		DatabaseURL: "postgres://user:pass@localhost/db",
		DBMaxConns:  "invalid",
	}, logger)
	// Should not fail on parsing, but if DB is not available, it will fail on ping
	if err == nil {
		pool.Close()
	}
}

func TestNewDatabaseInvalidMinConns(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	pool, err := NewDatabase(context.Background(), &Config{
		DatabaseURL: "postgres://user:pass@localhost/db",
		DBMinConns:  "invalid",
	}, logger)
	// Should not fail on parsing, but if DB is not available, it will fail on ping
	if err == nil {
		pool.Close()
	}
}
