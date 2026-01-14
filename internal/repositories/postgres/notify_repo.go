// Package postgres provides PostgreSQL-backed repository implementations.
package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

// PostgresNotifyRepository provides PostgreSQL-backed notify persistence.
type PostgresNotifyRepository struct {
	db DB
}

// NewPostgresNotifyRepository creates a notify repository using the provided DB.
func NewPostgresNotifyRepository(db DB) ports.NotifyRepository {
	return &PostgresNotifyRepository{db: db}
}

func (r *PostgresNotifyRepository) CreateTopic(ctx context.Context, topic *domain.Topic) error {
	query := `
		INSERT INTO topics (id, user_id, name, arn, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query,
		topic.ID,
		topic.UserID,
		topic.Name,
		topic.ARN,
		topic.CreatedAt,
		topic.UpdatedAt,
	)
	return err
}

func (r *PostgresNotifyRepository) GetTopicByID(ctx context.Context, id, userID uuid.UUID) (*domain.Topic, error) {
	query := `SELECT id, user_id, name, arn, created_at, updated_at FROM topics WHERE id = $1 AND user_id = $2`
	return r.scanTopic(r.db.QueryRow(ctx, query, id, userID))
}

func (r *PostgresNotifyRepository) GetTopicByName(ctx context.Context, name string, userID uuid.UUID) (*domain.Topic, error) {
	query := `SELECT id, user_id, name, arn, created_at, updated_at FROM topics WHERE name = $1 AND user_id = $2`
	return r.scanTopic(r.db.QueryRow(ctx, query, name, userID))
}

func (r *PostgresNotifyRepository) ListTopics(ctx context.Context, userID uuid.UUID) ([]*domain.Topic, error) {
	query := `SELECT id, user_id, name, arn, created_at, updated_at FROM topics WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	return r.scanTopics(rows)
}

func (r *PostgresNotifyRepository) DeleteTopic(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM topics WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *PostgresNotifyRepository) CreateSubscription(ctx context.Context, sub *domain.Subscription) error {
	query := `
		INSERT INTO subscriptions (id, user_id, topic_id, protocol, endpoint, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query,
		sub.ID,
		sub.UserID,
		sub.TopicID,
		sub.Protocol,
		sub.Endpoint,
		sub.CreatedAt,
		sub.UpdatedAt,
	)
	return err
}

func (r *PostgresNotifyRepository) GetSubscriptionByID(ctx context.Context, id, userID uuid.UUID) (*domain.Subscription, error) {
	query := `SELECT id, user_id, topic_id, protocol, endpoint, created_at, updated_at FROM subscriptions WHERE id = $1 AND user_id = $2`
	return r.scanSubscription(r.db.QueryRow(ctx, query, id, userID))
}

func (r *PostgresNotifyRepository) ListSubscriptions(ctx context.Context, topicID uuid.UUID) ([]*domain.Subscription, error) {
	query := `SELECT id, user_id, topic_id, protocol, endpoint, created_at, updated_at FROM subscriptions WHERE topic_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, topicID)
	if err != nil {
		return nil, err
	}
	return r.scanSubscriptions(rows)
}

func (r *PostgresNotifyRepository) scanTopic(row pgx.Row) (*domain.Topic, error) {
	var topic domain.Topic
	err := row.Scan(
		&topic.ID,
		&topic.UserID,
		&topic.Name,
		&topic.ARN,
		&topic.CreatedAt,
		&topic.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &topic, nil
}

func (r *PostgresNotifyRepository) scanTopics(rows pgx.Rows) ([]*domain.Topic, error) {
	defer rows.Close()
	var topics []*domain.Topic
	for rows.Next() {
		topic, err := r.scanTopic(rows)
		if err != nil {
			return nil, err
		}
		topics = append(topics, topic)
	}
	return topics, nil
}

func (r *PostgresNotifyRepository) scanSubscription(row pgx.Row) (*domain.Subscription, error) {
	var sub domain.Subscription
	var protocol string
	err := row.Scan(
		&sub.ID,
		&sub.UserID,
		&sub.TopicID,
		&protocol,
		&sub.Endpoint,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	sub.Protocol = domain.SubscriptionProtocol(protocol)
	return &sub, nil
}

func (r *PostgresNotifyRepository) scanSubscriptions(rows pgx.Rows) ([]*domain.Subscription, error) {
	defer rows.Close()
	var subs []*domain.Subscription
	for rows.Next() {
		sub, err := r.scanSubscription(rows)
		if err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	return subs, nil
}

func (r *PostgresNotifyRepository) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM subscriptions WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *PostgresNotifyRepository) SaveMessage(ctx context.Context, msg *domain.NotifyMessage) error {
	query := `INSERT INTO notify_messages (id, topic_id, body, created_at) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(ctx, query, msg.ID, msg.TopicID, msg.Body, msg.CreatedAt)
	return err
}
