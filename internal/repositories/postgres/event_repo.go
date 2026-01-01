package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/poyraz/cloud/internal/core/domain"
)

type EventRepository struct {
	db *pgxpool.Pool
}

func NewEventRepository(db *pgxpool.Pool) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(ctx context.Context, e *domain.Event) error {
	query := `INSERT INTO events (id, action, resource_id, resource_type, metadata, created_at) 
              VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(ctx, query, e.ID, e.Action, e.ResourceID, e.ResourceType, e.Metadata, e.CreatedAt)
	return err
}

func (r *EventRepository) List(ctx context.Context, limit int) ([]*domain.Event, error) {
	query := `SELECT id, action, resource_id, resource_type, metadata, created_at 
              FROM events 
              ORDER BY created_at DESC 
              LIMIT $1`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*domain.Event
	for rows.Next() {
		e := &domain.Event{}
		if err := rows.Scan(&e.ID, &e.Action, &e.ResourceID, &e.ResourceType, &e.Metadata, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}
