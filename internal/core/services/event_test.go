package services_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupEventServiceTest(t *testing.T) (*services.EventService, *postgres.EventRepository, context.Context) {
	t.Helper()
	db := setupDB(t)
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
		action := "TEST_ACTION_" + uuid.New().String()[:8]
		resID := "res-123"
		resType := "TEST"
		details := map[string]interface{}{"key": "value"}

		err := svc.RecordEvent(ctx, action, resID, resType, details)
		require.NoError(t, err)

		// Verify in DB
		result, err := repo.List(ctx, 10)
		require.NoError(t, err)
		
		found := false
		for _, e := range result {
			if e.Action == action {
				assert.Equal(t, resID, e.ResourceID)
				assert.Equal(t, resType, e.ResourceType)
				assert.Equal(t, userID, e.UserID)
				found = true
				break
			}
		}
		assert.True(t, found, "Recorded event not found in list")
	})

	t.Run("ListEvents_Limits", func(t *testing.T) {
		uniqueType := "LIMIT_TEST_" + uuid.New().String()[:8]
		
		for i := 0; i < 5; i++ {
			_ = svc.RecordEvent(ctx, "ACTION", "res", uniqueType, nil)
		}

		t.Run("SpecificLimit", func(t *testing.T) {
			result, err := svc.ListEvents(ctx, 2)
			require.NoError(t, err)
			assert.Len(t, result, 2)
		})

		t.Run("DefaultLimit", func(t *testing.T) {
			result, err := svc.ListEvents(ctx, 0)
			require.NoError(t, err)
			
			count := 0
			for _, e := range result {
				if e.ResourceType == uniqueType {
					count++
				}
			}
			assert.Equal(t, 5, count)
		})
	})
}
