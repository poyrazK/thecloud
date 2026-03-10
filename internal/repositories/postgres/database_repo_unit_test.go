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
	"github.com/stretchr/testify/require"
)

func TestDatabaseRepository_Create(t *testing.T) {
	t.Parallel()
	db := &domain.Database{
		ID:                  uuid.New(),
		UserID:              uuid.New(),
		Name:                "test-db",
		Engine:              domain.EnginePostgres,
		Version:             "16",
		Status:              domain.DatabaseStatusCreating,
		Role:                domain.RolePrimary,
		PrimaryID:           nil,
		VpcID:               nil,
		ContainerID:         "cid-1",
		Port:                5432,
		Username:            "admin",
		Password:            "password",
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
		AllocatedStorage:    20,
		Parameters:          map[string]string{"max_connections": "200"},
		MetricsEnabled:      true,
		MetricsPort:         9187,
		ExporterContainerID: "exp-cid",
		PoolingEnabled:      true,
		PoolingPort:         6432,
		PoolerContainerID:   "pool-cid",
	}

	t.Run("Success", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		repo := NewDatabaseRepository(mock)
		mock.ExpectExec("INSERT INTO databases").
			WithArgs(db.ID, db.UserID, db.Name, db.Engine, db.Version, db.Status, db.Role, db.PrimaryID, db.VpcID, db.ContainerID, db.Port, db.Username, db.Password, db.CreatedAt, db.UpdatedAt, db.AllocatedStorage, db.Parameters, db.MetricsEnabled, db.MetricsPort, db.ExporterContainerID, db.PoolingEnabled, db.PoolingPort, db.PoolerContainerID).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err := repo.Create(context.Background(), db)
		require.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		repo := NewDatabaseRepository(mock)
		mock.ExpectExec("INSERT INTO databases").WillReturnError(assert.AnError)
		err := repo.Create(context.Background(), db)
		assert.Error(t, err)
	})
}

func TestDatabaseRepository_GetByID(t *testing.T) {
	t.Parallel()
	id := uuid.New()
	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	now := time.Now()

	t.Run("Success", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		repo := NewDatabaseRepository(mock)
		mock.ExpectQuery("SELECT .* FROM databases").
			WithArgs(id, userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "engine", "version", "status", "role", "primary_id", "vpc_id", "container_id", "port", "username", "password", "created_at", "updated_at", "allocated_storage", "parameters", "metrics_enabled", "metrics_port", "exporter_container_id", "pooling_enabled", "pooling_port", "pooler_container_id"}).
				AddRow(id, userID, "test-db", string(domain.EnginePostgres), "16", string(domain.DatabaseStatusCreating), string(domain.RolePrimary), nil, nil, "cid-1", 5432, "admin", "password", now, now, 10, map[string]string{"k": "v"}, true, 9187, "exp-cid", true, 6432, "pool-cid"))

		db, err := repo.GetByID(ctx, id)
		require.NoError(t, err)
		assert.NotNil(t, db)
		assert.Equal(t, id, db.ID)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		repo := NewDatabaseRepository(mock)
		mock.ExpectQuery("SELECT .* FROM databases").WithArgs(id, userID).
			WillReturnRows(pgxmock.NewRows([]string{"id"}))
		db, err := repo.GetByID(ctx, id)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		assert.Nil(t, db)
	})

	t.Run("Error", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		repo := NewDatabaseRepository(mock)
		mock.ExpectQuery("SELECT .* FROM databases").WithArgs(id, userID).
			WillReturnError(assert.AnError)
		_, err := repo.GetByID(ctx, id)
		assert.Error(t, err)
	})
}

func TestDatabaseRepository_List(t *testing.T) {
	t.Parallel()
	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	now := time.Now()

	t.Run("Success", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		repo := NewDatabaseRepository(mock)
		mock.ExpectQuery("SELECT .* FROM databases").
			WithArgs(userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "engine", "version", "status", "role", "primary_id", "vpc_id", "container_id", "port", "username", "password", "created_at", "updated_at", "allocated_storage", "parameters", "metrics_enabled", "metrics_port", "exporter_container_id", "pooling_enabled", "pooling_port", "pooler_container_id"}).
				AddRow(uuid.New(), userID, "test-db", string(domain.EnginePostgres), "16", string(domain.DatabaseStatusCreating), string(domain.RolePrimary), nil, nil, "cid-1", 5432, "admin", "password", now, now, 20, map[string]string{}, false, 0, "", false, 0, ""))

		databases, err := repo.List(ctx)
		require.NoError(t, err)
		assert.Len(t, databases, 1)
	})

	t.Run("Error", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		repo := NewDatabaseRepository(mock)
		mock.ExpectQuery("SELECT .* FROM databases").WithArgs(userID).
			WillReturnError(assert.AnError)
		_, err := repo.List(ctx)
		assert.Error(t, err)
	})
}

func TestDatabaseRepository_ListReplicas(t *testing.T) {
	t.Parallel()
	primaryID := uuid.New()
	now := time.Now()

	t.Run("Success", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		repo := NewDatabaseRepository(mock)
		mock.ExpectQuery("SELECT .* FROM databases WHERE primary_id = \\$1").
			WithArgs(primaryID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "engine", "version", "status", "role", "primary_id", "vpc_id", "container_id", "port", "username", "password", "created_at", "updated_at", "allocated_storage", "parameters", "metrics_enabled", "metrics_port", "exporter_container_id", "pooling_enabled", "pooling_port", "pooler_container_id"}).
				AddRow(uuid.New(), uuid.New(), "replica-1", string(domain.EnginePostgres), "16", string(domain.DatabaseStatusRunning), string(domain.RoleReplica), &primaryID, nil, "cid-2", 5432, "admin", "password", now, now, 20, map[string]string{}, false, 0, "", false, 0, ""))

		replicas, err := repo.ListReplicas(context.Background(), primaryID)
		require.NoError(t, err)
		assert.Len(t, replicas, 1)
	})

	t.Run("Error", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		repo := NewDatabaseRepository(mock)
		mock.ExpectQuery("SELECT .* FROM databases WHERE primary_id = \\$1").WithArgs(primaryID).
			WillReturnError(assert.AnError)
		_, err := repo.ListReplicas(context.Background(), primaryID)
		assert.Error(t, err)
	})
}

func TestDatabaseRepository_Update(t *testing.T) {
	t.Parallel()
	db := &domain.Database{
		ID:                  uuid.New(),
		UserID:              uuid.New(),
		Name:                "updated-name",
		Status:              domain.DatabaseStatusRunning,
		Role:                domain.RolePrimary,
		ContainerID:         "cid-1",
		Port:                5432,
		Parameters:          map[string]string{"k": "v2"},
		MetricsEnabled:      true,
		MetricsPort:         9187,
		ExporterContainerID: "exp-cid",
		PoolingEnabled:      true,
		PoolingPort:         6432,
		PoolerContainerID:   "pool-cid",
		AllocatedStorage:    20,
	}

	t.Run("Success", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		repo := NewDatabaseRepository(mock)
		mock.ExpectExec("UPDATE databases").
			WithArgs(db.Name, db.Status, db.Role, db.PrimaryID, db.ContainerID, db.Port, pgxmock.AnyArg(), db.Parameters, db.MetricsEnabled, db.MetricsPort, db.ExporterContainerID, db.PoolingEnabled, db.PoolingPort, db.PoolerContainerID, db.AllocatedStorage, db.ID, db.UserID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := repo.Update(context.Background(), db)
		require.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		repo := NewDatabaseRepository(mock)
		mock.ExpectExec("UPDATE databases").WillReturnError(assert.AnError)
		err := repo.Update(context.Background(), db)
		assert.Error(t, err)
	})
}

func TestDatabaseRepository_Delete(t *testing.T) {
	t.Parallel()
	id := uuid.New()
	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)

	t.Run("Success", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		repo := NewDatabaseRepository(mock)
		mock.ExpectExec("DELETE FROM databases WHERE id = \\$1 AND user_id = \\$2").
			WithArgs(id, userID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err := repo.Delete(ctx, id)
		require.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		repo := NewDatabaseRepository(mock)
		mock.ExpectExec("DELETE FROM databases").WillReturnError(assert.AnError)
		err := repo.Delete(ctx, id)
		assert.Error(t, err)
	})
}
