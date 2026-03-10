package services_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupQueueServiceTest(t *testing.T) (ports.QueueService, *postgres.PostgresQueueRepository, context.Context) {
	t.Helper()
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewPostgresQueueRepository(db)
	rbacSvc := new(MockRBACService)
	rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	eventRepo := postgres.NewEventRepository(db)
	eventSvc := services.NewEventService(services.EventServiceParams{
		Repo:    eventRepo,
		RBACSvc: rbacSvc,
		Hub:     nil,
		Logger:  nil,
	})
	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(services.AuditServiceParams{
		Repo:    auditRepo,
		RBACSvc: rbacSvc,
	})

	svc := services.NewQueueService(repo, rbacSvc, eventSvc, auditSvc, slog.Default())
	return svc, repo.(*postgres.PostgresQueueRepository), ctx
}

func TestQueueService_Lifecycle(t *testing.T) {
	svc, repo, ctx := setupQueueServiceTest(t)

	t.Run("CreateAndGet", func(t *testing.T) {
		name := "test-queue"
		q, err := svc.CreateQueue(ctx, name, nil)
		require.NoError(t, err)
		assert.Equal(t, name, q.Name)

		fetched, err := svc.GetQueue(ctx, q.ID)
		require.NoError(t, err)
		assert.Equal(t, q.ID, fetched.ID)
	})

	t.Run("Messages", func(t *testing.T) {
		q, _ := svc.CreateQueue(ctx, "msg-queue", nil)

		payload := "hello world"
		msg, err := svc.SendMessage(ctx, q.ID, payload)
		require.NoError(t, err)
		assert.Equal(t, payload, msg.Body)

		// Receive
		received, err := svc.ReceiveMessages(ctx, q.ID, 1)
		require.NoError(t, err)
		assert.Len(t, received, 1)
		assert.Equal(t, msg.ID, received[0].ID)

		// Delete
		err = svc.DeleteMessage(ctx, q.ID, received[0].ReceiptHandle)
		require.NoError(t, err)

		// Verify gone
		visible, inFlight, _ := repo.GetQueueStats(ctx, q.ID)
		assert.Equal(t, 0, visible)
		assert.Equal(t, 0, inFlight)
	})
}
