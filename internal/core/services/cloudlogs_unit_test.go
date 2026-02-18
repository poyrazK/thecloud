package services_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

func TestCloudLogsService_IngestLogs_Unit(t *testing.T) {
	mockRepo := new(MockLogRepository)
	svc := services.NewCloudLogsService(mockRepo, slog.Default())
	ctx := context.Background()

	t.Run("success with entries", func(t *testing.T) {
		entries := []*domain.LogEntry{
			{ID: uuid.New(), Message: "test 1"},
		}
		mockRepo.On("Create", mock.Anything, entries).Return(nil).Once()
		err := svc.IngestLogs(ctx, entries)
		assert.NoError(t, err)
	})

	t.Run("empty entries", func(t *testing.T) {
		err := svc.IngestLogs(ctx, nil)
		assert.NoError(t, err)
	})

	t.Run("preserve existing trace id", func(t *testing.T) {
		entries := []*domain.LogEntry{
			{ID: uuid.New(), Message: "test trace", TraceID: "existing-trace"},
		}
		mockRepo.On("Create", mock.Anything, entries).Return(nil).Once()
		err := svc.IngestLogs(ctx, entries)
		assert.NoError(t, err)
		assert.Equal(t, "existing-trace", entries[0].TraceID)
	})

	t.Run("extract trace id from context", func(t *testing.T) {
		res, _ := resource.New(ctx, resource.WithAttributes(semconv.ServiceNameKey.String("test")))
		tp := sdktrace.NewTracerProvider(sdktrace.WithResource(res))
		tr := tp.Tracer("test")
		
		ctxWithTrace, span := tr.Start(ctx, "test-span")
		defer span.End()
		
		traceID := span.SpanContext().TraceID().String()

		entries := []*domain.LogEntry{
			{ID: uuid.New(), Message: "test context trace"},
		}
		mockRepo.On("Create", mock.Anything, entries).Return(nil).Once()
		
		err := svc.IngestLogs(ctxWithTrace, entries)
		assert.NoError(t, err)
		assert.Equal(t, traceID, entries[0].TraceID)
	})

	t.Run("no trace id in context", func(t *testing.T) {
		// Use a context with a no-op span (no trace ID)
		ctxNoTrace := trace.ContextWithSpan(ctx, trace.SpanFromContext(context.Background()))
		
		entries := []*domain.LogEntry{
			{ID: uuid.New(), Message: "no trace"},
		}
		mockRepo.On("Create", mock.Anything, entries).Return(nil).Once()
		
		err := svc.IngestLogs(ctxNoTrace, entries)
		assert.NoError(t, err)
		assert.Empty(t, entries[0].TraceID)
	})

	mockRepo.AssertExpectations(t)
}

func TestCloudLogsService_SearchLogs_Unit(t *testing.T) {
	mockRepo := new(MockLogRepository)
	svc := services.NewCloudLogsService(mockRepo, slog.Default())
	ctx := context.Background()

	query := domain.LogQuery{ResourceID: "res-1"}
	expectedLogs := []*domain.LogEntry{{Message: "found"}}
	
	mockRepo.On("List", mock.Anything, query).Return(expectedLogs, 1, nil)

	logs, total, err := svc.SearchLogs(ctx, query)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Equal(t, "found", logs[0].Message)
	mockRepo.AssertExpectations(t)
}

func TestCloudLogsService_RunRetentionPolicy_Unit(t *testing.T) {
	mockRepo := new(MockLogRepository)
	svc := services.NewCloudLogsService(mockRepo, slog.Default())
	ctx := context.Background()

	mockRepo.On("DeleteByAge", mock.Anything, 30).Return(nil)

	err := svc.RunRetentionPolicy(ctx, 30)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
