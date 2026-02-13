package main

// @title The Cloud API
// @version 1.0
// @description The Cloud - Multi-tenant Cloud Infrastructure Engine

// @securityDefinitions.apikey APIKeyAuth
// @in header
// @name X-API-Key

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
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

const (
	defaultDBInitTimeout = 120 * time.Second
)

// ErrMigrationDone signals that migrations have already completed.
var ErrMigrationDone = errors.New("migrations done")

// AppDeps holds the dependencies for the application, enabling DI for testing.
type AppDeps struct {
	LoadConfig         func(*slog.Logger) (*platform.Config, error)
	InitDatabase       func(context.Context, *platform.Config, *slog.Logger) (postgres.DB, error)
	RunMigrations      func(context.Context, postgres.DB, *slog.Logger) error
	InitRedis          func(context.Context, *platform.Config, *slog.Logger) (*redis.Client, error)
	InitComputeBackend func(*platform.Config, *slog.Logger) (ports.ComputeBackend, error)
	InitStorageBackend func(*platform.Config, *slog.Logger) (ports.StorageBackend, error)
	InitNetworkBackend func(*platform.Config, *slog.Logger) ports.NetworkBackend
	InitLBProxy        func(*platform.Config, ports.ComputeBackend, ports.InstanceRepository, ports.VpcRepository) (ports.LBProxyAdapter, error)
	InitRepositories   func(postgres.DB, *redis.Client) *setup.Repositories
	InitServices       func(setup.ServiceConfig) (*setup.Services, *setup.Workers, error)
	InitHandlers       func(*setup.Services, *platform.Config, *slog.Logger) *setup.Handlers
	SetupRouter        func(*platform.Config, *slog.Logger, *setup.Handlers, *setup.Services, ports.NetworkBackend) *gin.Engine
	NewHTTPServer      func(string, http.Handler) *http.Server
	StartHTTPServer    func(*http.Server) error
	ShutdownHTTPServer func(context.Context, *http.Server) error
	NotifySignals      func(chan<- os.Signal, ...os.Signal)
}

// DefaultDeps returns the production dependency wiring for the API server.
func DefaultDeps() AppDeps {
	return AppDeps{
		LoadConfig:         setup.LoadConfig,
		InitDatabase:       setup.InitDatabase,
		RunMigrations:      setup.RunMigrations,
		InitRedis:          setup.InitRedis,
		InitComputeBackend: setup.InitComputeBackend,
		InitStorageBackend: setup.InitStorageBackend,
		InitNetworkBackend: setup.InitNetworkBackend,
		InitLBProxy:        setup.InitLBProxy,
		InitRepositories:   setup.InitRepositories,
		InitServices:       setup.InitServices,
		InitHandlers:       setup.InitHandlers,
		SetupRouter:        setup.SetupRouter,
		NewHTTPServer: func(addr string, handler http.Handler) *http.Server {
			return &http.Server{Addr: addr, Handler: handler}
		},
		StartHTTPServer: func(s *http.Server) error {
			return s.ListenAndServe()
		},
		ShutdownHTTPServer: func(ctx context.Context, s *http.Server) error {
			return s.Shutdown(ctx)
		},
		NotifySignals: func(c chan<- os.Signal, sigs ...os.Signal) {
			signal.Notify(c, sigs...)
		},
	}
}

func main() {
	logger := setup.InitLogger()
	migrateOnly := flag.Bool("migrate-only", false, "run database migrations and exit")
	flag.Parse()

	deps := DefaultDeps()

	// Initialize Tracing (Opt-in)
	if tp := initTracing(logger); tp != nil {
		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				logger.Error("failed to shutdown tracer", "error", err)
			}
		}()
	}

	cfg, db, rdb, err := initInfrastructure(deps, logger, *migrateOnly)
	if err != nil {
		if *migrateOnly && errors.Is(err, ErrMigrationDone) {
			return
		}
		logger.Error("initialization failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	defer func() { _ = rdb.Close() }()

	compute, storage, network, lbProxy, err := initBackends(deps, cfg, logger, db, rdb)
	if err != nil {
		logger.Error("backend initialization failed", "error", err)
		os.Exit(1)
	}

	repos := deps.InitRepositories(db, rdb)
	svcs, workers, err := deps.InitServices(setup.ServiceConfig{
		Config: cfg, Repos: repos, Compute: compute, Storage: storage,
		Network: network, LBProxy: lbProxy, DB: db, RDB: rdb, Logger: logger,
	})
	if err != nil {
		logger.Error("services initialization failed", "error", err)
		os.Exit(1)
	}

	handlers := deps.InitHandlers(svcs, cfg, logger)
	r := deps.SetupRouter(cfg, logger, handlers, svcs, network)

	// Add Tracing Middleware if enabled
	if os.Getenv("TRACING_ENABLED") == "true" {
		r.Use(otelgin.Middleware("thecloud-api"))
	}

	runApplication(deps, cfg, logger, r, workers)
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

func initInfrastructure(deps AppDeps, logger *slog.Logger, migrateOnly bool) (*platform.Config, postgres.DB, *redis.Client, error) {
	cfg, err := deps.LoadConfig(logger)
	if err != nil {
		return nil, nil, nil, err
	}

	timeout := defaultDBInitTimeout
	if os.Getenv("DB_INIT_TIMEOUT") != "" {
		if t, err := time.ParseDuration(os.Getenv("DB_INIT_TIMEOUT")); err == nil {
			timeout = t
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	db, err := deps.InitDatabase(ctx, cfg, logger)
	if err != nil {
		return nil, nil, nil, err
	}

	if err := deps.RunMigrations(ctx, db, logger); err != nil {
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

	rdb, err := deps.InitRedis(ctx, cfg, logger)
	if err != nil {
		db.Close()
		return nil, nil, nil, err
	}

	return cfg, db, rdb, nil
}

func initBackends(deps AppDeps, cfg *platform.Config, logger *slog.Logger, db postgres.DB, rdb *redis.Client) (ports.ComputeBackend, ports.StorageBackend, ports.NetworkBackend, ports.LBProxyAdapter, error) {
	compute, err := deps.InitComputeBackend(cfg, logger)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	storage, err := deps.InitStorageBackend(cfg, logger)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	network := deps.InitNetworkBackend(cfg, logger)

	tmpRepos := deps.InitRepositories(db, rdb)
	lbProxy, err := deps.InitLBProxy(cfg, compute, tmpRepos.Instance, tmpRepos.Vpc)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return compute, storage, network, lbProxy, nil
}

func runApplication(deps AppDeps, cfg *platform.Config, logger *slog.Logger, r *gin.Engine, workers *setup.Workers) {
	role := os.Getenv("ROLE")
	if role == "" {
		role = "all"
	}

	wg := &sync.WaitGroup{}
	workerCtx, workerCancel := context.WithCancel(context.Background())

	if role == "worker" || role == "all" {
		runWorkers(workerCtx, wg, workers)
	}

	srv := deps.NewHTTPServer(":"+cfg.Port, r)

	if role == "api" || role == "all" {
		go func() {
			logger.Info("starting compute-api", "port", cfg.Port)
			if err := deps.StartHTTPServer(srv); err != nil && err != http.ErrServerClosed {
				logger.Error("failed to start server", "error", err)
				os.Exit(1)
			}
		}()
	} else {
		logger.Info("running in worker-only mode")
	}

	quit := make(chan os.Signal, 1)
	deps.NotifySignals(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := deps.ShutdownHTTPServer(ctx, srv); err != nil {
		logger.Error("server forced to shutdown", "error", err)
	}

	workerCancel()
	wg.Wait()
	logger.Info("server exited")
}

type runner interface {
	Run(context.Context, *sync.WaitGroup)
}

func startWorker(ctx context.Context, wg *sync.WaitGroup, w runner) {
	if w != nil {
		wg.Add(1)
		go w.Run(ctx, wg)
	}
}

func runWorkers(ctx context.Context, wg *sync.WaitGroup, workers *setup.Workers) {
	startWorker(ctx, wg, workers.LB)
	startWorker(ctx, wg, workers.AutoScaling)
	startWorker(ctx, wg, workers.Cron)
	startWorker(ctx, wg, workers.Container)
	startWorker(ctx, wg, workers.Provision)
	startWorker(ctx, wg, workers.Accounting)
	startWorker(ctx, wg, workers.Cluster)
	startWorker(ctx, wg, workers.Lifecycle)
	startWorker(ctx, wg, workers.ReplicaMonitor)
	startWorker(ctx, wg, workers.ClusterReconciler)
	startWorker(ctx, wg, workers.Healing)
}
