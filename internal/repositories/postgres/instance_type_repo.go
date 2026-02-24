// Package postgres provides PostgreSQL-backed repository implementations.
package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	stdlib_errors "errors"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/errors"
)

// InstanceTypeRepository provides a PostgreSQL implementation for instance types.
type InstanceTypeRepository struct {
	db DB
}

// NewInstanceTypeRepository creates a new InstanceTypeRepository.
func NewInstanceTypeRepository(db DB) *InstanceTypeRepository {
	return &InstanceTypeRepository{db: db}
}

// List returns all available instance types.
func (r *InstanceTypeRepository) List(ctx context.Context) ([]*domain.InstanceType, error) {
	query := `
		SELECT id, name, vcpus, memory_mb, disk_gb, network_mbps, price_per_hour, category
		FROM instance_types
		ORDER BY vcpus ASC, memory_mb ASC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to list instance types", err)
	}
	defer rows.Close()

	var types []*domain.InstanceType
	for rows.Next() {
		t, err := r.scanInstanceType(rows)
		if err != nil {
			return nil, err
		}
		types = append(types, t)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(errors.Internal, "iteration error in list instance types", err)
	}
	return types, nil
}

// GetByID retrieves an instance type by its ID.
func (r *InstanceTypeRepository) GetByID(ctx context.Context, id string) (*domain.InstanceType, error) {
	query := `
		SELECT id, name, vcpus, memory_mb, disk_gb, network_mbps, price_per_hour, category
		FROM instance_types
		WHERE id = $1
	`
	return r.scanInstanceType(r.db.QueryRow(ctx, query, id))
}

// Create persists a new instance type.
func (r *InstanceTypeRepository) Create(ctx context.Context, it *domain.InstanceType) (*domain.InstanceType, error) {
	query := `
		INSERT INTO instance_types (id, name, vcpus, memory_mb, disk_gb, network_mbps, price_per_hour, category)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, name, vcpus, memory_mb, disk_gb, network_mbps, price_per_hour, category
	`
	row := r.db.QueryRow(ctx, query,
		it.ID, it.Name, it.VCPUs, it.MemoryMB, it.DiskGB, it.NetworkMbps, it.PricePerHr, it.Category,
	)
	return r.scanInstanceType(row)
}

// Update modifies an existing instance type.
func (r *InstanceTypeRepository) Update(ctx context.Context, it *domain.InstanceType) (*domain.InstanceType, error) {
	query := `
		UPDATE instance_types
		SET name = $2, vcpus = $3, memory_mb = $4, disk_gb = $5, network_mbps = $6, price_per_hour = $7, category = $8
		WHERE id = $1
		RETURNING id, name, vcpus, memory_mb, disk_gb, network_mbps, price_per_hour, category
	`
	row := r.db.QueryRow(ctx, query,
		it.ID, it.Name, it.VCPUs, it.MemoryMB, it.DiskGB, it.NetworkMbps, it.PricePerHr, it.Category,
	)
	return r.scanInstanceType(row)
}

// Delete removes an instance type by its ID.
func (r *InstanceTypeRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM instance_types WHERE id = $1`
	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to delete instance type", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New(errors.NotFound, "instance type not found")
	}
	return nil
}

func (r *InstanceTypeRepository) scanInstanceType(row pgx.Row) (*domain.InstanceType, error) {
	var t domain.InstanceType
	err := row.Scan(
		&t.ID, &t.Name, &t.VCPUs, &t.MemoryMB, &t.DiskGB, &t.NetworkMbps, &t.PricePerHr, &t.Category,
	)
	if err != nil {
		if stdlib_errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(errors.NotFound, "instance type not found")
		}
		return nil, errors.Wrap(errors.Internal, "failed to scan instance type", err)
	}
	return &t, nil
}
