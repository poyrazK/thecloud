package services_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDatabaseRepo struct {
	mock.Mock
}

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
func (m *MockDatabaseRepo) ListReplicas(ctx context.Context, primaryID uuid.UUID) ([]*domain.Database, error) {
	args := m.Called(ctx, primaryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Database), args.Error(1)
}
func (m *MockDatabaseRepo) Update(ctx context.Context, db *domain.Database) error {
	return m.Called(ctx, db).Error(0)
}
func (m *MockDatabaseRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func TestDatabaseService_Unit_Extended(t *testing.T) {
	mockRepo := new(MockDatabaseRepo)
	mockCompute := new(MockComputeBackend)
	mockVpcRepo := new(MockVpcRepo)
	mockEventSvc := new(MockEventService)
	mockAuditSvc := new(MockAuditService)
	
	svc := services.NewDatabaseService(services.DatabaseServiceParams{
		Repo:     mockRepo,
		Compute:  mockCompute,
		VpcRepo:  mockVpcRepo,
		EventSvc: mockEventSvc,
		AuditSvc: mockAuditSvc,
		Logger:   slog.Default(),
	})

	ctx := context.Background()

	t.Run("CreateDatabase_Success", func(t *testing.T) {
		mockCompute.On("LaunchInstanceWithOptions", mock.Anything, mock.Anything).
			Return("cid", []string{"30001:5432"}, nil).Once()
		mockCompute.On("GetInstanceIP", mock.Anything, "cid").Return("10.0.0.5", nil).Once()
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		mockEventSvc.On("RecordEvent", mock.Anything, "DATABASE_CREATE", mock.Anything, "DATABASE", mock.Anything).
			Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, mock.Anything, "database.create", "database", mock.Anything, mock.Anything).
			Return(nil).Once()

		db, err := svc.CreateDatabase(ctx, "test-db", "postgres", "16", nil)
		assert.NoError(t, err)
		assert.NotNil(t, db)
		assert.Equal(t, 30001, db.Port)
	})

	t.Run("CreateReplica", func(t *testing.T) {
		primaryID := uuid.New()
		primary := &domain.Database{ID: primaryID, Engine: "postgres", Version: "16", Port: 5432, ContainerID: "primary-cid"}
		mockRepo.On("GetByID", mock.Anything, primaryID).Return(primary, nil).Once()
		mockCompute.On("GetInstanceIP", mock.Anything, "primary-cid").Return("10.0.0.5", nil).Once()
		mockCompute.On("LaunchInstanceWithOptions", mock.Anything, mock.Anything).
			Return("cid-rep", []string{"30002:5432"}, nil).Once()
		mockCompute.On("GetInstanceIP", mock.Anything, "cid-rep").Return("10.0.0.6", nil).Once()
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		mockEventSvc.On("RecordEvent", mock.Anything, "DATABASE_REPLICA_CREATE", mock.Anything, "DATABASE", mock.Anything).
			Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, mock.Anything, "database.create_replica", "database", mock.Anything, mock.Anything).
			Return(nil).Once()

		replica, err := svc.CreateReplica(ctx, primaryID, "test-rep")
		assert.NoError(t, err)
		assert.NotNil(t, replica)
		assert.Equal(t, domain.RoleReplica, replica.Role)
	})

	t.Run("PromoteToPrimary", func(t *testing.T) {
		dbID := uuid.New()
		db := &domain.Database{ID: dbID, Role: domain.RoleReplica}
		mockRepo.On("GetByID", mock.Anything, dbID).Return(db, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(d *domain.Database) bool {
			return d.Role == domain.RolePrimary
		})).Return(nil).Once()
		mockEventSvc.On("RecordEvent", mock.Anything, "DATABASE_PROMOTED", dbID.String(), "DATABASE", mock.Anything).
			Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, mock.Anything, "database.promote", "database", dbID.String(), mock.Anything).
			Return(nil).Once()

		err := svc.PromoteToPrimary(ctx, dbID)
		assert.NoError(t, err)
	})
	
	t.Run("GetConnectionString", func(t *testing.T) {
		dbID := uuid.New()
		db := &domain.Database{
			ID:       dbID,
			Engine:   domain.EnginePostgres,
			Username: "user",
			Password: "pass",
			Port:     5432,
			Name:     "mydb",
		}
		mockRepo.On("GetByID", mock.Anything, dbID).Return(db, nil).Once()
		
		conn, err := svc.GetConnectionString(ctx, dbID)
		assert.NoError(t, err)
		assert.Contains(t, conn, "postgres://user:pass@localhost:5432/mydb")
	})

	t.Run("DeleteDatabase", func(t *testing.T) {
		dbID := uuid.New()
		db := &domain.Database{ID: dbID, ContainerID: "cid"}
		mockRepo.On("GetByID", mock.Anything, dbID).Return(db, nil).Once()
		mockCompute.On("DeleteInstance", mock.Anything, "cid").Return(nil).Once()
		mockRepo.On("Delete", mock.Anything, dbID).Return(nil).Once()
		mockEventSvc.On("RecordEvent", mock.Anything, "DATABASE_DELETE", dbID.String(), "DATABASE", mock.Anything).
			Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, mock.Anything, "database.delete", "database", dbID.String(), mock.Anything).
			Return(nil).Once()

		err := svc.DeleteDatabase(ctx, dbID)
		assert.NoError(t, err)
	})
}
