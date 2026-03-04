package services

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
	"go.opentelemetry.io/otel/trace"
)

// CloudLogsService implements the LogService port.
type CloudLogsService struct {
	repo    ports.LogRepository
	rbacSvc ports.RBACService
	logger  *slog.Logger
}

// NewCloudLogsService creates a new CloudLogsService.
func NewCloudLogsService(repo ports.LogRepository, rbacSvc ports.RBACService, logger *slog.Logger) *CloudLogsService {
	return &CloudLogsService{
		repo:    repo,
		rbacSvc: rbacSvc,
		logger:  logger,
	}
}

func (s *CloudLogsService) IngestLogs(ctx context.Context, entries []*domain.LogEntry) error {
	if len(entries) == 0 {
		return nil
	}

	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	// Writing logs is typically a write or launch permission, but audit is safer for general logs
	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionInstanceUpdate); err != nil {
		return err
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
		// Ensure TenantID is set
		if entry.TenantID == uuid.Nil {
			entry.TenantID = tenantID
		}
	}

	// Validate or process entries if needed before persistence
	return s.repo.Create(ctx, entries)
}

func (s *CloudLogsService) SearchLogs(ctx context.Context, query domain.LogQuery) ([]*domain.LogEntry, int, error) {
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionAuditRead); err != nil {
		return nil, 0, err
	}

	return s.repo.List(ctx, query)
}

func (s *CloudLogsService) RunRetentionPolicy(ctx context.Context, days int) error {
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	// Privileged system operation
	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionFullAccess); err != nil {
		return err
	}

	if days <= 0 {
		return errors.New(errors.InvalidInput, "invalid retention days; must be > 0")
	}
	s.logger.Info("running log retention policy", "days", days)
	return s.repo.DeleteByAge(ctx, days)
}
