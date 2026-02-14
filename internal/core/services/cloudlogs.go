package services

import (
	"context"
	"log/slog"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"go.opentelemetry.io/otel/trace"
)

// CloudLogsService implements the LogService port.
type CloudLogsService struct {
	repo   ports.LogRepository
	logger *slog.Logger
}

// NewCloudLogsService creates a new CloudLogsService.
func NewCloudLogsService(repo ports.LogRepository, logger *slog.Logger) *CloudLogsService {
	return &CloudLogsService{
		repo:   repo,
		logger: logger,
	}
}

func (s *CloudLogsService) IngestLogs(ctx context.Context, entries []*domain.LogEntry) error {
	if len(entries) == 0 {
		return nil
	}

	// Extract TraceID from context if available
	span := trace.SpanFromContext(ctx)
	traceID := ""
	if span.SpanContext().HasTraceID() {
		traceID = span.SpanContext().TraceID().String()
	}

	for _, entry := range entries {
		if entry.TraceID == "" {
			entry.TraceID = traceID
		}
	}

	// Validate or process entries if needed before persistence
	return s.repo.Create(ctx, entries)
}

func (s *CloudLogsService) SearchLogs(ctx context.Context, query domain.LogQuery) ([]*domain.LogEntry, int, error) {
	return s.repo.List(ctx, query)
}

func (s *CloudLogsService) RunRetentionPolicy(ctx context.Context, days int) error {
	s.logger.Info("running log retention policy", "days", days)
	return s.repo.DeleteByAge(ctx, days)
}
