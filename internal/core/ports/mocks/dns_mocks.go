package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/stretchr/testify/mock"
)

// DNSRepository is a mock for ports.DNSRepository
type DNSRepository struct {
	mock.Mock
}

func NewDNSRepository(t mock.TestingT) *DNSRepository {
	m := &DNSRepository{}
	m.Test(t)
	return m
}

func (m *DNSRepository) CreateZone(ctx context.Context, zone *domain.DNSZone) error {
	args := m.Called(ctx, zone)
	return args.Error(0)
}

func (m *DNSRepository) GetZoneByID(ctx context.Context, id uuid.UUID) (*domain.DNSZone, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DNSZone), args.Error(1)
}

func (m *DNSRepository) GetZoneByName(ctx context.Context, name string) (*domain.DNSZone, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DNSZone), args.Error(1)
}

func (m *DNSRepository) GetZoneByVPC(ctx context.Context, vpcID uuid.UUID) (*domain.DNSZone, error) {
	args := m.Called(ctx, vpcID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DNSZone), args.Error(1)
}

func (m *DNSRepository) ListZones(ctx context.Context) ([]*domain.DNSZone, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.DNSZone), args.Error(1)
}

func (m *DNSRepository) UpdateZone(ctx context.Context, zone *domain.DNSZone) error {
	args := m.Called(ctx, zone)
	return args.Error(0)
}

func (m *DNSRepository) DeleteZone(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *DNSRepository) CreateRecord(ctx context.Context, record *domain.DNSRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *DNSRepository) GetRecordByID(ctx context.Context, id uuid.UUID) (*domain.DNSRecord, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DNSRecord), args.Error(1)
}

func (m *DNSRepository) ListRecordsByZone(ctx context.Context, zoneID uuid.UUID) ([]*domain.DNSRecord, error) {
	args := m.Called(ctx, zoneID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.DNSRecord), args.Error(1)
}

func (m *DNSRepository) GetRecordsByInstance(ctx context.Context, instanceID uuid.UUID) ([]*domain.DNSRecord, error) {
	args := m.Called(ctx, instanceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.DNSRecord), args.Error(1)
}

func (m *DNSRepository) UpdateRecord(ctx context.Context, record *domain.DNSRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *DNSRepository) DeleteRecord(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *DNSRepository) DeleteRecordsByInstance(ctx context.Context, instanceID uuid.UUID) error {
	args := m.Called(ctx, instanceID)
	return args.Error(0)
}

// DNSBackend is a mock for ports.DNSBackend
type DNSBackend struct {
	mock.Mock
}

func NewDNSBackend(t mock.TestingT) *DNSBackend {
	m := &DNSBackend{}
	m.Test(t)
	return m
}

func (m *DNSBackend) CreateZone(ctx context.Context, zoneName string, nameservers []string) error {
	args := m.Called(ctx, zoneName, nameservers)
	return args.Error(0)
}

func (m *DNSBackend) DeleteZone(ctx context.Context, zoneName string) error {
	args := m.Called(ctx, zoneName)
	return args.Error(0)
}

func (m *DNSBackend) GetZone(ctx context.Context, zoneName string) (*ports.ZoneInfo, error) {
	args := m.Called(ctx, zoneName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ports.ZoneInfo), args.Error(1)
}

func (m *DNSBackend) AddRecords(ctx context.Context, zoneName string, records []ports.RecordSet) error {
	args := m.Called(ctx, zoneName, records)
	return args.Error(0)
}

func (m *DNSBackend) UpdateRecords(ctx context.Context, zoneName string, records []ports.RecordSet) error {
	args := m.Called(ctx, zoneName, records)
	return args.Error(0)
}

func (m *DNSBackend) DeleteRecords(ctx context.Context, zoneName, name, recordType string) error {
	args := m.Called(ctx, zoneName, name, recordType)
	return args.Error(0)
}

func (m *DNSBackend) ListRecords(ctx context.Context, zoneName string) ([]ports.RecordSet, error) {
	args := m.Called(ctx, zoneName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ports.RecordSet), args.Error(1)
}

// DNSService is a mock for ports.DNSService
type DNSService struct {
	mock.Mock
}

func NewDNSService(t mock.TestingT) *DNSService {
	m := &DNSService{}
	m.Test(t)
	return m
}

func (m *DNSService) CreateZone(ctx context.Context, vpcID uuid.UUID, name, description string) (*domain.DNSZone, error) {
	args := m.Called(ctx, vpcID, name, description)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DNSZone), args.Error(1)
}

func (m *DNSService) GetZone(ctx context.Context, idOrName string) (*domain.DNSZone, error) {
	args := m.Called(ctx, idOrName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DNSZone), args.Error(1)
}

func (m *DNSService) GetZoneByVPC(ctx context.Context, vpcID uuid.UUID) (*domain.DNSZone, error) {
	args := m.Called(ctx, vpcID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DNSZone), args.Error(1)
}

func (m *DNSService) ListZones(ctx context.Context) ([]*domain.DNSZone, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.DNSZone), args.Error(1)
}

func (m *DNSService) DeleteZone(ctx context.Context, idOrName string) error {
	args := m.Called(ctx, idOrName)
	return args.Error(0)
}

func (m *DNSService) CreateRecord(ctx context.Context, zoneID uuid.UUID, name string, recordType domain.RecordType, content string, ttl int, priority *int) (*domain.DNSRecord, error) {
	args := m.Called(ctx, zoneID, name, recordType, content, ttl, priority)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DNSRecord), args.Error(1)
}

func (m *DNSService) GetRecord(ctx context.Context, id uuid.UUID) (*domain.DNSRecord, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DNSRecord), args.Error(1)
}

func (m *DNSService) ListRecords(ctx context.Context, zoneID uuid.UUID) ([]*domain.DNSRecord, error) {
	args := m.Called(ctx, zoneID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.DNSRecord), args.Error(1)
}

func (m *DNSService) UpdateRecord(ctx context.Context, id uuid.UUID, content string, ttl int, priority *int) (*domain.DNSRecord, error) {
	args := m.Called(ctx, id, content, ttl, priority)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DNSRecord), args.Error(1)
}

func (m *DNSService) DeleteRecord(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *DNSService) RegisterInstance(ctx context.Context, instance *domain.Instance, ipAddress string) error {
	args := m.Called(ctx, instance, ipAddress)
	return args.Error(0)
}

func (m *DNSService) UnregisterInstance(ctx context.Context, instanceID uuid.UUID) error {
	args := m.Called(ctx, instanceID)
	return args.Error(0)
}
