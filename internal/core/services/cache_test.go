package services_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/docker"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupCacheServiceTest(t *testing.T) (*services.CacheService, ports.CacheRepository, *docker.DockerAdapter, ports.VpcRepository, context.Context) {
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewCacheRepository(db)
	vpcRepo := postgres.NewVpcRepository(db)

	compute, err := docker.NewDockerAdapter(slog.Default())
	require.NoError(t, err)

	eventRepo := postgres.NewEventRepository(db)
	eventSvc := services.NewEventService(eventRepo, nil, slog.Default())

	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(auditRepo)

	logger := slog.Default()

	svc := services.NewCacheService(repo, compute, vpcRepo, eventSvc, auditSvc, logger)

	return svc, repo, compute, vpcRepo, ctx
}

func TestCreateCacheSuccess(t *testing.T) {
	svc, repo, compute, _, ctx := setupCacheServiceTest(t)
	name := "test-cache-success"

	cache, err := svc.CreateCache(ctx, name, "7.2", 128, nil)

	assert.NoError(t, err)
	assert.NotNil(t, cache)
	assert.Equal(t, name, cache.Name)
	assert.Equal(t, domain.EngineRedis, cache.Engine)
	assert.NotEmpty(t, cache.ContainerID)

	// Verify instance creation by checking connectivity
	ip, err := compute.GetInstanceIP(ctx, cache.ContainerID)
	// Note: It might take a moment or fail if not yet ready, but Adapter retries.
	// For integration test with real docker, this should work eventually.
	assert.NoError(t, err)
	assert.NotEmpty(t, ip)

	// Clean up created container
	_ = compute.DeleteInstance(ctx, cache.ContainerID)

	// Verify in DB
	fetched, err := repo.GetByID(ctx, cache.ID)
	assert.NoError(t, err)
	assert.Equal(t, cache.ID, fetched.ID)
}

func TestCreateCacheWithVpc(t *testing.T) {
	svc, _, compute, vpcRepo, ctx := setupCacheServiceTest(t)

	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	vpcID := uuid.New()
	// Create actual docker network
	networkName := "net-" + vpcID.String()
	netID, err := compute.CreateNetwork(ctx, networkName)
	require.NoError(t, err)
	defer func() { _ = compute.DeleteNetwork(ctx, netID) }()

	vpc := &domain.VPC{
		ID:        vpcID,
		UserID:    userID,
		TenantID:  tenantID,
		Name:      "test-cache-vpc",
		CIDRBlock: "10.0.0.0/16",
		NetworkID: netID, // Use the real docker network ID
		Status:    "ACTIVE",
		CreatedAt: time.Now(),
	}
	err = vpcRepo.Create(ctx, vpc)
	require.NoError(t, err)

	name := "test-cache-vpc"
	cache, err := svc.CreateCache(ctx, name, "7.2", 128, &vpcID)
	assert.NoError(t, err)
	assert.Equal(t, &vpcID, cache.VpcID)

	// Cleanup
	_ = compute.DeleteInstance(ctx, cache.ContainerID)
}

func TestDeleteCacheSuccess(t *testing.T) {
	svc, repo, _, _, ctx := setupCacheServiceTest(t)

	// Setup: Create a cache first
	name := "test-cache-delete"
	cache, err := svc.CreateCache(ctx, name, "7.2", 128, nil)
	require.NoError(t, err)

	// Execute
	err = svc.DeleteCache(ctx, cache.ID.String())
	assert.NoError(t, err)

	// Verify deleted from DB
	_, err = repo.GetByID(ctx, cache.ID)
	assert.Error(t, err) // Should return not found error
}
