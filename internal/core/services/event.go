// Package services implements core business workflows.
package services

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/handlers/ws"
)

// EventServiceParams defines the dependencies for EventService.
type EventServiceParams struct {
	Repo    ports.EventRepository
	RBACSvc ports.RBACService
	Hub     *ws.Hub
	Logger  *slog.Logger
}

// EventService records events and emits websocket notifications.
type EventService struct {
	repo    ports.EventRepository
	rbacSvc ports.RBACService
	hub     *ws.Hub
	logger  *slog.Logger
}

// NewEventService constructs an EventService with its dependencies.
func NewEventService(params EventServiceParams) *EventService {
	return &EventService{
		repo:    params.Repo,
		rbacSvc: params.RBACSvc,
		hub:     params.Hub,
		logger:  params.Logger,
	}
}

func (s *EventService) RecordEvent(ctx context.Context, action, resourceID, resourceType string, metadata map[string]interface{}) error {
	metaJSON, _ := json.Marshal(metadata)

	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	event := &domain.Event{
		ID:           uuid.New(),
		UserID:       userID,
		TenantID:     tenantID,
		Action:       action,
		ResourceID:   resourceID,
		ResourceType: resourceType,
		Metadata:     metaJSON,
		CreatedAt:    time.Now(),
	}

	// We don't want event recording to block the main flow or return error to user if it fails
	// But since this method returns error, the caller can decide.
	// Usually, we might want to run this in a goroutine or just log error and proceed.
	// For now, we execute synchronously but the caller (InstanceService) can ignore error.

	if err := s.repo.Create(ctx, event); err != nil {
		s.logger.Error("failed to record event", "action", action, "error", err)
		return err
	}

	// Real-time broadcast
	if s.hub != nil {
		wsEvent, err := domain.NewWSEvent(domain.WSEventAuditLog, event)
		if err == nil {
			s.hub.BroadcastEvent(wsEvent)
		}
	}

	return nil
}

func (s *EventService) ListEvents(ctx context.Context, limit int) ([]*domain.Event, error) {
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionAuditRead, "*"); err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 50
	}
	return s.repo.List(ctx, limit)
}
