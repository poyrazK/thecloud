package main

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
	"github.com/redis/go-redis/v9"
)

// ... (omitted comments) ...

var ErrMigrationDone = errors.New("migrations done")

func main() {
	logger := setup.InitLogger()
	migrateOnly := flag.Bool("migrate-only", false, "run database migrations and exit")
	flag.Parse()

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

	repos := setup.InitRepositories(db, rdb)
	svcs, workers, err := setup.InitServices(setup.ServiceConfig{
		Config: cfg, Repos: repos, Compute: compute, Storage: storage,
		Network: network, LBProxy: lbProxy, DB: db, RDB: rdb, Logger: logger,
	})
	if err != nil {
		logger.Error("services initialization failed", "error", err)
		os.Exit(1)
	}

	handlers := setup.InitHandlers(svcs, logger)
	r := setup.SetupRouter(cfg, logger, handlers, svcs, network)

	runApplication(cfg, logger, r, workers)
}

func initInfrastructure(logger *slog.Logger, migrateOnly bool) (*platform.Config, postgres.DB, *redis.Client, error) {
	cfg, err := setup.LoadConfig(logger)
	if err != nil {
		return nil, nil, nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := setup.InitDatabase(ctx, cfg, logger)
	if err != nil {
		return nil, nil, nil, err
	}

	if err := setup.RunMigrations(ctx, db, logger); err != nil {
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

	rdb, err := setup.InitRedis(ctx, cfg, logger)
	if err != nil {
		db.Close()
		return nil, nil, nil, err
	}

	return cfg, db, rdb, nil
}

func initBackends(cfg *platform.Config, logger *slog.Logger, db postgres.DB, rdb *redis.Client) (ports.ComputeBackend, ports.StorageBackend, ports.NetworkBackend, ports.LBProxyAdapter, error) {
	compute, err := setup.InitComputeBackend(cfg, logger)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	storage, err := setup.InitStorageBackend(cfg, logger)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	network := setup.InitNetworkBackend(cfg, logger)

	tmpRepos := setup.InitRepositories(db, rdb)
	lbProxy, err := setup.InitLBProxy(cfg, compute, tmpRepos.Instance, tmpRepos.Vpc)
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

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	if role == "api" || role == "all" {
		go func() {
			logger.Info("starting compute-api", "port", cfg.Port)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Error("failed to start server", "error", err)
				os.Exit(1)
			}
		}()
	} else {
		logger.Info("running in worker-only mode")
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
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
