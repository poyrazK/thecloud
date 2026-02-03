//go:build integration

package postgres

import (
	"testing"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresNotifyRepository(t *testing.T) {
	db := SetupDB(t)
	defer db.Close()
	repo := NewPostgresNotifyRepository(db)
	ctx := SetupTestUser(t, db)
	userID := appcontext.UserIDFromContext(ctx)

	t.Run("CreateAndGetTopic", func(t *testing.T) {
		topic := &domain.Topic{
			ID:        uuid.New(),
			UserID:    userID,
			Name:      "test-topic",
			ARN:       "arn:thecloud:notify:local:" + userID.String() + ":topic/test-topic",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := repo.CreateTopic(ctx, topic)
		require.NoError(t, err)

		found, err := repo.GetTopicByName(ctx, topic.Name, userID)
		require.NoError(t, err)
		assert.Equal(t, topic.ID, found.ID)
	})

	t.Run("Subscriptions", func(t *testing.T) {
		topicID := uuid.New()
		err := repo.CreateTopic(ctx, &domain.Topic{ID: topicID, UserID: userID, Name: "sub-topic", ARN: "arn..."})
		require.NoError(t, err)

		sub := &domain.Subscription{
			ID:        uuid.New(),
			UserID:    userID,
			TopicID:   topicID,
			Protocol:  domain.ProtocolWebhook,
			Endpoint:  "http://test",
			CreatedAt: time.Now(),
		}

		err = repo.CreateSubscription(ctx, sub)
		require.NoError(t, err)

		subs, err := repo.ListSubscriptions(ctx, topicID)
		require.NoError(t, err)
		assert.Len(t, subs, 1)
	})
}
