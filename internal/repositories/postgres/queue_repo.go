package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

type PostgresQueueRepository struct {
	db *sql.DB
}

func NewPostgresQueueRepository(db *sql.DB) ports.QueueRepository {
	return &PostgresQueueRepository{db: db}
}

func (r *PostgresQueueRepository) Create(ctx context.Context, q *domain.Queue) error {
	query := `
		INSERT INTO queues (id, user_id, name, arn, visibility_timeout, retention_days, max_message_size, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.ExecContext(ctx, query,
		q.ID, q.UserID, q.Name, q.ARN, q.VisibilityTimeout, q.RetentionDays, q.MaxMessageSize, q.Status, q.CreatedAt, q.UpdatedAt)
	return err
}

func (r *PostgresQueueRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Queue, error) {
	q := &domain.Queue{}
	query := `SELECT id, user_id, name, arn, visibility_timeout, retention_days, max_message_size, status, created_at, updated_at FROM queues WHERE id = $1 AND user_id = $2`
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&q.ID, &q.UserID, &q.Name, &q.ARN, &q.VisibilityTimeout, &q.RetentionDays, &q.MaxMessageSize, &q.Status, &q.CreatedAt, &q.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return q, err
}

func (r *PostgresQueueRepository) GetByName(ctx context.Context, name string, userID uuid.UUID) (*domain.Queue, error) {
	q := &domain.Queue{}
	query := `SELECT id, user_id, name, arn, visibility_timeout, retention_days, max_message_size, status, created_at, updated_at FROM queues WHERE name = $1 AND user_id = $2`
	err := r.db.QueryRowContext(ctx, query, name, userID).Scan(
		&q.ID, &q.UserID, &q.Name, &q.ARN, &q.VisibilityTimeout, &q.RetentionDays, &q.MaxMessageSize, &q.Status, &q.CreatedAt, &q.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return q, err
}

func (r *PostgresQueueRepository) List(ctx context.Context, userID uuid.UUID) ([]*domain.Queue, error) {
	query := `SELECT id, user_id, name, arn, visibility_timeout, retention_days, max_message_size, status, created_at, updated_at FROM queues WHERE user_id = $1`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var queues []*domain.Queue
	for rows.Next() {
		q := &domain.Queue{}
		if err := rows.Scan(&q.ID, &q.UserID, &q.Name, &q.ARN, &q.VisibilityTimeout, &q.RetentionDays, &q.MaxMessageSize, &q.Status, &q.CreatedAt, &q.UpdatedAt); err != nil {
			return nil, err
		}
		queues = append(queues, q)
	}
	return queues, nil
}

func (r *PostgresQueueRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM queues WHERE id = $1", id)
	return err
}

func (r *PostgresQueueRepository) SendMessage(ctx context.Context, queueID uuid.UUID, body string) (*domain.Message, error) {
	m := &domain.Message{
		ID:        uuid.New(),
		QueueID:   queueID,
		Body:      body,
		VisibleAt: time.Now(),
		CreatedAt: time.Now(),
	}
	query := `INSERT INTO queue_messages (id, queue_id, body, visible_at, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query, m.ID, m.QueueID, m.Body, m.VisibleAt, m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *PostgresQueueRepository) ReceiveMessages(ctx context.Context, queueID uuid.UUID, maxMessages, visibilityTimeout int) ([]*domain.Message, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 1. Select visible messages and lock them
	// We use SKIP LOCKED to allow parallel consumers
	query := `
		SELECT id, queue_id, body, received_count, created_at 
		FROM queue_messages 
		WHERE queue_id = $1 AND visible_at <= NOW() 
		FOR UPDATE SKIP LOCKED 
		LIMIT $2
	`
	rows, err := tx.QueryContext(ctx, query, queueID, maxMessages)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*domain.Message
	now := time.Now()
	for rows.Next() {
		m := &domain.Message{}
		if err := rows.Scan(&m.ID, &m.QueueID, &m.Body, &m.ReceivedCount, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.ReceiptHandle = uuid.New().String()
		m.VisibleAt = now.Add(time.Duration(visibilityTimeout) * time.Second)
		m.ReceivedCount++
		messages = append(messages, m)
	}

	// 2. Update status of received messages
	updateQuery := `UPDATE queue_messages SET receipt_handle = $1, visible_at = $2, received_count = received_count + 1 WHERE id = $3`
	for _, m := range messages {
		_, err := tx.ExecContext(ctx, updateQuery, m.ReceiptHandle, m.VisibleAt, m.ID)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *PostgresQueueRepository) DeleteMessage(ctx context.Context, queueID uuid.UUID, receiptHandle string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM queue_messages WHERE queue_id = $1 AND receipt_handle = $2", queueID, receiptHandle)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("message not found or receipt handle expired")
	}
	return nil
}

func (r *PostgresQueueRepository) PurgeMessages(ctx context.Context, queueID uuid.UUID) (int64, error) {
	result, err := r.db.ExecContext(ctx, "DELETE FROM queue_messages WHERE queue_id = $1", queueID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (r *PostgresQueueRepository) GetQueueStats(ctx context.Context, queueID uuid.UUID) (int, int, error) {
	var visible, inFlight int
	query := `
		SELECT 
			COUNT(*) FILTER (WHERE visible_at <= NOW()) as visible,
			COUNT(*) FILTER (WHERE visible_at > NOW()) as in_flight
		FROM queue_messages 
		WHERE queue_id = $1
	`
	err := r.db.QueryRowContext(ctx, query, queueID).Scan(&visible, &in_flight)
	return visible, in_flight, err
}
