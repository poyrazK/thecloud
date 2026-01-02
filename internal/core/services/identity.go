package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
)

type IdentityService struct {
	repo ports.IdentityRepository
}

func NewIdentityService(repo ports.IdentityRepository) *IdentityService {
	return &IdentityService{repo: repo}
}

func (s *IdentityService) CreateKey(ctx context.Context, userID uuid.UUID, name string) (*domain.ApiKey, error) {
	// Generate a secure random key
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to generate random key", err)
	}
	keyStr := "thecloud_" + hex.EncodeToString(b)

	apiKey := &domain.ApiKey{
		ID:        uuid.New(),
		UserID:    userID,
		Key:       keyStr,
		Name:      name,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateApiKey(ctx, apiKey); err != nil {
		return nil, err
	}

	return apiKey, nil
}

func (s *IdentityService) ValidateApiKey(ctx context.Context, key string) (*domain.ApiKey, error) {
	apiKey, err := s.repo.GetApiKeyByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	return apiKey, nil
}
