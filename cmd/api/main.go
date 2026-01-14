// Package main provides the API server entrypoint.
package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/poyrazk/thecloud/internal/api/setup"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/platform"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/poyrazk/thecloud/pkg/tracing"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// ... (omitted comments) ...

// ErrMigrationDone signals that migrations have already completed.
var ErrMigrationDone = errors.New("migrations done")

func main() {
	logger := setup.InitLogger()
	migrateOnly := flag.Bool("migrate-only", false, "run database migrations and exit")
	flag.Parse()

	// Initialize Tracing (Opt-in)
	if tp := initTracing(logger); tp != nil {
		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				logger.Error("failed to shutdown tracer", "error", err)
			}
		}()
	}

	cfg, db, rdb, err := initInfrastructure(logger, *migrateOnly)
	if err != nil {
		if *migrateOnly && errors.Is(err, ErrMigrationDone) {
			return
		}
		logger.Error("initialization failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	defer func() { _ = rdb.Close() }()

	compute, storage, network, lbProxy, err := initBackends(cfg, logger, db, rdb)
	if err != nil {
		logger.Error("backend initialization failed", "error", err)
		os.Exit(1)
	}

	repos := initRepositoriesFunc(db, rdb)
	svcs, workers, err := initServicesFunc(setup.ServiceConfig{
		Config: cfg, Repos: repos, Compute: compute, Storage: storage,
		Network: network, LBProxy: lbProxy, DB: db, RDB: rdb, Logger: logger,
	})
	if err != nil {
		logger.Error("services initialization failed", "error", err)
		os.Exit(1)
	}

	handlers := initHandlersFunc(svcs, logger)
	r := setupRouterFunc(cfg, logger, handlers, svcs, network)

	// Add Tracing Middleware if enabled
	if os.Getenv("TRACING_ENABLED") == "true" {
		r.Use(otelgin.Middleware("thecloud-api"))
	}

	runApplication(cfg, logger, r, workers)
}

func initTracing(logger *slog.Logger) *sdktrace.TracerProvider {
	if os.Getenv("TRACING_ENABLED") != "true" {
		return nil
	}

	serviceName := "thecloud-api"
	exporterType := os.Getenv("TRACING_EXPORTER")

	if exporterType == "console" {
		logger.Info("initializing console tracing (debug mode)")
		tp, err := tracing.InitConsoleTracer(serviceName)
		if err != nil {
			logger.Error("failed to init console tracer", "error", err)
			return nil
		}
		return tp
	}

	endpoint := os.Getenv("JAEGER_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:4318"
	}
	logger.Info("initializing distributed tracing (Jaeger)", "endpoint", endpoint)
	tp, err := tracing.InitTracer(context.Background(), serviceName, endpoint)
	if err != nil {
		logger.Error("failed to init tracer", "error", err)
		return nil
	}
	return tp
}

func initInfrastructure(logger *slog.Logger, migrateOnly bool) (*platform.Config, postgres.DB, *redis.Client, error) {
	cfg, err := loadConfigFunc(logger)
	if err != nil {
		return nil, nil, nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := initDatabaseFunc(ctx, cfg, logger)
	if err != nil {
		return nil, nil, nil, err
	}

	if err := runMigrationsFunc(ctx, db, logger); err != nil {
		logger.Warn("failed to run migrations", "error", err)
		if migrateOnly {
			db.Close()
			return nil, nil, nil, err
		}
	} else if migrateOnly {
		logger.Info("migrations completed")
		db.Close()
		return nil, nil, nil, ErrMigrationDone
	}

	rdb, err := initRedisFunc(ctx, cfg, logger)
	if err != nil {
		db.Close()
		return nil, nil, nil, err
	}

	return cfg, db, rdb, nil
}

func initBackends(cfg *platform.Config, logger *slog.Logger, db postgres.DB, rdb *redis.Client) (ports.ComputeBackend, ports.StorageBackend, ports.NetworkBackend, ports.LBProxyAdapter, error) {
	compute, err := initComputeBackendFunc(cfg, logger)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	storage, err := initStorageBackendFunc(cfg, logger)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	network := initNetworkBackendFunc(cfg, logger)

	tmpRepos := initRepositoriesFunc(db, rdb)
	lbProxy, err := initLBProxyFunc(cfg, compute, tmpRepos.Instance, tmpRepos.Vpc)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return compute, storage, network, lbProxy, nil
}

func runApplication(cfg *platform.Config, logger *slog.Logger, r *gin.Engine, workers *setup.Workers) {
	role := os.Getenv("ROLE")
	if role == "" {
		role = "all"
	}

	wg := &sync.WaitGroup{}
	workerCtx, workerCancel := context.WithCancel(context.Background())

	if role == "worker" || role == "all" {
		runWorkers(workerCtx, wg, workers)
	}

	srv := newHTTPServer(":"+cfg.Port, r)

	if role == "api" || role == "all" {
		go func() {
			logger.Info("starting compute-api", "port", cfg.Port)
			if err := startHTTPServer(srv); err != nil && err != http.ErrServerClosed {
				logger.Error("failed to start server", "error", err)
				os.Exit(1)
			}
		}()
	} else {
		logger.Info("running in worker-only mode")
	}

	quit := make(chan os.Signal, 1)
	notifySignals(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := shutdownHTTPServer(ctx, srv); err != nil {
		logger.Error("server forced to shutdown", "error", err)
	}

	workerCancel()
	wg.Wait()
	logger.Info("server exited")
}

func runWorkers(ctx context.Context, wg *sync.WaitGroup, workers *setup.Workers) {
	wg.Add(6)
	go workers.LB.Run(ctx, wg)
	go workers.AutoScaling.Run(ctx, wg)
	go workers.Cron.Run(ctx, wg)
	go workers.Container.Run(ctx, wg)
	go workers.Provision.Run(ctx, wg)
	go workers.Accounting.Run(ctx, wg)
}
