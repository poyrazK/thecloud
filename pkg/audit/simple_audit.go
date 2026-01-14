// Package simpleaudit provides a simple implementation of the AuditLogger interface.
package simpleaudit

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

// SimpleAuditLogger logs audit events to a structured logger.
type SimpleAuditLogger struct {
	logger *slog.Logger
}

// NewSimpleAuditLogger constructs a SimpleAuditLogger.
func NewSimpleAuditLogger(logger *slog.Logger) ports.AuditLogger {
	return &SimpleAuditLogger{logger: logger}
}

func (l *SimpleAuditLogger) Log(ctx context.Context, entry *domain.AuditLog) error {
	userID := entry.UserID.String()
	if entry.UserID == uuid.Nil {
		userID = "anonymous"
	}

	// We log at INFO level but marked as AUDIT for easy filtering
	l.logger.Info("AUDIT_LOG",
		slog.String("type", "security_audit"),
		slog.String("id", entry.ID.String()),
		slog.String("action", entry.Action),
		slog.String("user_id", userID),
		slog.String("resource_type", entry.ResourceType),
		slog.String("resource_id", entry.ResourceID),
		slog.String("ip", entry.IPAddress),
		slog.String("user_agent", entry.UserAgent),
		slog.Any("details", entry.Details),
	)
	return nil
}
