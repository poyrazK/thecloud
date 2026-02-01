package services_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAccountingRepository
type MockAccountingRepo struct {
	mock.Mock
}

func (m *MockAccountingRepo) CreateRecord(ctx context.Context, record domain.UsageRecord) error {
	return m.Called(ctx, record).Error(0)
}
func (m *MockAccountingRepo) GetUsageSummary(ctx context.Context, userID uuid.UUID, start, end time.Time) (map[domain.ResourceType]float64, error) {
	args := m.Called(ctx, userID, start, end)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[domain.ResourceType]float64), args.Error(1)
}
func (m *MockAccountingRepo) ListRecords(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]domain.UsageRecord, error) {
	args := m.Called(ctx, userID, start, end)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.UsageRecord), args.Error(1)
}

func TestTrackUsage(t *testing.T) {
	repo := new(MockAccountingRepo)
	instRepo := new(MockInstanceRepo) // Uses shared mock from services_test package
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := services.NewAccountingService(repo, instRepo, logger)

	record := domain.UsageRecord{
		UserID:       uuid.New(),
		ResourceID:   uuid.New(),
		ResourceType: domain.ResourceInstance,
		Quantity:     10,
	}

	repo.On("CreateRecord", mock.Anything, mock.MatchedBy(func(r domain.UsageRecord) bool {
		return r.UserID == record.UserID && r.ResourceType == record.ResourceType
	})).Return(nil)

	err := svc.TrackUsage(context.Background(), record)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestProcessHourlyBilling(t *testing.T) {
	repo := new(MockAccountingRepo)
	instRepo := new(MockInstanceRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := services.NewAccountingService(repo, instRepo, logger)

	userID := uuid.New()
	instID := uuid.New()
	instances := []*domain.Instance{
		{ID: instID, UserID: userID, Status: domain.StatusRunning},
		{ID: uuid.New(), UserID: userID, Status: domain.StatusStopped},
	}

	// Expect ListAll (which maps to List in shared mock)
	instRepo.On("List", mock.Anything).Return(instances, nil)

	// Expect CreateRecord for the running instance
	repo.On("CreateRecord", mock.Anything, mock.MatchedBy(func(r domain.UsageRecord) bool {
		return r.UserID == userID && r.ResourceID == instID && r.Quantity == 60
	})).Return(nil)

	err := svc.ProcessHourlyBilling(context.Background())
	assert.NoError(t, err)
	instRepo.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestGetSummary(t *testing.T) {
	repo := new(MockAccountingRepo)
	instRepo := new(MockInstanceRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := services.NewAccountingService(repo, instRepo, logger)

	userID := uuid.New()
	now := time.Now()

	usage := map[domain.ResourceType]float64{
		domain.ResourceInstance: 100, // 100 minutes * 0.01 = $1.00
		domain.ResourceStorage:  200, // 200 units * 0.005 = $1.00
	}

	repo.On("GetUsageSummary", mock.Anything, userID, mock.Anything, mock.Anything).Return(usage, nil)

	summary, err := svc.GetSummary(context.Background(), userID, now.Add(-24*time.Hour), now)
	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, 2.0, summary.TotalAmount)
	repo.AssertExpectations(t)
}

func TestProcessHourlyBillingError(t *testing.T) {
	repo := new(MockAccountingRepo)
	instRepo := new(MockInstanceRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := services.NewAccountingService(repo, instRepo, logger)

	instRepo.On("List", mock.Anything).Return(nil, assert.AnError)

	err := svc.ProcessHourlyBilling(context.Background())
	assert.Error(t, err)
}

func TestProcessHourlyBillingRecordError(t *testing.T) {
	repo := new(MockAccountingRepo)
	instRepo := new(MockInstanceRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := services.NewAccountingService(repo, instRepo, logger)

	userID := uuid.New()
	instID := uuid.New()
	instances := []*domain.Instance{
		{ID: instID, UserID: userID, Status: domain.StatusRunning},
	}

	instRepo.On("List", mock.Anything).Return(instances, nil)
	repo.On("CreateRecord", mock.Anything, mock.Anything).Return(assert.AnError)

	err := svc.ProcessHourlyBilling(context.Background())
	// It continues on error but logs it
	assert.NoError(t, err)
}

func TestListUsage(t *testing.T) {
	repo := new(MockAccountingRepo)
	instRepo := new(MockInstanceRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := services.NewAccountingService(repo, instRepo, logger)

	userID := uuid.New()
	records := []domain.UsageRecord{{ID: uuid.New()}}
	repo.On("ListRecords", mock.Anything, userID, mock.Anything, mock.Anything).Return(records, nil)

	res, err := svc.ListUsage(context.Background(), userID, time.Now(), time.Now())
	assert.NoError(t, err)
	assert.Equal(t, records, res)
}
