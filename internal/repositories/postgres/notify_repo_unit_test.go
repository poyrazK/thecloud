package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

func TestNotifyRepository_CreateTopic(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewPostgresNotifyRepository(mock)
	topic := &domain.Topic{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Name:      "test-topic",
		ARN:       "arn",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mock.ExpectExec("INSERT INTO topics").
		WithArgs(topic.ID, topic.UserID, topic.Name, topic.ARN, topic.CreatedAt, topic.UpdatedAt).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.CreateTopic(context.Background(), topic)
	assert.NoError(t, err)
}

func TestNotifyRepository_GetSubscriptionByID(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewPostgresNotifyRepository(mock)
	id := uuid.New()
	userID := uuid.New()
	now := time.Now()

	mock.ExpectQuery("SELECT id, user_id, topic_id, protocol, endpoint, created_at, updated_at FROM subscriptions").
		WithArgs(id, userID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "topic_id", "protocol", "endpoint", "created_at", "updated_at"}).
			AddRow(id, userID, uuid.New(), string(domain.ProtocolWebhook), "http://test", now, now))

	sub, err := repo.GetSubscriptionByID(context.Background(), id, userID)
	assert.NoError(t, err)
	assert.NotNil(t, sub)
	assert.Equal(t, id, sub.ID)
	assert.Equal(t, domain.ProtocolWebhook, sub.Protocol)
}

func TestNotifyRepository_ListSubscriptions(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewPostgresNotifyRepository(mock)
	topicID := uuid.New()
	now := time.Now()

	mock.ExpectQuery("SELECT id, user_id, topic_id, protocol, endpoint, created_at, updated_at FROM subscriptions").
		WithArgs(topicID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "topic_id", "protocol", "endpoint", "created_at", "updated_at"}).
			AddRow(uuid.New(), uuid.New(), topicID, string(domain.ProtocolWebhook), "http://test", now, now))

	subs, err := repo.ListSubscriptions(context.Background(), topicID)
	assert.NoError(t, err)
	assert.Len(t, subs, 1)
	assert.Equal(t, domain.ProtocolWebhook, subs[0].Protocol)
}
