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

const (
	testDBName    = "test-db"
	dbContainerID = "cont-123"
)

// MockDatabaseRepo
type MockDatabaseRepo struct{ mock.Mock }

func (m *MockDatabaseRepo) Create(ctx context.Context, db *domain.Database) error {
	return m.Called(ctx, db).Error(0)
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
	return m.Called(ctx, db).Error(0)
}
func (m *MockDatabaseRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupDatabaseServiceTest(_ *testing.T) (*MockDatabaseRepo, *MockComputeBackend, *MockVpcRepo, *MockEventService, *MockAuditService, ports.DatabaseService) {
	repo := new(MockDatabaseRepo)
	docker := new(MockComputeBackend)
	vpcRepo := new(MockVpcRepo)
	eventSvc := new(MockEventService)
	auditSvc := new(MockAuditService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := services.NewDatabaseService(repo, docker, vpcRepo, eventSvc, auditSvc, logger)
	return repo, docker, vpcRepo, eventSvc, auditSvc, svc
}

func TestCreateDatabaseSuccess(t *testing.T) {
	repo, docker, _, eventSvc, auditSvc, svc := setupDatabaseServiceTest(t)
	defer repo.AssertExpectations(t)
	defer docker.AssertExpectations(t)
	defer eventSvc.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	ctx := context.Background()
	name := testDBName
	engine := "postgres"
	version := "16"

	docker.On("CreateInstance", ctx, mock.MatchedBy(func(opts ports.CreateInstanceOptions) bool {
		return strings.HasPrefix(opts.Name, "cloud-db-") &&
			opts.ImageName == "postgres:16-alpine" &&
			len(opts.Ports) == 1 && opts.Ports[0] == "0:5432"
	})).Return(dbContainerID, nil)
	docker.On("GetInstancePort", ctx, dbContainerID, "5432").Return(54321, nil)
	repo.On("Create", ctx, mock.AnythingOfType("*domain.Database")).Return(nil)
	eventSvc.On("RecordEvent", ctx, "DATABASE_CREATE", mock.Anything, "DATABASE", mock.Anything).Return(nil)
	auditSvc.On("Log", ctx, mock.Anything, "database.create", "database", mock.Anything, mock.Anything).Return(nil)

	db, err := svc.CreateDatabase(ctx, name, engine, version, nil)

	assert.NoError(t, err)
	assert.NotNil(t, db)
	assert.Equal(t, name, db.Name)
	assert.Equal(t, domain.EnginePostgres, db.Engine)
	assert.Equal(t, 54321, db.Port)
	assert.Equal(t, dbContainerID, db.ContainerID)
}

func TestCreateDatabaseWithVpc(t *testing.T) {
	repo, docker, vpcRepo, eventSvc, auditSvc, svc := setupDatabaseServiceTest(t)
	defer repo.AssertExpectations(t)
	defer docker.AssertExpectations(t)
	defer vpcRepo.AssertExpectations(t)
	defer eventSvc.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	ctx := context.Background()
	name := testDBName
	engine := "postgres"
	version := "16"
	vpcID := uuid.New()

	vpcRepo.On("GetByID", ctx, vpcID).Return(&domain.VPC{ID: vpcID, NetworkID: "net-1"}, nil)
	docker.On("CreateInstance", ctx, mock.MatchedBy(func(opts ports.CreateInstanceOptions) bool {
		return opts.NetworkID == "net-1" && opts.ImageName == "postgres:16-alpine"
	})).Return(dbContainerID, nil)
	docker.On("GetInstancePort", ctx, dbContainerID, "5432").Return(54321, nil)
	repo.On("Create", ctx, mock.AnythingOfType("*domain.Database")).Return(nil)
	eventSvc.On("RecordEvent", ctx, "DATABASE_CREATE", mock.Anything, "DATABASE", mock.Anything).Return(nil)
	auditSvc.On("Log", ctx, mock.Anything, "database.create", "database", mock.Anything, mock.Anything).Return(nil)

	_, err := svc.CreateDatabase(ctx, name, engine, version, &vpcID)
	assert.NoError(t, err)
}

func TestCreateDatabaseVpcNotFound(t *testing.T) {
	_, _, vpcRepo, _, _, svc := setupDatabaseServiceTest(t)
	ctx := context.Background()
	vpcID := uuid.New()

	vpcRepo.On("GetByID", ctx, vpcID).Return(nil, assert.AnError)

	_, err := svc.CreateDatabase(ctx, testDBName, "postgres", "16", &vpcID)
	assert.Error(t, err)
}

func TestCreateDatabaseInvalidEngine(t *testing.T) {
	_, _, _, _, _, svc := setupDatabaseServiceTest(t)
	ctx := context.Background()

	_, err := svc.CreateDatabase(ctx, testDBName, "oracle", "16", nil)
	assert.Error(t, err)
}

func TestCreateDatabaseMySQL(t *testing.T) {
	repo, docker, _, eventSvc, auditSvc, svc := setupDatabaseServiceTest(t)
	defer repo.AssertExpectations(t)
	defer docker.AssertExpectations(t)
	defer eventSvc.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	ctx := context.Background()
	name := "mysql-db"
	engine := "mysql"
	version := "8.0"

	docker.On("CreateInstance", ctx, mock.MatchedBy(func(opts ports.CreateInstanceOptions) bool {
		return opts.ImageName == "mysql:8.0" && len(opts.Ports) == 1 && opts.Ports[0] == "0:3306"
	})).Return(dbContainerID, nil)
	docker.On("GetInstancePort", ctx, dbContainerID, "3306").Return(33060, nil)
	repo.On("Create", ctx, mock.AnythingOfType("*domain.Database")).Return(nil)
	eventSvc.On("RecordEvent", ctx, "DATABASE_CREATE", mock.Anything, "DATABASE", mock.Anything).Return(nil)
	auditSvc.On("Log", ctx, mock.Anything, "database.create", "database", mock.Anything, mock.Anything).Return(nil)

	db, err := svc.CreateDatabase(ctx, name, engine, version, nil)
	assert.NoError(t, err)
	assert.Equal(t, domain.EngineMySQL, db.Engine)
}

func TestDeleteDatabaseSuccess(t *testing.T) {
	repo, docker, _, eventSvc, auditSvc, svc := setupDatabaseServiceTest(t)
	defer repo.AssertExpectations(t)
	defer docker.AssertExpectations(t)
	defer eventSvc.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	ctx := context.Background()
	dbID := uuid.New()
	db := &domain.Database{
		ID:          dbID,
		Name:        testDBName,
		ContainerID: dbContainerID,
	}

	repo.On("GetByID", ctx, dbID).Return(db, nil)
	docker.On("DeleteInstance", ctx, dbContainerID).Return(nil)
	repo.On("Delete", ctx, dbID).Return(nil)
	eventSvc.On("RecordEvent", ctx, "DATABASE_DELETE", dbID.String(), "DATABASE", mock.Anything).Return(nil)
	auditSvc.On("Log", ctx, mock.Anything, "database.delete", "database", dbID.String(), mock.MatchedBy(func(details map[string]interface{}) bool {
		return details["name"] == testDBName
	})).Return(nil)

	err := svc.DeleteDatabase(ctx, dbID)

	assert.NoError(t, err)
}

func TestGetDatabaseByID(t *testing.T) {
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
