// Package postgres provides PostgreSQL-backed repository implementations.
package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/errors"
)

// CacheRepository provides PostgreSQL-backed cache persistence.
type CacheRepository struct {
	db DB
}

// NewCacheRepository creates a CacheRepository using the provided DB.
func NewCacheRepository(db DB) *CacheRepository {
	return &CacheRepository{db: db}
}

func (r *CacheRepository) Create(ctx context.Context, cache *domain.Cache) error {
	query := `
		INSERT INTO caches (
			id, user_id, name, engine, version, status, vpc_id, 
			container_id, port, password, memory_mb, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.db.Exec(ctx, query,
		cache.ID, cache.UserID, cache.Name, cache.Engine, cache.Version, cache.Status, cache.VpcID,
		cache.ContainerID, cache.Port, cache.Password, cache.MemoryMB, cache.CreatedAt, cache.UpdatedAt,
	)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to create cache", err)
	}
	return nil
}

func (r *CacheRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Cache, error) {
	query := `
		SELECT 
			id, user_id, name, engine, version, status, vpc_id,
			container_id, port, password, memory_mb, created_at, updated_at
		FROM caches
		WHERE id = $1
	`
	return r.scanCache(r.db.QueryRow(ctx, query, id))
}

func (r *CacheRepository) GetByName(ctx context.Context, userID uuid.UUID, name string) (*domain.Cache, error) {
	query := `
		SELECT 
			id, user_id, name, engine, version, status, vpc_id,
			container_id, port, password, memory_mb, created_at, updated_at
		FROM caches
		WHERE user_id = $1 AND name = $2
	`
	return r.scanCache(r.db.QueryRow(ctx, query, userID, name))
}

func (r *CacheRepository) List(ctx context.Context, userID uuid.UUID) ([]*domain.Cache, error) {
	query := `
		SELECT 
			id, user_id, name, engine, version, status, vpc_id,
			container_id, port, password, memory_mb, created_at, updated_at
		FROM caches
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to list caches", err)
	}
	return r.scanCaches(rows)
}

func (r *CacheRepository) scanCache(row pgx.Row) (*domain.Cache, error) {
	var cache domain.Cache
	var engine, status string
	err := row.Scan(
		&cache.ID, &cache.UserID, &cache.Name, &engine, &cache.Version, &status, &cache.VpcID,
		&cache.ContainerID, &cache.Port, &cache.Password, &cache.MemoryMB, &cache.CreatedAt, &cache.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New(errors.NotFound, "cache not found")
		}
		return nil, errors.Wrap(errors.Internal, "failed to scan cache", err)
	}
	cache.Engine = domain.CacheEngine(engine)
	cache.Status = domain.CacheStatus(status)
	return &cache, nil
}

func (r *CacheRepository) scanCaches(rows pgx.Rows) ([]*domain.Cache, error) {
	defer rows.Close()
	var caches []*domain.Cache
	for rows.Next() {
		cache, err := r.scanCache(rows)
		if err != nil {
			return nil, err
		}
		caches = append(caches, cache)
	}
	return caches, nil
}

func (r *CacheRepository) Update(ctx context.Context, cache *domain.Cache) error {
	query := `
		UPDATE caches SET
			status = $1,
			container_id = $2,
			port = $3,
			updated_at = $4
		WHERE id = $5
	`
	_, err := r.db.Exec(ctx, query,
		cache.Status, cache.ContainerID, cache.Port, cache.UpdatedAt, cache.ID,
	)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to update cache", err)
	}
	return nil
}

func (r *CacheRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM caches WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to delete cache", err)
	}
	return nil
}
