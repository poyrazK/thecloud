package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/errors"
)

// DNSRepository implements ports.DNSRepository using PostgreSQL.
type DNSRepository struct {
	db DB
}

// NewDNSRepository creates a new PostgreSQL DNS repository.
func NewDNSRepository(db DB) *DNSRepository {
	return &DNSRepository{db: db}
}

// --- Zone Operations ---

func (r *DNSRepository) CreateZone(ctx context.Context, zone *domain.DNSZone) error {
	query := `
		INSERT INTO dns_zones (
			id, user_id, tenant_id, vpc_id, name, description, 
			status, default_ttl, powerdns_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Exec(ctx, query,
		zone.ID, zone.UserID, zone.TenantID, zone.VpcID, zone.Name, zone.Description,
		zone.Status, zone.DefaultTTL, zone.PowerDNSID, zone.CreatedAt, zone.UpdatedAt,
	)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to create dns zone", err)
	}
	return nil
}

func (r *DNSRepository) GetZoneByID(ctx context.Context, id uuid.UUID) (*domain.DNSZone, error) {
	query := `
		SELECT id, user_id, tenant_id, vpc_id, name, description, 
		       status, default_ttl, powerdns_id, created_at, updated_at
		FROM dns_zones
		WHERE id = $1
	`
	// Typically we might also filter by tenant_id for security, but the service later should handle authz check or context-based filtering?
	// Existing repos often just get by ID. However, let's verify if we need to respect tenant scoping here.
	// For GetByID, usually we fetch the raw object and service checks permission.
	// But let's check if the pattern is to include tenant check.
	// Checking the implementation plan, it just says "Standard PostgreSQL repository".

	var z domain.DNSZone
	err := r.db.QueryRow(ctx, query, id).Scan(
		&z.ID, &z.UserID, &z.TenantID, &z.VpcID, &z.Name, &z.Description,
		&z.Status, &z.DefaultTTL, &z.PowerDNSID, &z.CreatedAt, &z.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New(errors.NotFound, "dns zone not found")
		}
		return nil, errors.Wrap(errors.Internal, "failed to get dns zone", err)
	}
	return &z, nil
}

func (r *DNSRepository) GetZoneByName(ctx context.Context, name string) (*domain.DNSZone, error) {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `
		SELECT id, user_id, tenant_id, vpc_id, name, description, 
		       status, default_ttl, powerdns_id, created_at, updated_at
		FROM dns_zones
		WHERE name = $1 AND tenant_id = $2
	`
	var z domain.DNSZone
	err := r.db.QueryRow(ctx, query, name, tenantID).Scan(
		&z.ID, &z.UserID, &z.TenantID, &z.VpcID, &z.Name, &z.Description,
		&z.Status, &z.DefaultTTL, &z.PowerDNSID, &z.CreatedAt, &z.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New(errors.NotFound, "dns zone not found")
		}
		return nil, errors.Wrap(errors.Internal, "failed to get dns zone by name", err)
	}
	return &z, nil
}

func (r *DNSRepository) GetZoneByVPC(ctx context.Context, vpcID uuid.UUID) (*domain.DNSZone, error) {
	query := `
		SELECT id, user_id, tenant_id, vpc_id, name, description, 
		       status, default_ttl, powerdns_id, created_at, updated_at
		FROM dns_zones
		WHERE vpc_id = $1
	`
	var z domain.DNSZone
	err := r.db.QueryRow(ctx, query, vpcID).Scan(
		&z.ID, &z.UserID, &z.TenantID, &z.VpcID, &z.Name, &z.Description,
		&z.Status, &z.DefaultTTL, &z.PowerDNSID, &z.CreatedAt, &z.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New(errors.NotFound, "dns zone for vpc not found")
		}
		return nil, errors.Wrap(errors.Internal, "failed to get dns zone by vpc", err)
	}
	return &z, nil
}

func (r *DNSRepository) ListZones(ctx context.Context) ([]*domain.DNSZone, error) {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `
		SELECT id, user_id, tenant_id, vpc_id, name, description, 
		       status, default_ttl, powerdns_id, created_at, updated_at
		FROM dns_zones
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to list dns zones", err)
	}
	defer rows.Close()

	var zones []*domain.DNSZone
	for rows.Next() {
		var z domain.DNSZone
		if err := rows.Scan(
			&z.ID, &z.UserID, &z.TenantID, &z.VpcID, &z.Name, &z.Description,
			&z.Status, &z.DefaultTTL, &z.PowerDNSID, &z.CreatedAt, &z.UpdatedAt,
		); err != nil {
			return nil, errors.Wrap(errors.Internal, "failed to scan dns zone", err)
		}
		zones = append(zones, &z)
	}
	return zones, nil
}

func (r *DNSRepository) UpdateZone(ctx context.Context, zone *domain.DNSZone) error {
	zone.UpdatedAt = time.Now()
	query := `
		UPDATE dns_zones
		SET description = $2, status = $3, default_ttl = $4, updated_at = $5
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query,
		zone.ID, zone.Description, zone.Status, zone.DefaultTTL, zone.UpdatedAt,
	)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to update dns zone", err)
	}
	return nil
}

func (r *DNSRepository) DeleteZone(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM dns_zones WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to delete dns zone", err)
	}
	return nil
}

// --- Record Operations ---

func (r *DNSRepository) CreateRecord(ctx context.Context, record *domain.DNSRecord) error {
	query := `
		INSERT INTO dns_records (
			id, zone_id, name, type, content, ttl, priority, 
			disabled, auto_managed, instance_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.Exec(ctx, query,
		record.ID, record.ZoneID, record.Name, record.Type, record.Content, record.TTL, record.Priority,
		record.Disabled, record.AutoManaged, record.InstanceID, record.CreatedAt, record.UpdatedAt,
	)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to create dns record", err)
	}
	return nil
}

func (r *DNSRepository) GetRecordByID(ctx context.Context, id uuid.UUID) (*domain.DNSRecord, error) {
	query := `
		SELECT id, zone_id, name, type, content, ttl, priority, 
		       disabled, auto_managed, instance_id, created_at, updated_at
		FROM dns_records
		WHERE id = $1
	`
	var rec domain.DNSRecord
	err := r.db.QueryRow(ctx, query, id).Scan(
		&rec.ID, &rec.ZoneID, &rec.Name, &rec.Type, &rec.Content, &rec.TTL, &rec.Priority,
		&rec.Disabled, &rec.AutoManaged, &rec.InstanceID, &rec.CreatedAt, &rec.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New(errors.NotFound, "dns record not found")
		}
		return nil, errors.Wrap(errors.Internal, "failed to get dns record", err)
	}
	return &rec, nil
}

func (r *DNSRepository) ListRecordsByZone(ctx context.Context, zoneID uuid.UUID) ([]*domain.DNSRecord, error) {
	query := `
		SELECT id, zone_id, name, type, content, ttl, priority, 
		       disabled, auto_managed, instance_id, created_at, updated_at
		FROM dns_records
		WHERE zone_id = $1
		ORDER BY name ASC, type ASC
	`
	rows, err := r.db.Query(ctx, query, zoneID)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to list dns records", err)
	}
	defer rows.Close()

	var records []*domain.DNSRecord
	for rows.Next() {
		var rec domain.DNSRecord
		if err := rows.Scan(
			&rec.ID, &rec.ZoneID, &rec.Name, &rec.Type, &rec.Content, &rec.TTL, &rec.Priority,
			&rec.Disabled, &rec.AutoManaged, &rec.InstanceID, &rec.CreatedAt, &rec.UpdatedAt,
		); err != nil {
			return nil, errors.Wrap(errors.Internal, "failed to scan dns record", err)
		}
		records = append(records, &rec)
	}
	return records, nil
}

func (r *DNSRepository) GetRecordsByInstance(ctx context.Context, instanceID uuid.UUID) ([]*domain.DNSRecord, error) {
	query := `
		SELECT id, zone_id, name, type, content, ttl, priority, 
		       disabled, auto_managed, instance_id, created_at, updated_at
		FROM dns_records
		WHERE instance_id = $1
	`
	rows, err := r.db.Query(ctx, query, instanceID)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to list dns records by instance", err)
	}
	defer rows.Close()

	var records []*domain.DNSRecord
	for rows.Next() {
		var rec domain.DNSRecord
		if err := rows.Scan(
			&rec.ID, &rec.ZoneID, &rec.Name, &rec.Type, &rec.Content, &rec.TTL, &rec.Priority,
			&rec.Disabled, &rec.AutoManaged, &rec.InstanceID, &rec.CreatedAt, &rec.UpdatedAt,
		); err != nil {
			return nil, errors.Wrap(errors.Internal, "failed to scan dns record", err)
		}
		records = append(records, &rec)
	}
	return records, nil
}

func (r *DNSRepository) UpdateRecord(ctx context.Context, record *domain.DNSRecord) error {
	record.UpdatedAt = time.Now()
	query := `
		UPDATE dns_records
		SET content = $2, ttl = $3, priority = $4, disabled = $5, updated_at = $6
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query,
		record.ID, record.Content, record.TTL, record.Priority, record.Disabled, record.UpdatedAt,
	)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to update dns record", err)
	}
	return nil
}

func (r *DNSRepository) DeleteRecord(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM dns_records WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to delete dns record", err)
	}
	return nil
}

func (r *DNSRepository) DeleteRecordsByInstance(ctx context.Context, instanceID uuid.UUID) error {
	query := `DELETE FROM dns_records WHERE instance_id = $1`
	_, err := r.db.Exec(ctx, query, instanceID)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to delete dns records by instance", err)
	}
	return nil
}
