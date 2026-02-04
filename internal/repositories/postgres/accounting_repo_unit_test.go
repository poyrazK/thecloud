package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

func TestAccountingRepository_CreateRecord(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewAccountingRepository(mock)
	record := domain.UsageRecord{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		ResourceID:   uuid.New(),
		ResourceType: domain.ResourceInstance,
		Quantity:     1.5,
		Unit:         "hours",
		StartTime:    time.Now(),
		EndTime:      time.Now().Add(time.Hour),
	}

	mock.ExpectExec("INSERT INTO usage_records").
		WithArgs(record.ID, record.UserID, record.ResourceID, record.ResourceType, record.Quantity, record.Unit, record.StartTime, record.EndTime).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.CreateRecord(context.Background(), record)
	assert.NoError(t, err)
}

func TestAccountingRepository_GetUsageSummary(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewAccountingRepository(mock)
	userID := uuid.New()
	start := time.Now()
	end := time.Now().Add(time.Hour)

	mock.ExpectQuery("SELECT resource_type, SUM\\(quantity\\) FROM usage_records").
		WithArgs(userID, start, end).
		WillReturnRows(pgxmock.NewRows([]string{"resource_type", "sum"}).
			AddRow(string(domain.ResourceInstance), 10.5).
			AddRow(string(domain.ResourceStorage), 5.0))

	summary, err := repo.GetUsageSummary(context.Background(), userID, start, end)
	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, 10.5, summary[domain.ResourceInstance])
	assert.Equal(t, 5.0, summary[domain.ResourceStorage])
}

func TestAccountingRepository_ListRecords(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewAccountingRepository(mock)
	userID := uuid.New()
	start := time.Now()
	end := time.Now().Add(time.Hour)

	mock.ExpectQuery("SELECT id, user_id, resource_id, resource_type, quantity, unit, start_time, end_time FROM usage_records").
		WithArgs(userID, start, end).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "resource_id", "resource_type", "quantity", "unit", "start_time", "end_time"}).
			AddRow(uuid.New(), userID, uuid.New(), string(domain.ResourceInstance), 1.0, "hour", start, end))

	records, err := repo.ListRecords(context.Background(), userID, start, end)
	assert.NoError(t, err)
	assert.Len(t, records, 1)
	assert.Equal(t, userID, records[0].UserID)
}
