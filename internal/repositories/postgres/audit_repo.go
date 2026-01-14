// Package postgres provides PostgreSQL-backed repository implementations.
package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/errors"
)

// AuditRepository provides PostgreSQL-backed audit log storage.
type AuditRepository struct {
	db DB
}

// NewAuditRepository creates an AuditRepository using the provided DB.
func NewAuditRepository(db DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	query := `
		INSERT INTO audit_logs (id, user_id, action, resource_type, resource_id, details, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query,
		log.ID, log.UserID, log.Action, log.ResourceType, log.ResourceID,
		log.Details, log.IPAddress, log.UserAgent, log.CreatedAt,
	)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to create audit log", err)
	}
	return nil
}

func (r *AuditRepository) ListByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.AuditLog, error) {
	query := `
		SELECT id, user_id, action, resource_type, resource_id, details, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := r.db.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to list audit logs", err)
	}
	return r.scanAuditLogs(rows)
}

func (r *AuditRepository) scanAuditLog(row pgx.Row) (*domain.AuditLog, error) {
	var log domain.AuditLog
	err := row.Scan(
		&log.ID, &log.UserID, &log.Action, &log.ResourceType, &log.ResourceID,
		&log.Details, &log.IPAddress, &log.UserAgent, &log.CreatedAt,
	)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to scan audit log", err)
	}
	return &log, nil
}

func (r *AuditRepository) scanAuditLogs(rows pgx.Rows) ([]*domain.AuditLog, error) {
	defer rows.Close()
	var logs []*domain.AuditLog
	for rows.Next() {
		log, err := r.scanAuditLog(rows)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}
