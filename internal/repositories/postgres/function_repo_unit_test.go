package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testFuncName    = "test-func"
	testFuncRuntime = "python3.9"
	testFuncHandler = "main.handler"
	testFuncPath    = "/tmp/code"
)

func TestFunctionRepositoryCreate(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewFunctionRepository(mock)
	f := &domain.Function{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Name:      testFuncName,
		Runtime:   testFuncRuntime,
		Handler:   testFuncHandler,
		CodePath:  testFuncPath,
		Timeout:   30,
		MemoryMB:  128,
		Status:    "ready",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mock.ExpectExec("INSERT INTO functions").
		WithArgs(f.ID, f.UserID, f.Name, f.Runtime, f.Handler, f.CodePath, f.Timeout, f.MemoryMB, f.Status, f.CreatedAt, f.UpdatedAt).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.Create(context.Background(), f)
	require.NoError(t, err)
}

func TestFunctionRepositoryGetByID(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewFunctionRepository(mock)
	id := uuid.New()
	now := time.Now()

	mock.ExpectQuery("SELECT id, user_id, name, runtime, handler, code_path, timeout_seconds, memory_mb, status, created_at, updated_at FROM functions WHERE id = \\$1").
		WithArgs(id).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "runtime", "handler", "code_path", "timeout_seconds", "memory_mb", "status", "created_at", "updated_at"}).
			AddRow(id, uuid.New(), testFuncName, testFuncRuntime, testFuncHandler, testFuncPath, 30, 128, "ready", now, now))

	f, err := repo.GetByID(context.Background(), id)
	require.NoError(t, err)
	assert.NotNil(t, f)
	assert.Equal(t, id, f.ID)
}

func TestFunctionRepositoryList(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewFunctionRepository(mock)
	userID := uuid.New()
	now := time.Now()

	mock.ExpectQuery("SELECT id, user_id, name, runtime, handler, code_path, timeout_seconds, memory_mb, status, created_at, updated_at FROM functions").
		WithArgs(userID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "runtime", "handler", "code_path", "timeout_seconds", "memory_mb", "status", "created_at", "updated_at"}).
			AddRow(uuid.New(), userID, testFuncName, testFuncRuntime, testFuncHandler, testFuncPath, 30, 128, "ready", now, now))

	functions, err := repo.List(context.Background(), userID)
	require.NoError(t, err)
	assert.Len(t, functions, 1)
}

func TestFunctionRepositoryDelete(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewFunctionRepository(mock)
	id := uuid.New()

	mock.ExpectExec("DELETE FROM functions WHERE id = \\$1").
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = repo.Delete(context.Background(), id)
	require.NoError(t, err)
}
