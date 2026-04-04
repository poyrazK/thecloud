package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	theclouderrors "github.com/poyrazk/thecloud/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	dummyChecksum = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
)

func TestStorageRepositorySaveMeta(t *testing.T) {
	t.Run("success without demotion", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewStorageRepository(mock)
		obj := &domain.Object{
			ID:           uuid.New(),
			UserID:       uuid.New(),
			ARN:          "arn:aws:s3:::mybucket/mykey",
			Bucket:       "mybucket",
			Key:          "mykey",
			IsLatest:     false,
			SizeBytes:    1024,
			ContentType:  "text/plain",
			Checksum:     dummyChecksum,
			UploadStatus: domain.UploadStatusAvailable,
			CreatedAt:    time.Now(),
		}

		mock.ExpectExec("INSERT INTO objects").
			WithArgs(obj.ID, obj.UserID, obj.ARN, obj.Bucket, obj.Key, obj.VersionID, obj.IsLatest, obj.SizeBytes, obj.ContentType, obj.Checksum, obj.UploadStatus, obj.CreatedAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = repo.SaveMeta(context.Background(), obj)
		require.NoError(t, err)
	})

	t.Run("success with demotion", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewStorageRepository(mock)
		obj := &domain.Object{
			ID:           uuid.New(),
			UserID:       uuid.New(),
			ARN:          "arn:aws:s3:::mybucket/mykey",
			Bucket:       "mybucket",
			Key:          "mykey",
			VersionID:    "v2",
			IsLatest:     true,
			SizeBytes:    1024,
			ContentType:  "text/plain",
			Checksum:     dummyChecksum,
			UploadStatus: domain.UploadStatusAvailable,
			CreatedAt:    time.Now(),
		}

		// Expect demotion UPDATE
		mock.ExpectExec("UPDATE objects SET is_latest = FALSE WHERE bucket = \\$1 AND key = \\$2 AND is_latest = TRUE AND version_id != \\$3").
			WithArgs(obj.Bucket, obj.Key, obj.VersionID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		// Expect INSERT
		mock.ExpectExec("INSERT INTO objects").
			WithArgs(obj.ID, obj.UserID, obj.ARN, obj.Bucket, obj.Key, obj.VersionID, obj.IsLatest, obj.SizeBytes, obj.ContentType, obj.Checksum, obj.UploadStatus, obj.CreatedAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = repo.SaveMeta(context.Background(), obj)
		require.NoError(t, err)
	})

	t.Run("success with empty checksum (NULL)", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewStorageRepository(mock)
		obj := &domain.Object{
			ID:           uuid.New(),
			UserID:       uuid.New(),
			ARN:          "arn:aws:s3:::mybucket/mykey",
			Bucket:       "mybucket",
			Key:          "mykey",
			IsLatest:     false,
			SizeBytes:    0,
			ContentType:  "text/plain",
			Checksum:     "", // Should be passed as nil to DB
			UploadStatus: domain.UploadStatusPending,
			CreatedAt:    time.Now(),
		}

		mock.ExpectExec("INSERT INTO objects").
			WithArgs(obj.ID, obj.UserID, obj.ARN, obj.Bucket, obj.Key, obj.VersionID, obj.IsLatest, obj.SizeBytes, obj.ContentType, nil, obj.UploadStatus, obj.CreatedAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = repo.SaveMeta(context.Background(), obj)
		require.NoError(t, err)
	})

	t.Run("db error on demotion", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewStorageRepository(mock)
		obj := &domain.Object{
			Bucket:       "b",
			Key:          "k",
			IsLatest:     true,
			UploadStatus: domain.UploadStatusAvailable,
		}

		mock.ExpectExec("UPDATE objects").WillReturnError(errors.New("demote error"))

		err = repo.SaveMeta(context.Background(), obj)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update previous latest")
	})

	t.Run("db error on insert", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewStorageRepository(mock)
		obj := &domain.Object{
			ID: uuid.New(),
		}

		mock.ExpectExec("INSERT INTO objects").
			WillReturnError(errors.New("db error"))

		err = repo.SaveMeta(context.Background(), obj)
		require.Error(t, err)
	})
}

func TestStorageRepositoryGetMeta(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewStorageRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		now := time.Now()

		mock.ExpectQuery("SELECT id, user_id, arn, bucket, key, version_id, is_latest, size_bytes, content_type, COALESCE\\(checksum, ''\\), upload_status, created_at, deleted_at FROM objects").
			WithArgs("mybucket", "mykey", userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "arn", "bucket", "key", "version_id", "is_latest", "size_bytes", "content_type", "checksum", "upload_status", "created_at", "deleted_at"}).
				AddRow(id, userID, "arn", "mybucket", "mykey", "v1", true, int64(1024), "text/plain", dummyChecksum, domain.UploadStatusAvailable, now, nil))

		obj, err := repo.GetMeta(ctx, "mybucket", "mykey")
		require.NoError(t, err)
		assert.NotNil(t, obj)
		assert.Equal(t, id, obj.ID)
	})

	t.Run("not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewStorageRepository(mock)
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectQuery("SELECT id, user_id, arn, bucket, key, version_id, is_latest, size_bytes, content_type, COALESCE\\(checksum, ''\\), upload_status, created_at, deleted_at FROM objects").
			WithArgs("mybucket", "mykey", userID).
			WillReturnError(pgx.ErrNoRows)

		obj, err := repo.GetMeta(ctx, "mybucket", "mykey")
		require.Error(t, err)
		assert.Nil(t, obj)
		var target *theclouderrors.Error
		ok := errors.As(err, &target)
		if ok {
			assert.Equal(t, theclouderrors.ObjectNotFound, target.Type)
		}
	})

	t.Run("db error", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewStorageRepository(mock)
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectQuery("SELECT id, user_id, arn, bucket, key, version_id, is_latest, size_bytes, content_type, COALESCE\\(checksum, ''\\), upload_status, created_at, deleted_at FROM objects").
			WithArgs("mybucket", "mykey", userID).
			WillReturnError(errors.New("db error"))

		obj, err := repo.GetMeta(ctx, "mybucket", "mykey")
		require.Error(t, err)
		assert.Nil(t, obj)
	})
}

func TestStorageRepositoryList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewStorageRepository(mock)
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		now := time.Now()

		mock.ExpectQuery("SELECT id, user_id, arn, bucket, key, version_id, is_latest, size_bytes, content_type, COALESCE\\(checksum, ''\\), upload_status, created_at, deleted_at FROM objects").
			WithArgs("mybucket", userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "arn", "bucket", "key", "version_id", "is_latest", "size_bytes", "content_type", "checksum", "upload_status", "created_at", "deleted_at"}).
				AddRow(uuid.New(), userID, "arn", "mybucket", "mykey", "v1", true, int64(1024), "text/plain", dummyChecksum, domain.UploadStatusAvailable, now, nil))

		objects, err := repo.List(ctx, "mybucket")
		require.NoError(t, err)
		assert.Len(t, objects, 1)
	})

	t.Run("db error", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewStorageRepository(mock)
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectQuery("SELECT id, user_id, arn, bucket, key, version_id, is_latest, size_bytes, content_type, COALESCE\\(checksum, ''\\), upload_status, created_at, deleted_at FROM objects").
			WithArgs("mybucket", userID).
			WillReturnError(errors.New("db error"))

		objects, err := repo.List(ctx, "mybucket")
		require.Error(t, err)
		assert.Nil(t, objects)
	})
}

func TestStorageRepositorySoftDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewStorageRepository(mock)
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		bucket := "mybucket"
		key := "mykey"

		mock.ExpectExec("UPDATE objects").
			WithArgs(pgxmock.AnyArg(), bucket, key, userID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err = repo.SoftDelete(ctx, bucket, key)
		require.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewStorageRepository(mock)
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		bucket := "mybucket"
		key := "mykey"

		mock.ExpectExec("UPDATE objects").
			WithArgs(pgxmock.AnyArg(), bucket, key, userID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 0))

		err = repo.SoftDelete(ctx, bucket, key)
		require.Error(t, err)
		var target *theclouderrors.Error
		ok := errors.As(err, &target)
		if ok {
			assert.Equal(t, theclouderrors.ObjectNotFound, target.Type)
		}
	})
}

func TestStorageRepositorySetBucketVersioning(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewStorageRepository(mock)
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		bucket := "mybucket"

		mock.ExpectExec("UPDATE buckets SET versioning_enabled = \\$1 WHERE name = \\$2 AND user_id = \\$3").
			WithArgs(true, bucket, userID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err = repo.SetBucketVersioning(ctx, bucket, true)
		require.NoError(t, err)
	})

	t.Run("not found or not owner", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewStorageRepository(mock)
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		bucket := "mybucket"

		mock.ExpectExec("UPDATE buckets").
			WithArgs(true, bucket, userID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 0))

		err = repo.SetBucketVersioning(ctx, bucket, true)
		require.Error(t, err)
		var target *theclouderrors.Error
		ok := errors.As(err, &target)
		if ok {
			assert.Equal(t, theclouderrors.BucketNotFound, target.Type)
		}
	})
}
