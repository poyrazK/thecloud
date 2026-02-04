package services_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
)

func setupCronServiceIntegrationTest(t *testing.T) (ports.CronService, ports.CronRepository, context.Context) {
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewPostgresCronRepository(db)
	eventRepo := postgres.NewEventRepository(db)
	eventSvc := services.NewEventService(eventRepo, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(auditRepo)

	svc := services.NewCronService(repo, eventSvc, auditSvc)

	return svc, repo, ctx
}

func TestCronService_Integration(t *testing.T) {
	svc, repo, ctx := setupCronServiceIntegrationTest(t)
	userID := appcontext.UserIDFromContext(ctx)

	t.Run("JobLifecycle", func(t *testing.T) {
		name := "backup-job"
		schedule := "0 0 * * *"
		job, err := svc.CreateJob(ctx, name, schedule, "http://api/backup", "POST", "")
		assert.NoError(t, err)
		assert.NotNil(t, job)
		assert.Equal(t, name, job.Name)
		assert.Equal(t, domain.CronStatusActive, job.Status)
		assert.NotNil(t, job.NextRunAt)

		// Get
		fetched, err := svc.GetJob(ctx, job.ID)
		assert.NoError(t, err)
		assert.Equal(t, job.ID, fetched.ID)

		// List
		jobs, err := svc.ListJobs(ctx)
		assert.NoError(t, err)
		assert.Len(t, jobs, 1)

		// Pause
		err = svc.PauseJob(ctx, job.ID)
		assert.NoError(t, err)

		paused, _ := repo.GetJobByID(ctx, job.ID, userID)
		assert.Equal(t, domain.CronStatusPaused, paused.Status)
		assert.Nil(t, paused.NextRunAt)

		// Resume
		err = svc.ResumeJob(ctx, job.ID)
		assert.NoError(t, err)

		resumed, _ := repo.GetJobByID(ctx, job.ID, userID)
		assert.Equal(t, domain.CronStatusActive, resumed.Status)
		assert.NotNil(t, resumed.NextRunAt)

		// Delete
		err = svc.DeleteJob(ctx, job.ID)
		assert.NoError(t, err)

		_, err = repo.GetJobByID(ctx, job.ID, userID)
		assert.Error(t, err)
	})

	t.Run("Validation", func(t *testing.T) {
		// Invalid cron
		_, err := svc.CreateJob(ctx, "bad", "invalid schedule", "http://u", "GET", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid schedule")
	})
}
