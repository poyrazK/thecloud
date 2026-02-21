package services_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAccountingServiceTest(t *testing.T) (ports.AccountingService, ports.AccountingRepository, *postgres.InstanceRepository, context.Context, *pgxpool.Pool) {
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewAccountingRepository(db)
	instRepo := postgres.NewInstanceRepository(db)
	logger := slog.Default()

	svc := services.NewAccountingService(repo, instRepo, logger)

	return svc, repo, instRepo, ctx, db
}

func TestTrackUsage(t *testing.T) {
	svc, repo, _, ctx, db := setupAccountingServiceTest(t)
	defer db.Close()
	userID := appcontext.UserIDFromContext(ctx)

	record := domain.UsageRecord{
		UserID:       userID,
		ResourceID:   uuid.New(),
		ResourceType: domain.ResourceInstance,
		Quantity:     10,
		StartTime:    time.Now(),
		EndTime:      time.Now().Add(10 * time.Minute),
	}

	err := svc.TrackUsage(ctx, record)
	assert.NoError(t, err)

	// Verify in DB
	records, err := repo.ListRecords(ctx, userID, time.Now().Add(-1*time.Hour), time.Now().Add(1*time.Hour))
	assert.NoError(t, err)
	assert.Len(t, records, 1)
	assert.InDelta(t, 10.0, records[0].Quantity, 0.001)
}

func TestProcessHourlyBilling(t *testing.T) {
	svc, repo, instRepo, ctx, db := setupAccountingServiceTest(t)
	defer db.Close()
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	// Create running instance
	instance := &domain.Instance{
		ID:           uuid.New(),
		UserID:       userID,
		TenantID:     tenantID,
		Name:         "test-inst",
		Image:        "ubuntu",
		InstanceType: "small", // Assuming InstanceType replaced Plan or is what was meant
		Status:       domain.StatusRunning,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := instRepo.Create(ctx, instance)
	require.NoError(t, err)

	err = svc.ProcessHourlyBilling(ctx)
	assert.NoError(t, err)

	// Verify usage record created
	records, err := repo.ListRecords(ctx, userID, time.Now().Add(-2*time.Hour), time.Now().Add(2*time.Hour))
	assert.NoError(t, err)
	assert.NotEmpty(t, records)

	// We expect one record for the running instance
	found := false
	for _, r := range records {
		if r.ResourceID == instance.ID && r.Quantity == 60 { // Hourly billing assumes 60 mins?
			found = true
			break
		}
	}
	assert.True(t, found, "should have found usage record for instance")
}

func TestGetSummary(t *testing.T) {
	svc, _, _, ctx, db := setupAccountingServiceTest(t)
	defer db.Close()
	userID := appcontext.UserIDFromContext(ctx)
	now := time.Now()

	// Add some usage manually via service or repo
	rec1 := domain.UsageRecord{
		UserID:       userID,
		ResourceID:   uuid.New(),
		ResourceType: domain.ResourceInstance,
		Quantity:     100, // 100 mins
		StartTime:    now.Add(-2 * time.Hour),
		EndTime:      now.Add(-1 * time.Hour),
	}
	err := svc.TrackUsage(ctx, rec1)
	require.NoError(t, err)

	rec2 := domain.UsageRecord{
		UserID:       userID,
		ResourceID:   uuid.New(),
		ResourceType: domain.ResourceStorage,
		Quantity:     10, // 10 GB
		StartTime:    now.Add(-2 * time.Hour),
		EndTime:      now.Add(-1 * time.Hour),
	}
	err = svc.TrackUsage(ctx, rec2)
	require.NoError(t, err)

	// Get Summary
	summary, err := svc.GetSummary(ctx, userID, now.Add(-24*time.Hour), now)
	assert.NoError(t, err)
	assert.NotNil(t, summary)

	// Verify TotalAmount
	assert.Positive(t, summary.TotalAmount)
}

func TestListUsage(t *testing.T) {
	svc, _, _, ctx, db := setupAccountingServiceTest(t)
	defer db.Close()
	userID := appcontext.UserIDFromContext(ctx)

	rec := domain.UsageRecord{
		UserID:       userID,
		ResourceID:   uuid.New(),
		ResourceType: domain.ResourceInstance,
		Quantity:     50,
		StartTime:    time.Now(),
		EndTime:      time.Now(),
	}
	_ = svc.TrackUsage(ctx, rec)

	res, err := svc.ListUsage(ctx, userID, time.Now().Add(-1*time.Hour), time.Now().Add(1*time.Hour))
	assert.NoError(t, err)
	assert.Len(t, res, 1)
}
