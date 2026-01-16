package services_test

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/mock"
)

func TestContainerWorkerReconcile(t *testing.T) {
	ctx := context.Background()

	t.Run("Reconcile Scale Up", func(t *testing.T) {
		repo := new(MockContainerRepo)
		instSvc := new(MockInstanceService)
		eventSvc := new(MockEventService)
		worker := services.NewContainerWorker(repo, instSvc, eventSvc)

		depID := uuid.New()
		dep := &domain.Deployment{
			ID:       depID,
			Name:     "test-dep",
			Replicas: 2,
			Image:    "nginx",
			Status:   domain.DeploymentStatusReady,
			UserID:   uuid.New(),
		}

		// ListAllDeployments returns our deployment
		repo.On("ListAllDeployments", ctx).Return([]*domain.Deployment{dep}, nil)

		// GetContainers returns 0 containers initially
		repo.On("GetContainers", mock.Anything, depID).Return([]uuid.UUID{}, nil).Once()

		// LaunchInstance called twice
		inst1 := &domain.Instance{ID: uuid.New(), Name: "dep-inst-1"}
		inst2 := &domain.Instance{ID: uuid.New(), Name: "dep-inst-2"}
		instSvc.On("LaunchInstance", mock.Anything, mock.Anything, "nginx", "", (*uuid.UUID)(nil), (*uuid.UUID)(nil), []domain.VolumeAttachment(nil)).Return(inst1, nil).Once()
		instSvc.On("LaunchInstance", mock.Anything, mock.Anything, "nginx", "", (*uuid.UUID)(nil), (*uuid.UUID)(nil), []domain.VolumeAttachment(nil)).Return(inst2, nil).Once()

		// AddContainer called twice
		repo.On("AddContainer", mock.Anything, depID, inst1.ID).Return(nil)
		repo.On("AddContainer", mock.Anything, depID, inst2.ID).Return(nil)

		// UpdateDeployment called to update status/count
		repo.On("UpdateDeployment", mock.Anything, mock.MatchedBy(func(d *domain.Deployment) bool {
			return d.ID == depID // relax check as status might change
		})).Return(nil)

		worker.Reconcile(ctx)

		repo.AssertExpectations(t)
		instSvc.AssertExpectations(t)
	})

	t.Run("Reconcile Scale Down", func(t *testing.T) {
		repo := new(MockContainerRepo)
		instSvc := new(MockInstanceService)
		eventSvc := new(MockEventService)
		worker := services.NewContainerWorker(repo, instSvc, eventSvc)

		depID := uuid.New()
		userID := uuid.New()
		dep := &domain.Deployment{
			ID:       depID,
			Name:     "test-dep-down",
			Replicas: 1,
			Image:    "nginx",
			Status:   domain.DeploymentStatusScaling,
			UserID:   userID,
		}

		repo.On("ListAllDeployments", ctx).Return([]*domain.Deployment{dep}, nil)

		// GetContainers returns 3 containers (excess of 2)
		c1, c2, c3 := uuid.New(), uuid.New(), uuid.New()
		repo.On("GetContainers", mock.Anything, depID).Return([]uuid.UUID{c1, c2, c3}, nil).Once()

		// TerminateContainer called twice
		// RemoveContainer called twice
		repo.On("RemoveContainer", mock.Anything, depID, c1).Return(nil)
		instSvc.On("TerminateInstance", mock.Anything, c1.String()).Return(nil)

		repo.On("RemoveContainer", mock.Anything, depID, c2).Return(nil)
		instSvc.On("TerminateInstance", mock.Anything, c2.String()).Return(nil)

		repo.On("UpdateDeployment", mock.Anything, mock.Anything).Return(nil)

		worker.Reconcile(ctx)

		repo.AssertExpectations(t)
		instSvc.AssertExpectations(t)
	})

	t.Run("Reconcile Deleting", func(t *testing.T) {
		repo := new(MockContainerRepo)
		instSvc := new(MockInstanceService)
		eventSvc := new(MockEventService)
		worker := services.NewContainerWorker(repo, instSvc, eventSvc)

		depID := uuid.New()
		dep := &domain.Deployment{
			ID:     depID,
			Name:   "test-dep-del",
			Status: domain.DeploymentStatusDeleting,
			UserID: uuid.New(),
		}

		repo.On("ListAllDeployments", ctx).Return([]*domain.Deployment{dep}, nil)

		// Case 1: Has containers -> terminate them
		c1 := uuid.New()
		repo.On("GetContainers", mock.Anything, depID).Return([]uuid.UUID{c1}, nil).Once()
		repo.On("RemoveContainer", mock.Anything, depID, c1).Return(nil)
		instSvc.On("TerminateInstance", mock.Anything, c1.String()).Return(nil)

		worker.Reconcile(ctx)

		// Case 2: No containers -> delete deployment
		repo.On("ListAllDeployments", ctx).Return([]*domain.Deployment{dep}, nil)
		repo.On("GetContainers", mock.Anything, depID).Return([]uuid.UUID{}, nil).Once()
		repo.On("DeleteDeployment", mock.Anything, depID).Return(nil)

		worker.Reconcile(ctx)

		repo.AssertExpectations(t)
		instSvc.AssertExpectations(t)
	})
}

func TestContainerWorkerLaunchError(t *testing.T) {
	repo := new(MockContainerRepo)
	instSvc := new(MockInstanceService)
	eventSvc := new(MockEventService)
	worker := services.NewContainerWorker(repo, instSvc, eventSvc)
	ctx := context.Background()

	depID := uuid.New()
	dep := &domain.Deployment{
		ID:       depID,
		Name:     "test-dep-fail",
		Replicas: 1,
		Status:   domain.DeploymentStatusReady,
	}

	repo.On("ListAllDeployments", ctx).Return([]*domain.Deployment{dep}, nil)
	repo.On("GetContainers", mock.Anything, depID).Return([]uuid.UUID{}, nil) // Launch fails
	instSvc.On("LaunchInstance", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, context.DeadlineExceeded)

	// Since launch failed, replica count is 0 < 1, so status becomes SCALING
	repo.On("UpdateDeployment", mock.Anything, mock.MatchedBy(func(d *domain.Deployment) bool {
		return d.Status == domain.DeploymentStatusScaling
	})).Return(nil)

	worker.Reconcile(ctx)

	instSvc.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestContainerWorkerRun(t *testing.T) {
	repo := new(MockContainerRepo)
	instSvc := new(MockInstanceService)
	eventSvc := new(MockEventService)
	worker := services.NewContainerWorker(repo, instSvc, eventSvc)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	cancel()
	worker.Run(ctx, &wg)
	wg.Wait()
}
