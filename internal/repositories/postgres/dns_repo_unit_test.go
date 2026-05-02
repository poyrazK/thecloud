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

func TestDNSRepository_Zones(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewDNSRepository(mock)
	ctx := context.Background()
	tenantID := uuid.New()
	ctx = appcontext.WithTenantID(ctx, tenantID)

	zone := &domain.DNSZone{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		TenantID:    tenantID,
		Name:        "example.com.",
		Description: "test zone",
		Status:      domain.ZoneStatusActive,
		DefaultTTL:  3600,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	t.Run("CreateZone", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO dns_zones").
			WithArgs(zone.ID, zone.UserID, zone.TenantID, zone.VpcID, zone.Name, zone.Description, zone.Status, zone.DefaultTTL, zone.PowerDNSID, zone.CreatedAt, zone.UpdatedAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err := repo.CreateZone(ctx, zone)
		require.NoError(t, err)
	})

	t.Run("GetZoneByID", func(t *testing.T) {
		mock.ExpectQuery("SELECT .* FROM dns_zones WHERE id = \\$1").
			WithArgs(zone.ID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "tenant_id", "vpc_id", "name", "description", "status", "default_ttl", "powerdns_id", "created_at", "updated_at"}).
				AddRow(zone.ID, zone.UserID, zone.TenantID, zone.VpcID, zone.Name, zone.Description, zone.Status, zone.DefaultTTL, zone.PowerDNSID, zone.CreatedAt, zone.UpdatedAt))

		res, err := repo.GetZoneByID(ctx, zone.ID)
		require.NoError(t, err)
		assert.Equal(t, zone.ID, res.ID)
	})

	t.Run("GetZoneByName", func(t *testing.T) {
		mock.ExpectQuery("SELECT .* FROM dns_zones WHERE name = \\$1 AND tenant_id = \\$2").
			WithArgs(zone.Name, tenantID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "tenant_id", "vpc_id", "name", "description", "status", "default_ttl", "powerdns_id", "created_at", "updated_at"}).
				AddRow(zone.ID, zone.UserID, zone.TenantID, zone.VpcID, zone.Name, zone.Description, zone.Status, zone.DefaultTTL, zone.PowerDNSID, zone.CreatedAt, zone.UpdatedAt))

		res, err := repo.GetZoneByName(ctx, zone.Name)
		require.NoError(t, err)
		assert.Equal(t, zone.Name, res.Name)
	})

	t.Run("GetZoneByVPC", func(t *testing.T) {
		vpcID := uuid.New()
		mock.ExpectQuery("SELECT .* FROM dns_zones WHERE vpc_id = \\$1").
			WithArgs(vpcID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "tenant_id", "vpc_id", "name", "description", "status", "default_ttl", "powerdns_id", "created_at", "updated_at"}).
				AddRow(zone.ID, zone.UserID, zone.TenantID, vpcID, zone.Name, zone.Description, zone.Status, zone.DefaultTTL, zone.PowerDNSID, zone.CreatedAt, zone.UpdatedAt))

		res, err := repo.GetZoneByVPC(ctx, vpcID)
		require.NoError(t, err)
		assert.Equal(t, vpcID, res.VpcID)
	})

	t.Run("ListZones", func(t *testing.T) {
		mock.ExpectQuery("SELECT .* FROM dns_zones WHERE tenant_id = \\$1").
			WithArgs(tenantID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "tenant_id", "vpc_id", "name", "description", "status", "default_ttl", "powerdns_id", "created_at", "updated_at"}).
				AddRow(zone.ID, zone.UserID, zone.TenantID, zone.VpcID, zone.Name, zone.Description, zone.Status, zone.DefaultTTL, zone.PowerDNSID, zone.CreatedAt, zone.UpdatedAt))

		res, err := repo.ListZones(ctx)
		require.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("UpdateZone", func(t *testing.T) {
		mock.ExpectExec("UPDATE dns_zones").
			WithArgs(zone.ID, zone.Description, zone.Status, zone.DefaultTTL, pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := repo.UpdateZone(ctx, zone)
		require.NoError(t, err)
	})

	t.Run("DeleteZone", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM dns_zones WHERE id = \\$1").
			WithArgs(zone.ID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err := repo.DeleteZone(ctx, zone.ID)
		require.NoError(t, err)
	})
}

func TestDNSRepository_Records(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewDNSRepository(mock)
	ctx := context.Background()
	zoneID := uuid.New()

	record := &domain.DNSRecord{
		ID:        uuid.New(),
		ZoneID:    zoneID,
		Name:      "www.example.com.",
		Type:      domain.RecordTypeA,
		Content:   "1.2.3.4",
		TTL:       3600,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Run("CreateRecord", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO dns_records").
			WithArgs(record.ID, record.ZoneID, record.Name, record.Type, record.Content, record.TTL, record.Priority, record.Disabled, record.AutoManaged, record.InstanceID, record.CreatedAt, record.UpdatedAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err := repo.CreateRecord(ctx, record)
		require.NoError(t, err)
	})

	t.Run("GetRecordByID", func(t *testing.T) {
		mock.ExpectQuery("SELECT .* FROM dns_records WHERE id = \\$1").
			WithArgs(record.ID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "zone_id", "name", "type", "content", "ttl", "priority", "disabled", "auto_managed", "instance_id", "created_at", "updated_at"}).
				AddRow(record.ID, record.ZoneID, record.Name, record.Type, record.Content, record.TTL, record.Priority, record.Disabled, record.AutoManaged, record.InstanceID, record.CreatedAt, record.UpdatedAt))

		res, err := repo.GetRecordByID(ctx, record.ID)
		require.NoError(t, err)
		assert.Equal(t, record.ID, res.ID)
	})

	t.Run("ListRecordsByZone", func(t *testing.T) {
		mock.ExpectQuery("SELECT .* FROM dns_records WHERE zone_id = \\$1").
			WithArgs(zoneID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "zone_id", "name", "type", "content", "ttl", "priority", "disabled", "auto_managed", "instance_id", "created_at", "updated_at"}).
				AddRow(record.ID, record.ZoneID, record.Name, record.Type, record.Content, record.TTL, record.Priority, record.Disabled, record.AutoManaged, record.InstanceID, record.CreatedAt, record.UpdatedAt))

		res, err := repo.ListRecordsByZone(ctx, zoneID)
		require.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("GetRecordsByInstance", func(t *testing.T) {
		instanceID := uuid.New()
		mock.ExpectQuery("SELECT .* FROM dns_records WHERE instance_id = \\$1").
			WithArgs(instanceID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "zone_id", "name", "type", "content", "ttl", "priority", "disabled", "auto_managed", "instance_id", "created_at", "updated_at"}).
				AddRow(record.ID, record.ZoneID, record.Name, record.Type, record.Content, record.TTL, record.Priority, record.Disabled, record.AutoManaged, &instanceID, record.CreatedAt, record.UpdatedAt))

		res, err := repo.GetRecordsByInstance(ctx, instanceID)
		require.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("UpdateRecord", func(t *testing.T) {
		mock.ExpectExec("UPDATE dns_records").
			WithArgs(record.ID, record.Content, record.TTL, record.Priority, record.Disabled, pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := repo.UpdateRecord(ctx, record)
		require.NoError(t, err)
	})

	t.Run("DeleteRecord", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM dns_records WHERE id = \\$1").
			WithArgs(record.ID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err := repo.DeleteRecord(ctx, record.ID)
		require.NoError(t, err)
	})

	t.Run("DeleteRecordsByInstance", func(t *testing.T) {
		instanceID := uuid.New()
		mock.ExpectExec("DELETE FROM dns_records WHERE instance_id = \\$1").
			WithArgs(instanceID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err := repo.DeleteRecordsByInstance(ctx, instanceID)
		require.NoError(t, err)
	})
}
