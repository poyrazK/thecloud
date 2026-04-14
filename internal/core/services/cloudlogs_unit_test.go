package services_test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCloudLogsService_Unit(t *testing.T) {
	mockRepo := new(MockLogRepository)
	mockRBAC := new(MockRBACService)
	svc := services.NewCloudLogsService(mockRepo, mockRBAC, slog.Default())

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	t.Run("IngestLogs_Success", func(t *testing.T) {
		mockRBAC.On("Authorize", mock.Anything, userID, tenantID, domain.PermissionInstanceUpdate, "*").Return(nil).Once()
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()

		entries := []*domain.LogEntry{{Message: "test"}}
		err := svc.IngestLogs(ctx, entries)
		require.NoError(t, err)
	})

	t.Run("IngestLogs_Empty", func(t *testing.T) {
		err := svc.IngestLogs(ctx, nil)
		require.NoError(t, err)
	})

	t.Run("SearchLogs_Success", func(t *testing.T) {
		mockRBAC.On("Authorize", mock.Anything, userID, tenantID, domain.PermissionAuditRead, "*").Return(nil).Once()
		mockRepo.On("List", mock.Anything, mock.Anything).Return([]*domain.LogEntry{}, 0, nil).Once()

		logs, total, err := svc.SearchLogs(ctx, domain.LogQuery{})
		require.NoError(t, err)
		assert.Equal(t, 0, total)
		assert.NotNil(t, logs)
	})

	t.Run("RunRetentionPolicy_Success", func(t *testing.T) {
		mockRBAC.On("Authorize", mock.Anything, userID, tenantID, domain.PermissionFullAccess, "*").Return(nil).Once()
		mockRepo.On("DeleteByAge", mock.Anything, 30).Return(nil).Once()

		err := svc.RunRetentionPolicy(ctx, 30)
		require.NoError(t, err)
	})

	t.Run("RunRetentionPolicy_Invalid", func(t *testing.T) {
		mockRBAC.On("Authorize", mock.Anything, userID, tenantID, domain.PermissionFullAccess, "*").Return(nil).Once()
		err := svc.RunRetentionPolicy(ctx, 0)
		require.Error(t, err)
	})

	t.Run("SearchLogs_RepoError", func(t *testing.T) {
		query := domain.LogQuery{ResourceID: "res-1"}
		mockRBAC.On("Authorize", mock.Anything, userID, tenantID, domain.PermissionAuditRead, "*").Return(nil).Once()
		mockRepo.On("List", mock.Anything, mock.Anything).Return(nil, 0, fmt.Errorf("db fail")).Once()
		_, _, err := svc.SearchLogs(ctx, query)
		require.Error(t, err)
	})

	mockRepo.AssertExpectations(t)
}

func TestCloudLogsServiceSearchLogsUnit(t *testing.T) {
	mockRepo := new(MockLogRepository)
	mockRBAC := new(MockRBACService)
	svc := services.NewCloudLogsService(mockRepo, mockRBAC, slog.Default())
	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	query := domain.LogQuery{ResourceID: "res-1"}
	
	t.Run("Success", func(t *testing.T) {
		mockRBAC.On("Authorize", mock.Anything, userID, tenantID, domain.PermissionAuditRead, "*").Return(nil).Once()
		expectedLogs := []*domain.LogEntry{{Message: "found"}}
		mockRepo.On("List", mock.Anything, query).Return(expectedLogs, 1, nil).Once()

		logs, total, err := svc.SearchLogs(ctx, query)
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Equal(t, "found", logs[0].Message)
	})

	t.Run("RepoError", func(t *testing.T) {
		mockRBAC.On("Authorize", mock.Anything, userID, tenantID, domain.PermissionAuditRead, "*").Return(nil).Once()
		mockRepo.On("List", mock.Anything, mock.Anything).Return(nil, 0, fmt.Errorf("db fail")).Once()
		_, _, err := svc.SearchLogs(ctx, query)
		require.Error(t, err)
	})

	mockRepo.AssertExpectations(t)
}

func TestCloudLogsServiceRunRetentionPolicyUnit(t *testing.T) {
	mockRepo := new(MockLogRepository)
	mockRBAC := new(MockRBACService)
	svc := services.NewCloudLogsService(mockRepo, mockRBAC, slog.Default())
	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	t.Run("Success", func(t *testing.T) {
		mockRBAC.On("Authorize", mock.Anything, userID, tenantID, domain.PermissionFullAccess, "*").Return(nil).Once()
		mockRepo.On("DeleteByAge", mock.Anything, 30).Return(nil)

		err := svc.RunRetentionPolicy(ctx, 30)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}
