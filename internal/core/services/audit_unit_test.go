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
	repo := new(MockAuditRepo)
	svc := services.NewAuditService(repo)
	ctx := context.Background()
	userID := uuid.New()

	t.Run("Log", func(t *testing.T) {
		repo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		err := svc.Log(ctx, userID, "test.action", "instance", "123", nil)
		require.NoError(t, err)
	})

	t.Run("ListLogs", func(t *testing.T) {
		repo.On("ListByUserID", mock.Anything, userID, 10).Return([]*domain.AuditLog{}, nil).Once()
		logs, err := svc.ListLogs(ctx, userID, 10)
		require.NoError(t, err)
		assert.NotNil(t, logs)
	})
}
