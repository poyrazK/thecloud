// Package postgres provides PostgreSQL-backed repository implementations.
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

type accountingRepository struct {
	db DB
}

// NewAccountingRepository returns a PostgreSQL-backed accounting repository.
func NewAccountingRepository(db DB) ports.AccountingRepository {
	return &accountingRepository{db: db}
}

func (r *accountingRepository) CreateRecord(ctx context.Context, record domain.UsageRecord) error {
	query := `
		INSERT INTO usage_records (id, user_id, resource_id, resource_type, quantity, unit, start_time, end_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(ctx, query,
		record.ID,
		record.UserID,
		record.ResourceID,
		record.ResourceType,
		record.Quantity,
		record.Unit,
		record.StartTime,
		record.EndTime,
	)
	if err != nil {
		return fmt.Errorf("failed to create usage record: %w", err)
	}
	return nil
}

func (r *accountingRepository) GetUsageSummary(ctx context.Context, userID uuid.UUID, start, end time.Time) (map[domain.ResourceType]float64, error) {
	query := `
		SELECT resource_type, SUM(quantity)
		FROM usage_records
		WHERE user_id = $1 AND start_time >= $2 AND end_time <= $3
		GROUP BY resource_type
	`
	rows, err := r.db.Query(ctx, query, userID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage summary: %w", err)
	}
	defer rows.Close()

	summary := make(map[domain.ResourceType]float64)
	for rows.Next() {
		var resType string
		var total float64
		if err := rows.Scan(&resType, &total); err != nil {
			return nil, err
		}
		summary[domain.ResourceType(resType)] = total
	}
	return summary, nil
}

func (r *accountingRepository) ListRecords(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]domain.UsageRecord, error) {
	query := `
		SELECT id, user_id, resource_id, resource_type, quantity, unit, start_time, end_time
		FROM usage_records
		WHERE user_id = $1 AND start_time >= $2 AND end_time <= $3
		ORDER BY start_time DESC
	`
	rows, err := r.db.Query(ctx, query, userID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to list usage records: %w", err)
	}
	return r.scanUsageRecords(rows)
}

func (r *accountingRepository) scanUsageRecord(row pgx.Row) (domain.UsageRecord, error) {
	var rec domain.UsageRecord
	var resType string
	err := row.Scan(
		&rec.ID,
		&rec.UserID,
		&rec.ResourceID,
		&resType,
		&rec.Quantity,
		&rec.Unit,
		&rec.StartTime,
		&rec.EndTime,
	)
	if err != nil {
		return domain.UsageRecord{}, err
	}
	rec.ResourceType = domain.ResourceType(resType)
	return rec, nil
}

func (r *accountingRepository) scanUsageRecords(rows pgx.Rows) ([]domain.UsageRecord, error) {
	defer rows.Close()
	var records []domain.UsageRecord
	for rows.Next() {
		rec, err := r.scanUsageRecord(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, nil
}
