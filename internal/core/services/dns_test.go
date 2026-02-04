package services_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/poyrazk/thecloud/internal/repositories/noop"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDNSServiceTest(t *testing.T) (*services.DNSService, ports.DNSRepository, ports.VpcRepository, *postgres.InstanceRepository, *pgxpool.Pool, context.Context) {
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewDNSRepository(db)
	vpcRepo := postgres.NewVpcRepository(db)
	instRepo := postgres.NewInstanceRepository(db)
	backend := noop.NewNoopDNSBackend()

	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(auditRepo)

	eventRepo := postgres.NewEventRepository(db)
	eventSvc := services.NewEventService(eventRepo, nil, slog.Default())

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := services.NewDNSService(services.DNSServiceParams{
		Repo:     repo,
		Backend:  backend,
		VpcRepo:  vpcRepo,
		AuditSvc: auditSvc,
		EventSvc: eventSvc,
		Logger:   logger,
	})

	return svc, repo, vpcRepo, instRepo, db, ctx
}

func TestDNSServiceCreateZone(t *testing.T) {
	svc, repo, vpcRepo, _, _, ctx := setupDNSServiceTest(t)
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	vpc := &domain.VPC{
		ID:       uuid.New(),
		UserID:   userID,
		TenantID: tenantID,
		Name:     "dns-vpc-" + uuid.New().String(),
	}
	err := vpcRepo.Create(ctx, vpc)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		name := "example.com"
		zone, err := svc.CreateZone(ctx, vpc.ID, name, "Test Zone")
		assert.NoError(t, err)
		assert.NotNil(t, zone)
		assert.Equal(t, name, zone.Name)
		assert.Equal(t, vpc.ID, zone.VpcID)

		// Verify DB
		fetched, err := repo.GetZoneByID(ctx, zone.ID)
		assert.NoError(t, err)
		assert.Equal(t, zone.ID, fetched.ID)
	})

	t.Run("duplicate zone", func(t *testing.T) {
		// Attempting to create a zone that already exists for the same VPC should result in a failure.
		// This verifies that the service enforcing unique constraints or handling DB unique violations correctly.
		_, err := svc.CreateZone(ctx, vpc.ID, "example.com", "Duplicate")
		assert.Error(t, err)
	})
}

func TestDNSServiceRegisterInstance_NoZone(t *testing.T) {
	svc, _, vpcRepo, instRepo, _, ctx := setupDNSServiceTest(t)
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	// VPC without Zone
	vpc := &domain.VPC{ID: uuid.New(), UserID: userID, TenantID: tenantID, Name: "no-zone-vpc-" + uuid.New().String()}
	_ = vpcRepo.Create(ctx, vpc)

	inst := &domain.Instance{
		ID:       uuid.New(),
		UserID:   userID,
		TenantID: tenantID,
		VpcID:    &vpc.ID,
		Name:     "no-zone-inst",
		Status:   domain.StatusRunning,
	}
	_ = instRepo.Create(ctx, inst)

	// Should not error, just silence
	err := svc.RegisterInstance(ctx, inst, "1.2.3.4")
	assert.NoError(t, err)
}

func TestDNSServiceDeleteZone(t *testing.T) {
	svc, repo, vpcRepo, _, _, ctx := setupDNSServiceTest(t)
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	vpc := &domain.VPC{ID: uuid.New(), UserID: userID, TenantID: tenantID, Name: "del-zone-vpc-" + uuid.New().String()}
	_ = vpcRepo.Create(ctx, vpc)

	zone, err := svc.CreateZone(ctx, vpc.ID, "delete.com", "")
	require.NoError(t, err)

	err = svc.DeleteZone(ctx, zone.ID.String())
	assert.NoError(t, err)

	// Verify Deleted
	_, err = repo.GetZoneByID(ctx, zone.ID)
	assert.Error(t, err)
}

func TestDNSServiceRecords(t *testing.T) {
	svc, _, vpcRepo, _, _, ctx := setupDNSServiceTest(t)
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	vpc := &domain.VPC{ID: uuid.New(), UserID: userID, TenantID: tenantID, Name: "rec-vpc-" + uuid.New().String()}
	_ = vpcRepo.Create(ctx, vpc)

	zone, err := svc.CreateZone(ctx, vpc.ID, "records.com", "")
	require.NoError(t, err)

	t.Run("Create A Record", func(t *testing.T) {
		rec, err := svc.CreateRecord(ctx, zone.ID, "www", domain.RecordTypeA, "1.2.3.4", 300, nil)
		assert.NoError(t, err)
		assert.Equal(t, "www", rec.Name)
		assert.Equal(t, "1.2.3.4", rec.Content)

		// Get
		fetched, err := svc.GetRecord(ctx, rec.ID)
		assert.NoError(t, err)
		assert.Equal(t, rec.ID, fetched.ID)
	})

	t.Run("Update Record", func(t *testing.T) {
		rec, _ := svc.CreateRecord(ctx, zone.ID, "api", domain.RecordTypeA, "5.6.7.8", 300, nil)

		updated, err := svc.UpdateRecord(ctx, rec.ID, "9.9.9.9", 600, nil)
		assert.NoError(t, err)
		assert.Equal(t, "9.9.9.9", updated.Content)
		assert.Equal(t, 600, updated.TTL)
	})

	t.Run("Delete Record", func(t *testing.T) {
		rec, _ := svc.CreateRecord(ctx, zone.ID, "del", domain.RecordTypeA, "1.1.1.1", 300, nil)

		err := svc.DeleteRecord(ctx, rec.ID)
		assert.NoError(t, err)

		_, err = svc.GetRecord(ctx, rec.ID)
		assert.Error(t, err)
	})

	t.Run("List Records", func(t *testing.T) {
		list, err := svc.ListRecords(ctx, zone.ID)
		assert.NoError(t, err)
		// We created 'www', 'api' (updated), and 'del' (deleted). So 2 remaining.
		assert.Len(t, list, 2)
	})
}

func TestDNSServiceRegisterInstance(t *testing.T) {
	svc, repo, vpcRepo, instRepo, _, ctx := setupDNSServiceTest(t)
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	vpc := &domain.VPC{ID: uuid.New(), UserID: userID, TenantID: tenantID, Name: "reg-vpc-" + uuid.New().String()}
	_ = vpcRepo.Create(ctx, vpc)

	// Ensure an internal DNS zone exists for the VPC.
	// In a complete system, this might be auto-provisioned, but for this integration test scope,
	// we explicitly create it to satisfy the dependency.
	zoneName := "internal.vpc"
	_, err := svc.CreateZone(ctx, vpc.ID, zoneName, "")
	require.NoError(t, err)

	inst := &domain.Instance{
		ID:       uuid.New(),
		UserID:   userID,
		TenantID: tenantID,
		VpcID:    &vpc.ID,
		Name:     "my-inst",
		Status:   domain.StatusRunning,
	}
	err = instRepo.Create(ctx, inst)
	require.NoError(t, err)

	// Register
	err = svc.RegisterInstance(ctx, inst, "10.0.0.5")
	assert.NoError(t, err)

	// Verify Record Created
	zone, err := repo.GetZoneByName(ctx, zoneName)
	require.NoError(t, err, "GetZoneByName failed")

	records, err := repo.ListRecordsByZone(ctx, zone.ID)
	assert.NoError(t, err)
	t.Logf("Found %d records in zone %s", len(records), zone.ID)

	found := false
	for _, r := range records {
		t.Logf("Record: %s %s %s", r.Name, r.Type, r.Content)
		if r.Name == "my-inst" && r.Content == "10.0.0.5" {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected A record for instance")

	// Unregister
	err = svc.UnregisterInstance(ctx, inst.ID)
	assert.NoError(t, err)

	// Verify gone
	recordsAfter, _ := repo.ListRecordsByZone(ctx, zone.ID)
	foundAfter := false
	for _, r := range recordsAfter {
		if r.Name == "my-inst" {
			foundAfter = true
			break
		}
	}
	assert.False(t, foundAfter, "Record should be removed")
}

type FaultyDNSBackend struct {
	ports.DNSBackend
	FailCreate bool
	FailAdd    bool
}

func (f *FaultyDNSBackend) CreateZone(ctx context.Context, zoneName string, ns []string) error {
	if f.FailCreate {
		return errors.New(errors.Internal, "simulated dns backend failure")
	}
	return f.DNSBackend.CreateZone(ctx, zoneName, ns)
}

func (f *FaultyDNSBackend) AddRecords(ctx context.Context, zoneName string, records []ports.RecordSet) error {
	if f.FailAdd {
		return errors.New(errors.Internal, "simulated dns backend failure")
	}
	return f.DNSBackend.AddRecords(ctx, zoneName, records)
}

func TestDNSService_BackendError(t *testing.T) {
	_, repo, vpcRepo, _, db, ctx := setupDNSServiceTest(t)
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	vpc := &domain.VPC{ID: uuid.New(), UserID: userID, TenantID: tenantID, Name: "fault-vpc-" + uuid.New().String()}
	_ = vpcRepo.Create(ctx, vpc)

	// Replace backend with a faulty one
	faulty := &FaultyDNSBackend{DNSBackend: noop.NewNoopDNSBackend(), FailCreate: true}

	auditSvc := services.NewAuditService(postgres.NewAuditRepository(db))
	eventSvc := services.NewEventService(postgres.NewEventRepository(db), nil, slog.Default())

	faultySvc := services.NewDNSService(services.DNSServiceParams{
		Repo:     repo,
		Backend:  faulty,
		VpcRepo:  vpcRepo,
		AuditSvc: auditSvc,
		EventSvc: eventSvc,
		Logger:   slog.Default(),
	})

	t.Run("CreateZone Failure", func(t *testing.T) {
		_, err := faultySvc.CreateZone(ctx, vpc.ID, "fail.com", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DNS backend")
	})

	t.Run("RegisterInstance Failure", func(t *testing.T) {
		// First create a successful zone with real backend
		realSvc := services.NewDNSService(services.DNSServiceParams{
			Repo:     repo,
			Backend:  noop.NewNoopDNSBackend(),
			VpcRepo:  vpcRepo,
			AuditSvc: auditSvc,
			EventSvc: eventSvc,
			Logger:   slog.Default(),
		})
		_, err := realSvc.CreateZone(ctx, vpc.ID, "ok.com", "")
		require.NoError(t, err)

		// Now use faulty backend for registration
		faulty.FailCreate = false
		faulty.FailAdd = true
		inst := &domain.Instance{ID: uuid.New(), VpcID: &vpc.ID, Name: "fail-inst"}
		err = faultySvc.RegisterInstance(ctx, inst, "1.1.1.1")
		assert.Error(t, err)
	})
}
