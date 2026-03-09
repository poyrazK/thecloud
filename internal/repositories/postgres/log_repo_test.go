package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogRepository_Create(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewLogRepository(mock)
	tenantID := uuid.New()

	t.Run("empty entries", func(t *testing.T) {
		err := repo.Create(context.Background(), nil)
		require.NoError(t, err)
	})

	t.Run("success batch", func(t *testing.T) {
		entries := []*domain.LogEntry{
			{
				ID: uuid.New(), TenantID: tenantID, ResourceID: "res-1", ResourceType: "instance",
				Level: "INFO", Message: "msg 1", Timestamp: time.Now(), TraceID: "t1",
			},
		}
		mock.ExpectExec(`INSERT INTO log_entries`).
			WithArgs(entries[0].ID, entries[0].TenantID, entries[0].ResourceID, entries[0].ResourceType, entries[0].Level, entries[0].Message, entries[0].Timestamp, entries[0].TraceID).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = repo.Create(context.Background(), entries)
		require.NoError(t, err)
	})
}

func TestLogRepository_List_Complex(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewLogRepository(mock)
	tenantID := uuid.New()
	now := time.Now()

	query := domain.LogQuery{
		TenantID:     tenantID,
		ResourceID:   "res-1",
		ResourceType: "instance",
		Level:        "ERROR",
		Search:       "fail",
		StartTime:    &now,
		EndTime:      &now,
		Limit:        50,
		Offset:       10,
	}

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM log_entries WHERE tenant_id = \$1 AND resource_id = \$2 AND resource_type = \$3 AND level = \$4 AND message ILIKE \$5 AND timestamp >= \$6 AND timestamp <= \$7`).
		WithArgs(tenantID, "res-1", "instance", "ERROR", "%fail%", now, now).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT id, tenant_id, resource_id, resource_type, level, message, timestamp, trace_id FROM log_entries WHERE tenant_id = \$1 AND resource_id = \$2 AND resource_type = \$3 AND level = \$4 AND message ILIKE \$5 AND timestamp >= \$6 AND timestamp <= \$7 ORDER BY timestamp DESC LIMIT \$8 OFFSET \$9`).
		WithArgs(tenantID, "res-1", "instance", "ERROR", "%fail%", now, now, 50, 10).
		WillReturnRows(pgxmock.NewRows([]string{"id", "tenant_id", "resource_id", "resource_type", "level", "message", "timestamp", "trace_id"}).
			AddRow(uuid.New(), tenantID, "res-1", "instance", "ERROR", "fail fast", now, "trace"))

	entries, total, err := repo.List(context.Background(), query)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, entries, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLogRepository_DeleteByAge(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewLogRepository(mock)
	days := 30

	mock.ExpectExec(`DELETE FROM log_entries WHERE timestamp < NOW\(\) - INTERVAL '1 day' \* \$1`).
		WithArgs(days).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = repo.DeleteByAge(context.Background(), days)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
