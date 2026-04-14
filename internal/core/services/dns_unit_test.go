package services_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockDNSRepository struct {
	mock.Mock
}

func (m *MockDNSRepository) CreateZone(ctx context.Context, zone *domain.DNSZone) error {
	return m.Called(ctx, zone).Error(0)
}
func (m *MockDNSRepository) GetZoneByID(ctx context.Context, id uuid.UUID) (*domain.DNSZone, error) {
	args := m.Called(ctx, id)
	r0, _ := args.Get(0).(*domain.DNSZone)
	return r0, args.Error(1)
}
func (m *MockDNSRepository) GetZoneByName(ctx context.Context, name string) (*domain.DNSZone, error) {
	args := m.Called(ctx, name)
	r0, _ := args.Get(0).(*domain.DNSZone)
	return r0, args.Error(1)
}
func (m *MockDNSRepository) GetZoneByVPC(ctx context.Context, vpcID uuid.UUID) (*domain.DNSZone, error) {
	args := m.Called(ctx, vpcID)
	r0, _ := args.Get(0).(*domain.DNSZone)
	return r0, args.Error(1)
}
func (m *MockDNSRepository) ListZones(ctx context.Context) ([]*domain.DNSZone, error) {
	args := m.Called(ctx)
	r0, _ := args.Get(0).([]*domain.DNSZone)
	return r0, args.Error(1)
}
func (m *MockDNSRepository) UpdateZone(ctx context.Context, zone *domain.DNSZone) error {
	return m.Called(ctx, zone).Error(0)
}
func (m *MockDNSRepository) DeleteZone(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockDNSRepository) CreateRecord(ctx context.Context, record *domain.DNSRecord) error {
	return m.Called(ctx, record).Error(0)
}
func (m *MockDNSRepository) GetRecordByID(ctx context.Context, id uuid.UUID) (*domain.DNSRecord, error) {
	args := m.Called(ctx, id)
	r0, _ := args.Get(0).(*domain.DNSRecord)
	return r0, args.Error(1)
}
func (m *MockDNSRepository) ListRecordsByZone(ctx context.Context, zoneID uuid.UUID) ([]*domain.DNSRecord, error) {
	args := m.Called(ctx, zoneID)
	r0, _ := args.Get(0).([]*domain.DNSRecord)
	return r0, args.Error(1)
}
func (m *MockDNSRepository) GetRecordsByInstance(ctx context.Context, instanceID uuid.UUID) ([]*domain.DNSRecord, error) {
	args := m.Called(ctx, instanceID)
	r0, _ := args.Get(0).([]*domain.DNSRecord)
	return r0, args.Error(1)
}
func (m *MockDNSRepository) UpdateRecord(ctx context.Context, record *domain.DNSRecord) error {
	return m.Called(ctx, record).Error(0)
}
func (m *MockDNSRepository) DeleteRecord(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockDNSRepository) DeleteRecordsByInstance(ctx context.Context, instanceID uuid.UUID) error {
	return m.Called(ctx, instanceID).Error(0)
}

type MockDNSBackend struct {
	mock.Mock
}

func (m *MockDNSBackend) CreateZone(ctx context.Context, name string, nameservers []string) error {
	return m.Called(ctx, name, nameservers).Error(0)
}
func (m *MockDNSBackend) DeleteZone(ctx context.Context, name string) error {
	return m.Called(ctx, name).Error(0)
}
func (m *MockDNSBackend) GetZone(ctx context.Context, name string) (*ports.ZoneInfo, error) {
	args := m.Called(ctx, name)
	r0, _ := args.Get(0).(*ports.ZoneInfo)
	return r0, args.Error(1)
}
func (m *MockDNSBackend) AddRecords(ctx context.Context, zoneID string, records []ports.RecordSet) error {
	return m.Called(ctx, zoneID, records).Error(0)
}
func (m *MockDNSBackend) UpdateRecords(ctx context.Context, zoneID string, records []ports.RecordSet) error {
	return m.Called(ctx, zoneID, records).Error(0)
}
func (m *MockDNSBackend) DeleteRecords(ctx context.Context, zoneID, name, rType string) error {
	return m.Called(ctx, zoneID, name, rType).Error(0)
}
func (m *MockDNSBackend) ListRecords(ctx context.Context, zoneID string) ([]ports.RecordSet, error) {
	args := m.Called(ctx, zoneID)
	r0, _ := args.Get(0).([]ports.RecordSet)
	return r0, args.Error(1)
}

func TestDNSService_Unit(t *testing.T) {
	t.Run("Extended", testDNSServiceUnitExtended)
	t.Run("GetZoneByVPC", testGetZoneByVPC)
}

func testDNSServiceUnitExtended(t *testing.T) {
	repo := new(MockDNSRepository)
	backend := new(MockDNSBackend)
	vpcRepo := new(MockVpcRepo)
	auditSvc := new(MockAuditService)
	eventSvc := new(MockEventService)

	rbacSvc := new(MockRBACService)
	svc := services.NewDNSService(services.DNSServiceParams{
		Repo:     repo,
		Backend:  backend,
		RBAC:     rbacSvc,
		VpcRepo:  vpcRepo,
		AuditSvc: auditSvc,
		EventSvc: eventSvc,
		Logger:   slog.Default(),
	})

	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)

	t.Run("CreateZone", func(t *testing.T) {
		vpcID := uuid.New()
		rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		vpcRepo.On("GetByID", mock.Anything, vpcID).Return(&domain.VPC{ID: vpcID, Name: "test-vpc"}, nil).Once()
		repo.On("GetZoneByVPC", mock.Anything, vpcID).Return(nil, nil).Once()
		backend.On("CreateZone", mock.Anything, "example.com.", mock.Anything).Return(nil).Once()
		repo.On("CreateZone", mock.Anything, mock.Anything).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "dns.zone.create", "dns_zone", mock.Anything, mock.Anything).Return(nil).Once()

		zone, err := svc.CreateZone(ctx, vpcID, "example.com", "my zone")
		require.NoError(t, err)
		assert.NotNil(t, zone)
		assert.Equal(t, "example.com", zone.Name)
	})

	t.Run("DeleteZone", func(t *testing.T) {
		zoneID := uuid.New()
		zone := &domain.DNSZone{ID: zoneID, Name: "example.com", PowerDNSID: "example.com.", UserID: userID}
		rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		repo.On("GetZoneByID", mock.Anything, zoneID).Return(zone, nil).Once()
		backend.On("DeleteZone", mock.Anything, "example.com.").Return(nil).Once()
		repo.On("DeleteZone", mock.Anything, zoneID).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "dns.zone.delete", "dns_zone", zoneID.String(), mock.Anything).Return(nil).Once()

		err := svc.DeleteZone(ctx, zoneID.String())
		require.NoError(t, err)
	})

	t.Run("CreateRecord", func(t *testing.T) {
		zoneID := uuid.New()
		zone := &domain.DNSZone{ID: zoneID, Name: "example.com", PowerDNSID: "example.com."}
		rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		repo.On("GetZoneByID", mock.Anything, zoneID).Return(zone, nil).Once()
		backend.On("AddRecords", mock.Anything, "example.com.", mock.Anything).Return(nil).Once()
		repo.On("CreateRecord", mock.Anything, mock.Anything).Return(nil).Once()

		record, err := svc.CreateRecord(ctx, zoneID, "www", domain.RecordTypeA, "1.2.3.4", 3600, nil)
		require.NoError(t, err)
		assert.NotNil(t, record)
		assert.Equal(t, "www", record.Name)
	})

	t.Run("UpdateRecord", func(t *testing.T) {
		recordID := uuid.New()
		zoneID := uuid.New()
		record := &domain.DNSRecord{ID: recordID, ZoneID: zoneID, Name: "www", Type: domain.RecordTypeA, Content: "1.2.3.4"}
		zone := &domain.DNSZone{ID: zoneID, Name: "example.com", PowerDNSID: "example.com."}

		repo.On("GetRecordByID", mock.Anything, recordID).Return(record, nil).Once()
		rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		repo.On("GetZoneByID", mock.Anything, zoneID).Return(zone, nil).Once()
		backend.On("UpdateRecords", mock.Anything, "example.com.", mock.Anything).Return(nil).Once()
		repo.On("UpdateRecord", mock.Anything, mock.Anything).Return(nil).Once()

		updated, err := svc.UpdateRecord(ctx, recordID, "5.6.7.8", 3600, nil)
		require.NoError(t, err)
		assert.Equal(t, "5.6.7.8", updated.Content)
	})

	t.Run("ListZones", func(t *testing.T) {
		expectedZones := []*domain.DNSZone{{ID: uuid.New()}, {ID: uuid.New()}}
		repo.On("ListZones", mock.Anything).Return(expectedZones, nil).Once()

		zones, err := svc.ListZones(ctx)
		require.NoError(t, err)
		assert.Equal(t, expectedZones, zones)
	})

	t.Run("GetRecord", func(t *testing.T) {
		recordID := uuid.New()
		expectedRecord := &domain.DNSRecord{ID: recordID, Name: "test"}
		repo.On("GetRecordByID", mock.Anything, recordID).Return(expectedRecord, nil).Once()

		record, err := svc.GetRecord(ctx, recordID)
		require.NoError(t, err)
		assert.Equal(t, expectedRecord, record)
	})

	t.Run("ListRecords", func(t *testing.T) {
		zoneID := uuid.New()
		expectedRecords := []*domain.DNSRecord{{ID: uuid.New()}, {ID: uuid.New()}}
		repo.On("ListRecordsByZone", mock.Anything, zoneID).Return(expectedRecords, nil).Once()

		records, err := svc.ListRecords(ctx, zoneID)
		require.NoError(t, err)
		assert.Equal(t, expectedRecords, records)
	})

	t.Run("DeleteRecord", func(t *testing.T) {
		recordID := uuid.New()
		zoneID := uuid.New()
		record := &domain.DNSRecord{ID: recordID, ZoneID: zoneID, Name: "www", Type: domain.RecordTypeA}
		zone := &domain.DNSZone{ID: zoneID, Name: "example.com", PowerDNSID: "example.com."}

		repo.On("GetRecordByID", mock.Anything, recordID).Return(record, nil).Once()
		rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		repo.On("GetZoneByID", mock.Anything, zoneID).Return(zone, nil).Once()
		backend.On("DeleteRecords", mock.Anything, "example.com.", mock.Anything, "A").Return(nil).Once()
		repo.On("DeleteRecord", mock.Anything, recordID).Return(nil).Once()

		err := svc.DeleteRecord(ctx, recordID)
		require.NoError(t, err)
	})

	t.Run("UnregisterInstance", func(t *testing.T) {
		instID := uuid.New()
		zoneID := uuid.New()
		records := []*domain.DNSRecord{
			{ID: uuid.New(), ZoneID: zoneID, Name: "web-1", Type: domain.RecordTypeA},
		}
		zone := &domain.DNSZone{ID: zoneID, Name: "example.com", PowerDNSID: "example.com."}

		repo.On("GetRecordsByInstance", mock.Anything, instID).Return(records, nil).Once()
		rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		repo.On("GetZoneByID", mock.Anything, zoneID).Return(zone, nil).Once()
		backend.On("DeleteRecords", mock.Anything, "example.com.", "web-1.example.com.", "A").Return(nil).Once()
		repo.On("DeleteRecordsByInstance", mock.Anything, instID).Return(nil).Once()

		err := svc.UnregisterInstance(ctx, instID)
		require.NoError(t, err)
	})

	t.Run("RegisterInstance", func(t *testing.T) {
		vpcID := uuid.New()
		instID := uuid.New()
		inst := &domain.Instance{ID: instID, Name: "web-1", VpcID: &vpcID}
		zone := &domain.DNSZone{ID: uuid.New(), Name: "example.com", PowerDNSID: "example.com.", DefaultTTL: 300}

		repo.On("GetZoneByVPC", mock.Anything, vpcID).Return(zone, nil).Once()
		backend.On("AddRecords", mock.Anything, "example.com.", mock.Anything).Return(nil).Once()
		repo.On("CreateRecord", mock.Anything, mock.MatchedBy(func(r *domain.DNSRecord) bool {
			return r.Name == "web-1" && r.Type == domain.RecordTypeA && *r.InstanceID == instID
		})).Return(nil).Once()

		err := svc.RegisterInstance(ctx, inst, "10.0.0.10")
		require.NoError(t, err)
	})
}

// testGetZoneByVPC tests the GetZoneByVPC method with table-driven cases
func testGetZoneByVPC(t *testing.T) {
	repo := new(MockDNSRepository)
	backend := new(MockDNSBackend)
	vpcRepo := new(MockVpcRepo)
	auditSvc := new(MockAuditService)
	eventSvc := new(MockEventService)
	rbacSvc := new(MockRBACService)

	svc := services.NewDNSService(services.DNSServiceParams{
		Repo:     repo,
		Backend:  backend,
		RBAC:     rbacSvc,
		VpcRepo:  vpcRepo,
		AuditSvc: auditSvc,
		EventSvc: eventSvc,
		Logger:   slog.Default(),
	})

	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)

	testCases := []struct {
		name       string
		rbacErr    error
		repoZone   *domain.DNSZone
		repoErr    error
		expectErr  bool
	}{
		{
			name:     "Success",
			rbacErr:  nil,
			repoZone: &domain.DNSZone{ID: uuid.New(), VpcID: uuid.New(), Name: "vpc.internal"},
			repoErr:  nil,
			expectErr: false,
		},
		{
			name:      "Unauthorized",
			rbacErr:   errors.New(errors.Forbidden, "access denied"),
			repoZone:  nil,
			repoErr:   nil,
			expectErr: true,
		},
		{
			name:      "RepoError",
			rbacErr:   nil,
			repoZone:  nil,
			repoErr:   errors.New(errors.Internal, "db error"),
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create fresh vpcID for each subtest
			vpcID := uuid.New()

			rbacSvc.On("Authorize",
				mock.Anything, // ctx
				mock.Anything, // userID - uuid.UUID (passed by value)
				mock.Anything, // tenantID - uuid.UUID (passed by value)
				mock.Anything, // permission - domain.Permission
				mock.Anything, // resource - string
			).Return(tc.rbacErr).Once()

			// Only set up repo mock if RBAC passes (repo won't be called on RBAC failure)
			if tc.rbacErr == nil {
				repo.On("GetZoneByVPC", mock.Anything, vpcID).Return(tc.repoZone, tc.repoErr).Once()
			}

			zone, err := svc.GetZoneByVPC(ctx, vpcID)

			if tc.expectErr {
				require.Error(t, err)
				assert.Nil(t, zone)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.repoZone, zone)
			}

			if tc.rbacErr == nil {
				repo.AssertExpectations(t)
			}
			rbacSvc.AssertExpectations(t)
		})
	}
}
