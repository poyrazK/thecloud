//go:build !test
// +build !test

// Package main provides the API server entrypoint.
package main

import (
	"context"
	"net/http"
	"os/signal"

	"github.com/poyrazk/thecloud/internal/api/setup"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/platform"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/redis/go-redis/v9"
)

// Default indirections for production builds.
var (
	loadConfigFunc         = setup.LoadConfig
	initDatabaseFunc       = setup.InitDatabase
	runMigrationsFunc      = setup.RunMigrations
	initRedisFunc          = setup.InitRedis
	initComputeBackendFunc = setup.InitComputeBackend
	initStorageBackendFunc = setup.InitStorageBackend
	initNetworkBackendFunc = setup.InitNetworkBackend
	initLBProxyFunc        = setup.InitLBProxy
	initRepositoriesFunc   = setup.InitRepositories
	initServicesFunc       = setup.InitServices
	initHandlersFunc       = setup.InitHandlers
	setupRouterFunc        = setup.SetupRouter

	newHTTPServer = func(addr string, handler http.Handler) *http.Server {
		return &http.Server{Addr: addr, Handler: handler}
	}
	startHTTPServer    = func(srv *http.Server) error { return srv.ListenAndServe() }
	shutdownHTTPServer = func(ctx context.Context, srv *http.Server) error { return srv.Shutdown(ctx) }
	notifySignals      = signal.Notify
)

// silence unused import linters when hooks are stubbed in tests
var (
	_ *platform.Config
	_ ports.ComputeBackend
	_ ports.StorageBackend
	_ ports.NetworkBackend
	_ ports.LBProxyAdapter
	_ postgres.DB
	_ *redis.Client
)
