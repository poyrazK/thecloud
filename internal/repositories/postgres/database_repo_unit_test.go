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
		CredentialPath:      "secret/rds/db1",
	}

	mock.ExpectExec("INSERT INTO databases").
		WithArgs(db.ID, db.UserID, db.TenantID, db.Name, db.Engine, db.Version, db.Status, db.Role, db.PrimaryID, db.VpcID, db.ContainerID, db.Port, db.Username, db.Password, db.CreatedAt, db.UpdatedAt, db.AllocatedStorage, db.Parameters, db.MetricsEnabled, db.MetricsPort, db.ExporterContainerID, db.PoolingEnabled, db.PoolingPort, db.PoolerContainerID, db.CredentialPath).
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
	ctx := appcontext.WithTenantID(context.Background(), tenantID)
	now := time.Now()

	mock.ExpectQuery("SELECT id, user_id, tenant_id, name, engine, version, status, role, primary_id, vpc_id, COALESCE\\(container_id, ''\\), port, username, password, created_at, updated_at, allocated_storage, parameters, metrics_enabled, COALESCE\\(metrics_port, 0\\), COALESCE\\(exporter_container_id, ''\\), pooling_enabled, COALESCE\\(pooling_port, 0\\), COALESCE\\(pooler_container_id, ''\\), COALESCE\\(credential_path, ''\\) FROM databases").
		WithArgs(id, tenantID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "tenant_id", "name", "engine", "version", "status", "role", "primary_id", "vpc_id", "container_id", "port", "username", "password", "created_at", "updated_at", "allocated_storage", "parameters", "metrics_enabled", "metrics_port", "exporter_container_id", "pooling_enabled", "pooling_port", "pooler_container_id", "credential_path"}).
			AddRow(id, userID, tenantID, "test-db", string(domain.EnginePostgres), "16", string(domain.DatabaseStatusCreating), string(domain.RolePrimary), nil, nil, "cid-1", 5432, "admin", "password", now, now, 10, map[string]string{"k": "v"}, true, 9187, "exp-cid", true, 6432, "pool-cid", "secret/rds/db1"))

	db, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.NotNil(t, db)
	assert.Equal(t, id, db.ID)
	assert.Equal(t, domain.EnginePostgres, db.Engine)
	assert.Equal(t, 10, db.AllocatedStorage)
	assert.Equal(t, "v", db.Parameters["k"])
	assert.True(t, db.MetricsEnabled)
	assert.Equal(t, 9187, db.MetricsPort)
	assert.True(t, db.PoolingEnabled)
	assert.Equal(t, 6432, db.PoolingPort)
	assert.Equal(t, "secret/rds/db1", db.CredentialPath)
}

func TestDatabaseRepository_List(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewDatabaseRepository(mock)
	userID := uuid.New()
	tenantID := uuid.New()
	ctx := appcontext.WithTenantID(context.Background(), tenantID)
	now := time.Now()

	mock.ExpectQuery("SELECT id, user_id, tenant_id, name, engine, version, status, role, primary_id, vpc_id, COALESCE\\(container_id, ''\\), port, username, password, created_at, updated_at, allocated_storage, parameters, metrics_enabled, COALESCE\\(metrics_port, 0\\), COALESCE\\(exporter_container_id, ''\\), pooling_enabled, COALESCE\\(pooling_port, 0\\), COALESCE\\(pooler_container_id, ''\\), COALESCE\\(credential_path, ''\\) FROM databases").
		WithArgs(tenantID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "tenant_id", "name", "engine", "version", "status", "role", "primary_id", "vpc_id", "container_id", "port", "username", "password", "created_at", "updated_at", "allocated_storage", "parameters", "metrics_enabled", "metrics_port", "exporter_container_id", "pooling_enabled", "pooling_port", "pooler_container_id", "credential_path"}).
			AddRow(uuid.New(), userID, tenantID, "test-db", string(domain.EnginePostgres), "16", string(domain.DatabaseStatusCreating), string(domain.RolePrimary), nil, nil, "cid-1", 5432, "admin", "password", now, now, 20, map[string]string{}, false, 0, "", false, 0, "", ""))

	databases, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, databases, 1)
	assert.Equal(t, domain.EnginePostgres, databases[0].Engine)
	assert.Equal(t, 20, databases[0].AllocatedStorage)
}

func TestDatabaseRepository_ListReplicas(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewDatabaseRepository(mock)
	primaryID := uuid.New()
	tenantID := uuid.New()
	now := time.Now()

	mock.ExpectQuery("SELECT id, user_id, tenant_id, name, engine, version, status, role, primary_id, vpc_id, COALESCE\\(container_id, ''\\), port, username, password, created_at, updated_at, allocated_storage, parameters, metrics_enabled, COALESCE\\(metrics_port, 0\\), COALESCE\\(exporter_container_id, ''\\), pooling_enabled, COALESCE\\(pooling_port, 0\\), COALESCE\\(pooler_container_id, ''\\), COALESCE\\(credential_path, ''\\) FROM databases WHERE primary_id = \\$1").
		WithArgs(primaryID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "tenant_id", "name", "engine", "version", "status", "role", "primary_id", "vpc_id", "container_id", "port", "username", "password", "created_at", "updated_at", "allocated_storage", "parameters", "metrics_enabled", "metrics_port", "exporter_container_id", "pooling_enabled", "pooling_port", "pooler_container_id", "credential_path"}).
			AddRow(uuid.New(), uuid.New(), tenantID, "replica-1", string(domain.EnginePostgres), "16", string(domain.DatabaseStatusRunning), string(domain.RoleReplica), &primaryID, nil, "cid-2", 5432, "admin", "password", now, now, 20, map[string]string{}, false, 0, "", false, 0, "", "secret/rds/replica1"))

	replicas, err := repo.ListReplicas(context.Background(), primaryID)
	require.NoError(t, err)
	assert.Len(t, replicas, 1)
	assert.Equal(t, domain.RoleReplica, replicas[0].Role)
	assert.Equal(t, 20, replicas[0].AllocatedStorage)
	assert.Equal(t, "secret/rds/replica1", replicas[0].CredentialPath)
}

func TestDatabaseRepository_Update(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewDatabaseRepository(mock)
	db := &domain.Database{
		ID:                  uuid.New(),
		UserID:              uuid.New(),
		TenantID:            uuid.New(),
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
		AllocatedStorage:    30,
		CredentialPath:      "secret/rds/updated",
	}

	mock.ExpectExec("UPDATE databases").
		WithArgs(db.Name, db.Status, db.Role, db.PrimaryID, db.ContainerID, db.Port, pgxmock.AnyArg(), db.Parameters, db.MetricsEnabled, db.MetricsPort, db.ExporterContainerID, db.PoolingEnabled, db.PoolingPort, db.PoolerContainerID, db.AllocatedStorage, db.CredentialPath, db.ID, db.TenantID).
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
	tenantID := uuid.New()
	ctx := appcontext.WithTenantID(context.Background(), tenantID)

	mock.ExpectExec("DELETE FROM databases WHERE id = \\$1 AND tenant_id = \\$2").
		WithArgs(id, tenantID).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = repo.Delete(ctx, id)
	require.NoError(t, err)
}
