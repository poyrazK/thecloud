package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockInstanceSvc reuses MockInstanceService to avoid duplicating mock behavior.
type mockInstanceSvc = MockInstanceService

// setupContainerWorkerTest creates a ContainerWorker with mock dependencies.
func setupContainerWorkerTest(t *testing.T) (*services.ContainerWorker, *MockContainerRepository, *MockInstanceService) {
	t.Helper()
	repo := new(MockContainerRepository)
	instanceSvc := new(mockInstanceSvc)
	eventSvc := new(MockEventService)
	worker := services.NewContainerWorker(repo, instanceSvc, eventSvc)
	return worker, repo, instanceSvc
}

func TestContainerWorker_Unit(t *testing.T) {
	t.Run("Reconcile", testContainerWorkerReconcile)
	t.Run("Reconcile_ListDeploymentsError", testContainerWorkerReconcileListDeploymentsError)
}

func testContainerWorkerReconcile(t *testing.T) {
	worker, repo, instanceSvc := setupContainerWorkerTest(t)
	ctx := context.Background()
	userID := uuid.New()
	depID := uuid.New()
	instID := uuid.New()

	dep := &domain.Deployment{
		ID:           depID,
		UserID:       userID,
		Name:         "test-deployment",
		Status:       domain.DeploymentStatusReady,
		Replicas:     1,
		CurrentCount: 1,
		Image:        "nginx:latest",
		InstanceType: "t2.micro",
	}
	inst := &domain.Instance{ID: instID, Status: domain.StatusRunning}

	repo.On("ListAllDeployments", mock.Anything).Return([]*domain.Deployment{dep}, nil).Once()
	repo.On("GetContainers", mock.Anything, depID).Return([]uuid.UUID{instID}, nil).Once()
	instanceSvc.On("GetInstance", mock.Anything, instID.String()).Return(inst, nil).Once()

	worker.Reconcile(ctx)

	repo.AssertExpectations(t)
	instanceSvc.AssertExpectations(t)
}

func testContainerWorkerReconcileListDeploymentsError(t *testing.T) {
	worker, repo, _ := setupContainerWorkerTest(t)
	ctx := context.Background()

	repo.On("ListAllDeployments", mock.Anything).Return(nil, assert.AnError).Once()

	worker.Reconcile(ctx)

	repo.AssertExpectations(t)
}
