package services_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type databaseMocks struct {
	repo         *MockDatabaseRepo
	rbacSvc      *MockRBACService
	compute      *MockComputeBackend
	vpcRepo      *MockVpcRepo
	auditSvc     *MockAuditService
	volSvc       *MockVolumeService
	snapshotSvc  *MockSnapshotService
	snapshotRepo *MockSnapshotRepo
	eventSvc     *MockEventService
}

func setupDatabaseUnit(t *testing.T) (databaseMocks, ports.DatabaseService) {
	t.Helper()
	m := databaseMocks{
		repo:         new(MockDatabaseRepo),
		compute:      new(MockComputeBackend),
		rbacSvc:      new(MockRBACService),
		vpcRepo:      new(MockVpcRepo),
		auditSvc:     new(MockAuditService),
		volSvc:       new(MockVolumeService),
		snapshotSvc:  new(MockSnapshotService),
		snapshotRepo: new(MockSnapshotRepo),
		eventSvc:     new(MockEventService),
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	m.rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	svc := services.NewDatabaseService(services.DatabaseServiceParams{
		Repo:         m.repo,
		RBAC:         m.rbacSvc,
		Compute:      m.compute,
		VpcRepo:      m.vpcRepo,
		VolumeSvc:    m.volSvc,
		SnapshotSvc:  m.snapshotSvc,
		SnapshotRepo: m.snapshotRepo,
		EventSvc:     m.eventSvc,
		AuditSvc:     m.auditSvc,
		Logger:       logger,
	})
	return m, svc
}

func TestDatabaseServiceUnit(t *testing.T) {
	m, svc := setupDatabaseUnit(t)
	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)

	t.Run("CreateDatabase", func(t *testing.T) {
		vpcID := uuid.New()
		m.vpcRepo.On("GetByID", mock.Anything, vpcID).Return(&domain.VPC{ID: vpcID}, nil).Once()
		m.repo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		m.volSvc.On("CreateVolume", mock.Anything, mock.Anything, mock.Anything).Return(&domain.Volume{ID: uuid.New()}, nil).Once()
		m.compute.On("LaunchInstanceWithOptions", mock.Anything, mock.Anything).Return("cid", []string{"3306:3306"}, nil).Once()
		m.repo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
		m.auditSvc.On("Log", mock.Anything, userID, "database.create", "database", mock.Anything, mock.Anything).Return(nil).Once()
		m.eventSvc.On("RecordEvent", mock.Anything, "DATABASE_CREATE", mock.Anything, "DATABASE", mock.Anything).Return(nil).Once()

		req := ports.CreateDatabaseRequest{
			Name:             "test-db",
			Engine:           "mysql",
			Version:          "8.0",
			VpcID:            &vpcID,
			AllocatedStorage: 10,
		}
		db, err := svc.CreateDatabase(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, db)
		assert.Equal(t, "test-db", db.Name)
	})

	t.Run("GetDatabase", func(t *testing.T) {
		dbID := uuid.New()
		m.repo.On("GetByID", mock.Anything, dbID).Return(&domain.Database{ID: dbID}, nil).Once()
		res, err := svc.GetDatabase(ctx, dbID)
		require.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("ListDatabases", func(t *testing.T) {
		m.repo.On("List", mock.Anything).Return([]*domain.Database{}, nil).Once()
		res, err := svc.ListDatabases(ctx)

		require.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("DeleteDatabase", func(t *testing.T) {
		dbID := uuid.New()
		db := &domain.Database{ID: dbID, UserID: userID, ContainerID: "cid"}
		m.repo.On("GetByID", mock.Anything, dbID).Return(db, nil).Once()
		m.compute.On("StopInstance", mock.Anything, "cid").Return(nil).Once()
		m.compute.On("DeleteInstance", mock.Anything, "cid").Return(nil).Once()
		m.volSvc.On("ListVolumes", mock.Anything).Return([]*domain.Volume{}, nil).Once()
		m.repo.On("Delete", mock.Anything, dbID).Return(nil).Once()
		m.eventSvc.On("RecordEvent", mock.Anything, "DATABASE_DELETE", dbID.String(), "DATABASE", mock.Anything).Return(nil).Once()
		m.auditSvc.On("Log", mock.Anything, userID, "database.delete", "database", dbID.String(), mock.Anything).Return(nil).Once()

		err := svc.DeleteDatabase(ctx, dbID)
		require.NoError(t, err)
	})
}
