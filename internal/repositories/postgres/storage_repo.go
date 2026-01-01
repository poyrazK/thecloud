package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/poyraz/cloud/internal/core/domain"
	"github.com/poyraz/cloud/internal/errors"
)

type StorageRepository struct {
	db *pgxpool.Pool
}

func NewStorageRepository(db *pgxpool.Pool) *StorageRepository {
	return &StorageRepository{db: db}
}

func (r *StorageRepository) SaveMeta(ctx context.Context, obj *domain.Object) error {
	query := `
		INSERT INTO objects (id, arn, bucket, key, size_bytes, content_type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (bucket, key) DO UPDATE SET
			size_bytes = EXCLUDED.size_bytes,
			content_type = EXCLUDED.content_type,
			created_at = EXCLUDED.created_at,
			deleted_at = NULL
	`
	_, err := r.db.Exec(ctx, query,
		obj.ID, obj.ARN, obj.Bucket, obj.Key, obj.SizeBytes, obj.ContentType, obj.CreatedAt,
	)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to save object metadata", err)
	}
	return nil
}

func (r *StorageRepository) GetMeta(ctx context.Context, bucket, key string) (*domain.Object, error) {
	query := `
		SELECT id, arn, bucket, key, size_bytes, content_type, created_at
		FROM objects
		WHERE bucket = $1 AND key = $2 AND deleted_at IS NULL
	`
	var obj domain.Object
	err := r.db.QueryRow(ctx, query, bucket, key).Scan(
		&obj.ID, &obj.ARN, &obj.Bucket, &obj.Key, &obj.SizeBytes, &obj.ContentType, &obj.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New(errors.ObjectNotFound, "object metadata not found")
		}
		return nil, errors.Wrap(errors.Internal, "failed to get object metadata", err)
	}
	return &obj, nil
}

func (r *StorageRepository) List(ctx context.Context, bucket string) ([]*domain.Object, error) {
	query := `
		SELECT id, arn, bucket, key, size_bytes, content_type, created_at
		FROM objects
		WHERE bucket = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, bucket)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to list objects", err)
	}
	defer rows.Close()

	var objects []*domain.Object
	for rows.Next() {
		var obj domain.Object
		err := rows.Scan(
			&obj.ID, &obj.ARN, &obj.Bucket, &obj.Key, &obj.SizeBytes, &obj.ContentType, &obj.CreatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(errors.Internal, "failed to scan object metadata", err)
		}
		objects = append(objects, &obj)
	}
	return objects, nil
}

func (r *StorageRepository) SoftDelete(ctx context.Context, bucket, key string) error {
	query := `
		UPDATE objects
		SET deleted_at = $1
		WHERE bucket = $2 AND key = $3 AND deleted_at IS NULL
	`
	cmd, err := r.db.Exec(ctx, query, time.Now(), bucket, key)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to soft delete object", err)
	}
	if cmd.RowsAffected() == 0 {
		return errors.New(errors.ObjectNotFound, "object not found or already deleted")
	}
	return nil
}
