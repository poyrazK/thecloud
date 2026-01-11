package domain

import (
	"time"

	"github.com/google/uuid"
)

type ImageStatus string

const (
	ImageStatusPending  ImageStatus = "PENDING"
	ImageStatusActive   ImageStatus = "ACTIVE"
	ImageStatusError    ImageStatus = "ERROR"
	ImageStatusDeleting ImageStatus = "DELETING"
)

type Image struct {
	ID          uuid.UUID   `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	OS          string      `json:"os"`
	Version     string      `json:"version"`
	SizeGB      int         `json:"size_gb"`
	FilePath    string      `json:"file_path"`
	Format      string      `json:"format"`
	IsPublic    bool        `json:"is_public"`
	UserID      uuid.UUID   `json:"user_id"`
	Status      ImageStatus `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}
