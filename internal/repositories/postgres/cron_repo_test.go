//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresCronRepository(t *testing.T) {
	db, _ := SetupDB(t)
	defer db.Close()
	repo := NewPostgresCronRepository(db)
	ctx := SetupTestUser(t, db)
	userID := appcontext.UserIDFromContext(ctx)

	t.Run("CreateAndGetJob", func(t *testing.T) {
		job := &domain.CronJob{
			ID:           uuid.New(),
			UserID:       userID,
			Name:         "test-job",
			Schedule:     "* * * * *",
			TargetURL:    "http://test",
			TargetMethod: "GET",
			Status:       domain.CronStatusActive,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err := repo.CreateJob(ctx, job)
		require.NoError(t, err)

		found, err := repo.GetJobByID(ctx, job.ID, userID)
		require.NoError(t, err)
		assert.Equal(t, job.Name, found.Name)
	})

	t.Run("ListJobs", func(t *testing.T) {
		jobs, err := repo.ListJobs(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, jobs, 1)
	})

	t.Run("GetNextJobs", func(t *testing.T) {
		job := &domain.CronJob{
			ID:           uuid.New(),
			UserID:       userID,
			Name:         "upcoming",
			Schedule:     "* * * * *",
			TargetURL:    "http://test",
			TargetMethod: "GET",
			Status:       domain.CronStatusActive,
			NextRunAt:    &[]time.Time{time.Now().Add(-1 * time.Minute)}[0],
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		require.NoError(t, repo.CreateJob(ctx, job))

		jobs, err := repo.GetNextJobsToRun(context.Background())
		require.NoError(t, err)
		assert.NotEmpty(t, jobs)
	})
}
