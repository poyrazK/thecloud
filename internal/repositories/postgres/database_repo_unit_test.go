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
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewDatabaseRepository(mock)
	tenantID := uuid.New()
	db := &domain.Database{
		ID:                  uuid.New(),
		UserID:              uuid.New(),
		TenantID:            tenantID,
		Name:                "test-db",
		Engine:              domain.EnginePostgres,
		Version:             "16",
		Status:              domain.DatabaseStatusCreating,
		Role:                domain.RolePrimary,
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

	mock.ExpectExec("INSERT INTO databases").
		WithArgs(db.ID, db.UserID, db.TenantID, db.Name, db.Engine, db.Version, db.Status, db.Role, db.PrimaryID, db.VpcID, db.ContainerID, db.Port, db.Username, db.Password, db.CreatedAt, db.UpdatedAt, db.AllocatedStorage, db.Parameters, db.MetricsEnabled, db.MetricsPort, db.ExporterContainerID, db.PoolingEnabled, db.PoolingPort, db.PoolerContainerID).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.Create(context.Background(), db)
	require.NoError(t, err)
}

func TestDatabaseRepository_GetByID(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewDatabaseRepository(mock)
	id := uuid.New()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), userID), tenantID)
	now := time.Now()

	mock.ExpectQuery("SELECT id, user_id, tenant_id, name, engine, version, status, role, primary_id, vpc_id, COALESCE\\(container_id, ''\\), port, username, password, created_at, updated_at, allocated_storage, parameters, metrics_enabled, COALESCE\\(metrics_port, 0\\), COALESCE\\(exporter_container_id, ''\\), pooling_enabled, COALESCE\\(pooling_port, 0\\), COALESCE\\(pooler_container_id, ''\\) FROM databases WHERE id = \\$1 AND tenant_id = \\$2").
		WithArgs(id, tenantID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "tenant_id", "name", "engine", "version", "status", "role", "primary_id", "vpc_id", "container_id", "port", "username", "password", "created_at", "updated_at", "allocated_storage", "parameters", "metrics_enabled", "metrics_port", "exporter_container_id", "pooling_enabled", "pooling_port", "pooler_container_id"}).
			AddRow(id, userID, tenantID, "test-db", string(domain.EnginePostgres), "16", string(domain.DatabaseStatusCreating), string(domain.RolePrimary), nil, nil, "cid-1", 5432, "admin", "password", now, now, 10, map[string]string{"k": "v"}, true, 9187, "exp-cid", true, 6432, "pool-cid"))

	db, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.NotNil(t, db)
	assert.Equal(t, id, db.ID)
	assert.Equal(t, tenantID, db.TenantID)
}

func TestDatabaseRepository_List(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewDatabaseRepository(mock)
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), userID), tenantID)
	now := time.Now()

	mock.ExpectQuery("SELECT id, user_id, tenant_id, name, engine, version, status, role, primary_id, vpc_id, COALESCE\\(container_id, ''\\), port, username, password, created_at, updated_at, allocated_storage, parameters, metrics_enabled, COALESCE\\(metrics_port, 0\\), COALESCE\\(exporter_container_id, ''\\), pooling_enabled, COALESCE\\(pooling_port, 0\\), COALESCE\\(pooler_container_id, ''\\) FROM databases WHERE tenant_id = \\$1").
		WithArgs(tenantID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "tenant_id", "name", "engine", "version", "status", "role", "primary_id", "vpc_id", "container_id", "port", "username", "password", "created_at", "updated_at", "allocated_storage", "parameters", "metrics_enabled", "metrics_port", "exporter_container_id", "pooling_enabled", "pooling_port", "pooler_container_id"}).
			AddRow(uuid.New(), userID, tenantID, "test-db", string(domain.EnginePostgres), "16", string(domain.DatabaseStatusCreating), string(domain.RolePrimary), nil, nil, "cid-1", 5432, "admin", "password", now, now, 20, map[string]string{}, false, 0, "", false, 0, ""))

	databases, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, databases, 1)
	assert.Equal(t, tenantID, databases[0].TenantID)
}

func TestDatabaseRepository_ListReplicas(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewDatabaseRepository(mock)
	primaryID := uuid.New()
	tenantID := uuid.New()
	ctx := appcontext.WithTenantID(context.Background(), tenantID)
	now := time.Now()

	mock.ExpectQuery("SELECT id, user_id, tenant_id, name, engine, version, status, role, primary_id, vpc_id, COALESCE\\(container_id, ''\\), port, username, password, created_at, updated_at, allocated_storage, parameters, metrics_enabled, COALESCE\\(metrics_port, 0\\), COALESCE\\(exporter_container_id, ''\\), pooling_enabled, COALESCE\\(pooling_port, 0\\), COALESCE\\(pooler_container_id, ''\\) FROM databases WHERE primary_id = \\$1 AND tenant_id = \\$2").
		WithArgs(primaryID, tenantID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "tenant_id", "name", "engine", "version", "status", "role", "primary_id", "vpc_id", "container_id", "port", "username", "password", "created_at", "updated_at", "allocated_storage", "parameters", "metrics_enabled", "metrics_port", "exporter_container_id", "pooling_enabled", "pooling_port", "pooler_container_id"}).
			AddRow(uuid.New(), uuid.New(), tenantID, "replica-1", string(domain.EnginePostgres), "16", string(domain.DatabaseStatusRunning), string(domain.RoleReplica), &primaryID, nil, "cid-2", 5432, "admin", "password", now, now, 20, map[string]string{}, false, 0, "", false, 0, ""))

	replicas, err := repo.ListReplicas(ctx, primaryID)
	require.NoError(t, err)
	assert.Len(t, replicas, 1)
	assert.Equal(t, domain.RoleReplica, replicas[0].Role)
	assert.Equal(t, tenantID, replicas[0].TenantID)
}

func TestDatabaseRepository_Update(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewDatabaseRepository(mock)
	tenantID := uuid.New()
	db := &domain.Database{
		ID:                  uuid.New(),
		UserID:              uuid.New(),
		TenantID:            tenantID,
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

	mock.ExpectExec("UPDATE databases").
		WithArgs(db.Name, db.Status, db.Role, db.PrimaryID, db.ContainerID, db.Port, pgxmock.AnyArg(), db.Parameters, db.MetricsEnabled, db.MetricsPort, db.ExporterContainerID, db.PoolingEnabled, db.PoolingPort, db.PoolerContainerID, db.AllocatedStorage, db.ID, db.TenantID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err = repo.Update(context.Background(), db)
	require.NoError(t, err)
}

func TestDatabaseRepository_Delete(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewDatabaseRepository(mock)
	id := uuid.New()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := appcontext.WithTenantID(appcontext.WithUserID(context.Background(), userID), tenantID)

	mock.ExpectExec("DELETE FROM databases WHERE id = \\$1 AND tenant_id = \\$2").
		WithArgs(id, tenantID).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = repo.Delete(ctx, id)
	require.NoError(t, err)
}
