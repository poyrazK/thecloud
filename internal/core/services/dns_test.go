package services

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	mocks "github.com/poyrazk/thecloud/internal/core/ports/mocks"
	theclouderrors "github.com/poyrazk/thecloud/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testDomain = "example.com"
	testPDNSID = "example.com."
	testIP     = "1.1.1.1"
)

type dnsTestContext struct {
	svc      *DNSService
	repo     *mocks.DNSRepository
	backend  *mocks.DNSBackend
	vpcRepo  *mocks.VpcRepository
	auditSvc *mocks.AuditService
	eventSvc *mocks.EventService
}

func setupDNSTest(t *testing.T) *dnsTestContext {
	repo := mocks.NewDNSRepository(t)
	backend := mocks.NewDNSBackend(t)
	vpcRepo := mocks.NewVpcRepository(t)
	auditSvc := mocks.NewAuditService(t)
	eventSvc := mocks.NewEventService(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := NewDNSService(DNSServiceParams{
		Repo:     repo,
		Backend:  backend,
		VpcRepo:  vpcRepo,
		AuditSvc: auditSvc,
		EventSvc: eventSvc,
		Logger:   logger,
	})

	return &dnsTestContext{
		svc:      svc,
		repo:     repo,
		backend:  backend,
		vpcRepo:  vpcRepo,
		auditSvc: auditSvc,
		eventSvc: eventSvc,
	}
}

func TestDNSServiceCreateZone(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tc := setupDNSTest(t)
		vpcID := uuid.New()
		tenantID := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), userID), tenantID)

		vpc := &domain.VPC{ID: vpcID, TenantID: tenantID, Name: "default-vpc"}
		tc.vpcRepo.On("GetByID", mock.Anything, vpcID).Return(vpc, nil)
		tc.repo.On("GetZoneByVPC", mock.Anything, vpcID).Return(nil, errors.New("not found"))
		tc.backend.On("CreateZone", mock.Anything, testPDNSID, mock.Anything).Return(nil)
		tc.repo.On("CreateZone", mock.Anything, mock.Anything).Return(nil)
		tc.auditSvc.On("Log", mock.Anything, userID, "dns.zone.create", "dns_zone", mock.Anything, mock.Anything).Return(nil)

		zone, err := tc.svc.CreateZone(ctx, vpcID, testDomain, "Test Zone")
		assert.NoError(t, err)
		assert.NotNil(t, zone)
		assert.Equal(t, "example.com", zone.Name)
	})

	t.Run("invalid name", func(t *testing.T) {
		tc := setupDNSTest(t)
		_, err := tc.svc.CreateZone(context.Background(), uuid.New(), "invalid_name", "")
		assert.Error(t, err)
		assert.True(t, theclouderrors.Is(err, theclouderrors.InvalidInput))
	})

	t.Run("vpc not found", func(t *testing.T) {
		tc := setupDNSTest(t)
		vpcID := uuid.New()
		tc.vpcRepo.On("GetByID", mock.Anything, vpcID).Return(nil, errors.New("not found"))

		_, err := tc.svc.CreateZone(context.Background(), vpcID, testDomain, "")
		assert.Error(t, err)
	})

	t.Run("zone already exists for vpc", func(t *testing.T) {
		tc := setupDNSTest(t)
		vpcID := uuid.New()
		tc.vpcRepo.On("GetByID", mock.Anything, vpcID).Return(&domain.VPC{ID: vpcID}, nil)
		tc.repo.On("GetZoneByVPC", mock.Anything, vpcID).Return(&domain.DNSZone{ID: uuid.New()}, nil)

		_, err := tc.svc.CreateZone(context.Background(), vpcID, testDomain, "")
		assert.Error(t, err)
		assert.True(t, theclouderrors.Is(err, theclouderrors.Conflict))
	})
}

func TestDNSServiceDeleteZone(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tc := setupDNSTest(t)
		zoneID := uuid.New()
		userID := uuid.New()
		zone := &domain.DNSZone{ID: zoneID, Name: testDomain, PowerDNSID: testPDNSID, UserID: userID}

		tc.repo.On("GetZoneByID", mock.Anything, zoneID).Return(zone, nil)
		tc.backend.On("DeleteZone", mock.Anything, zone.PowerDNSID).Return(nil)
		tc.repo.On("DeleteZone", mock.Anything, zoneID).Return(nil)
		tc.auditSvc.On("Log", mock.Anything, userID, "dns.zone.delete", "dns_zone", zoneID.String(), mock.Anything).Return(nil)

		err := tc.svc.DeleteZone(context.Background(), zoneID.String())
		assert.NoError(t, err)
	})
}

func TestDNSServiceCreateRecord(t *testing.T) {
	t.Run("success A record", func(t *testing.T) {
		tc := setupDNSTest(t)
		zoneID := uuid.New()
		zone := &domain.DNSZone{ID: zoneID, Name: testDomain, PowerDNSID: testPDNSID}

		tc.repo.On("GetZoneByID", mock.Anything, zoneID).Return(zone, nil)
		tc.backend.On("AddRecords", mock.Anything, zone.PowerDNSID, mock.MatchedBy(func(recs []ports.RecordSet) bool {
			return len(recs) == 1 && recs[0].Type == "A" && recs[0].Records[0] == testIP
		})).Return(nil)
		tc.repo.On("CreateRecord", mock.Anything, mock.Anything).Return(nil)

		rec, err := tc.svc.CreateRecord(context.Background(), zoneID, "web", domain.RecordTypeA, testIP, 3600, nil)
		assert.NoError(t, err)
		assert.NotNil(t, rec)
		assert.Equal(t, domain.RecordTypeA, rec.Type)
	})

	t.Run("success MX record", func(t *testing.T) {
		tc := setupDNSTest(t)
		zoneID := uuid.New()
		zone := &domain.DNSZone{ID: zoneID, Name: testDomain, PowerDNSID: testPDNSID}
		priority := 10
		mxContent := "mail." + testDomain

		tc.repo.On("GetZoneByID", mock.Anything, zoneID).Return(zone, nil)
		tc.backend.On("AddRecords", mock.Anything, zone.PowerDNSID, mock.MatchedBy(func(recs []ports.RecordSet) bool {
			return len(recs) == 1 && recs[0].Type == "MX" && recs[0].Records[0] == "10 "+mxContent
		})).Return(nil)
		tc.repo.On("CreateRecord", mock.Anything, mock.Anything).Return(nil)

		rec, err := tc.svc.CreateRecord(context.Background(), zoneID, "@", domain.RecordTypeMX, mxContent, 3600, &priority)
		assert.NoError(t, err)
		assert.NotNil(t, rec)
		assert.Equal(t, domain.RecordTypeMX, rec.Type)
	})
}

func TestDNSServiceRegisterInstance(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tc := setupDNSTest(t)
		vpcID := uuid.New()
		instID := uuid.New()
		instance := &domain.Instance{ID: instID, VpcID: &vpcID, Name: "my-vm"}
		zone := &domain.DNSZone{ID: uuid.New(), Name: "vpc.internal", PowerDNSID: "vpc.internal.", DefaultTTL: 300}

		tc.repo.On("GetZoneByVPC", mock.Anything, vpcID).Return(zone, nil)
		tc.backend.On("AddRecords", mock.Anything, zone.PowerDNSID, mock.MatchedBy(func(recs []ports.RecordSet) bool {
			return len(recs) == 1 && recs[0].Name == "my-vm.vpc.internal." && recs[0].Records[0] == "10.0.0.5"
		})).Return(nil)
		tc.repo.On("CreateRecord", mock.Anything, mock.Anything).Return(nil)

		err := tc.svc.RegisterInstance(context.Background(), instance, "10.0.0.5")
		assert.NoError(t, err)
	})

	t.Run("no vpc", func(t *testing.T) {
		tc := setupDNSTest(t)
		instance := &domain.Instance{ID: uuid.New(), Name: "my-vm", VpcID: nil}

		err := tc.svc.RegisterInstance(context.Background(), instance, "1.1.1.1")
		assert.NoError(t, err) // Should skip silently
	})
}

func TestDNSServiceUnregisterInstance(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tc := setupDNSTest(t)
		instID := uuid.New()
		zoneID := uuid.New()
		zone := &domain.DNSZone{ID: zoneID, Name: "vpc.internal", PowerDNSID: "vpc.internal."}
		records := []*domain.DNSRecord{
			{ID: uuid.New(), ZoneID: zoneID, Name: "my-vm", Type: domain.RecordTypeA},
		}

		tc.repo.On("GetRecordsByInstance", mock.Anything, instID).Return(records, nil)
		tc.repo.On("GetZoneByID", mock.Anything, zoneID).Return(zone, nil)
		tc.backend.On("DeleteRecords", mock.Anything, zone.PowerDNSID, "my-vm.vpc.internal.", "A").Return(nil)
		tc.repo.On("DeleteRecordsByInstance", mock.Anything, instID).Return(nil)

		err := tc.svc.UnregisterInstance(context.Background(), instID)
		assert.NoError(t, err)
	})
}
