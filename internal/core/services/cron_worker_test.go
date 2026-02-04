package services_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/mock"
)

func TestCronWorkerProcessJobs(t *testing.T) {
	t.Parallel()
	repo := new(MockCronRepo)
	worker := services.NewCronWorker(repo)

	// Setup a test server to be the target
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()

	ctx := context.Background()
	jobID := uuid.New()
	job := &domain.CronJob{
		ID:           jobID,
		UserID:       uuid.New(),
		TargetMethod: "GET",
		TargetURL:    server.URL,
		Schedule:     "* * * * *", // Every minute
	}

	repo.On("GetNextJobsToRun", ctx).Return([]*domain.CronJob{job}, nil)

	repo.On("SaveJobRun", mock.Anything, mock.MatchedBy(func(run *domain.CronJobRun) bool {
		return run.JobID == jobID && run.Status == "SUCCESS" && run.StatusCode == 200
	})).Return(nil)

	repo.On("UpdateJob", mock.Anything, mock.MatchedBy(func(j *domain.CronJob) bool {
		return j.ID == jobID && j.LastRunAt != nil && j.NextRunAt != nil
	})).Return(nil)

	worker.ProcessJobs(ctx)

	// Wait briefly for goroutines to finish since runJob is called in a goroutine
	time.Sleep(100 * time.Millisecond)

	repo.AssertExpectations(t)
}

func TestCronWorkerProcessJobs_RequestFailure(t *testing.T) {
	t.Parallel()
	repo := new(MockCronRepo)
	worker := services.NewCronWorker(repo)

	// No server, request should fail
	ctx := context.Background()
	jobID := uuid.New()
	job := &domain.CronJob{
		ID:           jobID,
		UserID:       uuid.New(),
		TargetMethod: "GET",
		TargetURL:    "http://localhost:12345/unreachable",
		Schedule:     "* * * * *",
	}

	repo.On("GetNextJobsToRun", ctx).Return([]*domain.CronJob{job}, nil)

	repo.On("SaveJobRun", mock.Anything, mock.MatchedBy(func(run *domain.CronJobRun) bool {
		return run.JobID == jobID && run.Status == "FAILED"
	})).Return(nil)

	repo.On("UpdateJob", mock.Anything, mock.MatchedBy(func(j *domain.CronJob) bool {
		return j.ID == jobID
	})).Return(nil)

	worker.ProcessJobs(ctx)

	time.Sleep(100 * time.Millisecond)

	repo.AssertExpectations(t)
}

func TestCronWorkerProcessJobs_HTTPError(t *testing.T) {
	t.Parallel()
	repo := new(MockCronRepo)
	worker := services.NewCronWorker(repo)

	// Server returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	ctx := context.Background()
	jobID := uuid.New()
	job := &domain.CronJob{
		ID:           jobID,
		TargetMethod: "GET",
		TargetURL:    server.URL,
		Schedule:     "* * * * *",
	}

	repo.On("GetNextJobsToRun", ctx).Return([]*domain.CronJob{job}, nil)

	repo.On("SaveJobRun", mock.Anything, mock.MatchedBy(func(run *domain.CronJobRun) bool {
		return run.JobID == jobID && run.Status == "FAILED" && run.StatusCode == 500
	})).Return(nil)

	repo.On("UpdateJob", mock.Anything, mock.Anything).Return(nil)

	worker.ProcessJobs(ctx)

	time.Sleep(100 * time.Millisecond)

	repo.AssertExpectations(t)
}

func TestCronWorkerRun(t *testing.T) {
	t.Parallel()
	repo := new(MockCronRepo)
	worker := services.NewCronWorker(repo)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	cancel()
	worker.Run(ctx, &wg)
	wg.Wait()
}
