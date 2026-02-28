package services_test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
	r0, _ := args.Get(0).(*domain.Database)
	return r0, args.Error(1)
}
func (m *MockDatabaseRepo) List(ctx context.Context) ([]*domain.Database, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).([]*domain.Database)
	return r0, args.Error(1)
}
func (m *MockDatabaseRepo) ListReplicas(ctx context.Context, primaryID uuid.UUID) ([]*domain.Database, error) {
	args := m.Called(ctx, primaryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).([]*domain.Database)
	return r0, args.Error(1)
}
func (m *MockDatabaseRepo) Update(ctx context.Context, db *domain.Database) error {
	return m.Called(ctx, db).Error(0)
}
func (m *MockDatabaseRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func TestDatabaseServiceUnitExtended(t *testing.T) {
	mockRepo := new(MockDatabaseRepo)
	mockCompute := new(MockComputeBackend)
	mockVpcRepo := new(MockVpcRepo)
	mockEventSvc := new(MockEventService)
	mockAuditSvc := new(MockAuditService)
	mockVolumeSvc := new(MockVolumeService)

	svc := services.NewDatabaseService(services.DatabaseServiceParams{
		Repo:      mockRepo,
		Compute:   mockCompute,
		VpcRepo:   mockVpcRepo,
		VolumeSvc: mockVolumeSvc,
		EventSvc:  mockEventSvc,
		AuditSvc:  mockAuditSvc,
		Logger:    slog.Default(),
	})

	ctx := context.Background()

	t.Run("CreateDatabase_Success", func(t *testing.T) {
		mockVolumeSvc.On("CreateVolume", mock.Anything, mock.Anything, 20).
			Return(&domain.Volume{ID: uuid.New(), Name: "db-vol"}, nil).Once()
		mockCompute.On("LaunchInstanceWithOptions", mock.Anything, mock.Anything).
			Return("cid", []string{"30001:5432"}, nil).Once()
		mockCompute.On("GetInstanceIP", mock.Anything, "cid").Return("10.0.0.5", nil).Once()
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		mockEventSvc.On("RecordEvent", mock.Anything, "DATABASE_CREATE", mock.Anything, "DATABASE", mock.Anything).
			Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, mock.Anything, "database.create", "database", mock.Anything, mock.Anything).
			Return(nil).Once()

		db, err := svc.CreateDatabase(ctx, "test-db", "postgres", "16", nil, 20)
		require.NoError(t, err)
		assert.NotNil(t, db)
		assert.Equal(t, 30001, db.Port)
		assert.Equal(t, 20, db.AllocatedStorage)
	})

	t.Run("CreateReplica", func(t *testing.T) {
		primaryID := uuid.New()
		primary := &domain.Database{ID: primaryID, Engine: "postgres", Version: "16", Port: 5432, ContainerID: "primary-cid", AllocatedStorage: 20}
		mockRepo.On("GetByID", mock.Anything, primaryID).Return(primary, nil).Once()
		mockCompute.On("GetInstanceIP", mock.Anything, "primary-cid").Return("10.0.0.5", nil).Once()
		mockVolumeSvc.On("CreateVolume", mock.Anything, mock.Anything, 20).
			Return(&domain.Volume{ID: uuid.New(), Name: "db-replica-vol"}, nil).Once()
		mockCompute.On("LaunchInstanceWithOptions", mock.Anything, mock.Anything).
			Return("cid-rep", []string{"30002:5432"}, nil).Once()
		mockCompute.On("GetInstanceIP", mock.Anything, "cid-rep").Return("10.0.0.6", nil).Once()
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		mockEventSvc.On("RecordEvent", mock.Anything, "DATABASE_REPLICA_CREATE", mock.Anything, "DATABASE", mock.Anything).
			Return(nil).Once()

		replica, err := svc.CreateReplica(ctx, primaryID, "test-rep")
		require.NoError(t, err)
		assert.NotNil(t, replica)
		assert.Equal(t, domain.RoleReplica, replica.Role)
		assert.Equal(t, 20, replica.AllocatedStorage)
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
		require.NoError(t, err)
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
		require.NoError(t, err)
		assert.Contains(t, conn, "postgres://user:pass@localhost:5432/mydb")
	})

	t.Run("DeleteDatabase", func(t *testing.T) {
		dbID := uuid.New()
		db := &domain.Database{ID: dbID, ContainerID: "cid", Role: domain.RolePrimary}
		mockRepo.On("GetByID", mock.Anything, dbID).Return(db, nil).Once()
		mockCompute.On("DeleteInstance", mock.Anything, "cid").Return(nil).Once()

		volID := uuid.New()
		mockVolumeSvc.On("ListVolumes", mock.Anything).Return([]*domain.Volume{
			{ID: volID, Name: fmt.Sprintf("db-vol-%s", dbID.String()[:8])},
		}, nil).Once()
		mockVolumeSvc.On("DeleteVolume", mock.Anything, volID.String()).Return(nil).Once()

		mockRepo.On("Delete", mock.Anything, dbID).Return(nil).Once()
		mockEventSvc.On("RecordEvent", mock.Anything, "DATABASE_DELETE", dbID.String(), "DATABASE", mock.Anything).
			Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, mock.Anything, "database.delete", "database", dbID.String(), mock.Anything).
			Return(nil).Once()

		err := svc.DeleteDatabase(ctx, dbID)
		require.NoError(t, err)
	})

	t.Run("GetDatabase", func(t *testing.T) {
		dbID := uuid.New()
		expected := &domain.Database{ID: dbID, Name: "test-db"}
		mockRepo.On("GetByID", mock.Anything, dbID).Return(expected, nil).Once()

		result, err := svc.GetDatabase(ctx, dbID)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("ListDatabases", func(t *testing.T) {
		expected := []*domain.Database{{ID: uuid.New()}}
		mockRepo.On("List", mock.Anything).Return(expected, nil).Once()

		result, err := svc.ListDatabases(ctx)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("CreateDatabase_RollbackOnLaunchFailure", func(t *testing.T) {
		volID := uuid.New()
		mockVolumeSvc.On("CreateVolume", mock.Anything, mock.Anything, 10).
			Return(&domain.Volume{ID: volID, Name: "db-vol"}, nil).Once()
		
		mockCompute.On("LaunchInstanceWithOptions", mock.Anything, mock.Anything).
			Return("", nil, fmt.Errorf("launch failed")).Once()
		
		mockVolumeSvc.On("DeleteVolume", mock.Anything, volID.String()).Return(nil).Once()

		db, err := svc.CreateDatabase(ctx, "fail-db", "postgres", "16", nil, 10)
		require.Error(t, err)
		assert.Nil(t, db)
		assert.Contains(t, err.Error(), "launch failed")
	})

	t.Run("CreateDatabase_RollbackOnRepoFailure", func(t *testing.T) {
		volID := uuid.New()
		mockVolumeSvc.On("CreateVolume", mock.Anything, mock.Anything, 10).
			Return(&domain.Volume{ID: volID, Name: "db-vol"}, nil).Once()
		
		mockCompute.On("LaunchInstanceWithOptions", mock.Anything, mock.Anything).
			Return("cid-123", []string{"30001:5432"}, nil).Once()
		
		mockRepo.On("Create", mock.Anything, mock.Anything).
			Return(fmt.Errorf("repo failed")).Once()
		
		mockCompute.On("DeleteInstance", mock.Anything, "cid-123").Return(nil).Once()
		mockVolumeSvc.On("DeleteVolume", mock.Anything, volID.String()).Return(nil).Once()

		db, err := svc.CreateDatabase(ctx, "repo-fail-db", "postgres", "16", nil, 10)
		require.Error(t, err)
		assert.Nil(t, db)
		assert.Contains(t, err.Error(), "repo failed")
	})

	t.Run("CreateReplica_Failure_PrimaryNotFound", func(t *testing.T) {
		primaryID := uuid.New()
		mockRepo.On("GetByID", mock.Anything, primaryID).
			Return(nil, fmt.Errorf("not found")).Once()

		replica, err := svc.CreateReplica(ctx, primaryID, "fail-rep")
		require.Error(t, err)
		assert.Nil(t, replica)
	})
}
