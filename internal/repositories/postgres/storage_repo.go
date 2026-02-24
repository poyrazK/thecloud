// Package postgres provides PostgreSQL-backed repository implementations.
package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	stdlib_errors "errors"
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
	// If this is the new latest version, mark previous one as not latest
	if obj.IsLatest {
		updateQuery := `UPDATE objects SET is_latest = FALSE WHERE bucket = $1 AND key = $2 AND is_latest = TRUE`
		_, err := r.db.Exec(ctx, updateQuery, obj.Bucket, obj.Key)
		if err != nil {
			return errors.Wrap(errors.Internal, "failed to update previous latest", err)
		}
	}

	query := `
		INSERT INTO objects (id, user_id, arn, bucket, key, version_id, is_latest, size_bytes, content_type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (bucket, key, version_id) DO UPDATE SET
			size_bytes = EXCLUDED.size_bytes,
			content_type = EXCLUDED.content_type,
			created_at = EXCLUDED.created_at,
			deleted_at = NULL,
			is_latest = EXCLUDED.is_latest,
			user_id = EXCLUDED.user_id
	`
	_, err := r.db.Exec(ctx, query,
		obj.ID, obj.UserID, obj.ARN, obj.Bucket, obj.Key, obj.VersionID, obj.IsLatest, obj.SizeBytes, obj.ContentType, obj.CreatedAt,
	)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to save object metadata", err)
	}
	return nil
}

func (r *StorageRepository) GetMeta(ctx context.Context, bucket, key string) (*domain.Object, error) {
	userID := appcontext.UserIDFromContext(ctx)
	query := `
		SELECT id, user_id, arn, bucket, key, version_id, is_latest, size_bytes, content_type, created_at, deleted_at
		FROM objects
		WHERE bucket = $1 AND key = $2 AND deleted_at IS NULL AND user_id = $3 AND is_latest = TRUE
	`
	return r.scanObject(r.db.QueryRow(ctx, query, bucket, key, userID))
}

func (r *StorageRepository) DeleteVersion(ctx context.Context, bucket, key, versionID string) error {
	userID := appcontext.UserIDFromContext(ctx)
	query := `DELETE FROM objects WHERE bucket = $1 AND key = $2 AND version_id = $3 AND user_id = $4`
	cmd, err := r.db.Exec(ctx, query, bucket, key, versionID, userID)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to delete object version", err)
	}
	if cmd.RowsAffected() == 0 {
		return errors.New(errors.ObjectNotFound, "object version not found")
	}
	return nil
}

func (r *StorageRepository) GetMetaByVersion(ctx context.Context, bucket, key, versionID string) (*domain.Object, error) {
	userID := appcontext.UserIDFromContext(ctx)
	query := `
		SELECT id, user_id, arn, bucket, key, version_id, is_latest, size_bytes, content_type, created_at, deleted_at
		FROM objects
		WHERE bucket = $1 AND key = $2 AND version_id = $3 AND deleted_at IS NULL AND user_id = $4
	`
	return r.scanObject(r.db.QueryRow(ctx, query, bucket, key, versionID, userID))
}

func (r *StorageRepository) List(ctx context.Context, bucket string) ([]*domain.Object, error) {
	userID := appcontext.UserIDFromContext(ctx)
	query := `
		SELECT id, user_id, arn, bucket, key, version_id, is_latest, size_bytes, content_type, created_at, deleted_at
		FROM objects
		WHERE bucket = $1 AND deleted_at IS NULL AND user_id = $2 AND is_latest = TRUE
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, bucket, userID)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to list objects", err)
	}
	return r.scanObjects(rows)
}

func (r *StorageRepository) ListVersions(ctx context.Context, bucket, key string) ([]*domain.Object, error) {
	userID := appcontext.UserIDFromContext(ctx)
	query := `
		SELECT id, user_id, arn, bucket, key, version_id, is_latest, size_bytes, content_type, created_at, deleted_at
		FROM objects
		WHERE bucket = $1 AND key = $2 AND deleted_at IS NULL AND user_id = $3
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, bucket, key, userID)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to list versions", err)
	}
	return r.scanObjects(rows)
}

func (r *StorageRepository) ListDeleted(ctx context.Context, limit int) ([]*domain.Object, error) {
	query := `
		SELECT id, user_id, arn, bucket, key, version_id, is_latest, size_bytes, content_type, created_at, deleted_at
		FROM objects
		WHERE deleted_at IS NOT NULL
		LIMIT $1
	`
	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to list deleted objects", err)
	}
	return r.scanObjects(rows)
}

func (r *StorageRepository) HardDelete(ctx context.Context, bucket, key, versionID string) error {
	query := `DELETE FROM objects WHERE bucket = $1 AND key = $2 AND version_id = $3`
	_, err := r.db.Exec(ctx, query, bucket, key, versionID)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to hard delete object", err)
	}
	return nil
}

func (r *StorageRepository) scanObject(row pgx.Row) (*domain.Object, error) {
	var obj domain.Object
	err := row.Scan(
		&obj.ID, &obj.UserID, &obj.ARN, &obj.Bucket, &obj.Key, &obj.VersionID, &obj.IsLatest, &obj.SizeBytes, &obj.ContentType, &obj.CreatedAt, &obj.DeletedAt,
	)
	if err != nil {
		if stdlib_errors.Is(err, pgx.ErrNoRows) {
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

// CreateBucket creates a new bucket.
func (r *StorageRepository) CreateBucket(ctx context.Context, bucket *domain.Bucket) error {
	query := `
		INSERT INTO buckets (id, name, user_id, is_public, versioning_enabled, encryption_enabled, encryption_key_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(ctx, query, bucket.ID, bucket.Name, bucket.UserID, bucket.IsPublic, bucket.VersioningEnabled, bucket.EncryptionEnabled, bucket.EncryptionKeyID, bucket.CreatedAt)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to create bucket", err)
	}
	return nil
}

// GetBucket retrieves a bucket by name.
func (r *StorageRepository) GetBucket(ctx context.Context, name string) (*domain.Bucket, error) {
	query := `
		SELECT id, name, user_id, is_public, versioning_enabled, encryption_enabled, encryption_key_id, created_at
		FROM buckets
		WHERE name = $1
	`
	var bucket domain.Bucket
	err := r.db.QueryRow(ctx, query, name).Scan(
		&bucket.ID, &bucket.Name, &bucket.UserID, &bucket.IsPublic, &bucket.VersioningEnabled,
		&bucket.EncryptionEnabled, &bucket.EncryptionKeyID, &bucket.CreatedAt,
	)
	if err != nil {
		if stdlib_errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(errors.ObjectNotFound, "bucket not found")
		}
		return nil, errors.Wrap(errors.Internal, "failed to get bucket", err)
	}
	return &bucket, nil
}

func (r *StorageRepository) SetBucketVersioning(ctx context.Context, name string, enabled bool) error {
	query := `UPDATE buckets SET versioning_enabled = $1 WHERE name = $2`
	_, err := r.db.Exec(ctx, query, enabled, name)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to set bucket versioning", err)
	}
	return nil
}

// DeleteBucket deletes a bucket by name.
func (r *StorageRepository) DeleteBucket(ctx context.Context, name string) error {
	query := `DELETE FROM buckets WHERE name = $1`
	cmd, err := r.db.Exec(ctx, query, name)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to delete bucket", err)
	}
	if cmd.RowsAffected() == 0 {
		return errors.New(errors.ObjectNotFound, "bucket not found")
	}
	return nil
}

// ListBuckets list buckets for a user.
func (r *StorageRepository) ListBuckets(ctx context.Context, userID string) ([]*domain.Bucket, error) {
	query := `
		SELECT id, name, user_id, is_public, versioning_enabled, created_at
		FROM buckets
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to list buckets", err)
	}
	defer rows.Close()

	var buckets []*domain.Bucket
	for rows.Next() {
		var b domain.Bucket
		if err := rows.Scan(&b.ID, &b.Name, &b.UserID, &b.IsPublic, &b.VersioningEnabled, &b.CreatedAt); err != nil {
			return nil, err
		}
		buckets = append(buckets, &b)
	}
	return buckets, nil
}

func (r *StorageRepository) SaveMultipartUpload(ctx context.Context, upload *domain.MultipartUpload) error {
	query := `INSERT INTO multipart_uploads (id, user_id, bucket, key, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(ctx, query, upload.ID, upload.UserID, upload.Bucket, upload.Key, upload.CreatedAt)
	return err
}

func (r *StorageRepository) GetMultipartUpload(ctx context.Context, uploadID uuid.UUID) (*domain.MultipartUpload, error) {
	query := `SELECT id, user_id, bucket, key, created_at FROM multipart_uploads WHERE id = $1`
	var u domain.MultipartUpload
	err := r.db.QueryRow(ctx, query, uploadID).Scan(&u.ID, &u.UserID, &u.Bucket, &u.Key, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *StorageRepository) DeleteMultipartUpload(ctx context.Context, uploadID uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM multipart_uploads WHERE id = $1", uploadID)
	return err
}

func (r *StorageRepository) SavePart(ctx context.Context, part *domain.Part) error {
	query := `INSERT INTO multipart_parts (upload_id, part_number, size_bytes, etag) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(ctx, query, part.UploadID, part.PartNumber, part.SizeBytes, part.ETag)
	return err
}

func (r *StorageRepository) ListParts(ctx context.Context, uploadID uuid.UUID) ([]*domain.Part, error) {
	query := `SELECT upload_id, part_number, size_bytes, etag FROM multipart_parts WHERE upload_id = $1 ORDER BY part_number ASC`
	rows, err := r.db.Query(ctx, query, uploadID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parts []*domain.Part
	for rows.Next() {
		var p domain.Part
		if err := rows.Scan(&p.UploadID, &p.PartNumber, &p.SizeBytes, &p.ETag); err != nil {
			return nil, err
		}
		parts = append(parts, &p)
	}
	return parts, nil
}
