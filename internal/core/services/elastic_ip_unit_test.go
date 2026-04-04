package services_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestElasticIPService_AllocateIP(t *testing.T) {
	repo := new(MockElasticIPRepo)
	auditSvc := new(MockAuditService)
	rbacSvc := new(MockRBACService)
	rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	svc := services.NewElasticIPService(services.ElasticIPServiceParams{
		Repo: repo, AuditSvc: auditSvc, RBAC: rbacSvc,
	})

	ctx := appcontext.WithUserID(context.Background(), uuid.New())
	userID := appcontext.UserIDFromContext(ctx)

	repo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
	auditSvc.On("Log", mock.Anything, userID, "eip.allocate", "eip", mock.Anything, mock.Anything).Return(nil).Once()

	eip, err := svc.AllocateIP(ctx)
	require.NoError(t, err)
	assert.NotNil(t, eip)
	assert.Equal(t, userID, eip.UserID)
	repo.AssertExpectations(t)
}

func TestElasticIPService_ReleaseIP(t *testing.T) {
	repo := new(MockElasticIPRepo)
	auditSvc := new(MockAuditService)
	rbacSvc := new(MockRBACService)
	rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	svc := services.NewElasticIPService(services.ElasticIPServiceParams{
		Repo: repo, AuditSvc: auditSvc, RBAC: rbacSvc,
	})

	id := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), uuid.New())
	userID := appcontext.UserIDFromContext(ctx)

	repo.On("GetByID", mock.Anything, id).Return(&domain.ElasticIP{ID: id, UserID: userID, Status: domain.EIPStatusAllocated}, nil).Once()
	repo.On("Delete", mock.Anything, id).Return(nil).Once()
	auditSvc.On("Log", mock.Anything, userID, "eip.release", "eip", id.String(), mock.Anything).Return(nil).Once()

	err := svc.ReleaseIP(ctx, id)
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestElasticIPService_AssociateIP(t *testing.T) {
	repo := new(MockElasticIPRepo)
	auditSvc := new(MockAuditService)
	rbacSvc := new(MockRBACService)
	instRepo := new(MockInstanceRepo)
	rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	svc := services.NewElasticIPService(services.ElasticIPServiceParams{
		Repo: repo, AuditSvc: auditSvc, RBAC: rbacSvc, InstanceRepo: instRepo,
	})

	id := uuid.New()
	instID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), uuid.New())
	userID := appcontext.UserIDFromContext(ctx)
	eip := &domain.ElasticIP{ID: id, UserID: userID, Status: domain.EIPStatusAllocated}
	inst := &domain.Instance{ID: instID, UserID: userID, Status: domain.StatusRunning}

	t.Run("success", func(t *testing.T) {
		repo.On("GetByID", mock.Anything, id).Return(eip, nil).Once()
		instRepo.On("GetByID", mock.Anything, instID).Return(inst, nil).Once()
		repo.On("GetByInstanceID", mock.Anything, instID).Return(nil, nil).Once()
		repo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "eip.associate", "eip", id.String(), mock.Anything).Return(nil).Once()

		result, err := svc.AssociateIP(context.Background(), id, instID)
		require.NoError(t, err)
		assert.Equal(t, domain.EIPStatusAssociated, result.Status)
	})

	t.Run("instance deleted", func(t *testing.T) {
		instDeleted := &domain.Instance{ID: instID, Status: domain.StatusDeleted}
		repo.On("GetByID", mock.Anything, id).Return(eip, nil).Once()
		instRepo.On("GetByID", mock.Anything, instID).Return(instDeleted, nil).Once()

		_, err := svc.AssociateIP(context.Background(), id, instID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "deleted instance")
	})

	t.Run("already has eip", func(t *testing.T) {
		repo.On("GetByID", mock.Anything, id).Return(eip, nil).Once()
		instRepo.On("GetByID", mock.Anything, instID).Return(inst, nil).Once()
		repo.On("GetByInstanceID", mock.Anything, instID).Return(&domain.ElasticIP{ID: uuid.New()}, nil).Once()

		_, err := svc.AssociateIP(context.Background(), id, instID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already has an associated")
	})
}

func TestElasticIPService_DisassociateIP(t *testing.T) {
	repo := new(MockElasticIPRepo)
	auditSvc := new(MockAuditService)
	svc := services.NewElasticIPService(services.ElasticIPServiceParams{
		Repo: repo, AuditSvc: auditSvc, Logger: slog.Default(),
	})

	id := uuid.New()
	userID := uuid.New()
	instID := uuid.New()
	eip := &domain.ElasticIP{ID: id, UserID: userID, Status: domain.EIPStatusAssociated, InstanceID: &instID}

	t.Run("success", func(t *testing.T) {
		repo.On("GetByID", mock.Anything, id).Return(eip, nil).Once()
		repo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "eip.disassociate", "eip", id.String(), mock.Anything).Return(nil).Once()

		result, err := svc.DisassociateIP(context.Background(), id)
		require.NoError(t, err)
		assert.Equal(t, domain.EIPStatusAllocated, result.Status)
		assert.Nil(t, result.InstanceID)
	})

	t.Run("not associated", func(t *testing.T) {
		eipAlloc := &domain.ElasticIP{ID: id, Status: domain.EIPStatusAllocated}
		repo.On("GetByID", mock.Anything, id).Return(eipAlloc, nil).Once()

		_, err := svc.DisassociateIP(context.Background(), id)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not associated")
	})
}

func TestElasticIPService_ListAndGet(t *testing.T) {
	repo := new(MockElasticIPRepo)
	svc := services.NewElasticIPService(services.ElasticIPServiceParams{
		Repo: repo, Logger: slog.Default(),
	})

	id := uuid.New()
	repo.On("List", mock.Anything).Return([]*domain.ElasticIP{{ID: id}}, nil).Once()
	repo.On("GetByID", mock.Anything, id).Return(&domain.ElasticIP{ID: id}, nil).Once()

	list, err := svc.ListElasticIPs(context.Background())
	require.NoError(t, err)
	assert.Len(t, list, 1)

	eip, err := svc.GetElasticIP(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, id, eip.ID)
}
