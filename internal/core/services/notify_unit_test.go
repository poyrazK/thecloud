package services_test

import (
	"context"
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

const testTopicName = "my-topic"

func TestNotifyServiceUnit(t *testing.T) {
	mockRepo := new(MockNotifyRepo)
	mockQueueSvc := new(MockQueueService)
	mockEventSvc := new(MockEventService)
	mockAuditSvc := new(MockAuditService)
	svc := services.NewNotifyService(mockRepo, mockQueueSvc, mockEventSvc, mockAuditSvc, slog.Default())

	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)

	t.Run("CreateTopic", func(t *testing.T) {
		mockRepo.On("GetTopicByName", mock.Anything, testTopicName, userID).Return(nil, nil).Once()
		mockRepo.On("CreateTopic", mock.Anything, mock.Anything).Return(nil).Once()
		mockEventSvc.On("RecordEvent", mock.Anything, "TOPIC_CREATED", mock.Anything, "TOPIC", mock.Anything).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "notify.topic_create", "topic", mock.Anything, mock.Anything).Return(nil).Once()

		topic, err := svc.CreateTopic(ctx, testTopicName)
		require.NoError(t, err)
		assert.NotNil(t, topic)
		assert.Equal(t, testTopicName, topic.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Subscribe", func(t *testing.T) {
		topicID := uuid.New()
		mockRepo.On("GetTopicByID", mock.Anything, topicID, userID).Return(&domain.Topic{ID: topicID}, nil).Once()
		mockRepo.On("CreateSubscription", mock.Anything, mock.Anything).Return(nil).Once()
		mockEventSvc.On("RecordEvent", mock.Anything, "SUBSCRIPTION_CREATED", mock.Anything, "SUBSCRIPTION", mock.Anything).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "notify.subscribe", "subscription", mock.Anything, mock.Anything).Return(nil).Once()

		sub, err := svc.Subscribe(ctx, topicID, domain.ProtocolWebhook, "https://example.com/hook")
		require.NoError(t, err)
		assert.NotNil(t, sub)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Publish", func(t *testing.T) {
		topicID := uuid.New()
		mockRepo.On("GetTopicByID", mock.Anything, topicID, userID).Return(&domain.Topic{ID: topicID, UserID: userID}, nil).Once()
		mockRepo.On("SaveMessage", mock.Anything, mock.Anything).Return(nil).Once()
		mockRepo.On("ListSubscriptions", mock.Anything, topicID).Return([]*domain.Subscription{}, nil).Once()
		mockEventSvc.On("RecordEvent", mock.Anything, "TOPIC_PUBLISHED", mock.Anything, "TOPIC", mock.Anything).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "notify.publish", "topic", topicID.String(), mock.Anything).Return(nil).Once()

		err := svc.Publish(ctx, topicID, "hello")
		require.NoError(t, err)
	})
}
