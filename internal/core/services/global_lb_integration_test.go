package services_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// StubGeoDNSBackend implements ports.GeoDNSBackend for testing purposes.
type StubGeoDNSBackend struct {
	CreatedRecords map[string][]domain.GlobalEndpoint
	DeletedRecords []string
}

func (s *StubGeoDNSBackend) CreateGeoRecord(ctx context.Context, hostname string, endpoints []domain.GlobalEndpoint) error {
	if s.CreatedRecords == nil {
		s.CreatedRecords = make(map[string][]domain.GlobalEndpoint)
	}
	// Copy to simulate persistence
	s.CreatedRecords[hostname] = endpoints
	return nil
}

func (s *StubGeoDNSBackend) DeleteGeoRecord(ctx context.Context, hostname string) error {
	s.DeletedRecords = append(s.DeletedRecords, hostname)
	delete(s.CreatedRecords, hostname)
	return nil
}

func TestGlobalLBServiceIntegration(t *testing.T) {
	db := setupDB(t)
	// Add global lb tables to cleanup if needed, but cleanDB in setup_test.go should be updated
	// For now we manually clean them or ensure cleanDB is updated.

	repo := postgres.NewGlobalLBRepository(db)
	lbRepo := postgres.NewLBRepository(db)
	geoDNS := &StubGeoDNSBackend{}

	// Real Audit Service
	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(auditRepo)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	svc := services.NewGlobalLBService(services.GlobalLBServiceParams{
		Repo:     repo,
		LBRepo:   lbRepo,
		GeoDNS:   geoDNS,
		AuditSvc: auditSvc,
		Logger:   logger,
	})

	t.Run("Scenario 1: Multi-User Isolation", func(t *testing.T) {
		cleanDB(t, db)
		ctxA := setupTestUser(t, db)
		userA := appcontext.UserIDFromContext(ctxA)

		ctxB := setupTestUser(t, db)
		userB := appcontext.UserIDFromContext(ctxB)

		hostname := "isolation.global.com"

		// 1. User A creates GLB
		glbA, err := svc.Create(ctxA, "user-a-lb", hostname, domain.RoutingLatency, domain.GlobalHealthCheckConfig{Protocol: "HTTP", Port: 80})
		require.NoError(t, err)

		// 2. User B tries to List (should not see User A's LB)
		listB, err := svc.List(ctxB, userB)
		require.NoError(t, err)
		assert.Empty(t, listB)

		// 3. User B tries to Delete User A's LB (should fail)
		err = svc.Delete(ctxB, glbA.ID, userB)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, errors.Unauthorized) || errors.Is(err, errors.NotFound))

		// 4. User A can see it
		listA, _ := svc.List(ctxA, userA)
		assert.Len(t, listA, 1)
	})

	t.Run("Scenario 2: Regional LB Ownership Validation", func(t *testing.T) {
		cleanDB(t, db)
		ctxA := setupTestUser(t, db)
		ctxB := setupTestUser(t, db)
		userB := appcontext.UserIDFromContext(ctxB)

		// 1. Create VPC and Subnet for User B (dependencies for LB)
		vpcRepo := postgres.NewVpcRepository(db)
		subnetRepo := postgres.NewSubnetRepository(db)

		tenantB := appcontext.TenantIDFromContext(ctxB)
		vpcB := &domain.VPC{
			ID:        uuid.New(),
			UserID:    userB,
			TenantID:  tenantB,
			Name:      "vpc-b",
			CIDRBlock: "10.0.0.0/16",
			Status:    "ACTIVE",
			NetworkID: "br-int",
			CreatedAt: time.Now(),
		}
		require.NoError(t, vpcRepo.Create(ctxB, vpcB))

		subnetB := &domain.Subnet{
			ID:        uuid.New(),
			VPCID:     vpcB.ID,
			UserID:    userB,
			Name:      "subnet-b",
			CIDRBlock: "10.0.1.0/24",
			Status:    "ACTIVE",
			CreatedAt: time.Now(),
		}
		require.NoError(t, subnetRepo.Create(ctxB, subnetB))

		// 2. Create Regional LB for User B
		lbB := &domain.LoadBalancer{
			ID:     uuid.New(),
			UserID: userB,
			VpcID:  vpcB.ID,
			Name:   "regional-b",
			Status: domain.LBStatusActive,
		}
		require.NoError(t, lbRepo.Create(ctxB, lbB))

		// 3. User A creates Global LB
		glbA, err := svc.Create(ctxA, "global-a", "multi-owner.com", domain.RoutingLatency, domain.GlobalHealthCheckConfig{Protocol: "HTTP", Port: 80})
		require.NoError(t, err)

		// 4. User A tries to add User B's Regional LB as endpoint (should fail)
		_, err = svc.AddEndpoint(ctxA, glbA.ID, "us-east-1", "LB", &lbB.ID, nil, 1, 1)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, errors.Unauthorized) || errors.Is(err, errors.NotFound))
	})

	t.Run("Scenario 3: DNS Sync Lifecycle", func(t *testing.T) {
		cleanDB(t, db)
		ctx := setupTestUser(t, db)
		hostname := "lifecycle.global.com"

		glb, _ := svc.Create(ctx, "life", hostname, domain.RoutingLatency, domain.GlobalHealthCheckConfig{Protocol: "HTTP", Port: 80})

		// Add 2 endpoints
		ip1 := "1.2.3.4"
		ip2 := "5.6.7.8"
		ep1, _ := svc.AddEndpoint(ctx, glb.ID, "us-east-1", "IP", nil, &ip1, 1, 1)
		_, err := svc.AddEndpoint(ctx, glb.ID, "eu-west-1", "IP", nil, &ip2, 1, 1)
		require.NoError(t, err)

		assert.Len(t, geoDNS.CreatedRecords[hostname], 2)

		// Remove 1 endpoint
		err = svc.RemoveEndpoint(ctx, glb.ID, ep1.ID)
		assert.NoError(t, err)

		// Verify DNS has exactly 1 remaining record
		assert.Len(t, geoDNS.CreatedRecords[hostname], 1)
		assert.Equal(t, ip2, *geoDNS.CreatedRecords[hostname][0].TargetIP)

		// Delete GLB
		err = svc.Delete(ctx, glb.ID, appcontext.UserIDFromContext(ctx))
		assert.NoError(t, err)

		// Verify DNS record is gone from stub
		_, exists := geoDNS.CreatedRecords[hostname]
		assert.False(t, exists)
	})
}
