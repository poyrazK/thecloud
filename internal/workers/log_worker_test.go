package workers

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockLogService struct {
	mock.Mock
}

func (m *mockLogService) IngestLogs(ctx context.Context, entries []*domain.LogEntry) error {
	return m.Called(ctx, entries).Error(0)
}
func (m *mockLogService) SearchLogs(ctx context.Context, query domain.LogQuery) ([]*domain.LogEntry, int, error) {
	args := m.Called(ctx, query)
	r0, _ := args.Get(0).([]*domain.LogEntry)
	return r0, args.Int(1), args.Error(2)
}
func (m *mockLogService) RunRetentionPolicy(ctx context.Context, days int) error {
	return m.Called(ctx, days).Error(0)
}

func TestNewLogWorker(t *testing.T) {
	mockSvc := new(mockLogService)
	logger := slog.Default()
	worker := NewLogWorker(mockSvc, logger)
	assert.NotNil(t, worker)
	assert.Equal(t, mockSvc, worker.logSvc)
}

func TestLogWorker_Run(t *testing.T) {
	t.Run("success loop", func(t *testing.T) {
		mockSvc := new(mockLogService)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		worker := &LogWorker{
			logSvc:        mockSvc,
			logger:        logger,
			interval:      10 * time.Millisecond,
			retentionDays: 30,
		}

		mockSvc.On("RunRetentionPolicy", mock.Anything, 30).Return(nil)

		ctx, cancel := context.WithCancel(context.Background())
		var wg sync.WaitGroup
		wg.Add(1)

		go worker.Run(ctx, &wg)

		time.Sleep(35 * time.Millisecond)
		cancel()
		wg.Wait()

		mockSvc.AssertExpectations(t)
	})

	t.Run("error in loop", func(t *testing.T) {
		mockSvc := new(mockLogService)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		worker := &LogWorker{
			logSvc:        mockSvc,
			logger:        logger,
			interval:      10 * time.Millisecond,
			retentionDays: 30,
		}

		// Initial run success, first tick error
		mockSvc.On("RunRetentionPolicy", mock.Anything, 30).Return(nil).Once()
		mockSvc.On("RunRetentionPolicy", mock.Anything, 30).Return(errors.New("fail")).Once()
		mockSvc.On("RunRetentionPolicy", mock.Anything, 30).Return(nil)

		ctx, cancel := context.WithCancel(context.Background())
		var wg sync.WaitGroup
		wg.Add(1)

		go worker.Run(ctx, &wg)

		time.Sleep(35 * time.Millisecond)
		cancel()
		wg.Wait()

		mockSvc.AssertExpectations(t)
	})

	t.Run("context cancel", func(t *testing.T) {
		mockSvc := new(mockLogService)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		worker := &LogWorker{
			logSvc:        mockSvc,
			logger:        logger,
			interval:      1 * time.Second,
			retentionDays: 30,
		}

		mockSvc.On("RunRetentionPolicy", mock.Anything, 30).Return(nil).Once()

		ctx, cancel := context.WithCancel(context.Background())
		var wg sync.WaitGroup
		wg.Add(1)

		go worker.Run(ctx, &wg)
		// Small sleep to ensure initial run completes
		time.Sleep(5 * time.Millisecond)
		cancel()
		wg.Wait()

		mockSvc.AssertExpectations(t)
	})
}
