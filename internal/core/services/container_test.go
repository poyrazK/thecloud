package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupContainerServiceTest(_ *testing.T) (*MockContainerRepository, *MockEventService, *MockAuditService, ports.ContainerService) {
	repo := new(MockContainerRepository)
	eventSvc := new(MockEventService)
	auditSvc := new(MockAuditService)
	svc := services.NewContainerService(repo, eventSvc, auditSvc)
	return repo, eventSvc, auditSvc, svc
}

func TestContainerService_CreateDeployment(t *testing.T) {
	repo, eventSvc, auditSvc, svc := setupContainerServiceTest(t)
	defer repo.AssertExpectations(t)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)

	repo.On("CreateDeployment", ctx, mock.Anything).Return(nil)
	eventSvc.On("RecordEvent", ctx, "DEPLOYMENT_CREATED", mock.Anything, "DEPLOYMENT", mock.Anything).Return(nil)
	auditSvc.On("Log", ctx, userID, "container.deployment_create", "deployment", mock.Anything, mock.Anything).Return(nil)

	dep, err := svc.CreateDeployment(ctx, "test", "nginx", 3, "80:80")

	assert.NoError(t, err)
	assert.NotNil(t, dep)
	assert.Equal(t, "test", dep.Name)
	assert.Equal(t, 3, dep.Replicas)
}

func TestContainerService_ListDeployments(t *testing.T) {
	repo, _, _, svc := setupContainerServiceTest(t)
	defer repo.AssertExpectations(t)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	deps := []*domain.Deployment{{ID: uuid.New(), UserID: userID}}

	repo.On("ListDeployments", ctx, userID).Return(deps, nil)

	res, err := svc.ListDeployments(ctx)
	assert.NoError(t, err)
	assert.Equal(t, deps, res)
}

func TestContainerService_GetDeployment(t *testing.T) {
	repo, _, _, svc := setupContainerServiceTest(t)
	defer repo.AssertExpectations(t)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	depID := uuid.New()
	dep := &domain.Deployment{ID: depID, UserID: userID}

	repo.On("GetDeploymentByID", ctx, depID, userID).Return(dep, nil)

	res, err := svc.GetDeployment(ctx, depID)
	assert.NoError(t, err)
	assert.Equal(t, dep, res)
}

func TestContainerService_ScaleDeployment(t *testing.T) {
	repo, _, auditSvc, svc := setupContainerServiceTest(t)
	defer repo.AssertExpectations(t)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	depID := uuid.New()
	dep := &domain.Deployment{ID: depID, UserID: userID}

	repo.On("GetDeploymentByID", ctx, depID, userID).Return(dep, nil)
	repo.On("UpdateDeployment", ctx, mock.MatchedBy(func(d *domain.Deployment) bool {
		return d.Replicas == 5 && d.Status == domain.DeploymentStatusScaling
	})).Return(nil)
	auditSvc.On("Log", ctx, userID, "container.deployment_scale", "deployment", depID.String(), mock.Anything).Return(nil)

	err := svc.ScaleDeployment(ctx, depID, 5)
	assert.NoError(t, err)
}

func TestContainerService_DeleteDeployment(t *testing.T) {
	repo, _, auditSvc, svc := setupContainerServiceTest(t)
	defer repo.AssertExpectations(t)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	depID := uuid.New()
	dep := &domain.Deployment{ID: depID, UserID: userID}

	repo.On("GetDeploymentByID", ctx, depID, userID).Return(dep, nil)
	repo.On("UpdateDeployment", ctx, mock.MatchedBy(func(d *domain.Deployment) bool {
		return d.Status == domain.DeploymentStatusDeleting
	})).Return(nil)
	auditSvc.On("Log", ctx, userID, "container.deployment_delete", "deployment", depID.String(), mock.Anything).Return(nil)

	err := svc.DeleteDeployment(ctx, depID)
	assert.NoError(t, err)
}

func TestContainerService_Errors(t *testing.T) {
	ctx := appcontext.WithUserID(context.Background(), uuid.New())
	userID := appcontext.UserIDFromContext(ctx)
	depID := uuid.New()

	t.Run("Create_Unauthorized", func(t *testing.T) {
		_, _, _, svc := setupContainerServiceTest(t)
		_, err := svc.CreateDeployment(context.Background(), "n", "i", 1, "p")
		assert.Error(t, err)
	})

	t.Run("Create_RepoError", func(t *testing.T) {
		repo, _, _, svc := setupContainerServiceTest(t)
		repo.On("CreateDeployment", ctx, mock.Anything).Return(assert.AnError)
		_, err := svc.CreateDeployment(ctx, "n", "i", 1, "p")
		assert.Error(t, err)
	})

	t.Run("List_Unauthorized", func(t *testing.T) {
		_, _, _, svc := setupContainerServiceTest(t)
		_, err := svc.ListDeployments(context.Background())
		assert.Error(t, err)
	})

	t.Run("Get_Unauthorized", func(t *testing.T) {
		_, _, _, svc := setupContainerServiceTest(t)
		_, err := svc.GetDeployment(context.Background(), depID)
		assert.Error(t, err)
	})

	t.Run("Scale_GetError", func(t *testing.T) {
		repo, _, _, svc := setupContainerServiceTest(t)
		repo.On("GetDeploymentByID", ctx, depID, userID).Return(nil, assert.AnError)
		err := svc.ScaleDeployment(ctx, depID, 5)
		assert.Error(t, err)
	})

	t.Run("Delete_GetError", func(t *testing.T) {
		repo, _, _, svc := setupContainerServiceTest(t)
		repo.On("GetDeploymentByID", ctx, depID, userID).Return(nil, assert.AnError)
		err := svc.DeleteDeployment(ctx, depID)
		assert.Error(t, err)
	})
}

func TestContainerService_MoreRepoErrors(t *testing.T) {
	ctx := appcontext.WithUserID(context.Background(), uuid.New())
	userID := appcontext.UserIDFromContext(ctx)
	depID := uuid.New()
	dep := &domain.Deployment{ID: depID, UserID: userID}

	t.Run("Scale_UpdateError", func(t *testing.T) {
		repo, _, _, svc := setupContainerServiceTest(t)
		repo.On("GetDeploymentByID", ctx, depID, userID).Return(dep, nil)
		repo.On("UpdateDeployment", ctx, mock.Anything).Return(assert.AnError)
		err := svc.ScaleDeployment(ctx, depID, 5)
		assert.Error(t, err)
	})

	t.Run("Delete_UpdateError", func(t *testing.T) {
		repo, _, _, svc := setupContainerServiceTest(t)
		repo.On("GetDeploymentByID", ctx, depID, userID).Return(dep, nil)
		repo.On("UpdateDeployment", ctx, mock.Anything).Return(assert.AnError)
		err := svc.DeleteDeployment(ctx, depID)
		assert.Error(t, err)
	})
}
