package domain

import (
	"time"

	"github.com/google/uuid"
)

type Object struct {
	ID          uuid.UUID  `json:"id"`
	ARN         string     `json:"arn"`
	Bucket      string     `json:"bucket"`
	Key         string     `json:"key"`
	SizeBytes   int64      `json:"size_bytes"`
	ContentType string     `json:"content_type"`
	CreatedAt   time.Time  `json:"created_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}
