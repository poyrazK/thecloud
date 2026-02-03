// Package platform provides infrastructure initialization helpers.
package platform

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// InitRedis initializes the Redis client
func InitRedis(ctx context.Context, cfg *Config, logger *slog.Logger) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisURL,
		DialTimeout:  1 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	})

	// Ping the Redis server to verify connectivity
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis at %s: %w", cfg.RedisURL, err)
	}

	logger.Info("connected to redis", "addr", cfg.RedisURL)
	return rdb, nil
}
