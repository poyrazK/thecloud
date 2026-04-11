// Package services implements core business workflows.
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/redis/go-redis/v9"
)

type cachedIdentityService struct {
	base   ports.IdentityService
	redis  *redis.Client
	logger *slog.Logger
	ttl    time.Duration
}

// NewCachedIdentityService wraps an IdentityService with a redis-backed cache.
func NewCachedIdentityService(base ports.IdentityService, redis *redis.Client, logger *slog.Logger) ports.IdentityService {
	if redis == nil {
		return base
	}
	return &cachedIdentityService{
		base:   base,
		redis:  redis,
		logger: logger,
		ttl:    5 * time.Minute,
	}
}
func (s *cachedIdentityService) CreateKey(ctx context.Context, userID uuid.UUID, name string) (*domain.APIKey, error) {
	return s.base.CreateKey(ctx, userID, name)
}

func (s *cachedIdentityService) ValidateAPIKey(ctx context.Context, key string) (*domain.APIKey, error) {
	keyHash := computeKeyHash(key)
	cacheKey := fmt.Sprintf("apikey:hash:%s", keyHash)

	// Try cache
	val, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var apiKey domain.APIKey
		if err := json.Unmarshal([]byte(val), &apiKey); err == nil {
			return &apiKey, nil
		}
	}

	// Cache miss or error
	apiKey, err := s.base.ValidateAPIKey(ctx, key)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if data, err := json.Marshal(apiKey); err == nil {
		s.redis.Set(ctx, cacheKey, data, s.ttl)
	}

	return apiKey, nil
}

func (s *cachedIdentityService) ListKeys(ctx context.Context, userID uuid.UUID) ([]*domain.APIKey, error) {
	return s.base.ListKeys(ctx, userID)
}

func (s *cachedIdentityService) GetAPIKeyByID(ctx context.Context, id uuid.UUID) (*domain.APIKey, error) {
	return s.base.GetAPIKeyByID(ctx, id)
}

func (s *cachedIdentityService) RevokeKey(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	// Fetch key by ID (base, not cached) to get raw key for cache invalidation
	key, err := s.base.GetAPIKeyByID(ctx, id)
	if err != nil {
		return err
	}
	keyHash := computeKeyHash(key.Key)
	s.redis.Del(ctx, fmt.Sprintf("apikey:hash:%s", keyHash))
	return s.base.RevokeKey(ctx, userID, id)
}

func (s *cachedIdentityService) RotateKey(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*domain.APIKey, error) {
	// Fetch old key to invalidate its cache entry
	oldKey, err := s.base.GetAPIKeyByID(ctx, id)
	if err != nil {
		return nil, err
	}
	oldHash := computeKeyHash(oldKey.Key)
	s.redis.Del(ctx, fmt.Sprintf("apikey:hash:%s", oldHash))

	newKey, err := s.base.RotateKey(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	// New key's hash is cached by ValidateAPIKey on next use — no action needed here
	return newKey, nil
}
