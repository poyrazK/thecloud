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

func TestNotifyService_CreateTopic(t *testing.T) {
	repo := new(MockNotifyRepo)
	eventSvc := new(MockEventService)
	auditSvc := new(services.MockAuditService)
	svc := services.NewNotifyService(repo, nil, eventSvc, auditSvc)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)

	repo.On("GetTopicByName", ctx, "test-topic", userID).Return(nil, nil)
	repo.On("CreateTopic", ctx, mock.AnythingOfType("*domain.Topic")).Return(nil)
	eventSvc.On("RecordEvent", ctx, "TOPIC_CREATED", mock.Anything, "TOPIC", mock.Anything).Return(nil)
	auditSvc.On("Log", ctx, userID, "notify.topic_create", "topic", mock.Anything, mock.Anything).Return(nil)

	topic, err := svc.CreateTopic(ctx, "test-topic")

	assert.NoError(t, err)
	assert.NotNil(t, topic)
	assert.Equal(t, "test-topic", topic.Name)
	assert.Equal(t, userID, topic.UserID)
	repo.AssertExpectations(t)
}

func TestNotifyService_Publish(t *testing.T) {
	repo := new(MockNotifyRepo)
	queueSvc := new(MockQueueService)
	eventSvc := new(MockEventService)
	auditSvc := new(services.MockAuditService)
	svc := services.NewNotifyService(repo, queueSvc, eventSvc, auditSvc)

	userID := uuid.New()
	topicID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)

	topic := &domain.Topic{ID: topicID, UserID: userID, Name: "test"}
	repo.On("GetTopicByID", ctx, topicID, userID).Return(topic, nil)
	repo.On("SaveMessage", ctx, mock.AnythingOfType("*domain.NotifyMessage")).Return(nil)

	sub := &domain.Subscription{
		ID:       uuid.New(),
		UserID:   userID,
		TopicID:  topicID,
		Protocol: domain.ProtocolQueue,
		Endpoint: uuid.New().String(),
	}
	repo.On("ListSubscriptions", ctx, topicID).Return([]*domain.Subscription{sub}, nil)

	queueSvc.On("SendMessage", mock.Anything, uuid.MustParse(sub.Endpoint), "hello").Return(&domain.Message{}, nil)
	eventSvc.On("RecordEvent", ctx, "TOPIC_PUBLISHED", topicID.String(), "TOPIC", mock.Anything).Return(nil)
	auditSvc.On("Log", ctx, userID, "notify.publish", "topic", topicID.String(), mock.Anything).Return(nil)

	err := svc.Publish(ctx, topicID, "hello")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}
