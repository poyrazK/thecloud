package services_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupQueueServiceTest(t *testing.T) (ports.QueueService, ports.QueueRepository, context.Context, *pgxpool.Pool) {
	db := setupDB(t)
	// We don't close DB here, caller should do it via defer

	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewPostgresQueueRepository(db)

	eventRepo := postgres.NewEventRepository(db)
	eventSvc := services.NewEventService(eventRepo, nil, slog.Default())

	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(auditRepo)

	svc := services.NewQueueService(repo, eventSvc, auditSvc)

	return svc, repo, ctx, db
}

func TestQueueServiceCreateQueue(t *testing.T) {
	svc, repo, ctx, db := setupQueueServiceTest(t)
	defer db.Close()
	userID := appcontext.UserIDFromContext(ctx)

	t.Run("success", func(t *testing.T) {
		opts := &ports.CreateQueueOptions{}
		q, err := svc.CreateQueue(ctx, "test-queue", opts)

		assert.NoError(t, err)
		assert.NotNil(t, q)
		assert.Equal(t, "test-queue", q.Name)
		assert.Equal(t, userID, q.UserID)

		// Verify DB
		fetched, err := repo.GetByName(ctx, "test-queue", userID)
		assert.NoError(t, err)
		assert.Equal(t, q.ID, fetched.ID)
	})

	t.Run("already exists", func(t *testing.T) {
		_, err := svc.CreateQueue(ctx, "test-queue", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("unauthorized", func(t *testing.T) {
		_, err := svc.CreateQueue(context.Background(), "no-auth", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})
}

func TestQueueServiceSendMessage(t *testing.T) {
	svc, _, ctx, db := setupQueueServiceTest(t)
	defer db.Close()

	q, err := svc.CreateQueue(ctx, "msg-queue", nil)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		msg, err := svc.SendMessage(ctx, q.ID, "hello world")
		assert.NoError(t, err)
		assert.NotNil(t, msg)
		assert.Equal(t, "hello world", msg.Body)
		// Receipt handle is only present when received
	})
}

func TestQueueServiceReceiveMessages(t *testing.T) {
	svc, _, ctx, db := setupQueueServiceTest(t)
	defer db.Close()

	q, err := svc.CreateQueue(ctx, "recv-queue", nil)
	require.NoError(t, err)

	_, err = svc.SendMessage(ctx, q.ID, "msg1")
	require.NoError(t, err)
	_, err = svc.SendMessage(ctx, q.ID, "msg2")
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		msgs, err := svc.ReceiveMessages(ctx, q.ID, 10)
		assert.NoError(t, err)
		assert.Len(t, msgs, 2)

		// Should be invisible now? Depends on repo impl (visibility timeout).
		// Try receiving again immediately
		msgs2, err := svc.ReceiveMessages(ctx, q.ID, 10)
		assert.NoError(t, err)
		assert.Len(t, msgs2, 0, "messages should be invisible")
	})
}

func TestQueueServiceDeleteMessage(t *testing.T) {
	svc, repo, ctx, db := setupQueueServiceTest(t)
	defer db.Close()

	q, err := svc.CreateQueue(ctx, "del-msg-queue", nil)
	require.NoError(t, err)

	_, err = svc.SendMessage(ctx, q.ID, "to delete")
	require.NoError(t, err)

	msgs, err := svc.ReceiveMessages(ctx, q.ID, 1)
	require.NoError(t, err)
	require.Len(t, msgs, 1)
	receipt := msgs[0].ReceiptHandle

	err = svc.DeleteMessage(ctx, q.ID, receipt)
	assert.NoError(t, err)

	v, inv, err := repo.(*postgres.PostgresQueueRepository).GetQueueStats(ctx, q.ID)
	assert.NoError(t, err)
	assert.Equal(t, 0, v)
	assert.Equal(t, 0, inv)
}

func TestQueueServicePurgeQueue(t *testing.T) {
	svc, repo, ctx, db := setupQueueServiceTest(t)
	defer db.Close()

	q, err := svc.CreateQueue(ctx, "purge-queue", nil)
	require.NoError(t, err)

	_, _ = svc.SendMessage(ctx, q.ID, "m1")
	_, _ = svc.SendMessage(ctx, q.ID, "m2")

	err = svc.PurgeQueue(ctx, q.ID)
	assert.NoError(t, err)

	v, inv, err := repo.(*postgres.PostgresQueueRepository).GetQueueStats(ctx, q.ID)
	assert.NoError(t, err)
	assert.Equal(t, 0, v)
	assert.Equal(t, 0, inv)
}

func TestQueueServiceDeleteQueue(t *testing.T) {
	svc, repo, ctx, db := setupQueueServiceTest(t)
	defer db.Close()
	userID := appcontext.UserIDFromContext(ctx)

	q, err := svc.CreateQueue(ctx, "to-delete", nil)
	require.NoError(t, err)

	err = svc.DeleteQueue(ctx, q.ID)
	assert.NoError(t, err)

	fetched, err := repo.GetByID(ctx, q.ID, userID)
	assert.NoError(t, err)
	assert.Nil(t, fetched)
}

func TestQueueServiceListQueues(t *testing.T) {
	svc, _, ctx, db := setupQueueServiceTest(t)
	defer db.Close()

	_, err := svc.CreateQueue(ctx, "q1", nil)
	require.NoError(t, err)
	_, err = svc.CreateQueue(ctx, "q2", nil)
	require.NoError(t, err)

	queues, err := svc.ListQueues(ctx)
	assert.NoError(t, err)
	assert.Len(t, queues, 2)
}
