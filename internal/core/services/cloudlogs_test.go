package services_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
)

func setupCloudLogsServiceTest(t *testing.T) (*services.CloudLogsService, *postgres.LogRepository, context.Context) {
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewLogRepository(db)
	svc := services.NewCloudLogsService(repo, slog.Default())
	return svc, repo, ctx
}

func TestCloudLogsService_IngestLogs(t *testing.T) {
	svc, repo, ctx := setupCloudLogsServiceTest(t)
	tenantID := appcontext.TenantIDFromContext(ctx)

	entries := []*domain.LogEntry{
		{
			ID:           uuid.New(),
			TenantID:     tenantID,
			ResourceID:   "res-1",
			ResourceType: "instance",
			Level:        "INFO",
			Message:      "Test log 1",
			Timestamp:    time.Now(),
		},
		{
			ID:           uuid.New(),
			TenantID:     tenantID,
			ResourceID:   "res-1",
			ResourceType: "instance",
			Level:        "ERROR",
			Message:      "Test log 2",
			Timestamp:    time.Now(),
		},
	}

	err := svc.IngestLogs(ctx, entries)
	assert.NoError(t, err)

	// Verify persistence via repo
	logs, total, err := repo.List(ctx, domain.LogQuery{TenantID: tenantID, ResourceID: "res-1"})
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, logs, 2)
}

func TestCloudLogsService_SearchLogs(t *testing.T) {
	svc, _, ctx := setupCloudLogsServiceTest(t)
	tenantID := appcontext.TenantIDFromContext(ctx)

	_ = svc.IngestLogs(ctx, []*domain.LogEntry{
		{
			ID:           uuid.New(),
			TenantID:     tenantID,
			ResourceID:   "res-1",
			ResourceType: "instance",
			Level:        "INFO",
			Message:      "Hello world",
			Timestamp:    time.Now().Add(-10 * time.Minute),
		},
		{
			ID:           uuid.New(),
			TenantID:     tenantID,
			ResourceID:   "res-2",
			ResourceType: "function",
			Level:        "ERROR",
			Message:      "Critical failure",
			Timestamp:    time.Now(),
		},
	})

	// Test 1: Search by resource type
	logs, total, err := svc.SearchLogs(ctx, domain.LogQuery{TenantID: tenantID, ResourceType: "function"})
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Equal(t, "Critical failure", logs[0].Message)

	// Test 2: Search by message
	logs, total, err = svc.SearchLogs(ctx, domain.LogQuery{TenantID: tenantID, Search: "world"})
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Equal(t, "Hello world", logs[0].Message)
}

func TestCloudLogsService_Retention(t *testing.T) {
	svc, repo, ctx := setupCloudLogsServiceTest(t)
	tenantID := appcontext.TenantIDFromContext(ctx)

	// Ingest one old log and one new log
	_ = svc.IngestLogs(ctx, []*domain.LogEntry{
		{
			ID:           uuid.New(),
			TenantID:     tenantID,
			ResourceID:   "old",
			ResourceType: "instance",
			Level:        "INFO",
			Message:      "Old log",
			Timestamp:    time.Now().AddDate(0, 0, -40), // 40 days ago
		},
		{
			ID:           uuid.New(),
			TenantID:     tenantID,
			ResourceID:   "new",
			ResourceType: "instance",
			Level:        "INFO",
			Message:      "New log",
			Timestamp:    time.Now(),
		},
	})

	// Run retention for 30 days
	err := svc.RunRetentionPolicy(ctx, 30)
	assert.NoError(t, err)

	// Verify only new log remains
	logs, total, err := repo.List(ctx, domain.LogQuery{TenantID: tenantID})
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Equal(t, "new", logs[0].ResourceID)
}
