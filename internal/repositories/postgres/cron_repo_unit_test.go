package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

func TestCronRepository_CreateJob(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewPostgresCronRepository(mock)
	job := &domain.CronJob{
		ID:            uuid.New(),
		UserID:        uuid.New(),
		Name:          "test-job",
		Schedule:      "* * * * *",
		TargetURL:     "http://test",
		TargetMethod:  "POST",
		TargetPayload: "{}",
		Status:        domain.CronStatusActive,
		NextRunAt:     nil,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	mock.ExpectExec("INSERT INTO cron_jobs").
		WithArgs(job.ID, job.UserID, job.Name, job.Schedule, job.TargetURL, job.TargetMethod, job.TargetPayload, job.Status, job.NextRunAt, job.CreatedAt, job.UpdatedAt).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.CreateJob(context.Background(), job)
	assert.NoError(t, err)
}

func TestCronRepository_GetJobByID(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewPostgresCronRepository(mock)
	id := uuid.New()
	userID := uuid.New()
	now := time.Now()
	var lastRunAt, nextRunAt *time.Time

	mock.ExpectQuery("SELECT id, user_id, name, schedule, target_url, target_method, target_payload, status, last_run_at, next_run_at, created_at, updated_at FROM cron_jobs").
		WithArgs(id, userID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "schedule", "target_url", "target_method", "target_payload", "status", "last_run_at", "next_run_at", "created_at", "updated_at"}).
			AddRow(id, userID, "test-job", "* * * * *", "http://test", "POST", "{}", string(domain.CronStatusActive), lastRunAt, nextRunAt, now, now))

	job, err := repo.GetJobByID(context.Background(), id, userID)
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, id, job.ID)
	assert.Equal(t, domain.CronStatusActive, job.Status)
}

func TestCronRepository_ListJobs(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewPostgresCronRepository(mock)
	userID := uuid.New()
	now := time.Now()
	var lastRunAt, nextRunAt *time.Time

	mock.ExpectQuery("SELECT id, user_id, name, schedule, target_url, target_method, target_payload, status, last_run_at, next_run_at, created_at, updated_at FROM cron_jobs").
		WithArgs(userID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "schedule", "target_url", "target_method", "target_payload", "status", "last_run_at", "next_run_at", "created_at", "updated_at"}).
			AddRow(uuid.New(), userID, "test-job", "* * * * *", "http://test", "POST", "{}", string(domain.CronStatusActive), lastRunAt, nextRunAt, now, now))

	jobs, err := repo.ListJobs(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, jobs, 1)
	assert.Equal(t, domain.CronStatusActive, jobs[0].Status)
}
