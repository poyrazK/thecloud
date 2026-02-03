package services_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/docker"
	"github.com/poyrazk/thecloud/internal/repositories/noop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupHealthServiceTest(t *testing.T) (*services.HealthServiceImpl, context.Context) {
	db := setupDB(t)
	// No cleanDB needed for health check as it only pings.
	ctx := setupTestUser(t, db)

	compute, err := docker.NewDockerAdapter(slog.Default())
	require.NoError(t, err)

	// Use noop for cluster as we don't have a simple way to test k8s ping without a real cluster.
	cluster := &noop.NoopClusterService{}

	svc := services.NewHealthServiceImpl(db, compute, cluster)
	return svc, ctx
}

func TestHealthServiceCheck(t *testing.T) {
	svc, ctx := setupHealthServiceTest(t)

	res := svc.Check(ctx)

	assert.Equal(t, "UP", res.Status)
	assert.Equal(t, "CONNECTED", res.Checks["database_primary"])
	assert.Equal(t, "CONNECTED", res.Checks["docker"])
	assert.Equal(t, "OK", res.Checks["kubernetes_service"])
}
