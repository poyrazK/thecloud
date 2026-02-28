package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAuditServiceTest(t *testing.T) (*services.AuditService, *postgres.AuditRepository, context.Context) {
	t.Helper()
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewAuditRepository(db)
	svc := services.NewAuditService(repo)
	return svc, repo, ctx
}

func TestAuditService_Integration(t *testing.T) {
	svc, repo, ctx := setupAuditServiceTest(t)
	userID := appcontext.UserIDFromContext(ctx)

	t.Run("LogAndRetrieve", func(t *testing.T) {
		action := "test.action"
		resType := "instance"
		resID := "123"
		details := map[string]interface{}{"key": "value"}

		err := svc.Log(ctx, userID, action, resType, resID, details)
		require.NoError(t, err)

		// Verify in DB via direct repo access
		logs, err := repo.ListByUserID(ctx, userID, 10)
		require.NoError(t, err)
		assert.Len(t, logs, 1)
		assert.Equal(t, action, logs[0].Action)
		assert.Equal(t, resID, logs[0].ResourceID)
		assert.Equal(t, details["key"], logs[0].Details["key"])
	})

	t.Run("ListLogs_Limits", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			_ = svc.Log(ctx, userID, "action", "res", "id", nil)
		}

		t.Run("SpecificLimit", func(t *testing.T) {
			logs, err := svc.ListLogs(ctx, userID, 2)
			require.NoError(t, err)
			assert.Len(t, logs, 2)
		})

		t.Run("DefaultLimit", func(t *testing.T) {
			// 0 should trigger default limit (50)
			logs, err := svc.ListLogs(ctx, userID, 0)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(logs), 5)
		})
	})

	t.Run("UserIsolation", func(t *testing.T) {
		otherUserID := uuid.New()
		_ = svc.Log(ctx, userID, "user1-action", "res", "1", nil)
		_ = svc.Log(ctx, otherUserID, "user2-action", "res", "2", nil)

		logs, err := svc.ListLogs(ctx, userID, 10)
		require.NoError(t, err)
		
		for _, log := range logs {
			assert.Equal(t, userID, log.UserID)
			assert.NotEqual(t, otherUserID, log.UserID)
		}
	})
}
