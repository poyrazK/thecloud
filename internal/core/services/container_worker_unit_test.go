package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockInstanceSvc is a minimal mock of ports.InstanceService.
type mockInstanceSvc struct {
	mock.Mock
}

func (m *mockInstanceSvc) LaunchInstance(ctx context.Context, params ports.LaunchParams) (*domain.Instance, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Instance), args.Error(1)
}

func (m *mockInstanceSvc) LaunchInstanceWithOptions(ctx context.Context, opts ports.CreateInstanceOptions) (*domain.Instance, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Instance), args.Error(1)
}

func (m *mockInstanceSvc) StartInstance(ctx context.Context, idOrName string) error {
	args := m.Called(ctx, idOrName)
	return args.Error(0)
}

func (m *mockInstanceSvc) StopInstance(ctx context.Context, idOrName string) error {
	args := m.Called(ctx, idOrName)
	return args.Error(0)
}

func (m *mockInstanceSvc) TerminateInstance(ctx context.Context, idOrName string) error {
	args := m.Called(ctx, idOrName)
	return args.Error(0)
}

func (m *mockInstanceSvc) ListInstances(ctx context.Context) ([]*domain.Instance, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Instance), args.Error(1)
}

func (m *mockInstanceSvc) GetInstance(ctx context.Context, idOrName string) (*domain.Instance, error) {
	args := m.Called(ctx, idOrName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Instance), args.Error(1)
}

func (m *mockInstanceSvc) GetInstanceLogs(ctx context.Context, idOrName string) (string, error) {
	args := m.Called(ctx, idOrName)
	return args.String(0), args.Error(1)
}

func (m *mockInstanceSvc) GetInstanceStats(ctx context.Context, idOrName string) (*domain.InstanceStats, error) {
	args := m.Called(ctx, idOrName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.InstanceStats), args.Error(1)
}

func (m *mockInstanceSvc) GetConsoleURL(ctx context.Context, idOrName string) (string, error) {
	args := m.Called(ctx, idOrName)
	return args.String(0), args.Error(1)
}

func (m *mockInstanceSvc) Exec(ctx context.Context, idOrName string, cmd []string) (string, error) {
	args := m.Called(ctx, idOrName, cmd)
	return args.String(0), args.Error(1)
}

func (m *mockInstanceSvc) UpdateInstanceMetadata(ctx context.Context, id uuid.UUID, metadata, labels map[string]string) error {
	args := m.Called(ctx, id, metadata, labels)
	return args.Error(0)
}

// mockEventSvc is a minimal mock of ports.EventService.
type mockEventSvc struct {
	mock.Mock
}

func (m *mockEventSvc) RecordEvent(ctx context.Context, eType, resourceID, resourceType string, meta map[string]interface{}) error {
	args := m.Called(ctx, eType, resourceID, resourceType, meta)
	return args.Error(0)
}

func (m *mockEventSvc) ListEvents(ctx context.Context, limit int) ([]*domain.Event, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Event), args.Error(1)
}

// setupContainerWorkerTest creates a ContainerWorker with mock dependencies.
func setupContainerWorkerTest(t *testing.T) (*services.ContainerWorker, *MockContainerRepository, *mockInstanceSvc, *mockEventSvc) {
	t.Helper()
	repo := new(MockContainerRepository)
	instanceSvc := new(mockInstanceSvc)
	eventSvc := new(mockEventSvc)
	worker := services.NewContainerWorker(repo, instanceSvc, eventSvc)
	return worker, repo, instanceSvc, eventSvc
}

func TestContainerWorker_Unit(t *testing.T) {
	t.Run("Reconcile", testContainerWorkerReconcile)
	t.Run("Reconcile_ListDeploymentsError", testContainerWorkerReconcileListDeploymentsError)
}

func testContainerWorkerReconcile(t *testing.T) {
	worker, repo, instanceSvc, _ := setupContainerWorkerTest(t)
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
	worker, repo, _, _ := setupContainerWorkerTest(t)
	ctx := context.Background()

	repo.On("ListAllDeployments", mock.Anything).Return(nil, assert.AnError).Maybe()

	// Reconcile logs the error and returns gracefully.
	worker.Reconcile(ctx)
}

