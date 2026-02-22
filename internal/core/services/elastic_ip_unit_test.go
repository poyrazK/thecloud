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
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
)

func TestElasticIPService_AllocateIP(t *testing.T) {
	repo := new(MockElasticIPRepo)
	auditSvc := new(MockAuditService)
	svc := services.NewElasticIPService(services.ElasticIPServiceParams{
		Repo: repo, AuditSvc: auditSvc, Logger: slog.Default(),
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
	svc := services.NewElasticIPService(services.ElasticIPServiceParams{
		Repo: repo, AuditSvc: auditSvc, Logger: slog.Default(),
	})

	id := uuid.New()
	userID := uuid.New()
	eip := &domain.ElasticIP{ID: id, UserID: userID, Status: domain.EIPStatusAllocated}

	t.Run("success", func(t *testing.T) {
		repo.On("GetByID", mock.Anything, id).Return(eip, nil).Once()
		repo.On("Delete", mock.Anything, id).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "eip.release", "eip", id.String(), mock.Anything).Return(nil).Once()

		err := svc.ReleaseIP(context.Background(), id)
		require.NoError(t, err)
	})

	t.Run("associated failure", func(t *testing.T) {
		eipAssoc := &domain.ElasticIP{ID: id, Status: domain.EIPStatusAssociated}
		repo.On("GetByID", mock.Anything, id).Return(eipAssoc, nil).Once()

		err := svc.ReleaseIP(context.Background(), id)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "disassociate it first")
	})
}

func TestElasticIPService_AssociateIP(t *testing.T) {
	repo := new(MockElasticIPRepo)
	instRepo := new(MockInstanceRepo)
	auditSvc := new(MockAuditService)
	svc := services.NewElasticIPService(services.ElasticIPServiceParams{
		Repo: repo, InstanceRepo: instRepo, AuditSvc: auditSvc, Logger: slog.Default(),
	})

	id := uuid.New()
	instID := uuid.New()
	userID := uuid.New()
	eip := &domain.ElasticIP{ID: id, UserID: userID, Status: domain.EIPStatusAllocated}
	inst := &domain.Instance{ID: instID, Status: domain.StatusRunning}

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

	repo.On("GetByID", mock.Anything, id).Return(eip, nil).Once()
	repo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
	auditSvc.On("Log", mock.Anything, userID, "eip.disassociate", "eip", id.String(), mock.Anything).Return(nil).Once()

	result, err := svc.DisassociateIP(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, domain.EIPStatusAllocated, result.Status)
	assert.Nil(t, result.InstanceID)
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
