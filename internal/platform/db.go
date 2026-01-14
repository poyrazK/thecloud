// Package platform provides infrastructure initialization helpers.
package platform

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewDatabase creates a PostgreSQL connection pool using config settings.
func NewDatabase(ctx context.Context, cfg *Config, logger *slog.Logger) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database url: %w", err)
	}

	if os.Getenv("TRACING_ENABLED") == "true" {
		config.ConnConfig.Tracer = otelpgx.NewTracer()
	}

	// Apply performance optimizations
	var maxConns int32
	if _, err := fmt.Sscanf(cfg.DBMaxConns, "%d", &maxConns); err == nil {
		config.MaxConns = maxConns
	}

	var minConns int32
	if _, err := fmt.Sscanf(cfg.DBMinConns, "%d", &minConns); err == nil {
		config.MinConns = minConns
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	logger.Info("connected to database")
	return pool, nil
}
