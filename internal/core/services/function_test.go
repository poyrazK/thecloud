package services

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDocker is duplicated here or should we export it from another test?
// Since tests can't share code easily without separate package, I'll redefine or update the existing usages.
// Actually, `function_test.go` uses `MockDocker` but it is NOT defined in this file.
// It relies on `MockDocker` being defined in `instance_test.go` (same package `services`).
// So I don't need to add it here, I just need to make sure `instance_test.go` MockDocker has Exec.
// Which I already did.

type MockFunctionRepo struct {
	mock.Mock
}

func (m *MockFunctionRepo) Create(ctx context.Context, f *domain.Function) error {
	args := m.Called(ctx, f)
	return args.Error(0)
}
func (m *MockFunctionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Function, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Function), args.Error(1)
}
func (m *MockFunctionRepo) GetByName(ctx context.Context, userID uuid.UUID, name string) (*domain.Function, error) {
	args := m.Called(ctx, userID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Function), args.Error(1)
}
func (m *MockFunctionRepo) List(ctx context.Context, userID uuid.UUID) ([]*domain.Function, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*domain.Function), args.Error(1)
}
func (m *MockFunctionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockFunctionRepo) CreateInvocation(ctx context.Context, i *domain.Invocation) error {
	args := m.Called(ctx, i)
	return args.Error(0)
}
func (m *MockFunctionRepo) GetInvocations(ctx context.Context, functionID uuid.UUID, limit int) ([]*domain.Invocation, error) {
	args := m.Called(ctx, functionID, limit)
	return args.Get(0).([]*domain.Invocation), args.Error(1)
}

type MockFileStore struct {
	mock.Mock
}

func (m *MockFileStore) Write(ctx context.Context, bucket, key string, r io.Reader) (int64, error) {
	args := m.Called(ctx, bucket, key, r)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockFileStore) Read(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	args := m.Called(ctx, bucket, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}
func (m *MockFileStore) Delete(ctx context.Context, bucket, key string) error {
	args := m.Called(ctx, bucket, key)
	return args.Error(0)
}

func TestFunctionService_InvokeFunction(t *testing.T) {
	repo := new(MockFunctionRepo)
	docker := new(MockDocker)
	fileStore := new(MockFileStore)
	logger := slog.Default()

	svc := NewFunctionService(repo, docker, fileStore, logger)
	assert.NotNil(t, svc)

	fID := uuid.New()
	f := &domain.Function{
		ID:       fID,
		Name:     "test-fn",
		Runtime:  "nodejs20",
		Handler:  "index.handler",
		CodePath: "user1/fn1/code.zip",
		Timeout:  30,
		MemoryMB: 128,
	}

	repo.On("GetByID", mock.Anything, fID).Return(f, nil)
	// We'd need to mock prepareCode which is internal...
	// For unit testing, we might want to extract internal logic to helpers or interfaces.
	// But let's see if we can test create first.
}

func TestCreateFunction_Success(t *testing.T) {
	repo := new(MockFunctionRepo)
	docker := new(MockDocker)
	fileStore := new(MockFileStore)
	logger := slog.Default()

	svc := NewFunctionService(repo, docker, fileStore, logger)

	ctx := appcontext.WithUserID(context.Background(), uuid.New())

	fileStore.On("Write", mock.Anything, "functions", mock.Anything, mock.Anything).Return(int64(100), nil)
	repo.On("Create", mock.Anything, mock.Anything).Return(nil)

	fn, err := svc.CreateFunction(ctx, "my-fn", "nodejs20", "index.handler", []byte("code"))

	assert.NoError(t, err)
	assert.NotNil(t, fn)
	assert.Equal(t, "my-fn", fn.Name)
	repo.AssertCalled(t, "Create", mock.Anything, mock.Anything)
}
