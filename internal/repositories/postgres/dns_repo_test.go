package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v3"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

func TestDNSRepositoryCreateZone(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewDNSRepository(mock)
	zone := &domain.DNSZone{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TenantID:  uuid.New(),
		VpcID:     uuid.New(),
		Name:      "example.com",
		Status:    domain.ZoneStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mock.ExpectExec("INSERT INTO dns_zones").
		WithArgs(zone.ID, zone.UserID, zone.TenantID, zone.VpcID, zone.Name, zone.Description, zone.Status, zone.DefaultTTL, zone.PowerDNSID, zone.CreatedAt, zone.UpdatedAt).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.CreateZone(context.Background(), zone)
	assert.NoError(t, err)
}

func TestDNSRepositoryGetZoneByID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewDNSRepository(mock)
	id := uuid.New()

	rows := pgxmock.NewRows([]string{"id", "user_id", "tenant_id", "vpc_id", "name", "description", "status", "default_ttl", "powerdns_id", "created_at", "updated_at"}).
		AddRow(id, uuid.New(), uuid.New(), uuid.New(), "example.com", "desc", "ACTIVE", 300, "pdns-1", time.Now(), time.Now())

	mock.ExpectQuery("SELECT (.+) FROM dns_zones WHERE id = \\$1").
		WithArgs(id).
		WillReturnRows(rows)

	zone, err := repo.GetZoneByID(context.Background(), id)
	assert.NoError(t, err)
	assert.NotNil(t, zone)
	assert.Equal(t, id, zone.ID)
}

func TestDNSRepositoryListZones(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewDNSRepository(mock)
	tenantID := uuid.New()
	ctx := appcontext.WithTenantID(context.Background(), tenantID)

	rows := pgxmock.NewRows([]string{"id", "user_id", "tenant_id", "vpc_id", "name", "description", "status", "default_ttl", "powerdns_id", "created_at", "updated_at"}).
		AddRow(uuid.New(), uuid.New(), tenantID, uuid.New(), "zone1.com", "desc", "ACTIVE", 300, "pdns-1", time.Now(), time.Now())

	mock.ExpectQuery("SELECT (.+) FROM dns_zones WHERE tenant_id = \\$1").
		WithArgs(tenantID).
		WillReturnRows(rows)

	zones, err := repo.ListZones(ctx)
	assert.NoError(t, err)
	assert.Len(t, zones, 1)
}

func TestDNSRepositoryCreateRecord(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewDNSRepository(mock)
	record := &domain.DNSRecord{
		ID:        uuid.New(),
		ZoneID:    uuid.New(),
		Name:      "www",
		Type:      domain.RecordTypeA,
		Content:   "1.1.1.1",
		TTL:       3600,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mock.ExpectExec("INSERT INTO dns_records").
		WithArgs(record.ID, record.ZoneID, record.Name, record.Type, record.Content, record.TTL, record.Priority, record.Disabled, record.AutoManaged, record.InstanceID, record.CreatedAt, record.UpdatedAt).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.CreateRecord(context.Background(), record)
	assert.NoError(t, err)
}

func TestDNSRepositoryDeleteRecord(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewDNSRepository(mock)
	id := uuid.New()

	mock.ExpectExec("DELETE FROM dns_records WHERE id = \\$1").
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = repo.DeleteRecord(context.Background(), id)
	assert.NoError(t, err)
}
