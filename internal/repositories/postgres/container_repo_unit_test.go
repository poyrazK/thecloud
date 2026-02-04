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

func TestContainerRepository_CreateDeployment(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewPostgresContainerRepository(mock)
		deployment := &domain.Deployment{
			ID:           uuid.New(),
			UserID:       uuid.New(),
			Name:         "test-dep",
			Image:        "nginx",
			Replicas:     3,
			CurrentCount: 0,
			Ports:        "80:80",
			Status:       domain.DeploymentStatusScaling,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		mock.ExpectExec("INSERT INTO deployments").
			WithArgs(deployment.ID, deployment.UserID, deployment.Name, deployment.Image, deployment.Replicas, deployment.CurrentCount, deployment.Ports, deployment.Status, deployment.CreatedAt, deployment.UpdatedAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = repo.CreateDeployment(context.Background(), deployment)
		assert.NoError(t, err)
	})
}

func TestContainerRepository_GetDeploymentByID(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewPostgresContainerRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		now := time.Now()

		mock.ExpectQuery("SELECT id, user_id, name, image, replicas, current_count, ports, status, created_at, updated_at FROM deployments").
			WithArgs(id, userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "image", "replicas", "current_count", "ports", "status", "created_at", "updated_at"}).
				AddRow(id, userID, "test-dep", "nginx", 3, 0, "80:80", string(domain.DeploymentStatusScaling), now, now))

		d, err := repo.GetDeploymentByID(context.Background(), id, userID)
		assert.NoError(t, err)
		assert.NotNil(t, d)
		assert.Equal(t, id, d.ID)
		assert.Equal(t, domain.DeploymentStatusScaling, d.Status)
	})
}

func TestContainerRepository_ListDeployments(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewPostgresContainerRepository(mock)
		userID := uuid.New()
		now := time.Now()

		mock.ExpectQuery("SELECT id, user_id, name, image, replicas, current_count, ports, status, created_at, updated_at FROM deployments").
			WithArgs(userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "image", "replicas", "current_count", "ports", "status", "created_at", "updated_at"}).
				AddRow(uuid.New(), userID, "test-dep", "nginx", 3, 0, "80:80", string(domain.DeploymentStatusScaling), now, now))

		deps, err := repo.ListDeployments(context.Background(), userID)
		assert.NoError(t, err)
		assert.Len(t, deps, 1)
		assert.Equal(t, domain.DeploymentStatusScaling, deps[0].Status)
	})
}

func TestContainerRepository_UpdateDeployment(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewPostgresContainerRepository(mock)
		deployment := &domain.Deployment{
			ID:           uuid.New(),
			Replicas:     5,
			CurrentCount: 2,
			Status:       domain.DeploymentStatusReady,
		}

		mock.ExpectExec("UPDATE deployments").
			WithArgs(deployment.Replicas, deployment.CurrentCount, deployment.Status, deployment.ID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err = repo.UpdateDeployment(context.Background(), deployment)
		assert.NoError(t, err)
	})
}

func TestContainerRepository_DeleteDeployment(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewPostgresContainerRepository(mock)
		id := uuid.New()

		mock.ExpectExec("DELETE FROM deployments").
			WithArgs(id).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = repo.DeleteDeployment(context.Background(), id)
		assert.NoError(t, err)
	})
}

func TestContainerRepository_AddContainer(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewPostgresContainerRepository(mock)
		depID := uuid.New()
		instID := uuid.New()

		mock.ExpectExec("INSERT INTO deployment_containers").
			WithArgs(depID, instID).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = repo.AddContainer(context.Background(), depID, instID)
		assert.NoError(t, err)
	})
}

func TestContainerRepository_RemoveContainer(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewPostgresContainerRepository(mock)
		depID := uuid.New()
		instID := uuid.New()

		mock.ExpectExec("DELETE FROM deployment_containers").
			WithArgs(depID, instID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = repo.RemoveContainer(context.Background(), depID, instID)
		assert.NoError(t, err)
	})
}

func TestContainerRepository_GetContainers(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewPostgresContainerRepository(mock)
		depID := uuid.New()
		instID := uuid.New()

		mock.ExpectQuery("SELECT instance_id FROM deployment_containers").
			WithArgs(depID).
			WillReturnRows(pgxmock.NewRows([]string{"instance_id"}).AddRow(instID))

		ids, err := repo.GetContainers(context.Background(), depID)
		assert.NoError(t, err)
		assert.Len(t, ids, 1)
		assert.Equal(t, instID, ids[0])
	})
}

func TestContainerRepository_ListAllDeployments(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewPostgresContainerRepository(mock)
		now := time.Now()

		mock.ExpectQuery("SELECT id, user_id, name, image, replicas, current_count, ports, status, created_at, updated_at FROM deployments").
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "image", "replicas", "current_count", "ports", "status", "created_at", "updated_at"}).
				AddRow(uuid.New(), uuid.New(), "test-dep", "nginx", 3, 0, "80:80", string(domain.DeploymentStatusScaling), now, now))

		deps, err := repo.ListAllDeployments(context.Background())
		assert.NoError(t, err)
		assert.Len(t, deps, 1)
	})
}
