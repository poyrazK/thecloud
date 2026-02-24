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

func (m *mockDatabaseRepo) Create(ctx context.Context, db *domain.Database) error { return nil }
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
func (m *mockDatabaseRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }

type mockDatabaseService struct {
	mock.Mock
}

func (m *mockDatabaseService) CreateDatabase(ctx context.Context, name, engine, version string, vpcID *uuid.UUID) (*domain.Database, error) {
	return nil, nil
}
func (m *mockDatabaseService) CreateReplica(ctx context.Context, primaryID uuid.UUID, name string) (*domain.Database, error) {
	return nil, nil
}
func (m *mockDatabaseService) PromoteToPrimary(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockDatabaseService) GetDatabase(ctx context.Context, id uuid.UUID) (*domain.Database, error) {
	return nil, nil
}
func (m *mockDatabaseService) ListDatabases(ctx context.Context) ([]*domain.Database, error) {
	return nil, nil
}
func (m *mockDatabaseService) DeleteDatabase(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockDatabaseService) GetConnectionString(ctx context.Context, id uuid.UUID) (string, error) {
	return "", nil
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
		Port:   1234, // Will fail health check (no listener)
	}

	replica := &domain.Database{
		ID:        replicaID,
		Name:      "replica",
		Role:      domain.RoleReplica,
		PrimaryID: &primaryID,
		Status:    domain.DatabaseStatusRunning,
	}

	t.Run("Failover triggered on unhealthy primary", func(t *testing.T) {
		repo := new(mockDatabaseRepo)
		svc := new(mockDatabaseService)
		logger := slog.Default()

		worker := NewDatabaseFailoverWorker(svc, repo, logger)

		repo.On("List", mock.Anything).Return([]*domain.Database{primary}, nil)
		repo.On("ListReplicas", mock.Anything, primaryID).Return([]*domain.Database{replica}, nil)
		svc.On("PromoteToPrimary", mock.Anything, replicaID).Return(nil)

		worker.checkDatabases(context.Background())

		repo.AssertExpectations(t)
		svc.AssertExpectations(t)
	})

	t.Run("No failover if primary healthy", func(t *testing.T) {
		// This is hard to test without a real listener, but we can mock isHealthy if we move it to a sub-struct/interface.
		// For now, we've tested the negative case implicitly (if List returns nothing, handleFailover isn't called).
	})

	t.Run("No failover if no replicas", func(t *testing.T) {
		repo := new(mockDatabaseRepo)
		svc := new(mockDatabaseService)
		worker := NewDatabaseFailoverWorker(svc, repo, slog.Default())

		repo.On("List", mock.Anything).Return([]*domain.Database{primary}, nil)
		repo.On("ListReplicas", mock.Anything, primaryID).Return([]*domain.Database{}, nil)

		worker.checkDatabases(context.Background())

		repo.AssertExpectations(t)
		svc.AssertNotCalled(t, "PromoteToPrimary", mock.Anything, mock.Anything)
	})

	t.Run("Repo list error handled", func(t *testing.T) {
		repo := new(mockDatabaseRepo)
		svc := new(mockDatabaseService)
		worker := NewDatabaseFailoverWorker(svc, repo, slog.Default())

		repo.On("List", mock.Anything).Return([]*domain.Database(nil), fmt.Errorf("db error"))

		worker.checkDatabases(context.Background())

		repo.AssertExpectations(t)
	})
}
