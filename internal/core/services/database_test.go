package services_test

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDatabaseRepo
type MockDatabaseRepo struct{ mock.Mock }

func (m *MockDatabaseRepo) Create(ctx context.Context, db *domain.Database) error {
	args := m.Called(ctx, db)
	return args.Error(0)
}
func (m *MockDatabaseRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Database, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Database), args.Error(1)
}
func (m *MockDatabaseRepo) List(ctx context.Context) ([]*domain.Database, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Database), args.Error(1)
}
func (m *MockDatabaseRepo) Update(ctx context.Context, db *domain.Database) error {
	args := m.Called(ctx, db)
	return args.Error(0)
}
func (m *MockDatabaseRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupDatabaseServiceTest(t *testing.T) (*MockDatabaseRepo, *MockComputeBackend, *MockVpcRepo, *MockEventService, *MockAuditService, ports.DatabaseService) {
	repo := new(MockDatabaseRepo)
	docker := new(MockComputeBackend)
	vpcRepo := new(MockVpcRepo)
	eventSvc := new(MockEventService)
	auditSvc := new(MockAuditService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := services.NewDatabaseService(repo, docker, vpcRepo, eventSvc, auditSvc, logger)
	return repo, docker, vpcRepo, eventSvc, auditSvc, svc
}

func TestCreateDatabase_Success(t *testing.T) {
	repo, docker, _, eventSvc, auditSvc, svc := setupDatabaseServiceTest(t)
	defer repo.AssertExpectations(t)
	defer docker.AssertExpectations(t)
	defer eventSvc.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	ctx := context.Background()
	name := "test-db"
	engine := "postgres"
	version := "16"

	docker.On("CreateInstance", ctx, mock.MatchedBy(func(name string) bool {
		return strings.HasPrefix(name, "cloud-db-")
	}), "postgres:16-alpine", []string{"0:5432"}, "", []string(nil), mock.Anything, []string(nil)).Return("cont-123", nil)
	docker.On("GetInstancePort", ctx, "cont-123", "5432").Return(54321, nil)
	repo.On("Create", ctx, mock.AnythingOfType("*domain.Database")).Return(nil)
	eventSvc.On("RecordEvent", ctx, "DATABASE_CREATE", mock.Anything, "DATABASE", mock.Anything).Return(nil)
	auditSvc.On("Log", ctx, mock.Anything, "database.create", "database", mock.Anything, mock.Anything).Return(nil)

	db, err := svc.CreateDatabase(ctx, name, engine, version, nil)

	assert.NoError(t, err)
	assert.NotNil(t, db)
	assert.Equal(t, name, db.Name)
	assert.Equal(t, domain.EnginePostgres, db.Engine)
	assert.Equal(t, 54321, db.Port)
	assert.Equal(t, "cont-123", db.ContainerID)
}

func TestDeleteDatabase_Success(t *testing.T) {
	repo, docker, _, eventSvc, auditSvc, svc := setupDatabaseServiceTest(t)
	defer repo.AssertExpectations(t)
	defer docker.AssertExpectations(t)
	defer eventSvc.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	ctx := context.Background()
	dbID := uuid.New()
	db := &domain.Database{
		ID:          dbID,
		Name:        "test-db",
		ContainerID: "cont-123",
	}

	repo.On("GetByID", ctx, dbID).Return(db, nil)
	docker.On("DeleteInstance", ctx, "cont-123").Return(nil)
	repo.On("Delete", ctx, dbID).Return(nil)
	eventSvc.On("RecordEvent", ctx, "DATABASE_DELETE", dbID.String(), "DATABASE", mock.Anything).Return(nil)
	auditSvc.On("Log", ctx, mock.Anything, "database.delete", "database", dbID.String(), mock.MatchedBy(func(details map[string]interface{}) bool {
		return details["name"] == "test-db"
	})).Return(nil)

	err := svc.DeleteDatabase(ctx, dbID)

	assert.NoError(t, err)
}

func TestGetDatabase_ByID(t *testing.T) {
	repo, _, _, _, _, svc := setupDatabaseServiceTest(t)
	defer repo.AssertExpectations(t)

	ctx := context.Background()
	dbID := uuid.New()
	db := &domain.Database{ID: dbID, Name: "my-db"}

	repo.On("GetByID", ctx, dbID).Return(db, nil)

	result, err := svc.GetDatabase(ctx, dbID) // Pass UUID directly

	assert.NoError(t, err)
	assert.Equal(t, dbID, result.ID)
}

func TestListDatabases(t *testing.T) {
	repo, _, _, _, _, svc := setupDatabaseServiceTest(t)
	defer repo.AssertExpectations(t)

	ctx := context.Background()
	dbs := []*domain.Database{{Name: "db1"}, {Name: "db2"}}

	repo.On("List", ctx).Return(dbs, nil)

	result, err := svc.ListDatabases(ctx)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestGetConnectionString(t *testing.T) {
	repo, _, _, _, _, svc := setupDatabaseServiceTest(t)
	defer repo.AssertExpectations(t)

	ctx := context.Background()
	dbID := uuid.New()
	db := &domain.Database{
		ID:       dbID,
		Name:     "conn-db",
		Engine:   domain.EnginePostgres,
		Port:     5432,
		Username: "admin",
		Password: "secret",
	}

	repo.On("GetByID", ctx, dbID).Return(db, nil)

	connStr, err := svc.GetConnectionString(ctx, dbID) // Correct method name

	assert.NoError(t, err)
	assert.Contains(t, connStr, "postgres://")
	assert.Contains(t, connStr, "admin:secret")
	assert.Contains(t, connStr, "5432")
}
