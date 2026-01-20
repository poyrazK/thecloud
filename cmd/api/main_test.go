package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/poyrazk/thecloud/internal/api/setup"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/platform"
	"github.com/poyrazk/thecloud/internal/repositories/noop"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestInitInfrastructureMigrateOnlyStopsAfterMigrations(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	fakeDB := &stubDB{}

	resetLoadConfig := stubLoadConfig(func(*slog.Logger) (*platform.Config, error) {
		return &platform.Config{}, nil
	})
	defer resetLoadConfig()

	resetInitDatabase := stubInitDatabase(func(context.Context, *platform.Config, *slog.Logger) (postgres.DB, error) {
		return fakeDB, nil
	})
	defer resetInitDatabase()

	migrationsRan := false
	resetRunMigrations := stubRunMigrations(func(context.Context, postgres.DB, *slog.Logger) error {
		migrationsRan = true
		return nil
	})
	defer resetRunMigrations()

	resetInitRedis := stubInitRedis(func(context.Context, *platform.Config, *slog.Logger) (*redis.Client, error) {
		t.Fatalf("initRedis should not be called when migrate-only completes")
		return nil, nil
	})
	defer resetInitRedis()

	cfg, db, rdb, err := initInfrastructure(logger, true)

	if !errors.Is(err, ErrMigrationDone) {
		t.Fatalf("expected ErrMigrationDone, got %v", err)
	}
	if cfg != nil || db != nil || rdb != nil {
		t.Fatalf("expected nil resources after migrate-only run, got cfg=%v db=%v rdb=%v", cfg, db, rdb)
	}
	if !migrationsRan {
		t.Fatalf("expected migrations to run")
	}
	if !fakeDB.closed {
		t.Fatalf("expected database to be closed")
	}
}

func TestInitInfrastructureConfigError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	resetLoadConfig := stubLoadConfig(func(*slog.Logger) (*platform.Config, error) {
		return nil, errors.New("boom")
	})
	defer resetLoadConfig()

	if _, _, _, err := initInfrastructure(logger, false); err == nil {
		t.Fatalf("expected error when config loading fails")
	}
}

func TestInitInfrastructureRedisErrorClosesDB(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	fakeDB := &stubDB{}

	resetLoadConfig := stubLoadConfig(func(*slog.Logger) (*platform.Config, error) {
		return &platform.Config{RedisURL: "127.0.0.1:1"}, nil
	})
	defer resetLoadConfig()

	resetInitDatabase := stubInitDatabase(func(context.Context, *platform.Config, *slog.Logger) (postgres.DB, error) {
		return fakeDB, nil
	})
	defer resetInitDatabase()

	resetRunMigrations := stubRunMigrations(func(context.Context, postgres.DB, *slog.Logger) error {
		return nil
	})
	defer resetRunMigrations()

	resetInitRedis := stubInitRedis(func(context.Context, *platform.Config, *slog.Logger) (*redis.Client, error) {
		return nil, errors.New("redis unavailable")
	})
	defer resetInitRedis()

	_, _, _, err := initInfrastructure(logger, false)
	if err == nil {
		t.Fatalf("expected error when redis init fails")
	}
	if !fakeDB.closed {
		t.Fatalf("expected database to be closed on redis init error")
	}
}

func TestInitBackendsLBProxyError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := &platform.Config{}

	resetCompute := stubInitComputeBackend(func(*platform.Config, *slog.Logger) (ports.ComputeBackend, error) {
		return noop.NewNoopComputeBackend(), nil
	})
	defer resetCompute()

	resetStorage := stubInitStorageBackend(func(*platform.Config, *slog.Logger) (ports.StorageBackend, error) {
		return noop.NewNoopStorageBackend(), nil
	})
	defer resetStorage()

	resetNetwork := stubInitNetworkBackend(func(*platform.Config, *slog.Logger) ports.NetworkBackend {
		return noop.NewNoopNetworkAdapter(logger)
	})
	defer resetNetwork()

	resetRepos := stubInitRepositories(func(postgres.DB, *redis.Client) *setup.Repositories {
		return &setup.Repositories{}
	})
	defer resetRepos()

	resetLBProxy := stubInitLBProxy(func(*platform.Config, ports.ComputeBackend, ports.InstanceRepository, ports.VpcRepository) (ports.LBProxyAdapter, error) {
		return nil, errors.New("lb proxy failed")
	})
	defer resetLBProxy()

	// Use nil for rdb since we shouldn't be calling it in this test if stubs work.
	_, _, _, _, err := initBackends(cfg, logger, &stubDB{}, nil)
	if err == nil {
		t.Fatalf("expected error when lb proxy init fails")
	}
}

func TestRunApplicationApiRoleStartsAndShutsDown(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	t.Setenv("ROLE", "api")

	started := make(chan struct{})
	shutdownCalled := make(chan struct{})

	resetNewHTTPServer := stubNewHTTPServer(func(addr string, handler http.Handler) *http.Server {
		return &http.Server{Addr: addr, Handler: handler}
	})
	defer resetNewHTTPServer()

	resetStartServer := stubStartHTTPServer(func(*http.Server) error {
		close(started)
		return http.ErrServerClosed
	})
	defer resetStartServer()

	resetShutdown := stubShutdownHTTPServer(func(context.Context, *http.Server) error {
		close(shutdownCalled)
		return nil
	})
	defer resetShutdown()

	resetNotify := stubNotifySignals(func(c chan<- os.Signal, _ ...os.Signal) {
		go func() {
			<-started
			c <- syscall.SIGTERM
		}()
	})
	defer resetNotify()

	runApplication(&platform.Config{Port: "0"}, logger, gin.New(), &setup.Workers{})

	select {
	case <-shutdownCalled:
	case <-time.After(time.Second):
		t.Fatalf("expected server shutdown to be called")
	}
}

// Stub helpers below keep main.go testable without altering production behavior.

type stubDB struct{ closed bool }

func (s *stubDB) Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (s *stubDB) Query(context.Context, string, ...interface{}) (pgx.Rows, error) { return nil, nil }
func (s *stubDB) QueryRow(context.Context, string, ...interface{}) pgx.Row        { return nil }
func (s *stubDB) Begin(context.Context) (pgx.Tx, error)                           { return nil, nil }
func (s *stubDB) Close()                                                          { s.closed = true }
func (s *stubDB) Ping(context.Context) error                                      { return nil }

func stubLoadConfig(fn func(*slog.Logger) (*platform.Config, error)) func() {
	prev := loadConfigFunc
	loadConfigFunc = fn
	return func() { loadConfigFunc = prev }
}

func stubInitDatabase(fn func(context.Context, *platform.Config, *slog.Logger) (postgres.DB, error)) func() {
	prev := initDatabaseFunc
	initDatabaseFunc = fn
	return func() { initDatabaseFunc = prev }
}

func stubRunMigrations(fn func(context.Context, postgres.DB, *slog.Logger) error) func() {
	prev := runMigrationsFunc
	runMigrationsFunc = fn
	return func() { runMigrationsFunc = prev }
}

func stubInitRedis(fn func(context.Context, *platform.Config, *slog.Logger) (*redis.Client, error)) func() {
	prev := initRedisFunc
	initRedisFunc = fn
	return func() { initRedisFunc = prev }
}

func stubInitComputeBackend(fn func(*platform.Config, *slog.Logger) (ports.ComputeBackend, error)) func() {
	prev := initComputeBackendFunc
	initComputeBackendFunc = fn
	return func() { initComputeBackendFunc = prev }
}

func stubInitStorageBackend(fn func(*platform.Config, *slog.Logger) (ports.StorageBackend, error)) func() {
	prev := initStorageBackendFunc
	initStorageBackendFunc = fn
	return func() { initStorageBackendFunc = prev }
}

func stubInitNetworkBackend(fn func(*platform.Config, *slog.Logger) ports.NetworkBackend) func() {
	prev := initNetworkBackendFunc
	initNetworkBackendFunc = fn
	return func() { initNetworkBackendFunc = prev }
}

func stubInitLBProxy(fn func(*platform.Config, ports.ComputeBackend, ports.InstanceRepository, ports.VpcRepository) (ports.LBProxyAdapter, error)) func() {
	prev := initLBProxyFunc
	initLBProxyFunc = fn
	return func() { initLBProxyFunc = prev }
}

func stubInitRepositories(fn func(postgres.DB, *redis.Client) *setup.Repositories) func() {
	prev := initRepositoriesFunc
	initRepositoriesFunc = fn
	return func() { initRepositoriesFunc = prev }
}

func stubNewHTTPServer(fn func(string, http.Handler) *http.Server) func() {
	prev := newHTTPServer
	newHTTPServer = fn
	return func() { newHTTPServer = prev }
}

func stubStartHTTPServer(fn func(*http.Server) error) func() {
	prev := startHTTPServer
	startHTTPServer = fn
	return func() { startHTTPServer = prev }
}

func stubShutdownHTTPServer(fn func(context.Context, *http.Server) error) func() {
	prev := shutdownHTTPServer
	shutdownHTTPServer = fn
	return func() { shutdownHTTPServer = prev }
}

func stubNotifySignals(fn func(chan<- os.Signal, ...os.Signal)) func() {
	prev := notifySignals
	notifySignals = fn
	return func() { notifySignals = prev }
}

func TestInitTracing(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("Disabled", func(t *testing.T) {
		t.Setenv("TRACING_ENABLED", "false")
		tp := initTracing(logger)
		assert.Nil(t, tp)
	})

	t.Run("Console", func(t *testing.T) {
		t.Setenv("TRACING_ENABLED", "true")
		t.Setenv("TRACING_EXPORTER", "console")
		tp := initTracing(logger)
		assert.NotNil(t, tp)
		_ = tp.Shutdown(context.Background())
	})

	t.Run("Jaeger", func(t *testing.T) {
		t.Setenv("TRACING_ENABLED", "true")
		t.Setenv("TRACING_EXPORTER", "jaeger")
		t.Setenv("JAEGER_ENDPOINT", "http://localhost:4318")
		// This might fail if it tries to connect, but let's see.
		// Actually initTracing just returns the provider, it doesn't necessarily block.
		tp := initTracing(logger)
		assert.NotNil(t, tp)
		if tp != nil {
			_ = tp.Shutdown(context.Background())
		}
	})
}
