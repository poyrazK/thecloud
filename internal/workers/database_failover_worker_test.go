package workers

import (
	"context"
	"fmt"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/mock"
)

type mockDatabaseRepo struct {
	mock.Mock
}

func (m *mockDatabaseRepo) Create(ctx context.Context, db *domain.Database) error {
	return m.Called(ctx, db).Error(0)
}
func (m *mockDatabaseRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Database, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).(*domain.Database)
	return r0, args.Error(1)
}
func (m *mockDatabaseRepo) List(ctx context.Context) ([]*domain.Database, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).([]*domain.Database)
	return r0, args.Error(1)
}
func (m *mockDatabaseRepo) ListReplicas(ctx context.Context, primaryID uuid.UUID) ([]*domain.Database, error) {
	args := m.Called(ctx, primaryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).([]*domain.Database)
	return r0, args.Error(1)
}
func (m *mockDatabaseRepo) Update(ctx context.Context, db *domain.Database) error {
	return m.Called(ctx, db).Error(0)
}
func (m *mockDatabaseRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

type mockDatabaseService struct {
	mock.Mock
}

func (m *mockDatabaseService) CreateDatabase(ctx context.Context, name, engine, version string, vpcID *uuid.UUID, allocatedStorage int, parameters map[string]string, metricsEnabled bool) (*domain.Database, error) {
	args := m.Called(ctx, name, engine, version, vpcID, allocatedStorage, parameters, metricsEnabled)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).(*domain.Database)
	return r0, args.Error(1)
}
func (m *mockDatabaseService) CreateReplica(ctx context.Context, primaryID uuid.UUID, name string) (*domain.Database, error) {
	args := m.Called(ctx, primaryID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Database), args.Error(1)
}
func (m *mockDatabaseService) PromoteToPrimary(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockDatabaseService) GetDatabase(ctx context.Context, id uuid.UUID) (*domain.Database, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Database), args.Error(1)
}
func (m *mockDatabaseService) ListDatabases(ctx context.Context) ([]*domain.Database, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Database), args.Error(1)
}
func (m *mockDatabaseService) DeleteDatabase(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockDatabaseService) GetConnectionString(ctx context.Context, id uuid.UUID) (string, error) {
	args := m.Called(ctx, id)
	return args.String(0), args.Error(1)
}
func (m *mockDatabaseService) CreateDatabaseSnapshot(ctx context.Context, databaseID uuid.UUID, description string) (*domain.Snapshot, error) {
	args := m.Called(ctx, databaseID, description)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Snapshot), args.Error(1)
}
func (m *mockDatabaseService) ListDatabaseSnapshots(ctx context.Context, databaseID uuid.UUID) ([]*domain.Snapshot, error) {
	args := m.Called(ctx, databaseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Snapshot), args.Error(1)
}
func (m *mockDatabaseService) RestoreDatabase(ctx context.Context, snapshotID uuid.UUID, newName, engine, version string, vpcID *uuid.UUID, allocatedStorage int, parameters map[string]string, metricsEnabled bool) (*domain.Database, error) {
	args := m.Called(ctx, snapshotID, newName, engine, version, vpcID, allocatedStorage, parameters, metricsEnabled)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Database), args.Error(1)
}

func TestDatabaseFailoverWorker(t *testing.T) {
	t.Parallel()

	primaryID := uuid.New()
	replicaID := uuid.New()

	primary := &domain.Database{
		ID:     primaryID,
		Name:   "primary",
		Role:   domain.RolePrimary,
		Status: domain.DatabaseStatusRunning,
		Port:   -1, // Invalid port to ensure DialTimeout fails
	}

	replica := &domain.Database{
		ID:        replicaID,
		Name:      "replica",
		Role:      domain.RoleReplica,
		PrimaryID: &primaryID,
		Status:    domain.DatabaseStatusRunning,
	}

	anyCtx := mock.MatchedBy(func(ctx context.Context) bool { return true })

	t.Run("Failover triggered on unhealthy primary", func(t *testing.T) {
		repo := new(mockDatabaseRepo)
		svc := new(mockDatabaseService)
		logger := slog.Default()

		worker := NewDatabaseFailoverWorker(svc, repo, logger)

		repo.On("List", anyCtx).Return([]*domain.Database{primary}, nil)
		repo.On("ListReplicas", anyCtx, primaryID).Return([]*domain.Database{replica}, nil)
		svc.On("PromoteToPrimary", anyCtx, replicaID).Return(nil)

		worker.checkDatabases(context.Background())

		repo.AssertExpectations(t)
		svc.AssertExpectations(t)
	})

	t.Run("No failover if primary healthy", func(t *testing.T) {
		// This logic would require mocking net.Dial which is handled by isHealthy
	})

	t.Run("No failover if no replicas", func(t *testing.T) {
		repo := new(mockDatabaseRepo)
		svc := new(mockDatabaseService)
		worker := NewDatabaseFailoverWorker(svc, repo, slog.Default())

		repo.On("List", anyCtx).Return([]*domain.Database{primary}, nil)
		repo.On("ListReplicas", anyCtx, primaryID).Return([]*domain.Database{}, nil)

		worker.checkDatabases(context.Background())

		repo.AssertExpectations(t)
		svc.AssertNotCalled(t, "PromoteToPrimary", anyCtx, mock.Anything)
	})

	t.Run("Repo list error handled", func(t *testing.T) {
		repo := new(mockDatabaseRepo)
		svc := new(mockDatabaseService)
		worker := NewDatabaseFailoverWorker(svc, repo, slog.Default())

		repo.On("List", anyCtx).Return([]*domain.Database(nil), fmt.Errorf("db error"))

		worker.checkDatabases(context.Background())

		repo.AssertExpectations(t)
	})
}
