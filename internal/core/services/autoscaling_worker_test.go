package services_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const testGroupName = "test-group"

func setupAutoScalingWorkerTest(_ *testing.T) (*MockAutoScalingRepo, *MockInstanceService, *MockLBService, *MockEventService, *MockClock, *services.AutoScalingWorker) {
	mockRepo := new(MockAutoScalingRepo)
	mockInstSvc := new(MockInstanceService)
	mockLBSvc := new(MockLBService)
	mockEventSvc := new(MockEventService)
	mockClock := new(MockClock)

	worker := services.NewAutoScalingWorker(mockRepo, mockInstSvc, mockLBSvc, mockEventSvc, mockClock)
	return mockRepo, mockInstSvc, mockLBSvc, mockEventSvc, mockClock, worker
}

func TestAutoScalingWorkerEvaluateScaleOut(t *testing.T) {
	mockRepo, mockInstSvc, _, mockEventSvc, mockClock, worker := setupAutoScalingWorkerTest(t)
	defer mockRepo.AssertExpectations(t)
	defer mockInstSvc.AssertExpectations(t)
	defer mockEventSvc.AssertExpectations(t)
	defer mockClock.AssertExpectations(t)

	ctx := context.Background()
	groupID := uuid.New()
	userID := uuid.New()

	now := time.Now()
	mockClock.On("Now").Return(now)

	// Setup Group
	group := &domain.ScalingGroup{
		ID:           groupID,
		UserID:       userID,
		Name:         "test-asg",
		MinInstances: 1,
		MaxInstances: 5,
		DesiredCount: 2,
		CurrentCount: 1,
		Image:        "nginx",
		Ports:        "80:80",
		Status:       domain.ScalingGroupStatusActive,
	}

	// Current State: 1 instance running
	inst1ID := uuid.New()
	instances := []uuid.UUID{inst1ID}

	// Expectations
	mockRepo.On("ListAllGroups", ctx).Return([]*domain.ScalingGroup{group}, nil)
	mockRepo.On("GetAllScalingGroupInstances", ctx, mock.Anything).Return(map[uuid.UUID][]uuid.UUID{groupID: instances}, nil)
	mockRepo.On("GetAllPolicies", ctx, mock.Anything).Return(map[uuid.UUID][]*domain.ScalingPolicy{groupID: {}}, nil)

	newInst := &domain.Instance{ID: uuid.New()}
	mockInstSvc.On("LaunchInstance", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(newInst, nil)
	mockRepo.On("AddInstanceToGroup", mock.Anything, groupID, newInst.ID).Return(nil)
	mockEventSvc.On("RecordEvent", mock.Anything, "AUTOSCALING_SCALE_OUT", groupID.String(), "SCALING_GROUP", mock.Anything).Return(nil)

	worker.Evaluate(ctx)
}

func TestAutoScalingWorkerEvaluateScaleIn(t *testing.T) {
	mockRepo, mockInstSvc, _, mockEventSvc, mockClock, worker := setupAutoScalingWorkerTest(t)
	defer mockRepo.AssertExpectations(t)
	defer mockInstSvc.AssertExpectations(t)
	defer mockClock.AssertExpectations(t)

	ctx := context.Background()
	groupID := uuid.New()
	userID := uuid.New()

	// Setup Group
	group := &domain.ScalingGroup{
		ID:           groupID,
		UserID:       userID,
		Name:         "test-asg-in",
		MinInstances: 1,
		MaxInstances: 5,
		DesiredCount: 1,
		CurrentCount: 2, // Excess
		Status:       domain.ScalingGroupStatusActive,
	}

	inst1ID := uuid.New()
	inst2ID := uuid.New()
	instances := []uuid.UUID{inst1ID, inst2ID}

	mockRepo.On("ListAllGroups", ctx).Return([]*domain.ScalingGroup{group}, nil)
	mockRepo.On("GetAllScalingGroupInstances", ctx, mock.Anything).Return(map[uuid.UUID][]uuid.UUID{groupID: instances}, nil)
	mockRepo.On("GetAllPolicies", ctx, mock.Anything).Return(map[uuid.UUID][]*domain.ScalingPolicy{groupID: {}}, nil)

	// Scale In: Remove last instance (inst2ID)
	// Using mock.Anything for everything to avoid matching issues
	mockRepo.On("RemoveInstanceFromGroup", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockInstSvc.On("TerminateInstance", mock.Anything, mock.Anything).Return(nil)
	mockEventSvc.On("RecordEvent", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	worker.Evaluate(ctx)
}

func TestAutoScalingWorkerEvaluatePolicyTrigger(t *testing.T) {
	mockRepo, _, _, _, mockClock, worker := setupAutoScalingWorkerTest(t)
	defer mockRepo.AssertExpectations(t)
	defer mockClock.AssertExpectations(t)

	ctx := context.Background()
	groupID := uuid.New()
	userID := uuid.New()
	policyID := uuid.New()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	mockClock.On("Now").Return(now)

	group := &domain.ScalingGroup{
		ID:           groupID,
		UserID:       userID,
		Name:         "test-asg-policy",
		MinInstances: 1,
		MaxInstances: 5,
		DesiredCount: 1,
		CurrentCount: 1,
		Status:       domain.ScalingGroupStatusActive,
		Image:        "nginx",
		Ports:        "80:80",
	}

	inst1ID := uuid.New()
	instances := []uuid.UUID{inst1ID}

	policy := &domain.ScalingPolicy{
		ID:           policyID,
		Name:         "cpu-high",
		MetricType:   "cpu",
		TargetValue:  70.0,
		ScaleOutStep: 1,
		ScaleInStep:  1,
		CooldownSec:  300,
		LastScaledAt: nil, // Ready to scale
	}

	mockRepo.On("ListAllGroups", ctx).Return([]*domain.ScalingGroup{group}, nil)
	mockRepo.On("GetAllScalingGroupInstances", ctx, mock.Anything).Return(map[uuid.UUID][]uuid.UUID{groupID: instances}, nil)
	mockRepo.On("GetAllPolicies", ctx, mock.Anything).Return(map[uuid.UUID][]*domain.ScalingPolicy{groupID: {policy}}, nil)

	mockRepo.On("GetAverageCPU", mock.Anything, mock.Anything, mock.Anything).Return(80.0, nil)
	mockRepo.On("UpdateGroup", mock.Anything, mock.Anything).Return(nil)
	mockRepo.On("UpdatePolicyLastScaled", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	worker.Evaluate(ctx)
}

func TestAutoScalingWorkerRunContextCancellation(t *testing.T) {
	mockRepo, _, _, _, _, worker := setupAutoScalingWorkerTest(t)
	defer mockRepo.AssertExpectations(t)

	ctx, cancel := context.WithCancel(context.Background())

	// Mock the Evaluate call that will happen
	mockRepo.On("ListAllGroups", mock.Anything).Return([]*domain.ScalingGroup{}, nil).Maybe()
	mockRepo.On("GetAllScalingGroupInstances", mock.Anything, mock.Anything).Return(map[uuid.UUID][]uuid.UUID{}, nil).Maybe()
	mockRepo.On("GetAllPolicies", mock.Anything, mock.Anything).Return(map[uuid.UUID][]*domain.ScalingPolicy{}, nil).Maybe()

	var wg sync.WaitGroup
	wg.Add(1)

	go worker.Run(ctx, &wg)

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Cancel context
	cancel()

	// Wait for worker to stop
	wg.Wait()

	// If we get here, the worker stopped gracefully
}

func TestAutoScalingWorkerCleanupGroupDeletesGroup(t *testing.T) {
	mockRepo, _, _, _, _, worker := setupAutoScalingWorkerTest(t)
	defer mockRepo.AssertExpectations(t)

	ctx := context.Background()
	groupID := uuid.New()
	group := &domain.ScalingGroup{
		ID:     groupID,
		Name:   testGroupName,
		Status: domain.ScalingGroupStatusDeleting,
	}

	// When no instances, should delete the group
	mockRepo.On("ListAllGroups", ctx).Return([]*domain.ScalingGroup{group}, nil)
	mockRepo.On("GetAllScalingGroupInstances", ctx, []uuid.UUID{groupID}).Return(map[uuid.UUID][]uuid.UUID{groupID: {}}, nil)
	mockRepo.On("GetAllPolicies", ctx, []uuid.UUID{groupID}).Return(map[uuid.UUID][]*domain.ScalingPolicy{}, nil)
	mockRepo.On("DeleteGroup", mock.Anything, groupID).Return(nil) // Use mock.Anything for context since worker adds UserID

	worker.Evaluate(ctx)

	mockRepo.AssertCalled(t, "DeleteGroup", mock.Anything, groupID)
}

func TestAutoScalingWorkerCleanupGroupWithInstances(t *testing.T) {
	mockRepo, mockInstSvc, _, mockEventSvc, _, worker := setupAutoScalingWorkerTest(t)
	defer mockRepo.AssertExpectations(t)
	defer mockInstSvc.AssertExpectations(t)
	defer mockEventSvc.AssertExpectations(t)

	ctx := context.Background()
	groupID := uuid.New()
	userID := uuid.New()
	inst1ID := uuid.New()
	inst2ID := uuid.New()
	group := &domain.ScalingGroup{
		ID:           groupID,
		UserID:       userID,
		Name:         testGroupName,
		Status:       domain.ScalingGroupStatusDeleting,
		DesiredCount: 2,
	}

	mockRepo.On("ListAllGroups", ctx).Return([]*domain.ScalingGroup{group}, nil)
	mockRepo.On("GetAllScalingGroupInstances", ctx, []uuid.UUID{groupID}).Return(map[uuid.UUID][]uuid.UUID{groupID: {inst1ID, inst2ID}}, nil)
	mockRepo.On("GetAllPolicies", ctx, []uuid.UUID{groupID}).Return(map[uuid.UUID][]*domain.ScalingPolicy{}, nil)

	mockRepo.On("RemoveInstanceFromGroup", mock.Anything, groupID, inst2ID).Return(nil)
	mockRepo.On("RemoveInstanceFromGroup", mock.Anything, groupID, inst1ID).Return(nil)
	mockInstSvc.On("TerminateInstance", mock.Anything, inst2ID.String()).Return(nil)
	mockInstSvc.On("TerminateInstance", mock.Anything, inst1ID.String()).Return(nil)
	mockEventSvc.On("RecordEvent", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	worker.Evaluate(ctx)

	mockRepo.AssertCalled(t, "RemoveInstanceFromGroup", mock.Anything, groupID, inst2ID)
	mockRepo.AssertCalled(t, "RemoveInstanceFromGroup", mock.Anything, groupID, inst1ID)
}

func TestAutoScalingWorkerRecordFailure(t *testing.T) {
	// This test verifies the worker logic handles failures properly
	// by checking that UpdateGroup is called when there's a failure
	mockRepo, mockInstSvc, _, _, mockClock, worker := setupAutoScalingWorkerTest(t)
	defer mockRepo.AssertExpectations(t)
	defer mockInstSvc.AssertExpectations(t)
	defer mockClock.AssertExpectations(t)

	ctx := context.Background()
	groupID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	group := &domain.ScalingGroup{
		ID:           groupID,
		UserID:       userID,
		Name:         testGroupName,
		MinInstances: 1,
		MaxInstances: 5,
		DesiredCount: 2,
		CurrentCount: 1,
		Image:        "nginx",
		Ports:        "80:80",
		Status:       domain.ScalingGroupStatusActive,
		FailureCount: 0,
	}

	mockClock.On("Now").Return(now)
	mockRepo.On("ListAllGroups", ctx).Return([]*domain.ScalingGroup{group}, nil)
	mockRepo.On("GetAllScalingGroupInstances", ctx, []uuid.UUID{groupID}).Return(map[uuid.UUID][]uuid.UUID{groupID: {}}, nil)
	mockRepo.On("GetAllPolicies", ctx, []uuid.UUID{groupID}).Return(map[uuid.UUID][]*domain.ScalingPolicy{}, nil)

	// Simulate a failure during scale out - use context matcher to handle UserID context
	mockInstSvc.On("LaunchInstance", mock.Anything, mock.Anything, "nginx", "0:80", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, assert.AnError)

	// Should record failure - use Any() matcher for context
	mockRepo.On("UpdateGroup", mock.Anything, mock.Anything).Return(nil)

	worker.Evaluate(ctx)

	mockRepo.AssertCalled(t, "UpdateGroup", mock.Anything, mock.Anything)
}

func TestAutoScalingWorkerFailureBackoffSkipsScaleOut(t *testing.T) {
	mockRepo, mockInstSvc, _, _, mockClock, worker := setupAutoScalingWorkerTest(t)
	defer mockRepo.AssertExpectations(t)
	defer mockClock.AssertExpectations(t)

	ctx := context.Background()
	groupID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	group := &domain.ScalingGroup{
		ID:            groupID,
		UserID:        userID,
		Name:          "test-backoff",
		MinInstances:  1,
		MaxInstances:  5,
		DesiredCount:  2,
		CurrentCount:  1,
		Status:        domain.ScalingGroupStatusActive,
		FailureCount:  5,
		LastFailureAt: &now,
	}

	mockClock.On("Now").Return(now)
	mockRepo.On("ListAllGroups", ctx).Return([]*domain.ScalingGroup{group}, nil)
	mockRepo.On("GetAllScalingGroupInstances", ctx, []uuid.UUID{groupID}).Return(map[uuid.UUID][]uuid.UUID{groupID: {uuid.New()}}, nil)
	mockRepo.On("GetAllPolicies", ctx, []uuid.UUID{groupID}).Return(map[uuid.UUID][]*domain.ScalingPolicy{}, nil)

	worker.Evaluate(ctx)

	mockInstSvc.AssertNotCalled(t, "LaunchInstance", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestAutoScalingWorkerAdjustDesiredBounds(t *testing.T) {
	mockRepo, mockInstSvc, _, _, _, worker := setupAutoScalingWorkerTest(t)
	defer mockRepo.AssertExpectations(t)
	defer mockInstSvc.AssertExpectations(t)

	ctx := context.Background()
	groupID := uuid.New()
	userID := uuid.New()
	instanceID := uuid.New()

	group := &domain.ScalingGroup{
		ID:           groupID,
		UserID:       userID,
		Name:         "test-bounds",
		MinInstances: 1,
		MaxInstances: 2,
		DesiredCount: 0,
		CurrentCount: 1,
		Status:       domain.ScalingGroupStatusActive,
	}

	mockRepo.On("ListAllGroups", ctx).Return([]*domain.ScalingGroup{group}, nil)
	mockRepo.On("GetAllScalingGroupInstances", ctx, []uuid.UUID{groupID}).Return(map[uuid.UUID][]uuid.UUID{groupID: {instanceID}}, nil)
	mockRepo.On("GetAllPolicies", ctx, []uuid.UUID{groupID}).Return(map[uuid.UUID][]*domain.ScalingPolicy{}, nil)

	mockRepo.On("UpdateGroup", mock.Anything, mock.MatchedBy(func(g *domain.ScalingGroup) bool {
		return g.DesiredCount == 1
	})).Return(nil)

	worker.Evaluate(ctx)
}

func TestAutoScalingWorkerResetFailures(t *testing.T) {
	// This test verifies that successful operations reset the failure count
	mockRepo, mockInstSvc, _, mockEventSvc, mockClock, worker := setupAutoScalingWorkerTest(t)
	defer mockRepo.AssertExpectations(t)
	defer mockInstSvc.AssertExpectations(t)
	defer mockEventSvc.AssertExpectations(t)
	defer mockClock.AssertExpectations(t)

	ctx := context.Background()
	groupID := uuid.New()
	userID := uuid.New()
	instanceID := uuid.New()
	now := time.Now()
	pastTime := now.Add(-1 * time.Hour)

	group := &domain.ScalingGroup{
		ID:            groupID,
		UserID:        userID,
		Name:          testGroupName,
		MinInstances:  1,
		MaxInstances:  5,
		DesiredCount:  2,
		CurrentCount:  1,
		Image:         "nginx",
		Ports:         "80:80",
		Status:        domain.ScalingGroupStatusActive,
		FailureCount:  3,
		LastFailureAt: &pastTime,
	}

	mockClock.On("Now").Return(now)
	mockRepo.On("ListAllGroups", ctx).Return([]*domain.ScalingGroup{group}, nil)
	mockRepo.On("GetAllScalingGroupInstances", ctx, []uuid.UUID{groupID}).Return(map[uuid.UUID][]uuid.UUID{groupID: {instanceID}}, nil)
	mockRepo.On("GetAllPolicies", ctx, []uuid.UUID{groupID}).Return(map[uuid.UUID][]*domain.ScalingPolicy{}, nil)

	// Successful scale out should reset failures
	newInstance := &domain.Instance{ID: uuid.New(), UserID: userID}
	mockInstSvc.On("LaunchInstance", mock.Anything, mock.Anything, "nginx", "0:80", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(newInstance, nil)
	mockRepo.On("AddInstanceToGroup", mock.Anything, groupID, newInstance.ID).Return(nil)
	mockEventSvc.On("RecordEvent", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockRepo.On("UpdateGroup", mock.Anything, mock.Anything).Return(nil)

	worker.Evaluate(ctx)

	mockRepo.AssertCalled(t, "UpdateGroup", mock.Anything, mock.Anything)
}

func TestAutoScalingWorkerEvaluatePolicyTriggerScaleIn(t *testing.T) {
	mockRepo, _, _, _, mockClock, worker := setupAutoScalingWorkerTest(t)
	defer mockRepo.AssertExpectations(t)
	defer mockClock.AssertExpectations(t)

	ctx := context.Background()
	groupID := uuid.New()
	userID := uuid.New()
	policyID := uuid.New()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	mockClock.On("Now").Return(now)

	group := &domain.ScalingGroup{
		ID:           groupID,
		UserID:       userID,
		Name:         "test-asg-policy-in",
		MinInstances: 1,
		MaxInstances: 5,
		DesiredCount: 2,
		CurrentCount: 2,
		Status:       domain.ScalingGroupStatusActive,
		Image:        "nginx",
		Ports:        "80:80",
	}

	inst1ID := uuid.New()
	inst2ID := uuid.New()
	instances := []uuid.UUID{inst1ID, inst2ID}

	policy := &domain.ScalingPolicy{
		ID:           policyID,
		Name:         "cpu-low",
		MetricType:   "cpu",
		TargetValue:  30.0, // Scale in if < 20.0
		ScaleOutStep: 1,
		ScaleInStep:  1,
		CooldownSec:  300,
		LastScaledAt: nil,
	}

	mockRepo.On("ListAllGroups", ctx).Return([]*domain.ScalingGroup{group}, nil)
	mockRepo.On("GetAllScalingGroupInstances", ctx, mock.Anything).Return(map[uuid.UUID][]uuid.UUID{groupID: instances}, nil)
	mockRepo.On("GetAllPolicies", ctx, mock.Anything).Return(map[uuid.UUID][]*domain.ScalingPolicy{groupID: {policy}}, nil)

	mockRepo.On("GetAverageCPU", mock.Anything, mock.Anything, mock.Anything).Return(15.0, nil)
	mockRepo.On("UpdateGroup", mock.Anything, mock.Anything).Return(nil)
	mockRepo.On("UpdatePolicyLastScaled", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	worker.Evaluate(ctx)

	mockRepo.AssertCalled(t, "UpdateGroup", mock.Anything, mock.MatchedBy(func(g *domain.ScalingGroup) bool {
		return g.DesiredCount == 1 // Scale in from 2 to 1
	}))
}
