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

	svc := services.NewGlobalLBService(services.GlobalLBServiceParams{
		Repo: repo, LBRepo: lbRepo, GeoDNS: geoDNS, AuditSvc: audit, Logger: logger,
	})
	return svc, repo, lbRepo, geoDNS.(*mock.MockGeoDNS)
}

func TestGlobalLBCreate(t *testing.T) {
	t.Parallel()
	svc, repo, _, geoDNS := setupGlobalLBTest(t)
	// Fix context usage
	ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), uuid.New()), uuid.New())

	t.Run("success", func(t *testing.T) {
		name := "global-api"
		hostname := "api.global.com"
		policy := domain.RoutingLatency
		hc := domain.GlobalHealthCheckConfig{
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
			UserID:   uuid.New(), // some other user
		}
		repo.GLBs[existing.ID] = existing

		_, err := svc.Create(ctx, "new", "duplicate.com", domain.RoutingLatency, domain.GlobalHealthCheckConfig{})
		assert.Error(t, err)
	})

	t.Run("list filtering", func(t *testing.T) {
		userID1 := uuid.New()
		userID2 := uuid.New()
		ctx1 := appcontext.WithUserID(context.Background(), userID1)
		ctx2 := appcontext.WithUserID(context.Background(), userID2)

		_, _ = svc.Create(ctx1, "lb1", "lb1.com", domain.RoutingLatency, domain.GlobalHealthCheckConfig{})
		_, _ = svc.Create(ctx2, "lb2", "lb2.com", domain.RoutingLatency, domain.GlobalHealthCheckConfig{})

		list1, _ := svc.List(ctx1, userID1)
		assert.Len(t, list1, 1)
		assert.Equal(t, "lb1.com", list1[0].Hostname)

		list2, _ := svc.List(ctx2, userID2)
		assert.Len(t, list2, 1)
		assert.Equal(t, "lb2.com", list2[0].Hostname)
	})
}

func TestGlobalLBAddEndpoint(t *testing.T) {
	t.Parallel()
	svc, repo, lbRepo, geoDNS := setupGlobalLBTest(t)
	// Fix context usage
	ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), uuid.New()), uuid.New())

	// Create GLB
	userID := appcontext.UserIDFromContext(ctx)
	glb := &domain.GlobalLoadBalancer{
		ID:        uuid.New(),
		UserID:    userID,
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
		lbRepo.LBs[lbID] = &domain.LoadBalancer{ID: lbID, UserID: appcontext.UserIDFromContext(ctx)}

		ep, err := svc.AddEndpoint(ctx, glb.ID, "eu-west-1", "LB", &lbID, nil, 1, 1)
		require.NoError(t, err)
		assert.NotNil(t, ep)
		assert.Equal(t, lbID, *ep.TargetID)
	})

	t.Run("unauthorized lb endpoint", func(t *testing.T) {
		otherUserID := uuid.New()
		lbID := uuid.New()
		lbRepo.LBs[lbID] = &domain.LoadBalancer{ID: lbID, UserID: otherUserID}

		_, err := svc.AddEndpoint(ctx, glb.ID, "us-west-2", "LB", &lbID, nil, 1, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized access")
	})
}

func TestGlobalLBRemoveEndpoint(t *testing.T) {
	t.Parallel()
	svc, repo, _, geoDNS := setupGlobalLBTest(t)
	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)

	glb, err := svc.Create(ctx, "delete-ep-test", "delete.test.com", domain.RoutingLatency, domain.GlobalHealthCheckConfig{})
	require.NoError(t, err)

	ip := "1.2.3.4"
	ep, err := svc.AddEndpoint(ctx, glb.ID, "us-east-1", "IP", nil, &ip, 1, 1)
	require.NoError(t, err)

	t.Run("success with dns sync", func(t *testing.T) {
		err := svc.RemoveEndpoint(ctx, glb.ID, ep.ID)
		assert.NoError(t, err)

		// Verify repo
		eps, _ := repo.ListEndpoints(ctx, glb.ID)
		assert.Len(t, eps, 0)

		// Verify DNS sync (should be empty now)
		dnsRecs := geoDNS.Records[glb.Hostname]
		assert.Len(t, dnsRecs, 0)
	})
}
