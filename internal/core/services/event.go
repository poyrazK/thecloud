package services

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/poyraz/cloud/internal/core/domain"
	"github.com/poyraz/cloud/internal/core/ports"
)

type EventService struct {
	repo   ports.EventRepository
	logger *slog.Logger
}

func NewEventService(repo ports.EventRepository, logger *slog.Logger) *EventService {
	return &EventService{
		repo:   repo,
		logger: logger,
	}
}

func (s *EventService) RecordEvent(ctx context.Context, action, resourceID, resourceType string, metadata map[string]interface{}) error {
	metaJSON, _ := json.Marshal(metadata)

	event := &domain.Event{
		ID:           uuid.New(),
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

	return nil
}

func (s *EventService) ListEvents(ctx context.Context, limit int) ([]*domain.Event, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.List(ctx, limit)
}
