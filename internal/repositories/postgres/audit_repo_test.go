package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuditRepositoryCreate(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewAuditRepository(mock)
	log := &domain.AuditLog{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		Action:       "CREATE_INSTANCE",
		ResourceType: "INSTANCE",
		ResourceID:   uuid.New().String(),
		Details:      map[string]interface{}{"name": "test"},
		IPAddress:    testutil.TestIPLocalhost,
		UserAgent:    testutil.TestUserAgent,
		CreatedAt:    time.Now(),
	}

	mock.ExpectExec("INSERT INTO audit_logs").
		WithArgs(log.ID, log.UserID, log.Action, log.ResourceType, log.ResourceID, log.Details, log.IPAddress, log.UserAgent, log.CreatedAt).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.Create(context.Background(), log)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAuditRepositoryListByUserID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewAuditRepository(mock)
	userID := uuid.New()
	limit := 10
	now := time.Now()

	mock.ExpectQuery("SELECT id, user_id, action, resource_type, resource_id, details, ip_address, user_agent, created_at FROM audit_logs").
		WithArgs(userID, limit).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "action", "resource_type", "resource_id", "details", "ip_address", "user_agent", "created_at"}).
			AddRow(uuid.New(), userID, "ACTION1", "RES1", "ID1", nil, "IP1", "UA1", now).
			AddRow(uuid.New(), userID, "ACTION2", "RES2", "ID2", nil, "IP2", "UA2", now))

	logs, err := repo.ListByUserID(context.Background(), userID, limit)
	require.NoError(t, err)
	assert.Len(t, logs, 2)
	assert.Equal(t, userID, logs[0].UserID)
	require.NoError(t, mock.ExpectationsWereMet())
}
