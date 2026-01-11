package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

type EventRepository struct {
	db DB
}

func NewEventRepository(db DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(ctx context.Context, e *domain.Event) error {
	query := `INSERT INTO events (id, user_id, action, resource_id, resource_type, metadata, created_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(ctx, query, e.ID, e.UserID, e.Action, e.ResourceID, e.ResourceType, e.Metadata, e.CreatedAt)
	return err
}

func (r *EventRepository) List(ctx context.Context, limit int) ([]*domain.Event, error) {
	userID := appcontext.UserIDFromContext(ctx)
	query := `SELECT id, user_id, action, resource_id, resource_type, metadata, created_at 
              FROM events 
              WHERE user_id = $1
              ORDER BY created_at DESC 
              LIMIT $2`

	rows, err := r.db.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	return r.scanEvents(rows)
}

func (r *EventRepository) scanEvent(row pgx.Row) (*domain.Event, error) {
	e := &domain.Event{}
	if err := row.Scan(&e.ID, &e.UserID, &e.Action, &e.ResourceID, &e.ResourceType, &e.Metadata, &e.CreatedAt); err != nil {
		return nil, err
	}
	return e, nil
}

func (r *EventRepository) scanEvents(rows pgx.Rows) ([]*domain.Event, error) {
	defer rows.Close()
	var events []*domain.Event
	for rows.Next() {
		e, err := r.scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}
