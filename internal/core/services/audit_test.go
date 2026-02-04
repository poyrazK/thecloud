package services_test

import (
	"context"
	"testing"

	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
)

func setupAuditServiceTest(t *testing.T) (*services.AuditService, *postgres.AuditRepository, context.Context) {
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewAuditRepository(db)
	svc := services.NewAuditService(repo)
	return svc, repo, ctx
}

func TestAuditServiceLog(t *testing.T) {
	svc, repo, ctx := setupAuditServiceTest(t)
	userID := appcontext.UserIDFromContext(ctx)

	action := "test.action"
	resType := "instance"
	resID := "123"
	details := map[string]interface{}{"key": "value"}

	err := svc.Log(ctx, userID, action, resType, resID, details)
	assert.NoError(t, err)

	// Verify in DB
	logs, err := repo.ListByUserID(ctx, userID, 10)
	assert.NoError(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, action, logs[0].Action)
	assert.Equal(t, resID, logs[0].ResourceID)
}

func TestAuditServiceListLogs(t *testing.T) {
	svc, _, ctx := setupAuditServiceTest(t)
	userID := appcontext.UserIDFromContext(ctx)

	_ = svc.Log(ctx, userID, "action1", "res", "1", nil)
	_ = svc.Log(ctx, userID, "action2", "res", "2", nil)

	logs, err := svc.ListLogs(ctx, userID, 10)
	assert.NoError(t, err)
	assert.Len(t, logs, 2)
}
