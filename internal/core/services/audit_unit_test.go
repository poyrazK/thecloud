package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuditService_Unit(t *testing.T) {
	mockRepo := new(MockAuditRepository)
	mockRBAC := new(MockRBACService)
	svc := services.NewAuditService(services.AuditServiceParams{
		Repo:    mockRepo,
		RBACSvc: mockRBAC,
	})

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	t.Run("Log_Success", func(t *testing.T) {
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		err := svc.Log(ctx, userID, "test.action", "resource", "res-123", nil)
		require.NoError(t, err)
	})

	t.Run("Log_Failure", func(t *testing.T) {
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(assert.AnError).Once()
		err := svc.Log(ctx, userID, "test.action", "resource", "res-123", nil)
		require.Error(t, err)
	})

	t.Run("ListLogs_Success", func(t *testing.T) {
		mockRBAC.On("Authorize", mock.Anything, userID, tenantID, domain.PermissionAuditRead, "*").Return(nil).Once()
		mockRepo.On("ListByUserID", mock.Anything, userID, 50).Return([]*domain.AuditLog{}, nil).Once()

		logs, err := svc.ListLogs(ctx, userID, 0)
		require.NoError(t, err)
		assert.NotNil(t, logs)
	})

	t.Run("ListLogs_Forbidden", func(t *testing.T) {
		mockRBAC.On("Authorize", mock.Anything, userID, tenantID, domain.PermissionAuditRead, "*").Return(assert.AnError).Once()

		logs, err := svc.ListLogs(ctx, userID, 50)
		require.Error(t, err)
		assert.Nil(t, logs)
	})
}
