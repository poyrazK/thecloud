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
)

func TestContainerService_Unit(t *testing.T) {
	repo := new(MockContainerRepository)
	eventSvc := new(MockEventService)
	auditSvc := new(MockAuditService)
	svc := services.NewContainerService(repo, eventSvc, auditSvc)
	
	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)

	t.Run("CreateDeployment", func(t *testing.T) {
		repo.On("CreateDeployment", mock.Anything, mock.Anything).Return(nil).Once()
		eventSvc.On("RecordEvent", mock.Anything, "DEPLOYMENT_CREATED", mock.Anything, "DEPLOYMENT", mock.Anything).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "container.deployment_create", "deployment", mock.Anything, mock.Anything).Return(nil).Once()

		dep, err := svc.CreateDeployment(ctx, "test-dep", "nginx", 2, "80:80")
		assert.NoError(t, err)
		assert.NotNil(t, dep)
		repo.AssertExpectations(t)
	})

	t.Run("ListDeployments", func(t *testing.T) {
		expectedDeps := []*domain.Deployment{{ID: uuid.New(), Name: "dep1"}}
		repo.On("ListDeployments", mock.Anything, userID).Return(expectedDeps, nil).Once()

		deps, err := svc.ListDeployments(ctx)
		assert.NoError(t, err)
		assert.Len(t, deps, 1)
		assert.Equal(t, "dep1", deps[0].Name)
	})

	t.Run("GetDeployment", func(t *testing.T) {
		depID := uuid.New()
		expectedDep := &domain.Deployment{ID: depID, Name: "dep1"}
		repo.On("GetDeploymentByID", mock.Anything, depID, userID).Return(expectedDep, nil).Once()

		dep, err := svc.GetDeployment(ctx, depID)
		assert.NoError(t, err)
		assert.Equal(t, depID, dep.ID)
	})

	t.Run("ScaleDeployment", func(t *testing.T) {
		depID := uuid.New()
		dep := &domain.Deployment{ID: depID, UserID: userID, Replicas: 2}
		repo.On("GetDeploymentByID", mock.Anything, depID, userID).Return(dep, nil).Once()
		repo.On("UpdateDeployment", mock.Anything, mock.MatchedBy(func(d *domain.Deployment) bool {
			return d.Replicas == 5
		})).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "container.deployment_scale", "deployment", depID.String(), mock.Anything).Return(nil).Once()

		err := svc.ScaleDeployment(ctx, depID, 5)
		assert.NoError(t, err)
	})

	t.Run("DeleteDeployment", func(t *testing.T) {
		depID := uuid.New()
		dep := &domain.Deployment{ID: depID, UserID: userID}
		repo.On("GetDeploymentByID", mock.Anything, depID, userID).Return(dep, nil).Once()
		repo.On("UpdateDeployment", mock.Anything, mock.MatchedBy(func(d *domain.Deployment) bool {
			return d.Status == domain.DeploymentStatusDeleting
		})).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "container.deployment_delete", "deployment", depID.String(), mock.Anything).Return(nil).Once()

		err := svc.DeleteDeployment(ctx, depID)
		assert.NoError(t, err)
	})
}
