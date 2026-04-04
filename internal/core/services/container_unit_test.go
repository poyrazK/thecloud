package services_test

import (
	"context"
	"fmt"
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

func TestContainerServiceUnit(t *testing.T) {
	repo := new(MockContainerRepository)
	eventSvc := new(MockEventService)
	auditSvc := new(MockAuditService)
	rbacSvc := new(MockRBACService)
	rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	svc := services.NewContainerService(repo, rbacSvc, eventSvc, auditSvc, slog.Default())

	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)

	t.Run("CreateDeployment", func(t *testing.T) {
		repo.On("CreateDeployment", mock.Anything, mock.Anything).Return(nil).Once()
		eventSvc.On("RecordEvent", mock.Anything, "DEPLOYMENT_CREATED", mock.Anything, "DEPLOYMENT", mock.Anything).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "container.deployment_create", "deployment", mock.Anything, mock.Anything).Return(nil).Once()

		dep, err := svc.CreateDeployment(ctx, "test-dep", "nginx", 2, "80:80")
		require.NoError(t, err)
		assert.NotNil(t, dep)
		repo.AssertExpectations(t)
	})

	t.Run("CreateDeployment_Unauthorized", func(t *testing.T) {
		emptyCtx := context.Background()
		_, err := svc.CreateDeployment(emptyCtx, "fail", "nginx", 1, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("ListDeployments", func(t *testing.T) {
		expectedDeps := []*domain.Deployment{{ID: uuid.New(), Name: "dep1"}}
		repo.On("ListDeployments", mock.Anything, userID).Return(expectedDeps, nil).Once()

		deps, err := svc.ListDeployments(ctx)
		require.NoError(t, err)
		assert.Len(t, deps, 1)
		assert.Equal(t, "dep1", deps[0].Name)
	})

	t.Run("GetDeployment", func(t *testing.T) {
		depID := uuid.New()
		expectedDep := &domain.Deployment{ID: depID, Name: "dep1"}
		repo.On("GetDeploymentByID", mock.Anything, depID, userID).Return(expectedDep, nil).Once()

		dep, err := svc.GetDeployment(ctx, depID)
		require.NoError(t, err)
		assert.Equal(t, depID, dep.ID)
	})

	t.Run("GetDeployment_Unauthorized", func(t *testing.T) {
		_, err := svc.GetDeployment(context.Background(), uuid.New())
		require.Error(t, err)
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
		require.NoError(t, err)
	})

	t.Run("ScaleDeployment_Error", func(t *testing.T) {
		depID := uuid.New()
		repo.On("GetDeploymentByID", mock.Anything, depID, userID).Return(nil, fmt.Errorf("not found")).Once()

		err := svc.ScaleDeployment(ctx, depID, 5)
		require.Error(t, err)
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
		require.NoError(t, err)
	})
}
