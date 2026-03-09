package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheRepository_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewCacheRepository(mock)
		now := time.Now()
		vpcID := uuid.New()
		cache := &domain.Cache{
			ID:          uuid.New(),
			UserID:      uuid.New(),
			Name:        "test-cache",
			Engine:      domain.EngineRedis,
			Version:     "6.2",
			Status:      domain.CacheStatusCreating,
			VpcID:       &vpcID,
			ContainerID: "cid-1",
			Port:        6379,
			Password:    "pass",
			MemoryMB:    1024,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		mock.ExpectExec("INSERT INTO caches").
			WithArgs(cache.ID, cache.UserID, cache.Name, cache.Engine, cache.Version, cache.Status, cache.VpcID,
				cache.ContainerID, cache.Port, cache.Password, cache.MemoryMB, cache.CreatedAt, cache.UpdatedAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = repo.Create(context.Background(), cache)
		require.NoError(t, err)
	})

	t.Run("db_error", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewCacheRepository(mock)
		cache := &domain.Cache{ID: uuid.New()}

		mock.ExpectExec("INSERT INTO caches").
			WillReturnError(assert.AnError)

		err = repo.Create(context.Background(), cache)
		require.Error(t, err)
		assert.True(t, errors.Is(err, errors.Internal))
	})
}

func TestCacheRepository_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewCacheRepository(mock)
		id := uuid.New()
		now := time.Now()
		vpcID := uuid.New()

		mock.ExpectQuery("SELECT.*FROM caches WHERE id = \\$1").
			WithArgs(id).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "engine", "version", "status", "vpc_id", "container_id", "port", "password", "memory_mb", "created_at", "updated_at"}).
				AddRow(id, uuid.New(), "test-cache", string(domain.EngineRedis), "6.2", string(domain.CacheStatusRunning), vpcID,
					"cid-1", 6379, "pass", 1024, now, now))

		cache, err := repo.GetByID(context.Background(), id)
		require.NoError(t, err)
		assert.NotNil(t, cache)
		assert.Equal(t, id, cache.ID)
		assert.Equal(t, domain.EngineRedis, cache.Engine)
	})

	t.Run("not_found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewCacheRepository(mock)
		id := uuid.New()

		mock.ExpectQuery("SELECT.*FROM caches").
			WithArgs(id).
			WillReturnError(pgx.ErrNoRows)

		cache, err := repo.GetByID(context.Background(), id)
		require.Error(t, err)
		assert.Nil(t, cache)
		assert.True(t, errors.Is(err, errors.NotFound))
	})
}

func TestCacheRepository_GetByName(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewCacheRepository(mock)
		userID := uuid.New()
		name := "test-cache"
		now := time.Now()

		mock.ExpectQuery("SELECT.*FROM caches WHERE user_id = \\$1 AND name = \\$2").
			WithArgs(userID, name).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "engine", "version", "status", "vpc_id", "container_id", "port", "password", "memory_mb", "created_at", "updated_at"}).
				AddRow(uuid.New(), userID, name, string(domain.EngineRedis), "6.2", string(domain.CacheStatusRunning), nil,
					"cid-1", 6379, "pass", 1024, now, now))

		cache, err := repo.GetByName(context.Background(), userID, name)
		require.NoError(t, err)
		assert.NotNil(t, cache)
		assert.Equal(t, name, cache.Name)
	})

	t.Run("not_found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewCacheRepository(mock)
		userID := uuid.New()
		name := "test-cache"

		mock.ExpectQuery("SELECT.*FROM caches").
			WithArgs(userID, name).
			WillReturnError(pgx.ErrNoRows)

		cache, err := repo.GetByName(context.Background(), userID, name)
		require.Error(t, err)
		assert.Nil(t, cache)
		assert.True(t, errors.Is(err, errors.NotFound))
	})
}

func TestCacheRepository_List(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewCacheRepository(mock)
		userID := uuid.New()
		now := time.Now()

		mock.ExpectQuery("SELECT.*FROM caches WHERE user_id = \\$1").
			WithArgs(userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "engine", "version", "status", "vpc_id", "container_id", "port", "password", "memory_mb", "created_at", "updated_at"}).
				AddRow(uuid.New(), userID, "cache-1", string(domain.EngineRedis), "6.2", string(domain.CacheStatusRunning), nil, "cid-1", 6379, "pass", 1024, now, now).
				AddRow(uuid.New(), userID, "cache-2", string(domain.EngineRedis), "6.2", string(domain.CacheStatusStopped), nil, "cid-2", 6380, "pass", 1024, now, now))

		caches, err := repo.List(context.Background(), userID)
		require.NoError(t, err)
		assert.Len(t, caches, 2)
	})

	t.Run("db_error", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewCacheRepository(mock)
		userID := uuid.New()

		mock.ExpectQuery("SELECT.*FROM caches").
			WithArgs(userID).
			WillReturnError(assert.AnError)

		caches, err := repo.List(context.Background(), userID)
		require.Error(t, err)
		assert.Nil(t, caches)
		assert.True(t, errors.Is(err, errors.Internal))
	})
}

func TestCacheRepository_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewCacheRepository(mock)
		cache := &domain.Cache{
			ID:          uuid.New(),
			Status:      domain.CacheStatusRunning,
			ContainerID: "cid-new",
			Port:        6379,
			UpdatedAt:   time.Now(),
		}

		mock.ExpectExec("UPDATE caches").
			WithArgs(cache.Status, cache.ContainerID, cache.Port, cache.UpdatedAt, cache.ID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err = repo.Update(context.Background(), cache)
		require.NoError(t, err)
	})
	t.Run("db_error", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewCacheRepository(mock)
		cache := &domain.Cache{ID: uuid.New()}

		mock.ExpectExec("UPDATE caches").
			WillReturnError(assert.AnError)

		err = repo.Update(context.Background(), cache)
		require.Error(t, err)
		assert.True(t, errors.Is(err, errors.Internal))
	})
}

func TestCacheRepository_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewCacheRepository(mock)
		id := uuid.New()

		mock.ExpectExec("DELETE FROM caches").
			WithArgs(id).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = repo.Delete(context.Background(), id)
		require.NoError(t, err)
	})

	t.Run("db_error", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewCacheRepository(mock)
		id := uuid.New()

		mock.ExpectExec("DELETE FROM caches").
			WithArgs(id).
			WillReturnError(assert.AnError)

		err = repo.Delete(context.Background(), id)
		require.Error(t, err)
		assert.True(t, errors.Is(err, errors.Internal))
	})
}
