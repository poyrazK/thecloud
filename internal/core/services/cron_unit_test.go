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
	"github.com/stretchr/testify/require"
)

func TestCronServiceUnit(t *testing.T) {
	repo := new(MockCronRepository)
	eventSvc := new(MockEventService)
	auditSvc := new(MockAuditService)
	svc := services.NewCronService(repo, eventSvc, auditSvc)

	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)

	t.Run("CreateJob", func(t *testing.T) {
		repo.On("CreateJob", mock.Anything, mock.Anything).Return(nil).Once()
		eventSvc.On("RecordEvent", mock.Anything, "CRON_JOB_CREATED", mock.Anything, "CRON_JOB", mock.Anything).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "cron.job_create", "cron_job", mock.Anything, mock.Anything).Return(nil).Once()

		job, err := svc.CreateJob(ctx, "test-job", "* * * * *", "http://example.com", "GET", "")
		require.NoError(t, err)
		assert.NotNil(t, job)
		assert.Equal(t, "test-job", job.Name)
		repo.AssertExpectations(t)
	})

	t.Run("ListJobs", func(t *testing.T) {
		expectedJobs := []*domain.CronJob{{ID: uuid.New(), Name: "job1"}}
		repo.On("ListJobs", mock.Anything, userID).Return(expectedJobs, nil).Once()

		jobs, err := svc.ListJobs(ctx)
		require.NoError(t, err)
		assert.Len(t, jobs, 1)
		assert.Equal(t, "job1", jobs[0].Name)
	})

	t.Run("GetJob", func(t *testing.T) {
		jobID := uuid.New()
		expectedJob := &domain.CronJob{ID: jobID, Name: "job1"}
		repo.On("GetJobByID", mock.Anything, jobID, userID).Return(expectedJob, nil).Once()

		job, err := svc.GetJob(ctx, jobID)
		require.NoError(t, err)
		assert.Equal(t, jobID, job.ID)
	})

	t.Run("PauseJob", func(t *testing.T) {
		jobID := uuid.New()
		job := &domain.CronJob{ID: jobID, UserID: userID, Status: domain.CronStatusActive}
		repo.On("GetJobByID", mock.Anything, jobID, userID).Return(job, nil).Once()
		repo.On("UpdateJob", mock.Anything, mock.MatchedBy(func(j *domain.CronJob) bool {
			return j.Status == domain.CronStatusPaused
		})).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "cron.job_pause", "cron_job", jobID.String(), mock.Anything).Return(nil).Once()

		err := svc.PauseJob(ctx, jobID)
		require.NoError(t, err)
	})

	t.Run("ResumeJob", func(t *testing.T) {
		jobID := uuid.New()
		job := &domain.CronJob{ID: jobID, UserID: userID, Status: domain.CronStatusPaused, Schedule: "* * * * *"}
		repo.On("GetJobByID", mock.Anything, jobID, userID).Return(job, nil).Once()
		repo.On("UpdateJob", mock.Anything, mock.MatchedBy(func(j *domain.CronJob) bool {
			return j.Status == domain.CronStatusActive
		})).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "cron.job_resume", "cron_job", jobID.String(), mock.Anything).Return(nil).Once()

		err := svc.ResumeJob(ctx, jobID)
		require.NoError(t, err)
	})

	t.Run("DeleteJob", func(t *testing.T) {
		jobID := uuid.New()
		job := &domain.CronJob{ID: jobID, UserID: userID}
		repo.On("GetJobByID", mock.Anything, jobID, userID).Return(job, nil).Once()
		repo.On("DeleteJob", mock.Anything, jobID).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "cron.job_delete", "cron_job", jobID.String(), mock.Anything).Return(nil).Once()

		err := svc.DeleteJob(ctx, jobID)
		require.NoError(t, err)
	})
}
