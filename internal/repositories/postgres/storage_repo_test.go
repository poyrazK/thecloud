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
)

func TestStorageRepository_SaveMeta(t *testing.T) {
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
		WithArgs(obj.ID, obj.UserID, obj.ARN, obj.Bucket, obj.Key, obj.SizeBytes, obj.ContentType, obj.CreatedAt).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.SaveMeta(context.Background(), obj)
	assert.NoError(t, err)
}

func TestStorageRepository_GetMeta(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewStorageRepository(mock)
	id := uuid.New()
	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	now := time.Now()

	mock.ExpectQuery("SELECT id, user_id, arn, bucket, key, size_bytes, content_type, created_at FROM objects").
		WithArgs("mybucket", "mykey", userID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "arn", "bucket", "key", "size_bytes", "content_type", "created_at"}).
			AddRow(id, userID, "arn", "mybucket", "mykey", int64(1024), "text/plain", now))

	obj, err := repo.GetMeta(ctx, "mybucket", "mykey")
	assert.NoError(t, err)
	assert.NotNil(t, obj)
	assert.Equal(t, id, obj.ID)
}

func TestStorageRepository_List(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewStorageRepository(mock)
	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	now := time.Now()

	mock.ExpectQuery("SELECT id, user_id, arn, bucket, key, size_bytes, content_type, created_at FROM objects").
		WithArgs("mybucket", userID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "arn", "bucket", "key", "size_bytes", "content_type", "created_at"}).
			AddRow(uuid.New(), userID, "arn", "mybucket", "mykey", int64(1024), "text/plain", now))

	objects, err := repo.List(ctx, "mybucket")
	assert.NoError(t, err)
	assert.Len(t, objects, 1)
}

func TestStorageRepository_SoftDelete(t *testing.T) {
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
}
