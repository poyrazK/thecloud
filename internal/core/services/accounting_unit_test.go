package services_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAccountingServiceUnit(t *testing.T) {
	mockRepo := new(MockAccountingRepo)
	mockInstRepo := new(MockInstanceRepo)
	rbacSvc := new(MockRBACService)
	rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	svc := services.NewAccountingService(mockRepo, rbacSvc, mockInstRepo, slog.Default())
	ctx := context.Background()

	t.Run("TrackUsage", func(t *testing.T) {
		record := domain.UsageRecord{
			UserID:   uuid.New(),
			Quantity: 10,
		}

		mockRepo.On("CreateRecord", mock.Anything, mock.MatchedBy(func(r domain.UsageRecord) bool {
			return r.UserID == record.UserID && r.ID != uuid.Nil
		})).Return(nil).Once()

		err := svc.TrackUsage(ctx, record)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("GetSummary", func(t *testing.T) {
		userID := uuid.New()
		start := time.Now().Add(-24 * time.Hour)
		end := time.Now()

		usage := map[domain.ResourceType]float64{
			domain.ResourceInstance: 100,
			domain.ResourceStorage:  200,
		}

		mockRepo.On("GetUsageSummary", mock.Anything, userID, mock.Anything, mock.Anything).Return(usage, nil).Once()

		summary, err := svc.GetSummary(ctx, userID, start, end)
		require.NoError(t, err)
		assert.InDelta(t, 2.0, summary.TotalAmount, 0.01)
		assert.Equal(t, userID, summary.UserID)
	})

	t.Run("ListUsage", func(t *testing.T) {
		userID := uuid.New()
		start := time.Now().Add(-24 * time.Hour)
		end := time.Now()

		mockRepo.On("ListRecords", mock.Anything, userID, mock.Anything, mock.Anything).Return([]domain.UsageRecord{}, nil).Once()

		res, err := svc.ListUsage(ctx, userID, start, end)
		require.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("ProcessHourlyBilling", func(t *testing.T) {
		inst1 := &domain.Instance{ID: uuid.New(), UserID: uuid.New(), Status: domain.StatusRunning}
		inst2 := &domain.Instance{ID: uuid.New(), UserID: uuid.New(), Status: domain.StatusStopped}
		instances := []*domain.Instance{inst1, inst2}

		// The mock implementation of ListAll calls List in the worker but here we mock ListAll directly
		mockInstRepo.On("ListAll", mock.Anything).Return(instances, nil).Once()
		mockRepo.On("CreateRecord", mock.Anything, mock.MatchedBy(func(r domain.UsageRecord) bool {
			return r.ResourceID == inst1.ID
		})).Return(nil).Once()

		err := svc.ProcessHourlyBilling(ctx)
		require.NoError(t, err)
		mockInstRepo.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}
