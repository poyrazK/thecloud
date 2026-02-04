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
)

func setupEventServiceTest(t *testing.T) (*services.EventService, *postgres.EventRepository, context.Context) {
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewEventRepository(db)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := services.NewEventService(repo, nil, logger)
	return svc, repo, ctx
}

func TestEventServiceRecordEventSuccess(t *testing.T) {
	svc, repo, ctx := setupEventServiceTest(t)
	userID := appcontext.UserIDFromContext(ctx)

	action := "TEST_ACTION"
	resID := "res-123"
	resType := "TEST"
	details := map[string]interface{}{"key": "value"}

	err := svc.RecordEvent(ctx, action, resID, resType, details)
	assert.NoError(t, err)

	// Verify in DB - wait, List doesn't filter by user ID in this implementation?
	// Let's check postgres.EventRepository.List
	result, err := repo.List(ctx, 10)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, action, result[0].Action)
	assert.Equal(t, userID, result[0].UserID)
}

func TestEventServiceListEvents(t *testing.T) {
	svc, _, ctx := setupEventServiceTest(t)

	_ = svc.RecordEvent(ctx, "A1", "r1", "T1", nil)
	_ = svc.RecordEvent(ctx, "A2", "r2", "T2", nil)

	result, err := svc.ListEvents(ctx, 10)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}
