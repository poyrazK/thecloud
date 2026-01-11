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

func setupCronServiceTest(_ *testing.T) (*MockCronRepository, *MockEventService, *MockAuditService, ports.CronService) {
	repo := new(MockCronRepository)
	eventSvc := new(MockEventService)
	auditSvc := new(MockAuditService)
	svc := services.NewCronService(repo, eventSvc, auditSvc)
	return repo, eventSvc, auditSvc, svc
}

func TestCreateJob_Success(t *testing.T) {
	repo, eventSvc, auditSvc, svc := setupCronServiceTest(t)
	defer repo.AssertExpectations(t)
	defer eventSvc.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)

	name := "test-job"
	schedule := "* * * * *" // Every minute
	targetURL := "http://example.com"
	targetMethod := "GET"
	targetPayload := ""

	repo.On("CreateJob", ctx, mock.MatchedBy(func(j *domain.CronJob) bool {
		return j.Name == name && j.Schedule == schedule && j.UserID == userID
	})).Return(nil)

	eventSvc.On("RecordEvent", ctx, "CRON_JOB_CREATED", mock.Anything, "CRON_JOB", mock.Anything).Return(nil)
	auditSvc.On("Log", ctx, userID, "cron.job_create", "cron_job", mock.Anything, mock.Anything).Return(nil)

	job, err := svc.CreateJob(ctx, name, schedule, targetURL, targetMethod, targetPayload)

	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, name, job.Name)
	assert.Equal(t, domain.CronStatusActive, job.Status)
	assert.NotNil(t, job.NextRunAt)
}

func TestPauseJob_Success(t *testing.T) {
	repo, _, _, svc := setupCronServiceTest(t)
	defer repo.AssertExpectations(t)

	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	jobID := uuid.New()

	existing := &domain.CronJob{
		ID:       jobID,
		UserID:   userID,
		Status:   domain.CronStatusActive,
		Schedule: "* * * * *",
	}

	repo.On("GetJobByID", ctx, jobID, userID).Return(existing, nil)
	repo.On("UpdateJob", ctx, mock.MatchedBy(func(j *domain.CronJob) bool {
		return j.Status == domain.CronStatusPaused && j.NextRunAt == nil
	})).Return(nil)

	err := svc.PauseJob(ctx, jobID)

	assert.NoError(t, err)
}

func TestResumeJob_Success(t *testing.T) {
	repo, _, _, svc := setupCronServiceTest(t)
	defer repo.AssertExpectations(t)

	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	jobID := uuid.New()

	existing := &domain.CronJob{
		ID:       jobID,
		UserID:   userID,
		Status:   domain.CronStatusPaused,
		Schedule: "@hourly",
	}

	repo.On("GetJobByID", ctx, jobID, userID).Return(existing, nil)
	repo.On("UpdateJob", ctx, mock.MatchedBy(func(j *domain.CronJob) bool {
		return j.Status == domain.CronStatusActive && j.NextRunAt != nil
	})).Return(nil)

	err := svc.ResumeJob(ctx, jobID)

	assert.NoError(t, err)
}

func TestCronWorker_DeleteJob(t *testing.T) {
	repo, _, auditSvc, svc := setupCronServiceTest(t)
	defer repo.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	jobID := uuid.New()

	existing := &domain.CronJob{ID: jobID, UserID: userID}

	repo.On("GetJobByID", ctx, jobID, userID).Return(existing, nil)
	repo.On("DeleteJob", ctx, jobID).Return(nil)
	auditSvc.On("Log", ctx, userID, "cron.job_delete", "cron_job", jobID.String(), mock.Anything).Return(nil)

	err := svc.DeleteJob(ctx, jobID)

	assert.NoError(t, err)
}

func TestCronService_ListJobs(t *testing.T) {
	repo, _, _, svc := setupCronServiceTest(t)
	defer repo.AssertExpectations(t)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	jobs := []*domain.CronJob{{ID: uuid.New(), UserID: userID}}

	repo.On("ListJobs", ctx, userID).Return(jobs, nil)

	res, err := svc.ListJobs(ctx)
	assert.NoError(t, err)
	assert.Equal(t, jobs, res)
}

func TestCronService_GetJob(t *testing.T) {
	repo, _, _, svc := setupCronServiceTest(t)
	defer repo.AssertExpectations(t)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	jobID := uuid.New()
	job := &domain.CronJob{ID: jobID, UserID: userID}

	repo.On("GetJobByID", ctx, jobID, userID).Return(job, nil)

	res, err := svc.GetJob(ctx, jobID)
	assert.NoError(t, err)
	assert.Equal(t, job, res)
}

func TestCronService_Errors(t *testing.T) {
	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	jobID := uuid.New()

	t.Run("Create_Unauthorized", func(t *testing.T) {
		_, _, _, svc := setupCronServiceTest(t)
		_, err := svc.CreateJob(context.Background(), "n", "* * * * *", "u", "m", "p")
		assert.Error(t, err)
	})

	t.Run("Create_InvalidSchedule", func(t *testing.T) {
		_, _, _, svc := setupCronServiceTest(t)
		_, err := svc.CreateJob(ctx, "n", "invalid", "u", "m", "p")
		assert.Error(t, err)
	})

	t.Run("List_Unauthorized", func(t *testing.T) {
		_, _, _, svc := setupCronServiceTest(t)
		_, err := svc.ListJobs(context.Background())
		assert.Error(t, err)
	})

	t.Run("Get_Unauthorized", func(t *testing.T) {
		_, _, _, svc := setupCronServiceTest(t)
		_, err := svc.GetJob(context.Background(), jobID)
		assert.Error(t, err)
	})

	t.Run("Resume_InvalidSchedule", func(t *testing.T) {
		repo, _, _, svc := setupCronServiceTest(t)
		repo.On("GetJobByID", ctx, jobID, userID).Return(&domain.CronJob{Schedule: "invalid"}, nil)
		err := svc.ResumeJob(ctx, jobID)
		assert.Error(t, err)
	})
}

func TestCronService_RepoErrors(t *testing.T) {
	ctx := appcontext.WithUserID(context.Background(), uuid.New())
	userID := appcontext.UserIDFromContext(ctx)
	jobID := uuid.New()

	t.Run("CreateJob", func(t *testing.T) {
		repo, _, _, svc := setupCronServiceTest(t)
		repo.On("CreateJob", ctx, mock.Anything).Return(assert.AnError)
		_, err := svc.CreateJob(ctx, "n", "* * * * *", "u", "m", "p")
		assert.Error(t, err)
	})

	t.Run("PauseJob_GetError", func(t *testing.T) {
		repo, _, _, svc := setupCronServiceTest(t)
		repo.On("GetJobByID", ctx, jobID, userID).Return(nil, assert.AnError)
		err := svc.PauseJob(ctx, jobID)
		assert.Error(t, err)
	})

	t.Run("ResumeJob_GetError", func(t *testing.T) {
		repo, _, _, svc := setupCronServiceTest(t)
		repo.On("GetJobByID", ctx, jobID, userID).Return(nil, assert.AnError)
		err := svc.ResumeJob(ctx, jobID)
		assert.Error(t, err)
	})

	t.Run("DeleteJob_GetError", func(t *testing.T) {
		repo, _, _, svc := setupCronServiceTest(t)
		repo.On("GetJobByID", ctx, jobID, userID).Return(nil, assert.AnError)
		err := svc.DeleteJob(ctx, jobID)
		assert.Error(t, err)
	})

	t.Run("DeleteJob_DeleteError", func(t *testing.T) {
		repo, _, _, svc := setupCronServiceTest(t)
		repo.On("GetJobByID", ctx, jobID, userID).Return(&domain.CronJob{ID: jobID}, nil)
		repo.On("DeleteJob", ctx, jobID).Return(assert.AnError)
		err := svc.DeleteJob(ctx, jobID)
		assert.Error(t, err)
	})
}
