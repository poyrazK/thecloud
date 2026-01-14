// Package setup wires API dependencies and routes.
package setup

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/platform"
	"github.com/poyrazk/thecloud/internal/repositories/docker"
	"github.com/poyrazk/thecloud/internal/repositories/libvirt"
	"github.com/poyrazk/thecloud/internal/repositories/lvm"
	"github.com/poyrazk/thecloud/internal/repositories/noop"
	"github.com/poyrazk/thecloud/internal/repositories/ovs"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/redis/go-redis/v9"
)

// InitDatabase initializes the database connection, including optional read replicas
func InitDatabase(ctx context.Context, cfg *platform.Config, logger *slog.Logger) (postgres.DB, error) {
	primary, err := platform.NewDatabase(ctx, cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to primary database: %w", err)
	}

	var replica postgres.DB // Initialize as nil interface
	if cfg.DatabaseReadURL != "" {
		readCfg := *cfg
		readCfg.DatabaseURL = cfg.DatabaseReadURL

		pool, err := platform.NewDatabase(ctx, &readCfg, logger)
		if err != nil {
			logger.Warn("failed to connect to read replica, falling back to primary", "error", err)
			// We fallback to primary if replica fails, or strict fail?
			// Let's fallback for resilience, but in production we might want to know.
			// Given it's a scaling feature, let's just warn and proceed without replica (DualDB handles nil replica)
		} else {
			replica = pool
			logger.Info("connected to read replica database")
		}
	}

	return postgres.NewDualDB(primary, replica), nil
}

// InitRedis initializes the redis connection
func InitRedis(ctx context.Context, cfg *platform.Config, logger *slog.Logger) (*redis.Client, error) {
	return platform.InitRedis(ctx, cfg, logger)
}

// RunMigrations runs database migrations
func RunMigrations(ctx context.Context, db postgres.DB, logger *slog.Logger) error {
	return postgres.RunMigrations(ctx, db, logger)
}

// InitComputeBackend initializes the compute backend based on configuration.
func InitComputeBackend(cfg *platform.Config, logger *slog.Logger) (ports.ComputeBackend, error) {
	if cfg.ComputeBackend == "noop" {
		logger.Info("using no-op compute backend")
		return noop.NewNoopComputeBackend(), nil
	}
	if cfg.ComputeBackend == "libvirt" {
		logger.Info("using libvirt compute backend")
		return libvirt.NewLibvirtAdapter(logger, "")
	}
	logger.Info("using docker compute backend")
	return docker.NewDockerAdapter()
}

// InitStorageBackend initializes the storage backend (LVM or Noop)
func InitStorageBackend(cfg *platform.Config, logger *slog.Logger) (ports.StorageBackend, error) {
	if cfg.StorageBackend == "lvm" {
		logger.Info("using LVM storage backend", "vg", cfg.LvmVgName)
		return lvm.NewLvmAdapter(cfg.LvmVgName), nil
	}

	if cfg.StorageBackend == "noop" {
		logger.Info("using no-op storage backend")
		return noop.NewNoopStorageBackend(), nil
	}

	logger.Info("using no-op storage backend (default)")
	return noop.NewNoopStorageBackend(), nil
}

// InitNetworkBackend initializes the network backend based on configuration.
func InitNetworkBackend(cfg *platform.Config, logger *slog.Logger) ports.NetworkBackend {
	if cfg.NetworkBackend == "noop" {
		logger.Info("using no-op network backend")
		return noop.NewNoopNetworkAdapter(logger)
	}

	ovsAdapter, err := ovs.NewOvsAdapter(logger)
	if err != nil {
		logger.Warn("failed to initialize OVS adapter, using no-op network backend", "error", err)
		return noop.NewNoopNetworkAdapter(logger)
	}
	return ovsAdapter
}

// InitLBProxy initializes the Load Balancer proxy adapter
func InitLBProxy(cfg *platform.Config, computeBackend ports.ComputeBackend, instanceRepo ports.InstanceRepository, vpcRepo ports.VpcRepository) (ports.LBProxyAdapter, error) {
	if cfg.ComputeBackend == "libvirt" {
		return libvirt.NewLBProxyAdapter(computeBackend), nil
	}
	return docker.NewLBProxyAdapter(instanceRepo, vpcRepo)
}
