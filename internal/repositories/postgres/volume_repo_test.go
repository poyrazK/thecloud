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
	testMountPath    = "/mock/path"
	testSelectVolume = "SELECT id, user_id, tenant_id, name, size_gb, status, instance_id, backend_path, mount_path, created_at, updated_at FROM volumes"
)

func TestVolumeRepositoryCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewVolumeRepository(mock)
		instanceID := uuid.New()
		tenantID := uuid.New()
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
			WithArgs(vol.ID, vol.UserID, tenantID, vol.Name, vol.SizeGB, string(vol.Status), vol.InstanceID, vol.BackendPath, vol.MountPath, vol.CreatedAt, vol.UpdatedAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		ctx := appcontext.WithTenantID(context.Background(), tenantID)
		err = repo.Create(ctx, vol)
		require.NoError(t, err)
	})

	t.Run(testDBError, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewVolumeRepository(mock)
		tenantID := uuid.New()
		vol := &domain.Volume{
			ID: uuid.New(),
		}

		mock.ExpectExec("INSERT INTO volumes").
			WillReturnError(errors.New(testDBError))

		ctx := appcontext.WithTenantID(context.Background(), tenantID)
		err = repo.Create(ctx, vol)
		require.Error(t, err)
	})
}

func TestVolumeRepositoryGetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewVolumeRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		tenantID := uuid.New()
		ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), userID), tenantID)
		now := time.Now()

		mock.ExpectQuery(testSelectVolume).
			WithArgs(id, tenantID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "tenant_id", "name", "size_gb", "status", "instance_id", "backend_path", "mount_path", "created_at", "updated_at"}).
				AddRow(id, userID, tenantID, "vol-1", 10, string(domain.VolumeStatusAvailable), &id, "", testMountPath, now, now))

		vol, err := repo.GetByID(ctx, id)
		require.NoError(t, err)
		assert.NotNil(t, vol)
		assert.Equal(t, id, vol.ID)
	})

	t.Run(testNotFound, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewVolumeRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		tenantID := uuid.New()
		ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), userID), tenantID)

		mock.ExpectQuery(testSelectVolume).
			WithArgs(id, tenantID).
			WillReturnError(pgx.ErrNoRows)

		vol, err := repo.GetByID(ctx, id)
		require.Error(t, err)
		assert.Nil(t, vol)
		var target *theclouderrors.Error
		if errors.As(err, &target) {
			assert.Equal(t, theclouderrors.NotFound, target.Type)
		}
	})
}

func TestVolumeRepositoryGetByName(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewVolumeRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		tenantID := uuid.New()
		ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), userID), tenantID)
		now := time.Now()
		name := "vol-1"

		mock.ExpectQuery(testSelectVolume).
			WithArgs(name, tenantID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "tenant_id", "name", "size_gb", "status", "instance_id", "backend_path", "mount_path", "created_at", "updated_at"}).
				AddRow(id, userID, tenantID, name, 10, string(domain.VolumeStatusAvailable), &id, "", testMountPath, now, now))

		vol, err := repo.GetByName(ctx, name)
		require.NoError(t, err)
		assert.NotNil(t, vol)
		assert.Equal(t, id, vol.ID)
	})

	t.Run(testNotFound, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewVolumeRepository(mock)
		userID := uuid.New()
		tenantID := uuid.New()
		ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), userID), tenantID)
		name := "vol-1"

		mock.ExpectQuery(testSelectVolume).
			WithArgs(name, tenantID).
			WillReturnError(pgx.ErrNoRows)

		vol, err := repo.GetByName(ctx, name)
		require.Error(t, err)
		assert.Nil(t, vol)
		var target *theclouderrors.Error
		ok := errors.As(err, &target)
		if ok {
			assert.Equal(t, theclouderrors.NotFound, target.Type)
		}
	})
}

func TestVolumeRepositoryList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewVolumeRepository(mock)
		userID := uuid.New()
		tenantID := uuid.New()
		ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), userID), tenantID)
		now := time.Now()

		instID := uuid.New()
		mock.ExpectQuery(testSelectVolume).
			WithArgs(tenantID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "tenant_id", "name", "size_gb", "status", "instance_id", "backend_path", "mount_path", "created_at", "updated_at"}).
				AddRow(uuid.New(), userID, tenantID, "vol-1", 10, string(domain.VolumeStatusAvailable), &instID, "", testMountPath, now, now))

		vols, err := repo.List(ctx)
		require.NoError(t, err)
		assert.Len(t, vols, 1)
	})

	t.Run(testDBError, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewVolumeRepository(mock)
		userID := uuid.New()
		tenantID := uuid.New()
		ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), userID), tenantID)

		mock.ExpectQuery(testSelectVolume).
			WithArgs(tenantID).
			WillReturnError(errors.New(testDBError))

		vols, err := repo.List(ctx)
		require.Error(t, err)
		assert.Nil(t, vols)
	})
}

func TestVolumeRepositoryListByInstanceID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewVolumeRepository(mock)
		userID := uuid.New()
		instanceID := uuid.New()
		tenantID := uuid.New()
		ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), userID), tenantID)
		now := time.Now()

		mock.ExpectQuery(testSelectVolume).
			WithArgs(instanceID, tenantID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "tenant_id", "name", "size_gb", "status", "instance_id", "backend_path", "mount_path", "created_at", "updated_at"}).
				AddRow(uuid.New(), userID, tenantID, "vol-1", 10, string(domain.VolumeStatusAvailable), &instanceID, "", testMountPath, now, now))

		vols, err := repo.ListByInstanceID(ctx, instanceID)
		require.NoError(t, err)
		assert.Len(t, vols, 1)
	})

	t.Run(testDBError, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewVolumeRepository(mock)
		userID := uuid.New()
		instanceID := uuid.New()
		tenantID := uuid.New()
		ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), userID), tenantID)

		mock.ExpectQuery(testSelectVolume).
			WithArgs(instanceID, tenantID).
			WillReturnError(errors.New(testDBError))

		vols, err := repo.ListByInstanceID(ctx, instanceID)
		require.Error(t, err)
		assert.Nil(t, vols)
	})
}

func TestVolumeRepositoryUpdate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewVolumeRepository(mock)
		vol := &domain.Volume{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			Status:    domain.VolumeStatusInUse,
			UpdatedAt: time.Now(),
		}
		tenantID := uuid.New()

		mock.ExpectExec("UPDATE volumes").
			WithArgs(string(vol.Status), vol.InstanceID, vol.BackendPath, vol.MountPath, vol.UpdatedAt, vol.ID, tenantID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		ctx := appcontext.WithTenantID(context.Background(), tenantID)
		err = repo.Update(ctx, vol)
		require.NoError(t, err)
	})

	t.Run(testDBError, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewVolumeRepository(mock)
		vol := &domain.Volume{
			ID: uuid.New(),
		}
		tenantID := uuid.New()

		mock.ExpectExec("UPDATE volumes").
			WillReturnError(errors.New(testDBError))

		ctx := appcontext.WithTenantID(context.Background(), tenantID)
		err = repo.Update(ctx, vol)
		require.Error(t, err)
	})
}

func TestVolumeRepositoryDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewVolumeRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		tenantID := uuid.New()
		ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), userID), tenantID)

		mock.ExpectExec("DELETE FROM volumes").
			WithArgs(id, tenantID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = repo.Delete(ctx, id)
		require.NoError(t, err)
	})

	t.Run(testNotFound, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewVolumeRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		tenantID := uuid.New()
		ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), userID), tenantID)

		mock.ExpectExec("DELETE FROM volumes").
			WithArgs(id, tenantID).
			WillReturnResult(pgxmock.NewResult("DELETE", 0))

		err = repo.Delete(ctx, id)
		require.Error(t, err)
		var target *theclouderrors.Error
		ok := errors.As(err, &target)
		if ok {
			assert.Equal(t, theclouderrors.NotFound, target.Type)
		}
	})
}
func TestVolumeRepositoryConcurrentList(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewVolumeRepository(mock)
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), userID), tenantID)
	now := time.Now()

	// Expect multiple queries
	concurrency := 10
	for i := 0; i < concurrency; i++ {
		mock.ExpectQuery(testSelectVolume).
			WithArgs(tenantID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "tenant_id", "name", "size_gb", "status", "instance_id", "backend_path", "mount_path", "created_at", "updated_at"}).
				AddRow(uuid.New(), userID, tenantID, "vol-1", 10, string(domain.VolumeStatusAvailable), nil, "", testMountPath, now, now))
	}

	done := make(chan bool)
	for i := 0; i < concurrency; i++ {
		go func() {
			vols, err := repo.List(ctx)
			assert.NoError(t, err)
			assert.NotNil(t, vols)
			done <- true
		}()
	}

	for i := 0; i < concurrency; i++ {
		<-done
	}
}
