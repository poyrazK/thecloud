package services_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupEventServiceTest(t *testing.T) (*services.EventService, *postgres.EventRepository, context.Context) {
	t.Helper()
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewEventRepository(db)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := services.NewEventService(repo, nil, logger)
	return svc, repo, ctx
}

func TestEventService_Integration(t *testing.T) {
	svc, repo, ctx := setupEventServiceTest(t)
	userID := appcontext.UserIDFromContext(ctx)

	t.Run("RecordEventSuccess", func(t *testing.T) {
		action := "TEST_ACTION"
		resID := "res-123"
		resType := "TEST"
		details := map[string]interface{}{"key": "value"}

		err := svc.RecordEvent(ctx, action, resID, resType, details)
		require.NoError(t, err)

		// Verify in DB
		result, err := repo.List(ctx, 10)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, action, result[0].Action)
		assert.Equal(t, resID, result[0].ResourceID)
		assert.Equal(t, resType, result[0].ResourceType)
		assert.Equal(t, userID, result[0].UserID)
	})

	t.Run("ListEvents_Limits", func(t *testing.T) {
		// Create a fresh user for this subtest to avoid interference
		db := setupDB(t)
		cleanDB(t, db)
		subCtx := setupTestUser(t, db)

		for i := 0; i < 5; i++ {
			_ = svc.RecordEvent(subCtx, "ACTION", "res", "TYPE", nil)
		}

		t.Run("SpecificLimit", func(t *testing.T) {
			result, err := svc.ListEvents(subCtx, 2)
			require.NoError(t, err)
			assert.Len(t, result, 2)
		})

		t.Run("DefaultLimit", func(t *testing.T) {
			result, err := svc.ListEvents(subCtx, 0)
			require.NoError(t, err)
			assert.Len(t, result, 5)
		})
	})
}
