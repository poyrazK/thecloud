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
)

func TestInitInfrastructureMigrateOnlyStopsAfterMigrations(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	fakeDB := &stubDB{}
	deps := DefaultDeps()

	deps.LoadConfig = func(*slog.Logger) (*platform.Config, error) {
		return &platform.Config{}, nil
	}
	deps.InitDatabase = func(context.Context, *platform.Config, *slog.Logger) (postgres.DB, error) {
		return fakeDB, nil
	}

	migrationsRan := false
	deps.RunMigrations = func(context.Context, postgres.DB, *slog.Logger) error {
		migrationsRan = true
		return nil
	}
	deps.InitRedis = func(context.Context, *platform.Config, *slog.Logger) (*redis.Client, error) {
		t.Fatalf("initRedis should not be called when migrate-only completes")
		return nil, nil
	}

	cfg, db, rdb, err := initInfrastructure(deps, logger, true)

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
	deps := DefaultDeps()

	deps.LoadConfig = func(*slog.Logger) (*platform.Config, error) {
		return nil, errors.New("boom")
	}

	if _, _, _, err := initInfrastructure(deps, logger, false); err == nil {
		t.Fatalf("expected error when config loading fails")
	}
}

func TestInitInfrastructureRedisErrorClosesDB(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	fakeDB := &stubDB{}
	deps := DefaultDeps()

	deps.LoadConfig = func(*slog.Logger) (*platform.Config, error) {
		return &platform.Config{RedisURL: "127.0.0.1:1"}, nil
	}
	deps.InitDatabase = func(context.Context, *platform.Config, *slog.Logger) (postgres.DB, error) {
		return fakeDB, nil
	}
	deps.RunMigrations = func(context.Context, postgres.DB, *slog.Logger) error {
		return nil
	}
	deps.InitRedis = func(context.Context, *platform.Config, *slog.Logger) (*redis.Client, error) {
		return nil, errors.New("redis unavailable")
	}

	_, _, _, err := initInfrastructure(deps, logger, false)
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
	deps := DefaultDeps()

	deps.InitComputeBackend = func(*platform.Config, *slog.Logger) (ports.ComputeBackend, error) {
		return noop.NewNoopComputeBackend(), nil
	}
	deps.InitStorageBackend = func(*platform.Config, *slog.Logger) (ports.StorageBackend, error) {
		return noop.NewNoopStorageBackend(), nil
	}
	deps.InitNetworkBackend = func(*platform.Config, *slog.Logger) ports.NetworkBackend {
		return noop.NewNoopNetworkAdapter(logger)
	}
	deps.InitRepositories = func(postgres.DB, *redis.Client) *setup.Repositories {
		return &setup.Repositories{}
	}
	deps.InitLBProxy = func(*platform.Config, ports.ComputeBackend, ports.InstanceRepository, ports.VpcRepository) (ports.LBProxyAdapter, error) {
		return nil, errors.New("lb proxy failed")
	}

	_, _, _, _, err := initBackends(deps, cfg, logger, &stubDB{}, redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"}))
	if err == nil {
		t.Fatalf("expected error when lb proxy init fails")
	}
}

func TestRunApplicationApiRoleStartsAndShutsDown(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	t.Setenv("ROLE", "api")

	started := make(chan struct{})
	shutdownCalled := make(chan struct{})
	deps := DefaultDeps()

	deps.NewHTTPServer = func(addr string, handler http.Handler) *http.Server {
		return &http.Server{Addr: addr, Handler: handler}
	}
	deps.StartHTTPServer = func(*http.Server) error {
		close(started)
		return http.ErrServerClosed
	}
	deps.ShutdownHTTPServer = func(context.Context, *http.Server) error {
		close(shutdownCalled)
		return nil
	}
	deps.NotifySignals = func(c chan<- os.Signal, _ ...os.Signal) {
		go func() {
			<-started
			c <- syscall.SIGTERM
		}()
	}

	runApplication(deps, &platform.Config{Port: "0"}, logger, gin.New(), &setup.Workers{})

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
