package services_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testTopicName     = "test-topic"
	existingTopicName = "existing-topic"
	testMsg           = "test message"
	auditPublish      = "notify.publish"
)

func setupNotifyServiceTest(_ *testing.T) (*MockNotifyRepo, *MockQueueService, *MockEventService, *MockAuditService, ports.NotifyService) {
	notifyRepo := new(MockNotifyRepo)
	queueSvc := new(MockQueueService)
	eventSvc := new(MockEventService)
	auditSvc := new(MockAuditService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := services.NewNotifyService(notifyRepo, queueSvc, eventSvc, auditSvc, logger)
	return notifyRepo, queueSvc, eventSvc, auditSvc, svc
}

func TestNotifyServiceCreateTopic(t *testing.T) {
	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)

	t.Run("Success", func(t *testing.T) {
		notifyRepo, _, eventSvc, auditSvc, svc := setupNotifyServiceTest(t)
		defer notifyRepo.AssertExpectations(t)
		defer eventSvc.AssertExpectations(t)
		defer auditSvc.AssertExpectations(t)

		topicName := testTopicName

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

		existingTopic := &domain.Topic{ID: uuid.New(), Name: existingTopicName}
		notifyRepo.On("GetTopicByName", ctx, existingTopicName, userID).Return(existingTopic, nil).Once()

		_, err := svc.CreateTopic(ctx, existingTopicName)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}

func TestNotifyServiceSubscribe(t *testing.T) {
	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	topicID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		notifyRepo, _, eventSvc, auditSvc, svc := setupNotifyServiceTest(t)
		defer notifyRepo.AssertExpectations(t)
		defer eventSvc.AssertExpectations(t)
		defer auditSvc.AssertExpectations(t)

		topic := &domain.Topic{ID: topicID, UserID: userID, Name: testTopicName}

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

func TestNotifyServicePublish(t *testing.T) {
	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	topicID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		notifyRepo, _, eventSvc, auditSvc, svc := setupNotifyServiceTest(t)
		defer notifyRepo.AssertExpectations(t)
		defer eventSvc.AssertExpectations(t)
		defer auditSvc.AssertExpectations(t)

		topic := &domain.Topic{ID: topicID, UserID: userID, Name: testTopicName}
		messageBody := testMsg

		notifyRepo.On("GetTopicByID", ctx, topicID, userID).Return(topic, nil).Once()
		notifyRepo.On("SaveMessage", ctx, mock.MatchedBy(func(m *domain.NotifyMessage) bool {
			return m.TopicID == topicID && m.Body == messageBody
		})).Return(nil).Once()
		notifyRepo.On("ListSubscriptions", ctx, topicID).Return([]*domain.Subscription{}, nil).Once()
		eventSvc.On("RecordEvent", ctx, "TOPIC_PUBLISHED", topicID.String(), "TOPIC", mock.Anything).Return(nil).Once()
		auditSvc.On("Log", ctx, userID, auditPublish, "topic", topicID.String(), mock.Anything).Return(nil).Once()

		err := svc.Publish(ctx, topicID, messageBody)
		require.NoError(t, err)
	})
}

func TestNotifyServiceListTopics(t *testing.T) {
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

func TestNotifyServiceDeleteTopic(t *testing.T) {
	notifyRepo, _, eventSvc, auditSvc, svc := setupNotifyServiceTest(t)
	defer notifyRepo.AssertExpectations(t)
	defer eventSvc.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	topicID := uuid.New()

	topic := &domain.Topic{ID: topicID, UserID: userID, Name: testTopicName}

	notifyRepo.On("GetTopicByID", ctx, topicID, userID).Return(topic, nil).Once()
	notifyRepo.On("DeleteTopic", ctx, topicID).Return(nil).Once()
	eventSvc.On("RecordEvent", ctx, "TOPIC_DELETED", topicID.String(), "TOPIC", mock.Anything).Return(nil).Once()
	auditSvc.On("Log", ctx, userID, "notify.topic_delete", "topic", topicID.String(), mock.Anything).Return(nil).Once()

	err := svc.DeleteTopic(ctx, topicID)
	require.NoError(t, err)
}

func TestNotifyServiceUnsubscribe(t *testing.T) {
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
func TestNotifyServiceListSubscriptions(t *testing.T) {
	notifyRepo, _, _, _, svc := setupNotifyServiceTest(t)
	defer notifyRepo.AssertExpectations(t)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	topicID := uuid.New()
	topic := &domain.Topic{ID: topicID, UserID: userID}
	subs := []*domain.Subscription{{ID: uuid.New(), TopicID: topicID}}

	notifyRepo.On("GetTopicByID", ctx, topicID, userID).Return(topic, nil).Once()
	notifyRepo.On("ListSubscriptions", ctx, topicID).Return(subs, nil).Once()

	result, err := svc.ListSubscriptions(ctx, topicID)
	require.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestNotifyServiceDeliverToQueue(t *testing.T) {
	notifyRepo, queueSvc, eventSvc, auditSvc, svc := setupNotifyServiceTest(t)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	topicID := uuid.New()
	qID := uuid.New()
	topic := &domain.Topic{ID: topicID, UserID: userID, Name: testTopicName}
	sub := &domain.Subscription{
		ID:       uuid.New(),
		UserID:   userID,
		TopicID:  topicID,
		Protocol: domain.ProtocolQueue,
		Endpoint: qID.String(),
	}
	body := testMsg

	notifyRepo.On("GetTopicByID", ctx, topicID, userID).Return(topic, nil).Once()
	notifyRepo.On("SaveMessage", ctx, mock.Anything).Return(nil).Once()
	notifyRepo.On("ListSubscriptions", ctx, topicID).Return([]*domain.Subscription{sub}, nil).Once()
	queueSvc.On("SendMessage", mock.Anything, qID, body).Return(&domain.Message{}, nil).Once()
	eventSvc.On("RecordEvent", ctx, "TOPIC_PUBLISHED", topicID.String(), "TOPIC", mock.Anything).Return(nil).Once()
	auditSvc.On("Log", ctx, userID, auditPublish, "topic", topicID.String(), mock.Anything).Return(nil).Once()

	err := svc.Publish(ctx, topicID, body)
	require.NoError(t, err)

	// Wait for goroutine
	time.Sleep(100 * time.Millisecond)
	queueSvc.AssertExpectations(t)
	eventSvc.AssertExpectations(t)
	auditSvc.AssertExpectations(t)
}

func TestNotifyServiceDeliverToWebhook(t *testing.T) {
	notifyRepo, _, eventSvc, auditSvc, svc := setupNotifyServiceTest(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	topicID := uuid.New()
	topic := &domain.Topic{ID: topicID, UserID: userID, Name: testTopicName}
	sub := &domain.Subscription{
		ID:       uuid.New(),
		UserID:   userID,
		TopicID:  topicID,
		Protocol: domain.ProtocolWebhook,
		Endpoint: server.URL,
	}
	body := testMsg

	notifyRepo.On("GetTopicByID", ctx, topicID, userID).Return(topic, nil).Once()
	notifyRepo.On("SaveMessage", ctx, mock.Anything).Return(nil).Once()
	notifyRepo.On("ListSubscriptions", ctx, topicID).Return([]*domain.Subscription{sub}, nil).Once()
	eventSvc.On("RecordEvent", ctx, "TOPIC_PUBLISHED", topicID.String(), "TOPIC", mock.Anything).Return(nil).Once()
	auditSvc.On("Log", ctx, userID, auditPublish, "topic", topicID.String(), mock.Anything).Return(nil).Once()

	err := svc.Publish(ctx, topicID, body)
	require.NoError(t, err)

	// Wait for goroutine
	time.Sleep(100 * time.Millisecond)
	notifyRepo.AssertExpectations(t)
}
