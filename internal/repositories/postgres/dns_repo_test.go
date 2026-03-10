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
	"github.com/stretchr/testify/require"
)

func TestDNSRepositoryCreateZone(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
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
	require.NoError(t, err)
}

func TestDNSRepositoryGetZoneByID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewDNSRepository(mock)
	id := uuid.New()

	rows := pgxmock.NewRows([]string{"id", "user_id", "tenant_id", "vpc_id", "name", "description", "status", "default_ttl", "powerdns_id", "created_at", "updated_at"}).
		AddRow(id, uuid.New(), uuid.New(), uuid.New(), "example.com", "desc", "ACTIVE", 300, "pdns-1", time.Now(), time.Now())

	mock.ExpectQuery("SELECT (.+) FROM dns_zones WHERE id = \\$1").
		WithArgs(id).
		WillReturnRows(rows)

	zone, err := repo.GetZoneByID(context.Background(), id)
	require.NoError(t, err)
	assert.NotNil(t, zone)
	assert.Equal(t, id, zone.ID)
}

func TestDNSRepositoryListZones(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
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
	require.NoError(t, err)
	assert.Len(t, zones, 1)
}

func TestDNSRepositoryCreateRecord(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
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
	require.NoError(t, err)
}

func TestDNSRepositoryDeleteRecord(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewDNSRepository(mock)
	id := uuid.New()

	mock.ExpectExec("DELETE FROM dns_records WHERE id = \\$1").
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = repo.DeleteRecord(context.Background(), id)
	require.NoError(t, err)
}

func TestDNSRepositoryGetZoneByName(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewDNSRepository(mock)
	tenantID := uuid.New()
	ctx := appcontext.WithTenantID(context.Background(), tenantID)
	name := "example.com"

	mock.ExpectQuery("SELECT (.+) FROM dns_zones WHERE name = \\$1 AND tenant_id = \\$2").
		WithArgs(name, tenantID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "tenant_id", "vpc_id", "name", "description", "status", "default_ttl", "powerdns_id", "created_at", "updated_at"}).
			AddRow(uuid.New(), uuid.New(), tenantID, uuid.New(), name, "desc", "ACTIVE", 300, "pdns-1", time.Now(), time.Now()))

	zone, err := repo.GetZoneByName(ctx, name)
	require.NoError(t, err)
	assert.NotNil(t, zone)
	assert.Equal(t, name, zone.Name)
}

func TestDNSRepositoryGetZoneByVPC(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewDNSRepository(mock)
	vpcID := uuid.New()

	mock.ExpectQuery("SELECT (.+) FROM dns_zones WHERE vpc_id = \\$1").
		WithArgs(vpcID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "tenant_id", "vpc_id", "name", "description", "status", "default_ttl", "powerdns_id", "created_at", "updated_at"}).
			AddRow(uuid.New(), uuid.New(), uuid.New(), vpcID, "vpc.com", "desc", "ACTIVE", 300, "pdns-1", time.Now(), time.Now()))

	zone, err := repo.GetZoneByVPC(context.Background(), vpcID)
	require.NoError(t, err)
	assert.NotNil(t, zone)
	assert.Equal(t, vpcID, zone.VpcID)
}

func TestDNSRepositoryUpdateZone(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewDNSRepository(mock)
	zone := &domain.DNSZone{
		ID:          uuid.New(),
		Description: "updated",
		Status:      domain.ZoneStatusActive,
		DefaultTTL:  600,
	}

	mock.ExpectExec("UPDATE dns_zones").
		WithArgs(zone.ID, zone.Description, zone.Status, zone.DefaultTTL, pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err = repo.UpdateZone(context.Background(), zone)
	require.NoError(t, err)
}

func TestDNSRepositoryDeleteZone(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewDNSRepository(mock)
	id := uuid.New()

	mock.ExpectExec("DELETE FROM dns_zones WHERE id = \\$1").
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = repo.DeleteZone(context.Background(), id)
	require.NoError(t, err)
}

func TestDNSRepositoryGetRecordByID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewDNSRepository(mock)
	id := uuid.New()

	mock.ExpectQuery("SELECT (.+) FROM dns_records WHERE id = \\$1").
		WithArgs(id).
		WillReturnRows(pgxmock.NewRows([]string{"id", "zone_id", "name", "type", "content", "ttl", "priority", "disabled", "auto_managed", "instance_id", "created_at", "updated_at"}).
			AddRow(id, uuid.New(), "www", domain.RecordTypeA, "1.2.3.4", 300, nil, false, false, nil, time.Now(), time.Now()))

	rec, err := repo.GetRecordByID(context.Background(), id)
	require.NoError(t, err)
	assert.NotNil(t, rec)
	assert.Equal(t, id, rec.ID)
}

func TestDNSRepositoryListRecordsByZone(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewDNSRepository(mock)
	zoneID := uuid.New()

	mock.ExpectQuery("SELECT (.+) FROM dns_records WHERE zone_id = \\$1").
		WithArgs(zoneID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "zone_id", "name", "type", "content", "ttl", "priority", "disabled", "auto_managed", "instance_id", "created_at", "updated_at"}).
			AddRow(uuid.New(), zoneID, "a", domain.RecordTypeA, "1.1.1.1", 300, nil, false, false, nil, time.Now(), time.Now()).
			AddRow(uuid.New(), zoneID, "b", domain.RecordTypeAAAA, "::1", 300, nil, false, false, nil, time.Now(), time.Now()))

	recs, err := repo.ListRecordsByZone(context.Background(), zoneID)
	require.NoError(t, err)
	assert.Len(t, recs, 2)
}

func TestDNSRepositoryGetRecordsByInstance(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewDNSRepository(mock)
	instanceID := uuid.New()

	mock.ExpectQuery("SELECT (.+) FROM dns_records WHERE instance_id = \\$1").
		WithArgs(instanceID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "zone_id", "name", "type", "content", "ttl", "priority", "disabled", "auto_managed", "instance_id", "created_at", "updated_at"}).
			AddRow(uuid.New(), uuid.New(), "inst", domain.RecordTypeA, "10.0.0.1", 300, nil, false, true, &instanceID, time.Now(), time.Now()))

	recs, err := repo.GetRecordsByInstance(context.Background(), instanceID)
	require.NoError(t, err)
	assert.Len(t, recs, 1)
}

func TestDNSRepositoryUpdateRecord(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewDNSRepository(mock)
	record := &domain.DNSRecord{
		ID:      uuid.New(),
		Content: "updated",
		TTL:     1200,
	}

	mock.ExpectExec("UPDATE dns_records").
		WithArgs(record.ID, record.Content, record.TTL, record.Priority, record.Disabled, pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err = repo.UpdateRecord(context.Background(), record)
	require.NoError(t, err)
}

func TestDNSRepositoryDeleteRecordsByInstance(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewDNSRepository(mock)
	instanceID := uuid.New()

	mock.ExpectExec("DELETE FROM dns_records WHERE instance_id = \\$1").
		WithArgs(instanceID).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = repo.DeleteRecordsByInstance(context.Background(), instanceID)
	require.NoError(t, err)
}
