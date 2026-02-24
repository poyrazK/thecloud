// Package services implements core business workflows.
package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/poyrazk/thecloud/internal/platform"
)

// IdentityService manages API key lifecycle and validation.
type IdentityService struct {
	repo     ports.IdentityRepository
	auditSvc ports.AuditService
	logger   *slog.Logger
}

// NewIdentityService constructs an IdentityService with its dependencies.
func NewIdentityService(repo ports.IdentityRepository, auditSvc ports.AuditService, logger *slog.Logger) *IdentityService {
	return &IdentityService{repo: repo, auditSvc: auditSvc, logger: logger}
}

func (s *IdentityService) CreateKey(ctx context.Context, userID uuid.UUID, name string) (*domain.APIKey, error) {
	// Generate a secure random key
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to generate secure key", err)
	}
	keyStr := "thecloud_" + hex.EncodeToString(b)

	apiKey := &domain.APIKey{
		ID:        uuid.New(),
		UserID:    userID,
		Key:       keyStr,
		Name:      name,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateAPIKey(ctx, apiKey); err != nil {
		return nil, err
	}

	// Log audit event
	_ = s.auditSvc.Log(ctx, userID, "api_key.create", "api_key", apiKey.ID.String(), map[string]interface{}{
		"name": name,
	})

	return apiKey, nil
}

func (s *IdentityService) ValidateAPIKey(ctx context.Context, key string) (*domain.APIKey, error) {
	apiKey, err := s.repo.GetAPIKeyByKey(ctx, key)
	if err != nil {
		return nil, errors.New(errors.Unauthorized, "invalid api key")
	}

	// Check expiration if set
	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return nil, errors.New(errors.Unauthorized, "api key has expired")
	}

	platform.AuthAttemptsTotal.WithLabelValues("success_api_key").Inc()

	return apiKey, nil
}

func (s *IdentityService) ListKeys(ctx context.Context, userID uuid.UUID) ([]*domain.APIKey, error) {
	return s.repo.ListAPIKeysByUserID(ctx, userID)
}

func (s *IdentityService) RevokeKey(ctx context.Context, userID, id uuid.UUID) error {
	// Verify ownership
	key, err := s.repo.GetAPIKeyByID(ctx, id)
	if err != nil {
		return err
	}

	if key.UserID != userID {
		return errors.New(errors.Forbidden, "unauthorized access to api key")
	}

	if err := s.repo.DeleteAPIKey(ctx, id); err != nil {
		return err
	}

	// Log audit event
	_ = s.auditSvc.Log(ctx, userID, "api_key.revoke", "api_key", id.String(), map[string]interface{}{
		"name": key.Name,
	})

	return nil
}

func (s *IdentityService) RotateKey(ctx context.Context, userID, id uuid.UUID) (*domain.APIKey, error) {
	// Verify ownership
	key, err := s.repo.GetAPIKeyByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if key.UserID != userID {
		return nil, errors.New(errors.Forbidden, "unauthorized access to api key")
	}

	// Create new key with same name
	newKey, err := s.CreateKey(ctx, userID, key.Name)
	if err != nil {
		return nil, err
	}

	// Delete old key
	if err := s.repo.DeleteAPIKey(ctx, id); err != nil {
		// Log error but we already have a new key
		if s.logger != nil {
			s.logger.Error("failed to delete old api key during rotation", "id", id, "error", err)
		}
		return newKey, nil
	}

	// Log audit event
	_ = s.auditSvc.Log(ctx, userID, "api_key.rotate", "api_key", id.String(), map[string]interface{}{
		"name":   key.Name,
		"new_id": newKey.ID.String(),
	})

	return newKey, nil
}
