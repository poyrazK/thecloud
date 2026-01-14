package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuditRepository is a mock implementation of ports.AuditRepository
type MockAuditRepository struct {
	mock.Mock
}

func (m *MockAuditRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockAuditRepository) ListByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.AuditLog, error) {
	args := m.Called(ctx, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.AuditLog), args.Error(1)
}

const auditLogType = "*domain.AuditLog"

func TestNewAuditService(t *testing.T) {
	repo := new(MockAuditRepository)
	svc := services.NewAuditService(repo)

	assert.NotNil(t, svc)
}

func TestAuditServiceLogSuccess(t *testing.T) {
	repo := new(MockAuditRepository)
	svc := services.NewAuditService(repo)

	ctx := context.Background()
	userID := uuid.New()
	action := "instance.create"
	resourceType := "instance"
	resourceID := uuid.New().String()
	details := map[string]interface{}{
		"image": "alpine",
		"size":  "small",
	}

	repo.On("Create", mock.Anything, mock.AnythingOfType(auditLogType)).Return(nil)

	err := svc.Log(ctx, userID, action, resourceType, resourceID, details)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	repo.AssertCalled(t, "Create", mock.Anything, mock.AnythingOfType(auditLogType))
}

func TestAuditServiceLogRepositoryError(t *testing.T) {
	repo := new(MockAuditRepository)
	svc := services.NewAuditService(repo)

	ctx := context.Background()
	userID := uuid.New()

	repo.On("Create", mock.Anything, mock.AnythingOfType(auditLogType)).Return(assert.AnError)

	err := svc.Log(ctx, userID, "test.action", "test", "123", nil)

	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
	repo.AssertExpectations(t)
}

func TestAuditServiceLogWithNilDetails(t *testing.T) {
	repo := new(MockAuditRepository)
	svc := services.NewAuditService(repo)

	ctx := context.Background()
	userID := uuid.New()

	repo.On("Create", mock.Anything, mock.AnythingOfType(auditLogType)).Return(nil)

	err := svc.Log(ctx, userID, "test.action", "test", "123", nil)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestAuditServiceListLogsSuccess(t *testing.T) {
	repo := new(MockAuditRepository)
	svc := services.NewAuditService(repo)

	ctx := context.Background()
	userID := uuid.New()
	limit := 10

	expectedLogs := []*domain.AuditLog{
		{
			ID:           uuid.New(),
			UserID:       userID,
			Action:       "instance.create",
			ResourceType: "instance",
			ResourceID:   "inst-123",
		},
		{
			ID:           uuid.New(),
			UserID:       userID,
			Action:       "volume.delete",
			ResourceType: "volume",
			ResourceID:   "vol-456",
		},
	}

	repo.On("ListByUserID", mock.Anything, userID, limit).Return(expectedLogs, nil)

	logs, err := svc.ListLogs(ctx, userID, limit)

	assert.NoError(t, err)
	assert.Len(t, logs, 2)
	assert.Equal(t, expectedLogs, logs)
	repo.AssertExpectations(t)
}

func TestAuditServiceListLogsEmptyResult(t *testing.T) {
	repo := new(MockAuditRepository)
	svc := services.NewAuditService(repo)

	ctx := context.Background()
	userID := uuid.New()

	repo.On("ListByUserID", mock.Anything, userID, 50).Return([]*domain.AuditLog{}, nil)

	logs, err := svc.ListLogs(ctx, userID, 50)

	assert.NoError(t, err)
	assert.Empty(t, logs)
	repo.AssertExpectations(t)
}

func TestAuditServiceListLogsDefaultLimit(t *testing.T) {
	repo := new(MockAuditRepository)
	svc := services.NewAuditService(repo)

	ctx := context.Background()
	userID := uuid.New()

	// When limit is 0 or negative, it should default to 50
	repo.On("ListByUserID", mock.Anything, userID, 50).Return([]*domain.AuditLog{}, nil)

	logs, err := svc.ListLogs(ctx, userID, 0)

	assert.NoError(t, err)
	assert.NotNil(t, logs)
	repo.AssertExpectations(t)
	repo.AssertCalled(t, "ListByUserID", mock.Anything, userID, 50)

	// Test with negative limit
	repo2 := new(MockAuditRepository)
	svc2 := services.NewAuditService(repo2)
	repo2.On("ListByUserID", mock.Anything, userID, 50).Return([]*domain.AuditLog{}, nil)

	logs, err = svc2.ListLogs(ctx, userID, -10)

	assert.NoError(t, err)
	assert.NotNil(t, logs)
	repo2.AssertExpectations(t)
}

func TestAuditServiceListLogsRepositoryError(t *testing.T) {
	repo := new(MockAuditRepository)
	svc := services.NewAuditService(repo)

	ctx := context.Background()
	userID := uuid.New()

	repo.On("ListByUserID", mock.Anything, userID, 50).Return(nil, assert.AnError)

	logs, err := svc.ListLogs(ctx, userID, 50)

	assert.Error(t, err)
	assert.Nil(t, logs)
	assert.Equal(t, assert.AnError, err)
	repo.AssertExpectations(t)
}
