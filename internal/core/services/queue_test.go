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
)

func TestQueueService_CreateQueue(t *testing.T) {
	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockQueueRepo)
		mockEventSvc := new(MockEventService)
		auditSvc := new(services.MockAuditService)
		svc := services.NewQueueService(mockRepo, mockEventSvc, auditSvc)

		mockRepo.On("GetByName", ctx, "test-queue", userID).Return(nil, nil)
		mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Queue")).Return(nil)
		mockEventSvc.On("RecordEvent", ctx, "QUEUE_CREATED", mock.Anything, "QUEUE", mock.Anything).Return(nil)
		auditSvc.On("Log", ctx, userID, "queue.create", "queue", mock.Anything, mock.Anything).Return(nil)

		q, err := svc.CreateQueue(ctx, "test-queue", nil)

		assert.NoError(t, err)
		assert.NotNil(t, q)
		assert.Equal(t, "test-queue", q.Name)
		assert.Equal(t, userID, q.UserID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("AlreadyExists", func(t *testing.T) {
		mockRepo := new(MockQueueRepo)
		mockEventSvc := new(MockEventService)
		auditSvc := new(services.MockAuditService)
		svc := services.NewQueueService(mockRepo, mockEventSvc, auditSvc)

		existing := &domain.Queue{ID: uuid.New(), Name: "test-queue", UserID: userID}
		mockRepo.On("GetByName", ctx, "test-queue", userID).Return(existing, nil)

		_, err := svc.CreateQueue(ctx, "test-queue", nil)

		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "already exists")
		}
	})

	t.Run("Unauthorized", func(t *testing.T) {
		mockRepo := new(MockQueueRepo)
		mockEventSvc := new(MockEventService)
		auditSvc := new(services.MockAuditService)
		svc := services.NewQueueService(mockRepo, mockEventSvc, auditSvc)

		_, err := svc.CreateQueue(context.Background(), "test-queue", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})
}

func TestQueueService_SendMessage(t *testing.T) {
	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	queueID := uuid.New()
	queue := &domain.Queue{ID: queueID, UserID: userID, MaxMessageSize: 100}

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockQueueRepo)
		mockEventSvc := new(MockEventService)
		auditSvc := new(services.MockAuditService)
		svc := services.NewQueueService(mockRepo, mockEventSvc, auditSvc)

		mockRepo.On("GetByID", ctx, queueID, userID).Return(queue, nil)
		mockRepo.On("SendMessage", ctx, queueID, "test message").Return(&domain.Message{ID: uuid.New()}, nil)
		mockEventSvc.On("RecordEvent", ctx, "MESSAGE_SENT", mock.Anything, "MESSAGE", mock.Anything).Return(nil)
		auditSvc.On("Log", ctx, userID, "queue.message_send", "queue", queueID.String(), mock.Anything).Return(nil)

		msg, err := svc.SendMessage(ctx, queueID, "test message")

		assert.NoError(t, err)
		assert.NotNil(t, msg)
		mockRepo.AssertExpectations(t)
	})

	t.Run("MessageTooLarge", func(t *testing.T) {
		mockRepo := new(MockQueueRepo)
		mockEventSvc := new(MockEventService)
		auditSvc := new(services.MockAuditService)
		svc := services.NewQueueService(mockRepo, mockEventSvc, auditSvc)

		mockRepo.On("GetByID", ctx, queueID, userID).Return(queue, nil)

		largeBody := string(make([]byte, 101))
		_, err := svc.SendMessage(ctx, queueID, largeBody)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "message size exceeds limit")
	})
}

func TestQueueService_ReceiveMessages(t *testing.T) {
	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	queueID := uuid.New()
	queue := &domain.Queue{ID: queueID, UserID: userID, VisibilityTimeout: 30}

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockQueueRepo)
		mockEventSvc := new(MockEventService)
		auditSvc := new(services.MockAuditService)
		svc := services.NewQueueService(mockRepo, mockEventSvc, auditSvc)

		mockRepo.On("GetByID", ctx, queueID, userID).Return(queue, nil)
		msgs := []*domain.Message{{ID: uuid.New()}, {ID: uuid.New()}}
		mockRepo.On("ReceiveMessages", ctx, queueID, 2, 30).Return(msgs, nil)
		mockEventSvc.On("RecordEvent", ctx, "MESSAGE_RECEIVED", mock.Anything, "MESSAGE", mock.Anything).Return(nil)

		received, err := svc.ReceiveMessages(ctx, queueID, 2)

		assert.NoError(t, err)
		assert.Len(t, received, 2)
		mockRepo.AssertExpectations(t)
	})
}
