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
)

type mockFunctionRepo struct {
	mock.Mock
}

func (m *mockFunctionRepo) Create(ctx context.Context, f *domain.Function) error {
	return m.Called(ctx, f).Error(0)
}
func (m *mockFunctionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Function, error) {
	args := m.Called(ctx, id)
	r0, _ := args.Get(0).(*domain.Function)
	return r0, args.Error(1)
}
func (m *mockFunctionRepo) GetByName(ctx context.Context, userID uuid.UUID, name string) (*domain.Function, error) {
	args := m.Called(ctx, userID, name)
	r0, _ := args.Get(0).(*domain.Function)
	return r0, args.Error(1)
}
func (m *mockFunctionRepo) List(ctx context.Context, userID uuid.UUID) ([]*domain.Function, error) {
	args := m.Called(ctx, userID)
	r0, _ := args.Get(0).([]*domain.Function)
	return r0, args.Error(1)
}
func (m *mockFunctionRepo) Update(ctx context.Context, f *domain.Function) error {
	return m.Called(ctx, f).Error(0)
}
func (m *mockFunctionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockFunctionRepo) CreateInvocation(ctx context.Context, i *domain.Invocation) error {
	return m.Called(ctx, i).Error(0)
}
func (m *mockFunctionRepo) GetInvocations(ctx context.Context, functionID uuid.UUID, limit int) ([]*domain.Invocation, error) {
	args := m.Called(ctx, functionID, limit)
	r0, _ := args.Get(0).([]*domain.Invocation)
	return r0, args.Error(1)
}

type mockFileStore struct {
	mock.Mock
}

func (m *mockFileStore) Write(ctx context.Context, bucket, key string, r io.Reader) (int64, error) {
	args := m.Called(ctx, bucket, key, r)
	r0, _ := args.Get(0).(int64)
	return r0, args.Error(1)
}
func (m *mockFileStore) Read(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	args := m.Called(ctx, bucket, key)
	r0, _ := args.Get(0).(io.ReadCloser)
	return r0, args.Error(1)
}
func (m *mockFileStore) Delete(ctx context.Context, bucket, key string) error {
	return m.Called(ctx, bucket, key).Error(0)
}
func (m *mockFileStore) GetClusterStatus(ctx context.Context) (*domain.StorageCluster, error) {
	args := m.Called(ctx)
	r0, _ := args.Get(0).(*domain.StorageCluster)
	return r0, args.Error(1)
}
func (m *mockFileStore) Assemble(ctx context.Context, bucket, key string, parts []string) (int64, error) {
	args := m.Called(ctx, bucket, key, parts)
	r0, _ := args.Get(0).(int64)
	return r0, args.Error(1)
}

func TestFunctionService_BasicOps(t *testing.T) {
	repo := new(mockFunctionRepo)
	compute := new(MockComputeBackend)
	fileStore := new(mockFileStore)
	auditSvc := new(MockAuditService)
	svc := services.NewFunctionService(repo, compute, fileStore, auditSvc, slog.Default())

	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	id := uuid.New()

	t.Run("GetFunction", func(t *testing.T) {
		expected := &domain.Function{ID: id, Name: "test-fn"}
		repo.On("GetByID", ctx, id).Return(expected, nil).Once()

		result, err := svc.GetFunction(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("ListFunctions", func(t *testing.T) {
		expected := []*domain.Function{{ID: id}}
		repo.On("List", ctx, userID).Return(expected, nil).Once()

		result, err := svc.ListFunctions(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("DeleteFunction", func(t *testing.T) {
		f := &domain.Function{ID: id, UserID: userID, Name: "test-fn", CodePath: "path/to/code"}
		repo.On("GetByID", ctx, id).Return(f, nil).Once()
		repo.On("Delete", ctx, id).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "function.delete", "function", id.String(), mock.Anything).Return(nil).Once()
		fileStore.On("Delete", mock.Anything, "functions", "path/to/code").Return(nil).Maybe()

		err := svc.DeleteFunction(ctx, id)
		assert.NoError(t, err)
	})

	t.Run("GetFunctionLogs", func(t *testing.T) {
		expected := []*domain.Invocation{{ID: uuid.New()}}
		repo.On("GetInvocations", ctx, id, 10).Return(expected, nil).Once()

		result, err := svc.GetFunctionLogs(ctx, id, 10)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})
}

func TestFunctionService_CreateFunction(t *testing.T) {
	repo := new(mockFunctionRepo)
	compute := new(MockComputeBackend)
	fileStore := new(mockFileStore)
	auditSvc := new(MockAuditService)
	svc := services.NewFunctionService(repo, compute, fileStore, auditSvc, slog.Default())

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
		assert.NoError(t, err)
		assert.NotNil(t, f)
		assert.Equal(t, name, f.Name)
	})
}

func TestFunctionService_InvokeFunction(t *testing.T) {
	repo := new(mockFunctionRepo)
	compute := new(MockComputeBackend)
	fileStore := new(mockFileStore)
	auditSvc := new(MockAuditService)
	svc := services.NewFunctionService(repo, compute, fileStore, auditSvc, slog.Default())

	ctx := context.Background()
	id := uuid.New()
	userID := uuid.New()
	f := &domain.Function{
		ID: id, 
		UserID: userID, 
		Name: "test-fn", 
		Runtime: "nodejs20", 
		CodePath: "path/v1.zip",
		Timeout: 30,
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
		assert.NoError(t, err)
		assert.NotNil(t, inv)
		assert.Equal(t, "SUCCESS", inv.Status)
	})
}
