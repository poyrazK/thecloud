package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/poyrazk/thecloud/internal/core/domain"
	theclouderrors "github.com/poyrazk/thecloud/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueueRepositoryCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewPostgresQueueRepository(mock)
		q := &domain.Queue{
			ID:                uuid.New(),
			UserID:            uuid.New(),
			TenantID:          uuid.New(),
			Name:              "test-queue",
			ARN:               "arn:aws:sqs:us-east-1:123456789012:test-queue",
			VisibilityTimeout: 30,
			RetentionDays:     4,
			MaxMessageSize:    262144,
			Status:            domain.QueueStatusActive,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		mock.ExpectExec("INSERT INTO queues").
			WithArgs(q.ID, q.UserID, q.TenantID, q.Name, q.ARN, q.VisibilityTimeout, q.RetentionDays, q.MaxMessageSize, q.Status, q.CreatedAt, q.UpdatedAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = repo.Create(context.Background(), q)
		require.NoError(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewPostgresQueueRepository(mock)
		q := &domain.Queue{ID: uuid.New()}

		mock.ExpectExec("INSERT INTO queues").
			WillReturnError(errors.New("db error"))

		err = repo.Create(context.Background(), q)
		require.Error(t, err)
	})
}

func TestQueueRepositoryGetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewPostgresQueueRepository(mock)
		id := uuid.New()
		tenantID := uuid.New()
		now := time.Now()

		mock.ExpectQuery("SELECT id, user_id, tenant_id, name, arn, visibility_timeout, retention_days, max_message_size, status, created_at, updated_at FROM queues WHERE id = \\$1 AND tenant_id = \\$2").
			WithArgs(id, tenantID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "tenant_id", "name", "arn", "visibility_timeout", "retention_days", "max_message_size", "status", "created_at", "updated_at"}).
				AddRow(id, uuid.New(), tenantID, "test-queue", "arn", 30, 4, 262144, string(domain.QueueStatusActive), now, now))

		q, err := repo.GetByID(context.Background(), id, tenantID)
		require.NoError(t, err)
		assert.NotNil(t, q)
		assert.Equal(t, id, q.ID)
	})

	t.Run("not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewPostgresQueueRepository(mock)
		id := uuid.New()
		tenantID := uuid.New()

		mock.ExpectQuery("SELECT id, user_id, tenant_id, name, arn, visibility_timeout, retention_days, max_message_size, status, created_at, updated_at FROM queues WHERE id = \\$1 AND tenant_id = \\$2").
			WithArgs(id, tenantID).
			WillReturnError(pgx.ErrNoRows)

		q, err := repo.GetByID(context.Background(), id, tenantID)
		require.NoError(t, err)
		assert.Nil(t, q)
	})
}

func TestQueueRepositoryList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewPostgresQueueRepository(mock)
		tenantID := uuid.New()
		now := time.Now()

		mock.ExpectQuery("SELECT id, user_id, tenant_id, name, arn, visibility_timeout, retention_days, max_message_size, status, created_at, updated_at FROM queues WHERE tenant_id = \\$1").
			WithArgs(tenantID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "tenant_id", "name", "arn", "visibility_timeout", "retention_days", "max_message_size", "status", "created_at", "updated_at"}).
				AddRow(uuid.New(), uuid.New(), tenantID, "test-queue", "arn", 30, 4, 262144, string(domain.QueueStatusActive), now, now))

		queues, err := repo.List(context.Background(), tenantID)
		require.NoError(t, err)
		assert.Len(t, queues, 1)
	})
}

func TestQueueRepositoryDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewPostgresQueueRepository(mock)
		id := uuid.New()
		tenantID := uuid.New()

		mock.ExpectExec("DELETE FROM queues").
			WithArgs(id, tenantID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = repo.Delete(context.Background(), id, tenantID)
		require.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewPostgresQueueRepository(mock)
		id := uuid.New()
		tenantID := uuid.New()

		mock.ExpectExec("DELETE FROM queues").
			WithArgs(id, tenantID).
			WillReturnResult(pgxmock.NewResult("DELETE", 0))

		err = repo.Delete(context.Background(), id, tenantID)
		require.Error(t, err)
		var theCloudErr *theclouderrors.Error
		if errors.As(err, &theCloudErr) {
			assert.Equal(t, theclouderrors.NotFound, theCloudErr.Type)
		}
	})
}

func TestQueueRepositorySendMessage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewPostgresQueueRepository(mock)
		queueID := uuid.New()
		body := "test-message"

		// 3 args used when using NOW()
		mock.ExpectExec("(?s)INSERT INTO queue_messages.*").
			WithArgs(pgxmock.AnyArg(), queueID, body).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		m, err := repo.SendMessage(context.Background(), queueID, body)
		require.NoError(t, err)
		assert.NotNil(t, m)
		assert.Equal(t, body, m.Body)
	})
}

func TestQueueRepositoryReceiveMessages(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewPostgresQueueRepository(mock)
		queueID := uuid.New()
		maxMessages := 1
		visibilityTimeout := 30
		now := time.Now()

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT id, queue_id, body, received_count, created_at FROM queue_messages").
			WithArgs(queueID, maxMessages).
			WillReturnRows(pgxmock.NewRows([]string{"id", "queue_id", "body", "received_count", "created_at"}).
				AddRow(uuid.New(), queueID, "test-message", 0, now))
		mock.ExpectExec("UPDATE queue_messages").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))
		mock.ExpectCommit()

		messages, err := repo.ReceiveMessages(context.Background(), queueID, maxMessages, visibilityTimeout)
		require.NoError(t, err)
		assert.Len(t, messages, 1)
	})
}

func TestQueueRepositoryDeleteMessage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewPostgresQueueRepository(mock)
		queueID := uuid.New()
		receiptHandle := "handle"

		mock.ExpectExec("DELETE FROM queue_messages").
			WithArgs(queueID, receiptHandle).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = repo.DeleteMessage(context.Background(), queueID, receiptHandle)
		require.NoError(t, err)
	})
}

func TestQueueRepositoryPurgeMessages(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewPostgresQueueRepository(mock)
		queueID := uuid.New()

		mock.ExpectExec("DELETE FROM queue_messages").
			WithArgs(queueID).
			WillReturnResult(pgxmock.NewResult("DELETE", 5))

		count, err := repo.PurgeMessages(context.Background(), queueID)
		require.NoError(t, err)
		assert.Equal(t, int64(5), count)
	})
}

func TestQueueRepositoryGetQueueStats(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewPostgresQueueRepository(mock)
		queueID := uuid.New()

		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FILTER.+as visible, COUNT\\(\\*\\) FILTER.+as in_flight FROM queue_messages").
			WithArgs(queueID).
			WillReturnRows(pgxmock.NewRows([]string{"visible", "in_flight"}).
				AddRow(10, 5))

		postgresRepo, ok := repo.(*PostgresQueueRepository)
		require.True(t, ok)
		visible, inFlight, err := postgresRepo.GetQueueStats(context.Background(), queueID)
		require.NoError(t, err)
		assert.Equal(t, 10, visible)
		assert.Equal(t, 5, inFlight)
	})
}
