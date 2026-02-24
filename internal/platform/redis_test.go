package platform

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"github.com/stretchr/testify/require"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func TestInitRedis_Success(t *testing.T) {
	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := &Config{RedisURL: server.Addr()}

	client, err := InitRedis(context.Background(), cfg, logger)
	require.NoError(t, err)
	if assert.NotNil(t, client) {
		assert.NoError(t, client.Close())
	}
}

func TestInitRedis_InvalidAddress(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := &Config{RedisURL: "127.0.0.1:1"}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	client, err := InitRedis(ctx, cfg, logger)
	require.Error(t, err)
	assert.Nil(t, client)
}
