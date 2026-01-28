// Package domain defines core business entities.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// LifecycleRule describes expiration and prefix rules for bucket objects.
type LifecycleRule struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	BucketName     string    `json:"bucket_name"`
	Prefix         string    `json:"prefix"`
	ExpirationDays int       `json:"expiration_days"`
	Enabled        bool      `json:"enabled"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
