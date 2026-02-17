package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAccountingRepository struct {
	mock.Mock
}

func (m *MockAccountingRepository) CreateRecord(ctx context.Context, record domain.UsageRecord) error {
	return m.Called(ctx, record).Error(0)
}
func (m *MockAccountingRepository) ListRecords(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]domain.UsageRecord, error) {
	return nil, nil
}
func (m *MockAccountingRepository) GetUsageSummary(ctx context.Context, userID uuid.UUID, start, end time.Time) (map[domain.ResourceType]float64, error) {
	return nil, nil
}

func TestAccountingService_Internal(t *testing.T) {
	repo := new(MockAccountingRepository)
	s := &accountingService{repo: repo}
	ctx := context.Background()

	t.Run("TrackUsage generates ID", func(t *testing.T) {
		record := domain.UsageRecord{UserID: uuid.New()}
		repo.On("CreateRecord", mock.Anything, mock.MatchedBy(func(r domain.UsageRecord) bool {
			return r.ID != uuid.Nil
		})).Return(nil).Once()

		err := s.TrackUsage(ctx, record)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})
}
