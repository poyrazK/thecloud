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
)

func TestStorageRepository_SaveMeta(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewStorageRepository(mock)
		obj := &domain.Object{
			ID:          uuid.New(),
			UserID:      uuid.New(),
			ARN:         "arn:aws:s3:::mybucket/mykey",
			Bucket:      "mybucket",
			Key:         "mykey",
			SizeBytes:   1024,
			ContentType: "text/plain",
			CreatedAt:   time.Now(),
		}

		mock.ExpectExec("INSERT INTO objects").
			WithArgs(obj.ID, obj.UserID, obj.ARN, obj.Bucket, obj.Key, obj.VersionID, obj.IsLatest, obj.SizeBytes, obj.ContentType, obj.CreatedAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = repo.SaveMeta(context.Background(), obj)
		assert.NoError(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewStorageRepository(mock)
		obj := &domain.Object{
			ID: uuid.New(),
		}

		mock.ExpectExec("INSERT INTO objects").
			WillReturnError(errors.New("db error"))

		err = repo.SaveMeta(context.Background(), obj)
		assert.Error(t, err)
	})
}

func TestStorageRepository_GetMeta(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewStorageRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		now := time.Now()

		mock.ExpectQuery("SELECT id, user_id, arn, bucket, key, version_id, is_latest, size_bytes, content_type, created_at, deleted_at FROM objects").
			WithArgs("mybucket", "mykey", userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "arn", "bucket", "key", "version_id", "is_latest", "size_bytes", "content_type", "created_at", "deleted_at"}).
				AddRow(id, userID, "arn", "mybucket", "mykey", "v1", true, int64(1024), "text/plain", now, nil))

		obj, err := repo.GetMeta(ctx, "mybucket", "mykey")
		assert.NoError(t, err)
		assert.NotNil(t, obj)
		assert.Equal(t, id, obj.ID)
	})

	t.Run("not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewStorageRepository(mock)
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectQuery("SELECT id, user_id, arn, bucket, key, version_id, is_latest, size_bytes, content_type, created_at FROM objects").
			WithArgs("mybucket", "mykey", userID).
			WillReturnError(pgx.ErrNoRows)

		obj, err := repo.GetMeta(ctx, "mybucket", "mykey")
		assert.Error(t, err)
		assert.Nil(t, obj)
		var target *theclouderrors.Error
		ok := errors.As(err, &target)
		if ok {
			assert.Equal(t, theclouderrors.NotFound, target.Type)
		}
	})

	t.Run("db error", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewStorageRepository(mock)
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectQuery("SELECT id, user_id, arn, bucket, key, version_id, is_latest, size_bytes, content_type, created_at FROM objects").
			WithArgs("mybucket", "mykey", userID).
			WillReturnError(errors.New("db error"))

		obj, err := repo.GetMeta(ctx, "mybucket", "mykey")
		assert.Error(t, err)
		assert.Nil(t, obj)
	})
}

func TestStorageRepository_List(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewStorageRepository(mock)
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		now := time.Now()

		mock.ExpectQuery("SELECT id, user_id, arn, bucket, key, version_id, is_latest, size_bytes, content_type, created_at, deleted_at FROM objects").
			WithArgs("mybucket", userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "arn", "bucket", "key", "version_id", "is_latest", "size_bytes", "content_type", "created_at", "deleted_at"}).
				AddRow(uuid.New(), userID, "arn", "mybucket", "mykey", "v1", true, int64(1024), "text/plain", now, nil))

		objects, err := repo.List(ctx, "mybucket")
		assert.NoError(t, err)
		assert.Len(t, objects, 1)
	})

	t.Run("db error", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewStorageRepository(mock)
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectQuery("SELECT id, user_id, arn, bucket, key, version_id, is_latest, size_bytes, content_type, created_at FROM objects").
			WithArgs("mybucket", userID).
			WillReturnError(errors.New("db error"))

		objects, err := repo.List(ctx, "mybucket")
		assert.Error(t, err)
		assert.Nil(t, objects)
	})
}

func TestStorageRepository_SoftDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewStorageRepository(mock)
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		bucket := "mybucket"
		key := "mykey"

		mock.ExpectExec("UPDATE objects SET deleted_at = \\$1 WHERE bucket = \\$2 AND key = \\$3 AND deleted_at IS NULL AND user_id = \\$4").
			WithArgs(pgxmock.AnyArg(), bucket, key, userID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err = repo.SoftDelete(ctx, bucket, key)
		assert.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewStorageRepository(mock)
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		bucket := "mybucket"
		key := "mykey"

		mock.ExpectExec("UPDATE objects SET deleted_at = \\$1 WHERE bucket = \\$2 AND key = \\$3 AND deleted_at IS NULL AND user_id = \\$4").
			WithArgs(pgxmock.AnyArg(), bucket, key, userID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 0))

		err = repo.SoftDelete(ctx, bucket, key)
		assert.Error(t, err)
		var target *theclouderrors.Error
		ok := errors.As(err, &target)
		if ok {
			assert.Equal(t, theclouderrors.NotFound, target.Type)
		}
	})

	t.Run("db error", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewStorageRepository(mock)
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		bucket := "mybucket"
		key := "mykey"

		mock.ExpectExec("UPDATE objects SET deleted_at = \\$1 WHERE bucket = \\$2 AND key = \\$3 AND deleted_at IS NULL AND user_id = \\$4").
			WithArgs(pgxmock.AnyArg(), bucket, key, userID).
			WillReturnError(errors.New("db error"))

		err = repo.SoftDelete(ctx, bucket, key)
		assert.Error(t, err)
	})
}
