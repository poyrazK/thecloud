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

func TestDatabaseRepository_Create(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewDatabaseRepository(mock)
	db := &domain.Database{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		Name:        "test-db",
		Engine:      domain.EnginePostgres,
		Version:     "16",
		Status:      domain.DatabaseStatusCreating,
		VpcID:       nil,
		ContainerID: "cid-1",
		Port:        5432,
		Username:    "admin",
		Password:    "password",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mock.ExpectExec("INSERT INTO databases").
		WithArgs(db.ID, db.UserID, db.Name, db.Engine, db.Version, db.Status, db.VpcID, db.ContainerID, db.Port, db.Username, db.Password, db.CreatedAt, db.UpdatedAt).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.Create(context.Background(), db)
	assert.NoError(t, err)
}

func TestDatabaseRepository_GetByID(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewDatabaseRepository(mock)
	id := uuid.New()
	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	now := time.Now()

	mock.ExpectQuery("SELECT id, user_id, name, engine, version, status, vpc_id, COALESCE\\(container_id, ''\\), port, username, password, created_at, updated_at FROM databases").
		WithArgs(id, userID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "engine", "version", "status", "vpc_id", "container_id", "port", "username", "password", "created_at", "updated_at"}).
			AddRow(id, userID, "test-db", string(domain.EnginePostgres), "16", string(domain.DatabaseStatusCreating), nil, "cid-1", 5432, "admin", "password", now, now))

	db, err := repo.GetByID(ctx, id)
	assert.NoError(t, err)
	assert.NotNil(t, db)
	assert.Equal(t, id, db.ID)
	assert.Equal(t, domain.EnginePostgres, db.Engine)
}

func TestDatabaseRepository_List(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewDatabaseRepository(mock)
	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	now := time.Now()

	mock.ExpectQuery("SELECT id, user_id, name, engine, version, status, vpc_id, COALESCE\\(container_id, ''\\), port, username, password, created_at, updated_at FROM databases").
		WithArgs(userID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "engine", "version", "status", "vpc_id", "container_id", "port", "username", "password", "created_at", "updated_at"}).
			AddRow(uuid.New(), userID, "test-db", string(domain.EnginePostgres), "16", string(domain.DatabaseStatusCreating), nil, "cid-1", 5432, "admin", "password", now, now))

	databases, err := repo.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, databases, 1)
	assert.Equal(t, domain.EnginePostgres, databases[0].Engine)
}
