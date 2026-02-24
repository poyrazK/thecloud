package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/poyrazk/thecloud/internal/core/domain"
	theclouderrors "github.com/poyrazk/thecloud/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStackRepositoryCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewStackRepository(mock)
		s := &domain.Stack{
			ID:         uuid.New(),
			UserID:     uuid.New(),
			Name:       "test-stack",
			Template:   "{}",
			Parameters: []byte(`{"foo": "bar"}`),
			Status:     "CREATE_IN_PROGRESS",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		mock.ExpectExec("INSERT INTO stacks").
			WithArgs(s.ID, s.UserID, s.Name, s.Template, s.Parameters, string(s.Status), s.CreatedAt, s.UpdatedAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = repo.Create(context.Background(), s)
		require.NoError(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewStackRepository(mock)
		s := &domain.Stack{
			ID: uuid.New(),
		}

		mock.ExpectExec("INSERT INTO stacks").
			WillReturnError(errors.New("db error"))

		err = repo.Create(context.Background(), s)
		require.Error(t, err)
	})
}

func TestStackRepositoryGetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewStackRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		now := time.Now()

		mock.ExpectQuery("SELECT id, user_id, name, template, parameters, status, status_reason, created_at, updated_at FROM stacks").
			WithArgs(id).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "template", "parameters", "status", "status_reason", "created_at", "updated_at"}).
				AddRow(id, userID, "test", "{}", nil, "ACTIVE", "", now, now))

		mock.ExpectQuery("SELECT id, stack_id, logical_id, physical_id, resource_type, status, created_at FROM stack_resources").
			WithArgs(id).
			WillReturnRows(pgxmock.NewRows([]string{"id", "stack_id", "logical_id", "physical_id", "resource_type", "status", "created_at"}).
				AddRow(uuid.New(), id, "res1", "phys1", "type1", "status1", now))

		s, err := repo.GetByID(context.Background(), id)
		require.NoError(t, err)
		assert.NotNil(t, s)
		assert.Len(t, s.Resources, 1)
	})

	t.Run("not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewStackRepository(mock)
		id := uuid.New()

		mock.ExpectQuery("SELECT id, user_id, name, template, parameters, status, status_reason, created_at, updated_at FROM stacks").
			WithArgs(id).
			WillReturnError(pgx.ErrNoRows)

		s, err := repo.GetByID(context.Background(), id)
		require.Error(t, err)
		assert.Nil(t, s)
		var target *theclouderrors.Error
		if errors.As(err, &target) {
			assert.Equal(t, theclouderrors.NotFound, target.Type)
		}
	})
}

func TestStackRepositoryGetByName(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewStackRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		name := "test-stack"
		now := time.Now()

		mock.ExpectQuery("SELECT id, user_id, name, template, parameters, status, status_reason, created_at, updated_at FROM stacks").
			WithArgs(userID, name).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "template", "parameters", "status", "status_reason", "created_at", "updated_at"}).
				AddRow(id, userID, name, "{}", nil, "ACTIVE", "", now, now))

		mock.ExpectQuery("SELECT id, stack_id, logical_id, physical_id, resource_type, status, created_at FROM stack_resources").
			WithArgs(id).
			WillReturnRows(pgxmock.NewRows([]string{"id", "stack_id", "logical_id", "physical_id", "resource_type", "status", "created_at"}).
				AddRow(uuid.New(), id, "res1", "phys1", "type1", "status1", now))

		s, err := repo.GetByName(context.Background(), userID, name)
		require.NoError(t, err)
		assert.NotNil(t, s)
		assert.Equal(t, id, s.ID)
	})

	t.Run("not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewStackRepository(mock)
		userID := uuid.New()
		name := "test-stack"

		mock.ExpectQuery("SELECT id, user_id, name, template, parameters, status, status_reason, created_at, updated_at FROM stacks").
			WithArgs(userID, name).
			WillReturnError(pgx.ErrNoRows)

		s, err := repo.GetByName(context.Background(), userID, name)
		require.Error(t, err)
		assert.Nil(t, s)
		var target *theclouderrors.Error
		ok := errors.As(err, &target)
		if ok {
			assert.Equal(t, theclouderrors.NotFound, target.Type)
		}
	})
}

func TestStackRepositoryListByUserID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewStackRepository(mock)
		userID := uuid.New()
		now := time.Now()

		mock.ExpectQuery("SELECT id, user_id, name, template, parameters, status, status_reason, created_at, updated_at FROM stacks").
			WithArgs(userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "template", "parameters", "status", "status_reason", "created_at", "updated_at"}).
				AddRow(uuid.New(), userID, "s1", "{}", nil, "ACTIVE", "", now, now))

		stacks, err := repo.ListByUserID(context.Background(), userID)
		require.NoError(t, err)
		assert.Len(t, stacks, 1)
	})

	t.Run("db error", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewStackRepository(mock)
		userID := uuid.New()

		mock.ExpectQuery("SELECT id, user_id, name, template, parameters, status, status_reason, created_at, updated_at FROM stacks").
			WithArgs(userID).
			WillReturnError(errors.New("db error"))

		stacks, err := repo.ListByUserID(context.Background(), userID)
		require.Error(t, err)
		assert.Nil(t, stacks)
	})
}

func TestStackRepositoryAddResource(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewStackRepository(mock)
		res := &domain.StackResource{
			ID:           uuid.New(),
			StackID:      uuid.New(),
			LogicalID:    "res1",
			PhysicalID:   "phys1",
			ResourceType: "type1",
			Status:       "CREATE_COMPLETE",
			CreatedAt:    time.Now(),
		}

		mock.ExpectExec("INSERT INTO stack_resources").
			WithArgs(res.ID, res.StackID, res.LogicalID, res.PhysicalID, res.ResourceType, res.Status, res.CreatedAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = repo.AddResource(context.Background(), res)
		require.NoError(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewStackRepository(mock)
		res := &domain.StackResource{
			ID: uuid.New(),
		}

		mock.ExpectExec("INSERT INTO stack_resources").
			WillReturnError(errors.New("db error"))

		err = repo.AddResource(context.Background(), res)
		require.Error(t, err)
	})
}

func TestStackRepositoryUpdate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewStackRepository(mock)
		s := &domain.Stack{
			ID:           uuid.New(),
			Status:       "UPDATE_COMPLETE",
			StatusReason: "reason",
			UpdatedAt:    time.Now(),
		}

		mock.ExpectExec("UPDATE stacks").
			WithArgs(string(s.Status), s.StatusReason, s.UpdatedAt, s.ID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err = repo.Update(context.Background(), s)
		require.NoError(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewStackRepository(mock)
		s := &domain.Stack{
			ID: uuid.New(),
		}

		mock.ExpectExec("UPDATE stacks").
			WillReturnError(errors.New("db error"))

		err = repo.Update(context.Background(), s)
		require.Error(t, err)
	})
}

func TestStackRepositoryDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewStackRepository(mock)
		id := uuid.New()

		mock.ExpectExec("DELETE FROM stacks").
			WithArgs(id).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = repo.Delete(context.Background(), id)
		require.NoError(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewStackRepository(mock)
		id := uuid.New()

		mock.ExpectExec("DELETE FROM stacks").
			WithArgs(id).
			WillReturnError(errors.New("db error"))

		err = repo.Delete(context.Background(), id)
		require.Error(t, err)
	})
}

func TestStackRepositoryDeleteResources(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewStackRepository(mock)
		stackID := uuid.New()

		mock.ExpectExec("DELETE FROM stack_resources").
			WithArgs(stackID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = repo.DeleteResources(context.Background(), stackID)
		require.NoError(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewStackRepository(mock)
		stackID := uuid.New()

		mock.ExpectExec("DELETE FROM stack_resources").
			WithArgs(stackID).
			WillReturnError(errors.New("db error"))

		err = repo.DeleteResources(context.Background(), stackID)
		require.Error(t, err)
	})
}
