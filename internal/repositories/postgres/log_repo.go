package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/errors"
)

// LogRepository provides PostgreSQL-backed log storage.
type LogRepository struct {
	db DB
}

// NewLogRepository creates a LogRepository using the provided DB.
func NewLogRepository(db DB) *LogRepository {
	return &LogRepository{db: db}
}

func (r *LogRepository) Create(ctx context.Context, entries []*domain.LogEntry) error {
	if len(entries) == 0 {
		return nil
	}

	query := `INSERT INTO log_entries (id, tenant_id, resource_id, resource_type, level, message, timestamp, trace_id) VALUES `
	values := []interface{}{}
	placeholders := []string{}

	for i, entry := range entries {
		offset := i * 8
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", 
			offset+1, offset+2, offset+3, offset+4, offset+5, offset+6, offset+7, offset+8))
		values = append(values, entry.ID, entry.TenantID, entry.ResourceID, entry.ResourceType, entry.Level, entry.Message, entry.Timestamp, entry.TraceID)
	}

	query += strings.Join(placeholders, ",")
	_, err := r.db.Exec(ctx, query, values...)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to batch create log entries", err)
	}
	return nil
}

func (r *LogRepository) List(ctx context.Context, query domain.LogQuery) ([]*domain.LogEntry, int, error) {
	sqlQuery := `SELECT id, tenant_id, resource_id, resource_type, level, message, timestamp, trace_id FROM log_entries WHERE tenant_id = $1`
	countQuery := `SELECT COUNT(*) FROM log_entries WHERE tenant_id = $1`
	args := []interface{}{query.TenantID}
	placeholderIdx := 2

	filters := ""
	if query.ResourceID != "" {
		filters += fmt.Sprintf(" AND resource_id = $%d", placeholderIdx)
		args = append(args, query.ResourceID)
		placeholderIdx++
	}
	if query.ResourceType != "" {
		filters += fmt.Sprintf(" AND resource_type = $%d", placeholderIdx)
		args = append(args, query.ResourceType)
		placeholderIdx++
	}
	if query.Level != "" {
		filters += fmt.Sprintf(" AND level = $%d", placeholderIdx)
		args = append(args, query.Level)
		placeholderIdx++
	}
	if query.Search != "" {
		filters += fmt.Sprintf(" AND message ILIKE $%d", placeholderIdx)
		args = append(args, "%"+query.Search+"%")
		placeholderIdx++
	}
	if query.StartTime != nil {
		filters += fmt.Sprintf(" AND timestamp >= $%d", placeholderIdx)
		args = append(args, *query.StartTime)
		placeholderIdx++
	}
	if query.EndTime != nil {
		filters += fmt.Sprintf(" AND timestamp <= $%d", placeholderIdx)
		args = append(args, *query.EndTime)
		placeholderIdx++
	}

	sqlQuery += filters
	countQuery += filters

	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, errors.Wrap(errors.Internal, "failed to count logs", err)
	}

	sqlQuery += " ORDER BY timestamp DESC"
	
	limit := query.Limit
	if limit <= 0 {
		limit = 100
	}
	sqlQuery += fmt.Sprintf(" LIMIT $%d", placeholderIdx)
	args = append(args, limit)
	placeholderIdx++

	if query.Offset > 0 {
		sqlQuery += fmt.Sprintf(" OFFSET $%d", placeholderIdx)
		args = append(args, query.Offset)
	}

	rows, err := r.db.Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, 0, errors.Wrap(errors.Internal, "failed to query logs", err)
	}
	defer rows.Close()

	var entries []*domain.LogEntry
	for rows.Next() {
		var entry domain.LogEntry
		err := rows.Scan(&entry.ID, &entry.TenantID, &entry.ResourceID, &entry.ResourceType, &entry.Level, &entry.Message, &entry.Timestamp, &entry.TraceID)
		if err != nil {
			return nil, 0, errors.Wrap(errors.Internal, "failed to scan log entry", err)
		}
		entries = append(entries, &entry)
	}

	return entries, total, nil
}

func (r *LogRepository) DeleteByAge(ctx context.Context, days int) error {
	query := `DELETE FROM log_entries WHERE timestamp < NOW() - INTERVAL '1 day' * $1`
	_, err := r.db.Exec(ctx, query, days)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to delete old logs", err)
	}
	return nil
}
