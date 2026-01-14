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

const (
	testMountPath    = "/mock/path"
	testSelectVolume = "SELECT id, user_id, name, size_gb, status, instance_id, backend_path, mount_path, created_at, updated_at FROM volumes"
)

func TestVolumeRepositoryCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewVolumeRepository(mock)
		instanceID := uuid.New()
		vol := &domain.Volume{
			ID:         uuid.New(),
			UserID:     uuid.New(),
			Name:       "vol-1",
			SizeGB:     10,
			Status:     domain.VolumeStatusAvailable,
			InstanceID: &instanceID,
			MountPath:  testMountPath,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		mock.ExpectExec("INSERT INTO volumes").
			WithArgs(vol.ID, vol.UserID, vol.Name, vol.SizeGB, string(vol.Status), vol.InstanceID, vol.BackendPath, vol.MountPath, vol.CreatedAt, vol.UpdatedAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = repo.Create(context.Background(), vol)
		assert.NoError(t, err)
	})

	t.Run(testDBError, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewVolumeRepository(mock)
		vol := &domain.Volume{
			ID: uuid.New(),
		}

		mock.ExpectExec("INSERT INTO volumes").
			WillReturnError(errors.New(testDBError))

		err = repo.Create(context.Background(), vol)
		assert.Error(t, err)
	})
}

func TestVolumeRepositoryGetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewVolumeRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		now := time.Now()

		mock.ExpectQuery(testSelectVolume).
			WithArgs(id, userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "size_gb", "status", "instance_id", "backend_path", "mount_path", "created_at", "updated_at"}).
				AddRow(id, userID, "vol-1", 10, string(domain.VolumeStatusAvailable), &id, "", testMountPath, now, now))

		vol, err := repo.GetByID(ctx, id)
		assert.NoError(t, err)
		assert.NotNil(t, vol)
		assert.Equal(t, id, vol.ID)
	})

	t.Run(testNotFound, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewVolumeRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectQuery(testSelectVolume).
			WithArgs(id, userID).
			WillReturnError(pgx.ErrNoRows)

		vol, err := repo.GetByID(ctx, id)
		assert.Error(t, err)
		assert.Nil(t, vol)
		theCloudErr, ok := err.(*theclouderrors.Error)
		if ok {
			assert.Equal(t, theclouderrors.NotFound, theCloudErr.Type)
		}
	})
}

func TestVolumeRepositoryGetByName(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewVolumeRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		now := time.Now()
		name := "vol-1"

		mock.ExpectQuery(testSelectVolume).
			WithArgs(name, userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "size_gb", "status", "instance_id", "backend_path", "mount_path", "created_at", "updated_at"}).
				AddRow(id, userID, name, 10, string(domain.VolumeStatusAvailable), &id, "", testMountPath, now, now))

		vol, err := repo.GetByName(ctx, name)
		assert.NoError(t, err)
		assert.NotNil(t, vol)
		assert.Equal(t, id, vol.ID)
	})

	t.Run(testNotFound, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewVolumeRepository(mock)
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		name := "vol-1"

		mock.ExpectQuery(testSelectVolume).
			WithArgs(name, userID).
			WillReturnError(pgx.ErrNoRows)

		vol, err := repo.GetByName(ctx, name)
		assert.Error(t, err)
		assert.Nil(t, vol)
		theCloudErr, ok := err.(*theclouderrors.Error)
		if ok {
			assert.Equal(t, theclouderrors.NotFound, theCloudErr.Type)
		}
	})
}

func TestVolumeRepositoryList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewVolumeRepository(mock)
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		now := time.Now()

		instID := uuid.New()
		mock.ExpectQuery(testSelectVolume).
			WithArgs(userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "size_gb", "status", "instance_id", "backend_path", "mount_path", "created_at", "updated_at"}).
				AddRow(uuid.New(), userID, "vol-1", 10, string(domain.VolumeStatusAvailable), &instID, "", testMountPath, now, now))

		vols, err := repo.List(ctx)
		assert.NoError(t, err)
		assert.Len(t, vols, 1)
	})

	t.Run(testDBError, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewVolumeRepository(mock)
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectQuery(testSelectVolume).
			WithArgs(userID).
			WillReturnError(errors.New(testDBError))

		vols, err := repo.List(ctx)
		assert.Error(t, err)
		assert.Nil(t, vols)
	})
}

func TestVolumeRepositoryListByInstanceID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewVolumeRepository(mock)
		userID := uuid.New()
		instanceID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		now := time.Now()

		mock.ExpectQuery(testSelectVolume).
			WithArgs(instanceID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "size_gb", "status", "instance_id", "backend_path", "mount_path", "created_at", "updated_at"}).
				AddRow(uuid.New(), userID, "vol-1", 10, string(domain.VolumeStatusAvailable), &instanceID, "", testMountPath, now, now))

		vols, err := repo.ListByInstanceID(ctx, instanceID)
		assert.NoError(t, err)
		assert.Len(t, vols, 1)
	})

	t.Run(testDBError, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewVolumeRepository(mock)
		userID := uuid.New()
		instanceID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectQuery(testSelectVolume).
			WithArgs(instanceID, userID).
			WillReturnError(errors.New(testDBError))

		vols, err := repo.ListByInstanceID(ctx, instanceID)
		assert.Error(t, err)
		assert.Nil(t, vols)
	})
}

func TestVolumeRepositoryUpdate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewVolumeRepository(mock)
		vol := &domain.Volume{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			Status:    domain.VolumeStatusInUse,
			UpdatedAt: time.Now(),
		}

		mock.ExpectExec("UPDATE volumes").
			WithArgs(string(vol.Status), vol.InstanceID, vol.BackendPath, vol.MountPath, vol.UpdatedAt, vol.ID, vol.UserID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err = repo.Update(context.Background(), vol)
		assert.NoError(t, err)
	})

	t.Run(testDBError, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewVolumeRepository(mock)
		vol := &domain.Volume{
			ID: uuid.New(),
		}

		mock.ExpectExec("UPDATE volumes").
			WillReturnError(errors.New(testDBError))

		err = repo.Update(context.Background(), vol)
		assert.Error(t, err)
	})
}

func TestVolumeRepositoryDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewVolumeRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectExec("DELETE FROM volumes").
			WithArgs(id, userID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = repo.Delete(ctx, id)
		assert.NoError(t, err)
	})

	t.Run(testNotFound, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewVolumeRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectExec("DELETE FROM volumes").
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
