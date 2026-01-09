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

func TestSnapshotRepository_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSnapshotRepository(mock)
		s := &domain.Snapshot{
			ID:          uuid.New(),
			UserID:      uuid.New(),
			VolumeID:    uuid.New(),
			VolumeName:  "vol1",
			SizeGB:      10,
			Status:      "available",
			Description: "desc",
			CreatedAt:   time.Now(),
		}

		mock.ExpectExec("INSERT INTO snapshots").
			WithArgs(s.ID, s.UserID, s.VolumeID, s.VolumeName, s.SizeGB, string(s.Status), s.Description, s.CreatedAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = repo.Create(context.Background(), s)
		assert.NoError(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSnapshotRepository(mock)
		s := &domain.Snapshot{
			ID: uuid.New(),
		}

		mock.ExpectExec("INSERT INTO snapshots").
			WillReturnError(errors.New("db error"))

		err = repo.Create(context.Background(), s)
		assert.Error(t, err)
	})
}

func TestSnapshotRepository_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSnapshotRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		now := time.Now()

		mock.ExpectQuery("SELECT id, user_id, volume_id, volume_name, size_gb, status, description, created_at FROM snapshots").
			WithArgs(id, userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "volume_id", "volume_name", "size_gb", "status", "description", "created_at"}).
				AddRow(id, userID, uuid.New(), "vol1", 10, string(domain.SnapshotStatusAvailable), "desc", now))

		s, err := repo.GetByID(ctx, id)
		assert.NoError(t, err)
		assert.NotNil(t, s)
		assert.Equal(t, id, s.ID)

	})

	t.Run("not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSnapshotRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectQuery("SELECT id, user_id, volume_id, volume_name, size_gb, status, description, created_at FROM snapshots").
			WithArgs(id, userID).
			WillReturnError(pgx.ErrNoRows)

		s, err := repo.GetByID(ctx, id)
		assert.Error(t, err)
		assert.Nil(t, s)
		theCloudErr, ok := err.(*theclouderrors.Error)
		if ok {
			assert.Equal(t, theclouderrors.NotFound, theCloudErr.Type)
		}
	})
}

func TestSnapshotRepository_ListByUserID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSnapshotRepository(mock)
		userID := uuid.New()
		now := time.Now()

		mock.ExpectQuery("SELECT id, user_id, volume_id, volume_name, size_gb, status, description, created_at FROM snapshots").
			WithArgs(userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "volume_id", "volume_name", "size_gb", "status", "description", "created_at"}).
				AddRow(uuid.New(), userID, uuid.New(), "vol1", 10, string(domain.SnapshotStatusAvailable), "desc", now))

		snapshots, err := repo.ListByUserID(context.Background(), userID)
		assert.NoError(t, err)
		assert.Len(t, snapshots, 1)
	})

	t.Run("db error", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSnapshotRepository(mock)
		userID := uuid.New()

		mock.ExpectQuery("SELECT id, user_id, volume_id, volume_name, size_gb, status, description, created_at FROM snapshots").
			WithArgs(userID).
			WillReturnError(errors.New("db error"))

		snapshots, err := repo.ListByUserID(context.Background(), userID)
		assert.Error(t, err)
		assert.Nil(t, snapshots)
	})
}

func TestSnapshotRepository_ListByVolumeID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSnapshotRepository(mock)
		volumeID := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		now := time.Now()

		mock.ExpectQuery("SELECT id, user_id, volume_id, volume_name, size_gb, status, description, created_at FROM snapshots").
			WithArgs(volumeID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "volume_id", "volume_name", "size_gb", "status", "description", "created_at"}).
				AddRow(uuid.New(), userID, volumeID, "vol1", 10, string(domain.SnapshotStatusAvailable), "desc", now))

		snapshots, err := repo.ListByVolumeID(ctx, volumeID)
		assert.NoError(t, err)
		assert.Len(t, snapshots, 1)
	})

	t.Run("db error", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSnapshotRepository(mock)
		volumeID := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectQuery("SELECT id, user_id, volume_id, volume_name, size_gb, status, description, created_at FROM snapshots").
			WithArgs(volumeID, userID).
			WillReturnError(errors.New("db error"))

		snapshots, err := repo.ListByVolumeID(ctx, volumeID)
		assert.Error(t, err)
		assert.Nil(t, snapshots)
	})
}

func TestSnapshotRepository_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSnapshotRepository(mock)
		s := &domain.Snapshot{
			ID:          uuid.New(),
			UserID:      uuid.New(),
			Status:      "available",
			Description: "updated desc",
		}

		mock.ExpectExec("UPDATE snapshots").
			WithArgs(string(s.Status), s.Description, s.ID, s.UserID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err = repo.Update(context.Background(), s)
		assert.NoError(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSnapshotRepository(mock)
		s := &domain.Snapshot{
			ID: uuid.New(),
		}

		mock.ExpectExec("UPDATE snapshots").
			WillReturnError(errors.New("db error"))

		err = repo.Update(context.Background(), s)
		assert.Error(t, err)
	})
}

func TestSnapshotRepository_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSnapshotRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectExec("DELETE FROM snapshots").
			WithArgs(id, userID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = repo.Delete(ctx, id)
		assert.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSnapshotRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectExec("DELETE FROM snapshots").
			WithArgs(id, userID).
			WillReturnResult(pgxmock.NewResult("DELETE", 0))

		err = repo.Delete(ctx, id)
		assert.Error(t, err)
		theCloudErr, ok := err.(*theclouderrors.Error)
		if ok {
			assert.Equal(t, theclouderrors.NotFound, theCloudErr.Type)
		}
	})
}
