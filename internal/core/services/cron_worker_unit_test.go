package services_test

import (
	"context"
	"testing"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupCronWorkerTest(t *testing.T) (*services.CronWorker, *MockCronRepository) {
	t.Helper()
	repo := new(MockCronRepository)
	worker := services.NewCronWorker(repo)
	return worker, repo
}

func TestCronWorker_Unit(t *testing.T) {
	t.Run("ProcessJobs_Empty", testCronWorkerProcessJobsEmpty)
	t.Run("ProcessJobs_ClaimError", testCronWorkerProcessJobsClaimError)
}

func testCronWorkerProcessJobsEmpty(t *testing.T) {
	worker, repo := setupCronWorkerTest(t)
	ctx := context.Background()

	repo.On("ClaimNextJobsToRun", mock.Anything, services.ClaimTimeout).Return([]*domain.CronJob{}, nil).Once()

	worker.ProcessJobs(ctx)

	repo.AssertExpectations(t)
}

func testCronWorkerProcessJobsClaimError(t *testing.T) {
	worker, repo := setupCronWorkerTest(t)
	ctx := context.Background()

	repo.On("ClaimNextJobsToRun", mock.Anything, services.ClaimTimeout).Return(nil, assert.AnError).Once()

	worker.ProcessJobs(ctx)

	repo.AssertExpectations(t)
}
