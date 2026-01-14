// Package postgres provides PostgreSQL-backed repository implementations.
package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/errors"
)

// StorageRepository provides PostgreSQL-backed object storage metadata persistence.
type StorageRepository struct {
	db DB
}

// NewStorageRepository creates a StorageRepository using the provided DB.
func NewStorageRepository(db DB) *StorageRepository {
	return &StorageRepository{db: db}
}

func (r *StorageRepository) SaveMeta(ctx context.Context, obj *domain.Object) error {
	query := `
		INSERT INTO objects (id, user_id, arn, bucket, key, size_bytes, content_type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (bucket, key) DO UPDATE SET
			size_bytes = EXCLUDED.size_bytes,
			content_type = EXCLUDED.content_type,
			created_at = EXCLUDED.created_at,
			deleted_at = NULL,
			user_id = EXCLUDED.user_id
	`
	_, err := r.db.Exec(ctx, query,
		obj.ID, obj.UserID, obj.ARN, obj.Bucket, obj.Key, obj.SizeBytes, obj.ContentType, obj.CreatedAt,
	)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to save object metadata", err)
	}
	return nil
}

func (r *StorageRepository) GetMeta(ctx context.Context, bucket, key string) (*domain.Object, error) {
	userID := appcontext.UserIDFromContext(ctx)
	query := `
		SELECT id, user_id, arn, bucket, key, size_bytes, content_type, created_at
		FROM objects
		WHERE bucket = $1 AND key = $2 AND deleted_at IS NULL AND user_id = $3
	`
	return r.scanObject(r.db.QueryRow(ctx, query, bucket, key, userID))
}

func (r *StorageRepository) List(ctx context.Context, bucket string) ([]*domain.Object, error) {
	userID := appcontext.UserIDFromContext(ctx)
	query := `
		SELECT id, user_id, arn, bucket, key, size_bytes, content_type, created_at
		FROM objects
		WHERE bucket = $1 AND deleted_at IS NULL AND user_id = $2
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, bucket, userID)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to list objects", err)
	}
	return r.scanObjects(rows)
}

func (r *StorageRepository) scanObject(row pgx.Row) (*domain.Object, error) {
	var obj domain.Object
	err := row.Scan(
		&obj.ID, &obj.UserID, &obj.ARN, &obj.Bucket, &obj.Key, &obj.SizeBytes, &obj.ContentType, &obj.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New(errors.ObjectNotFound, "object metadata not found")
		}
		return nil, errors.Wrap(errors.Internal, "failed to scan object metadata", err)
	}
	return &obj, nil
}

func (r *StorageRepository) scanObjects(rows pgx.Rows) ([]*domain.Object, error) {
	defer rows.Close()
	var objects []*domain.Object
	for rows.Next() {
		obj, err := r.scanObject(rows)
		if err != nil {
			return nil, err
		}
		objects = append(objects, obj)
	}
	return objects, nil
}

func (r *StorageRepository) SoftDelete(ctx context.Context, bucket, key string) error {
	userID := appcontext.UserIDFromContext(ctx)
	query := `
		UPDATE objects
		SET deleted_at = $1
		WHERE bucket = $2 AND key = $3 AND deleted_at IS NULL AND user_id = $4
	`
	cmd, err := r.db.Exec(ctx, query, time.Now(), bucket, key, userID)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to soft delete object", err)
	}
	if cmd.RowsAffected() == 0 {
		return errors.New(errors.ObjectNotFound, "object not found or already deleted")
	}
	return nil
}
