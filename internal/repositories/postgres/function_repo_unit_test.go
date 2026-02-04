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

func TestFunctionRepository_Create(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewFunctionRepository(mock)
	f := &domain.Function{
		ID:             uuid.New(),
		UserID:         uuid.New(),
		Name:           "test-func",
		Runtime:        "python3.9",
		Handler:        "main.handler",
		CodePath:       "/tmp/code",
		Timeout:        30,
		MemoryMB:       128,
		Status:         "ready",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	mock.ExpectExec("INSERT INTO functions").
		WithArgs(f.ID, f.UserID, f.Name, f.Runtime, f.Handler, f.CodePath, f.Timeout, f.MemoryMB, f.Status, f.CreatedAt, f.UpdatedAt).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.Create(context.Background(), f)
	assert.NoError(t, err)
}

func TestFunctionRepository_GetByID(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewFunctionRepository(mock)
	id := uuid.New()
	now := time.Now()

	mock.ExpectQuery("SELECT id, user_id, name, runtime, handler, code_path, timeout_seconds, memory_mb, status, created_at, updated_at FROM functions WHERE id = \\$1").
		WithArgs(id).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "runtime", "handler", "code_path", "timeout_seconds", "memory_mb", "status", "created_at", "updated_at"}).
			AddRow(id, uuid.New(), "test-func", "python3.9", "main.handler", "/tmp/code", 30, 128, "ready", now, now))

	f, err := repo.GetByID(context.Background(), id)
	assert.NoError(t, err)
	assert.NotNil(t, f)
	assert.Equal(t, id, f.ID)
}

func TestFunctionRepository_List(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewFunctionRepository(mock)
	userID := uuid.New()
	now := time.Now()

	mock.ExpectQuery("SELECT id, user_id, name, runtime, handler, code_path, timeout_seconds, memory_mb, status, created_at, updated_at FROM functions").
		WithArgs(userID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "runtime", "handler", "code_path", "timeout_seconds", "memory_mb", "status", "created_at", "updated_at"}).
			AddRow(uuid.New(), userID, "test-func", "python3.9", "main.handler", "/tmp/code", 30, 128, "ready", now, now))

	functions, err := repo.List(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, functions, 1)
}

func TestFunctionRepository_Delete(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewFunctionRepository(mock)
	id := uuid.New()

	mock.ExpectExec("DELETE FROM functions WHERE id = \\$1").
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = repo.Delete(context.Background(), id)
	assert.NoError(t, err)
}
