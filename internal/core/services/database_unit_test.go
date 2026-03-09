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

func setupDatabaseUnit(t *testing.T) (*MockDatabaseRepo, *MockRBACService, *MockComputeBackend, *MockVpcRepo, *MockAuditService, ports.DatabaseService) {
	t.Helper()
	repo := new(MockDatabaseRepo)
	compute := new(MockComputeBackend)
	rbacSvc := new(MockRBACService)
	vpcRepo := new(MockVpcRepo)
	auditSvc := new(MockAuditService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	svc := services.NewDatabaseService(services.DatabaseServiceParams{
		Repo:     repo,
		RBAC:     rbacSvc,
		Compute:  compute,
		VpcRepo:  vpcRepo,
		AuditSvc: auditSvc,
		Logger:   logger,
	})
	return repo, rbacSvc, compute, vpcRepo, auditSvc, svc
}

func TestDatabaseServiceUnit(t *testing.T) {
	repo, rbacSvc, compute, vpcRepo, auditSvc, svc := setupDatabaseUnit(t)
	_ = rbacSvc // Use it to avoid unused warning
	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)

	t.Run("CreateDatabase", func(t *testing.T) {
		vpcID := uuid.New()
		vpcRepo.On("GetByID", mock.Anything, vpcID).Return(&domain.VPC{ID: vpcID}, nil).Once()
		repo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		compute.On("LaunchInstanceWithOptions", mock.Anything, mock.Anything).Return("cid", []string{"3306:3306"}, nil).Once()
		repo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "database.create", "database", mock.Anything, mock.Anything).Return(nil).Once()

		req := ports.CreateDatabaseRequest{
			Name:    "test-db",
			Engine:  "mysql",
			Version: "8.0",
			VpcID:   &vpcID,
		}
		db, err := svc.CreateDatabase(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, db)
		assert.Equal(t, "test-db", db.Name)
	})

	t.Run("GetDatabase", func(t *testing.T) {
		dbID := uuid.New()
		repo.On("GetByID", mock.Anything, dbID).Return(&domain.Database{ID: dbID}, nil).Once()
		res, err := svc.GetDatabase(ctx, dbID)
		require.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("ListDatabases", func(t *testing.T) {
		repo.On("List", mock.Anything).Return([]*domain.Database{}, nil).Once()
		res, err := svc.ListDatabases(ctx)

		require.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("DeleteDatabase", func(t *testing.T) {
		dbID := uuid.New()
		db := &domain.Database{ID: dbID, UserID: userID, ContainerID: "cid"}
		repo.On("GetByID", mock.Anything, dbID).Return(db, nil).Once()
		compute.On("StopInstance", mock.Anything, "cid").Return(nil).Once()
		compute.On("DeleteInstance", mock.Anything, "cid").Return(nil).Once()
		repo.On("Delete", mock.Anything, dbID).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "database.delete", "database", dbID.String(), mock.Anything).Return(nil).Once()

		err := svc.DeleteDatabase(ctx, dbID)
		require.NoError(t, err)
	})
}
