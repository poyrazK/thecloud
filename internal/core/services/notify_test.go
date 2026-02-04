package services_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupNotifyServiceIntegrationTest(t *testing.T) (ports.NotifyService, ports.NotifyRepository, ports.QueueService, context.Context) {
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	notifyRepo := postgres.NewPostgresNotifyRepository(db)
	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(auditRepo)

	eventRepo := postgres.NewEventRepository(db)
	eventSvc := services.NewEventService(eventRepo, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))

	queueRepo := postgres.NewPostgresQueueRepository(db)
	queueSvc := services.NewQueueService(queueRepo, eventSvc, auditSvc)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := services.NewNotifyService(notifyRepo, queueSvc, eventSvc, auditSvc, logger)

	return svc, notifyRepo, queueSvc, ctx
}

func TestNotifyService_Integration(t *testing.T) {
	svc, _, queueSvc, ctx := setupNotifyServiceIntegrationTest(t)
	_ = appcontext.UserIDFromContext(ctx)

	t.Run("TopicLifecycle", func(t *testing.T) {
		name := "integration-topic"

		topic, err := svc.CreateTopic(ctx, name)
		assert.NoError(t, err)
		assert.Equal(t, name, topic.Name)

		// List
		topics, err := svc.ListTopics(ctx)
		assert.NoError(t, err)
		assert.Len(t, topics, 1)

		// Delete
		err = svc.DeleteTopic(ctx, topic.ID)
		assert.NoError(t, err)

		topics, _ = svc.ListTopics(ctx)
		assert.Len(t, topics, 0)
	})

	t.Run("SubscriptionAndPublishing", func(t *testing.T) {
		topic, _ := svc.CreateTopic(ctx, "pub-sub-topic")

		// 1. Queue Subscription
		q, err := queueSvc.CreateQueue(ctx, "sub-queue", nil)
		require.NoError(t, err)

		sub, err := svc.Subscribe(ctx, topic.ID, domain.ProtocolQueue, q.ID.String())
		assert.NoError(t, err)
		assert.NotNil(t, sub)

		// 2. Webhook Subscription
		webhookReceived := make(chan string, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			webhookReceived <- string(body)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		_, err = svc.Subscribe(ctx, topic.ID, domain.ProtocolWebhook, server.URL)
		assert.NoError(t, err)

		// 3. Publish
		msgBody := "hello integration"
		err = svc.Publish(ctx, topic.ID, msgBody)
		assert.NoError(t, err)

		// Wait for async delivery
		time.Sleep(200 * time.Millisecond)

		// Verify Webhook delivery
		select {
		case received := <-webhookReceived:
			assert.Equal(t, msgBody, received)
		case <-time.After(2 * time.Second):
			t.Error("webhook delivery timed out")
		}

		// Verify Queue delivery
		msgs, err := queueSvc.ReceiveMessages(ctx, q.ID, 1)
		assert.NoError(t, err)
		assert.Len(t, msgs, 1)
		assert.Equal(t, msgBody, msgs[0].Body)

		// Unsubscribe
		err = svc.Unsubscribe(ctx, sub.ID)
		assert.NoError(t, err)

		subs, _ := svc.ListSubscriptions(ctx, topic.ID)
		assert.Len(t, subs, 1) // Only webhook left
	})
}
