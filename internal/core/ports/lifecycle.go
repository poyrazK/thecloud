// Package ports defines interfaces for adapters and services.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

// LifecycleRepository stores lifecycle rules for buckets.
type LifecycleRepository interface {
	Create(ctx context.Context, rule *domain.LifecycleRule) error
	Get(ctx context.Context, id uuid.UUID) (*domain.LifecycleRule, error)
	List(ctx context.Context, bucketName string) ([]*domain.LifecycleRule, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetEnabledRules(ctx context.Context) ([]*domain.LifecycleRule, error)
}

// LifecycleService manages lifecycle rules for storage buckets.
type LifecycleService interface {
	CreateRule(ctx context.Context, bucket string, prefix string, expirationDays int, enabled bool) (*domain.LifecycleRule, error)
	ListRules(ctx context.Context, bucket string) ([]*domain.LifecycleRule, error)
	DeleteRule(ctx context.Context, bucket string, ruleID string) error
}
