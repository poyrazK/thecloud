package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuditService_Unit(t *testing.T) {
	mockRepo := new(MockAuditRepo)
	svc := services.NewAuditService(mockRepo)

	ctx := context.Background()
	userID := uuid.New()

	t.Run("Log_Success", func(t *testing.T) {
		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(l *domain.AuditLog) bool {
			return l.UserID == userID && l.Action == "test.action"
		})).Return(nil).Once()

		err := svc.Log(ctx, userID, "test.action", "instance", "inst-1", map[string]interface{}{"foo": "bar"})
		require.NoError(t, err)
	})

	t.Run("ListLogs_LimitDefault", func(t *testing.T) {
		mockRepo.On("ListByUserID", mock.Anything, userID, 50).Return([]*domain.AuditLog{}, nil).Once()

		logs, err := svc.ListLogs(ctx, userID, 0)
		require.NoError(t, err)
		assert.NotNil(t, logs)
	})

	t.Run("ListLogs_CustomLimit", func(t *testing.T) {
		mockRepo.On("ListByUserID", mock.Anything, userID, 10).Return([]*domain.AuditLog{}, nil).Once()

		_, err := svc.ListLogs(ctx, userID, 10)
		require.NoError(t, err)
	})
}
