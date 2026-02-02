package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupGlobalLBTest(t *testing.T) (*services.GlobalLBService, *mock.MockGlobalLBRepo, *mock.MockLBRepo, *mock.MockGeoDNS) {
	repo := mock.NewMockGlobalLBRepo()
	lbRepo := mock.NewMockLBRepo()
	geoDNS := mock.NewMockGeoDNS()
	audit := mock.NewMockAuditService()
	logger := mock.NewNoopLogger()

	svc := services.NewGlobalLBService(repo, lbRepo, geoDNS, audit, logger)
	return svc, repo, lbRepo, geoDNS.(*mock.MockGeoDNS)
}

func TestGlobalLBCreate(t *testing.T) {
	svc, repo, _, geoDNS := setupGlobalLBTest(t)
	ctx := appcontext.WithUser(context.Background(), uuid.New(), uuid.New())

	t.Run("success", func(t *testing.T) {
		name := "global-api"
		hostname := "api.global.com"
		policy := domain.RoutingLatency
		hc := domain.HealthCheckConfig{
			Protocol:    "HTTP",
			Port:        80,
			Path:        "/health",
			IntervalSec: 30,
		}

		glb, err := svc.Create(ctx, name, hostname, policy, hc)
		require.NoError(t, err)
		assert.NotNil(t, glb)
		assert.Equal(t, name, glb.Name)
		assert.Equal(t, hostname, glb.Hostname)
		assert.Equal(t, "ACTIVE", glb.Status)

		// Check Repo
		assert.NotNil(t, repo.GLBs[glb.ID])

		// Check DNS
		_, exists := geoDNS.Records[hostname]
		assert.True(t, exists, "DNS record should be created")
	})

	t.Run("duplicate hostname", func(t *testing.T) {
		// Existing
		existing := &domain.GlobalLoadBalancer{
			ID:       uuid.New(),
			Hostname: "duplicate.com",
		}
		repo.GLBs[existing.ID] = existing

		_, err := svc.Create(ctx, "new", "duplicate.com", domain.RoutingLatency, domain.HealthCheckConfig{})
		assert.Error(t, err)
	})
}

func TestGlobalLBAddEndpoint(t *testing.T) {
	svc, repo, lbRepo, geoDNS := setupGlobalLBTest(t)
	ctx := appcontext.WithUser(context.Background(), uuid.New(), uuid.New())

	// Create GLB
	glb := &domain.GlobalLoadBalancer{
		ID:        uuid.New(),
		Hostname:  "api.test.com",
		Status:    "ACTIVE",
		Endpoints: []*domain.GlobalEndpoint{},
	}
	repo.GLBs[glb.ID] = glb
	geoDNS.Records[glb.Hostname] = nil

	t.Run("add ip endpoint", func(t *testing.T) {
		ip := "1.2.3.4"
		ep, err := svc.AddEndpoint(ctx, glb.ID, "us-east-1", "IP", nil, &ip, 1, 1)
		require.NoError(t, err)
		assert.NotNil(t, ep)
		assert.Equal(t, ip, *ep.TargetIP)

		// Verify repo has endpoint
		assert.Len(t, repo.Endpoints[glb.ID], 1)

		// Verify DNS updated (mock logic depends on implementation detail, assuming it updates)
		// Our mock implementation updates `Records` map.
		records := geoDNS.Records[glb.Hostname]
		assert.Len(t, records, 1)
		assert.Equal(t, ip, *records[0].TargetIP)
	})

	t.Run("add lb endpoint", func(t *testing.T) {
		// Mock existing LB
		lbID := uuid.New()
		lbRepo.LBs[lbID] = &domain.LoadBalancer{ID: lbID}

		ep, err := svc.AddEndpoint(ctx, glb.ID, "eu-west-1", "LB", &lbID, nil, 1, 1)
		require.NoError(t, err)
		assert.NotNil(t, ep)
		assert.Equal(t, lbID, *ep.TargetID)
	})
}
