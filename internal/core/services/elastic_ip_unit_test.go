package services_test

import (
	"context"
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

	repo.On("GetByID", mock.Anything, id).Return(&domain.ElasticIP{ID: id, UserID: userID, Status: domain.EIPStatusAllocated}, nil).Once()
	repo.On("GetByInstanceID", mock.Anything, instID).Return(nil, nil).Once()
	instRepo.On("GetByID", mock.Anything, instID).Return(&domain.Instance{ID: instID, UserID: userID}, nil).Once()
	repo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
	auditSvc.On("Log", mock.Anything, userID, "eip.associate", "eip", id.String(), mock.Anything).Return(nil).Once()

	_, err := svc.AssociateIP(ctx, id, instID)
	require.NoError(t, err)
	repo.AssertExpectations(t)
}
