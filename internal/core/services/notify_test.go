package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupNotifyServiceTest(t *testing.T) (*MockNotifyRepo, *MockQueueService, *MockEventService, *MockAuditService, ports.NotifyService) {
	notifyRepo := new(MockNotifyRepo)
	queueSvc := new(MockQueueService)
	eventSvc := new(MockEventService)
	auditSvc := new(MockAuditService)
	svc := services.NewNotifyService(notifyRepo, queueSvc, eventSvc, auditSvc)
	return notifyRepo, queueSvc, eventSvc, auditSvc, svc
}

func TestNotifyService_CreateTopic(t *testing.T) {
	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)

	t.Run("Success", func(t *testing.T) {
		notifyRepo, _, eventSvc, auditSvc, svc := setupNotifyServiceTest(t)
		defer notifyRepo.AssertExpectations(t)
		defer eventSvc.AssertExpectations(t)
		defer auditSvc.AssertExpectations(t)

		topicName := "test-topic"

		notifyRepo.On("GetTopicByName", ctx, topicName, userID).Return(nil, assert.AnError).Once()
		notifyRepo.On("CreateTopic", ctx, mock.MatchedBy(func(t *domain.Topic) bool {
			return t.Name == topicName && t.UserID == userID
		})).Return(nil).Once()
		eventSvc.On("RecordEvent", ctx, "TOPIC_CREATED", mock.Anything, "TOPIC", mock.Anything).Return(nil).Once()
		auditSvc.On("Log", ctx, userID, "notify.topic_create", "topic", mock.Anything, mock.Anything).Return(nil).Once()

		topic, err := svc.CreateTopic(ctx, topicName)
		require.NoError(t, err)
		assert.Equal(t, topicName, topic.Name)
		assert.Equal(t, userID, topic.UserID)
		assert.NotEmpty(t, topic.ARN)
	})

	t.Run("TopicAlreadyExists", func(t *testing.T) {
		notifyRepo, _, _, _, svc := setupNotifyServiceTest(t)
		defer notifyRepo.AssertExpectations(t)

		existingTopic := &domain.Topic{ID: uuid.New(), Name: "existing-topic"}
		notifyRepo.On("GetTopicByName", ctx, "existing-topic", userID).Return(existingTopic, nil).Once()

		_, err := svc.CreateTopic(ctx, "existing-topic")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}

func TestNotifyService_Subscribe(t *testing.T) {
	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	topicID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		notifyRepo, _, eventSvc, auditSvc, svc := setupNotifyServiceTest(t)
		defer notifyRepo.AssertExpectations(t)
		defer eventSvc.AssertExpectations(t)
		defer auditSvc.AssertExpectations(t)

		topic := &domain.Topic{ID: topicID, UserID: userID, Name: "test-topic"}

		notifyRepo.On("GetTopicByID", ctx, topicID, userID).Return(topic, nil).Once()
		notifyRepo.On("CreateSubscription", ctx, mock.MatchedBy(func(s *domain.Subscription) bool {
			return s.TopicID == topicID && s.Protocol == domain.ProtocolQueue
		})).Return(nil).Once()
		eventSvc.On("RecordEvent", ctx, "SUBSCRIPTION_CREATED", mock.Anything, "SUBSCRIPTION", mock.Anything).Return(nil).Once()
		auditSvc.On("Log", ctx, userID, "notify.subscribe", "subscription", mock.Anything, mock.Anything).Return(nil).Once()

		sub, err := svc.Subscribe(ctx, topicID, domain.ProtocolQueue, "queue-endpoint")
		require.NoError(t, err)
		assert.Equal(t, topicID, sub.TopicID)
		assert.Equal(t, domain.ProtocolQueue, sub.Protocol)
	})
}

func TestNotifyService_Publish(t *testing.T) {
	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	topicID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		notifyRepo, _, eventSvc, auditSvc, svc := setupNotifyServiceTest(t)
		defer notifyRepo.AssertExpectations(t)
		defer eventSvc.AssertExpectations(t)
		defer auditSvc.AssertExpectations(t)

		topic := &domain.Topic{ID: topicID, UserID: userID, Name: "test-topic"}
		messageBody := "test message"

		notifyRepo.On("GetTopicByID", ctx, topicID, userID).Return(topic, nil).Once()
		notifyRepo.On("SaveMessage", ctx, mock.MatchedBy(func(m *domain.NotifyMessage) bool {
			return m.TopicID == topicID && m.Body == messageBody
		})).Return(nil).Once()
		notifyRepo.On("ListSubscriptions", ctx, topicID).Return([]*domain.Subscription{}, nil).Once()
		eventSvc.On("RecordEvent", ctx, "TOPIC_PUBLISHED", topicID.String(), "TOPIC", mock.Anything).Return(nil).Once()
		auditSvc.On("Log", ctx, userID, "notify.publish", "topic", topicID.String(), mock.Anything).Return(nil).Once()

		err := svc.Publish(ctx, topicID, messageBody)
		require.NoError(t, err)
	})
}

func TestNotifyService_ListTopics(t *testing.T) {
	notifyRepo, _, _, _, svc := setupNotifyServiceTest(t)
	defer notifyRepo.AssertExpectations(t)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)

	topics := []*domain.Topic{
		{ID: uuid.New(), Name: "topic1", UserID: userID},
		{ID: uuid.New(), Name: "topic2", UserID: userID},
	}

	notifyRepo.On("ListTopics", ctx, userID).Return(topics, nil).Once()

	result, err := svc.ListTopics(ctx)
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestNotifyService_DeleteTopic(t *testing.T) {
	notifyRepo, _, eventSvc, auditSvc, svc := setupNotifyServiceTest(t)
	defer notifyRepo.AssertExpectations(t)
	defer eventSvc.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	topicID := uuid.New()

	topic := &domain.Topic{ID: topicID, UserID: userID, Name: "test-topic"}

	notifyRepo.On("GetTopicByID", ctx, topicID, userID).Return(topic, nil).Once()
	notifyRepo.On("DeleteTopic", ctx, topicID).Return(nil).Once()
	eventSvc.On("RecordEvent", ctx, "TOPIC_DELETED", topicID.String(), "TOPIC", mock.Anything).Return(nil).Once()
	auditSvc.On("Log", ctx, userID, "notify.topic_delete", "topic", topicID.String(), mock.Anything).Return(nil).Once()

	err := svc.DeleteTopic(ctx, topicID)
	require.NoError(t, err)
}

func TestNotifyService_Unsubscribe(t *testing.T) {
	notifyRepo, _, eventSvc, auditSvc, svc := setupNotifyServiceTest(t)
	defer notifyRepo.AssertExpectations(t)
	defer eventSvc.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	subID := uuid.New()
	topicID := uuid.New()

	sub := &domain.Subscription{ID: subID, UserID: userID, TopicID: topicID}

	notifyRepo.On("GetSubscriptionByID", ctx, subID, userID).Return(sub, nil).Once()
	notifyRepo.On("DeleteSubscription", ctx, subID).Return(nil).Once()
	eventSvc.On("RecordEvent", ctx, "SUBSCRIPTION_DELETED", subID.String(), "SUBSCRIPTION", mock.Anything).Return(nil).Once()
	auditSvc.On("Log", ctx, userID, "notify.unsubscribe", "subscription", subID.String(), mock.Anything).Return(nil).Once()

	err := svc.Unsubscribe(ctx, subID)
	require.NoError(t, err)
}
