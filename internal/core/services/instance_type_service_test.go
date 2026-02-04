package services_test

import (
	"context"
	"testing"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupInstanceTypeServiceTest(t *testing.T) (ports.InstanceTypeService, *postgres.InstanceTypeRepository, context.Context) {
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewInstanceTypeRepository(db)
	svc := services.NewInstanceTypeService(repo)
	return svc, repo, ctx
}

func TestInstanceTypeService_List(t *testing.T) {
	svc, repo, ctx := setupInstanceTypeServiceTest(t)

	it1 := &domain.InstanceType{ID: "c1.medium", Name: "Medium", VCPUs: 2, MemoryMB: 4096}
	it2 := &domain.InstanceType{ID: "c1.large", Name: "Large", VCPUs: 4, MemoryMB: 8192}

	_, err := repo.Create(ctx, it1)
	require.NoError(t, err)
	_, err = repo.Create(ctx, it2)
	require.NoError(t, err)

	result, err := svc.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}
