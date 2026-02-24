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

func TestQueueServiceUnit(t *testing.T) {
	mockRepo := new(MockQueueRepository)
	mockEventSvc := new(MockEventService)
	mockAuditSvc := new(MockAuditService)
	svc := services.NewQueueService(mockRepo, mockEventSvc, mockAuditSvc)

	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)

	t.Run("CreateQueue", func(t *testing.T) {
		mockRepo.On("GetByName", mock.Anything, "my-queue", userID).Return(nil, nil).Once()
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		mockEventSvc.On("RecordEvent", mock.Anything, "QUEUE_CREATED", mock.Anything, "QUEUE", mock.Anything).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "queue.create", "queue", mock.Anything, mock.Anything).Return(nil).Once()

		q, err := svc.CreateQueue(ctx, "my-queue", nil)
		require.NoError(t, err)
		assert.NotNil(t, q)
		assert.Equal(t, "my-queue", q.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("SendMessage", func(t *testing.T) {
		qID := uuid.New()
		mockRepo.On("GetByID", mock.Anything, qID, userID).Return(&domain.Queue{ID: qID, MaxMessageSize: 100}, nil).Once()
		mockRepo.On("SendMessage", mock.Anything, qID, "hello").Return(&domain.Message{ID: uuid.New()}, nil).Once()
		mockEventSvc.On("RecordEvent", mock.Anything, "MESSAGE_SENT", mock.Anything, "MESSAGE", mock.Anything).Return(nil).Once()

		msg, err := svc.SendMessage(ctx, qID, "hello")
		require.NoError(t, err)
		assert.NotNil(t, msg)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ReceiveMessages", func(t *testing.T) {
		qID := uuid.New()
		mockRepo.On("GetByID", mock.Anything, qID, userID).Return(&domain.Queue{ID: qID, VisibilityTimeout: 30}, nil).Once()
		mockRepo.On("ReceiveMessages", mock.Anything, qID, 5, 30).Return([]*domain.Message{{ID: uuid.New()}}, nil).Once()
		mockEventSvc.On("RecordEvent", mock.Anything, "MESSAGE_RECEIVED", mock.Anything, "MESSAGE", mock.Anything).Return(nil).Once()

		msgs, err := svc.ReceiveMessages(ctx, qID, 5)
		require.NoError(t, err)
		assert.Len(t, msgs, 1)
	})
}
