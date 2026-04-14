package services_test

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFunctionService_Unit(t *testing.T) {
	t.Run("BasicOps", TestFunctionServiceBasicOps)
	t.Run("CreateFunction", TestFunctionServiceCreateFunction)
	t.Run("InvokeFunction", TestFunctionServiceInvokeFunction)
	t.Run("CreateFunction_UnsupportedRuntime", TestFunctionService_CreateFunction_UnsupportedRuntime)
}

func TestFunctionServiceBasicOps(t *testing.T) {
	repo := new(MockFunctionRepo)
	compute := new(MockComputeBackend)
	fileStore := new(MockFileStore)
	auditSvc := new(MockAuditService)
	rbacSvc := new(MockRBACService)
	rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	svc := services.NewFunctionService(repo, rbacSvc, compute, fileStore, auditSvc, slog.Default())

	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	id := uuid.New()

	t.Run("GetFunction", func(t *testing.T) {
		expected := &domain.Function{ID: id, Name: "test-fn"}
		repo.On("GetByID", ctx, id).Return(expected, nil).Once()

		result, err := svc.GetFunction(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("ListFunctions", func(t *testing.T) {
		expected := []*domain.Function{{ID: id}}
		repo.On("List", ctx, userID).Return(expected, nil).Once()

		result, err := svc.ListFunctions(ctx)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("DeleteFunction", func(t *testing.T) {
		f := &domain.Function{ID: id, UserID: userID, Name: "test-fn", CodePath: "path/to/code"}
		repo.On("GetByID", ctx, id).Return(f, nil).Once()
		repo.On("Delete", ctx, id).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "function.delete", "function", id.String(), mock.Anything).Return(nil).Once()
		fileStore.On("Delete", mock.Anything, "functions", "path/to/code").Return(nil).Maybe()

		err := svc.DeleteFunction(ctx, id)
		require.NoError(t, err)
	})

	t.Run("GetFunctionLogs", func(t *testing.T) {
		expected := []*domain.Invocation{{ID: uuid.New()}}
		repo.On("GetByID", ctx, id).Return(&domain.Function{ID: id}, nil).Once()
		repo.On("GetInvocations", ctx, id, 10).Return(expected, nil).Once()

		result, err := svc.GetFunctionLogs(ctx, id, 10)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})
}

func TestFunctionServiceCreateFunction(t *testing.T) {
	repo := new(MockFunctionRepo)
	compute := new(MockComputeBackend)
	fileStore := new(MockFileStore)
	auditSvc := new(MockAuditService)
	rbacSvc := new(MockRBACService)
	rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	svc := services.NewFunctionService(repo, rbacSvc, compute, fileStore, auditSvc, slog.Default())

	ctx := appcontext.WithUserID(context.Background(), uuid.New())
	userID := appcontext.UserIDFromContext(ctx)

	t.Run("success", func(t *testing.T) {
		name := "my-func"
		runtime := "nodejs20"
		handler := "index.handler"
		code := []byte("console.log('hi')")

		repo.On("GetByName", ctx, userID, name).Return(nil, nil).Once()
		fileStore.On("Write", ctx, "functions", mock.Anything, mock.Anything).Return(int64(len(code)), nil).Once()
		repo.On("Create", ctx, mock.Anything).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "function.create", "function", mock.Anything, mock.Anything).Return(nil).Once()

		f, err := svc.CreateFunction(ctx, name, runtime, handler, code)
		require.NoError(t, err)
		assert.NotNil(t, f)
		assert.Equal(t, name, f.Name)
	})
}

func TestFunctionServiceInvokeFunction(t *testing.T) {
	repo := new(MockFunctionRepo)
	compute := new(MockComputeBackend)
	fileStore := new(MockFileStore)
	auditSvc := new(MockAuditService)
	rbacSvc := new(MockRBACService)
	rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	svc := services.NewFunctionService(repo, rbacSvc, compute, fileStore, auditSvc, slog.Default())

	ctx := context.Background()
	id := uuid.New()
	userID := uuid.New()
	f := &domain.Function{
		ID:       id,
		UserID:   userID,
		Name:     "test-fn",
		Runtime:  "nodejs20",
		CodePath: "path/v1.zip",
		Timeout:  30,
	}

	// Create valid dummy zip
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	fw, _ := zw.Create("index.js")
	_, _ = fw.Write([]byte("console.log('hi')"))
	_ = zw.Close()

	t.Run("sync success", func(t *testing.T) {
		repo.On("GetByID", mock.Anything, id).Return(f, nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "function.invoke", "function", id.String(), mock.Anything).Return(nil).Once()

		// runInvocation details
		fileStore.On("Read", mock.Anything, "functions", f.CodePath).Return(io.NopCloser(bytes.NewReader(buf.Bytes())), nil).Once()
		compute.On("RunTask", mock.Anything, mock.Anything).Return("task-1", []string{}, nil).Once()
		compute.On("WaitTask", mock.Anything, "task-1").Return(int64(0), nil).Once()
		compute.On("GetInstanceLogs", mock.Anything, "task-1").Return(io.NopCloser(strings.NewReader("output")), nil).Once()
		compute.On("DeleteInstance", mock.Anything, "task-1").Return(nil).Once()
		repo.On("CreateInvocation", mock.Anything, mock.Anything).Return(nil).Once()

		inv, err := svc.InvokeFunction(ctx, id, []byte("{}"), false)
		require.NoError(t, err)
		assert.NotNil(t, inv)
		assert.Equal(t, "SUCCESS", inv.Status)
	})

	t.Run("async success", func(t *testing.T) {
		repo.On("GetByID", mock.Anything, id).Return(f, nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "function.invoke_async", "function", id.String(), mock.Anything).Return(nil).Once()

		// Minimal expectations for the goroutine part (non-deterministic timing)
		fileStore.On("Read", mock.Anything, mock.Anything, mock.Anything).Return(io.NopCloser(bytes.NewReader(buf.Bytes())), nil).Maybe()
		compute.On("RunTask", mock.Anything, mock.Anything).Return("task-async", []string{}, nil).Maybe()
		compute.On("WaitTask", mock.Anything, "task-async").Return(int64(0), nil).Maybe()
		compute.On("GetInstanceLogs", mock.Anything, "task-async").Return(io.NopCloser(strings.NewReader("logs")), nil).Maybe()
		compute.On("DeleteInstance", mock.Anything, "task-async").Return(nil).Maybe()
		repo.On("CreateInvocation", mock.Anything, mock.Anything).Return(nil).Maybe()

		inv, err := svc.InvokeFunction(ctx, id, []byte("{}"), true)
		require.NoError(t, err)
		assert.NotNil(t, inv)
		assert.Equal(t, "PENDING", inv.Status)
	})
}

func TestFunctionService_CreateFunction_UnsupportedRuntime(t *testing.T) {
	repo := new(MockFunctionRepo)
	rbacSvc := new(MockRBACService)
	rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	svc := services.NewFunctionService(repo, rbacSvc, nil, nil, nil, slog.Default())
	ctx := appcontext.WithUserID(context.Background(), uuid.New())

	_, err := svc.CreateFunction(ctx, "fail", "cobol99", "handler", []byte("code"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported runtime")
}
