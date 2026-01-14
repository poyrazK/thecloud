// Package domain defines core business entities.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// ImageStatus represents the lifecycle state of a machine image.
type ImageStatus string

const (
	// ImageStatusPending indicates the image is being imported or created.
	ImageStatusPending ImageStatus = "PENDING"
	// ImageStatusActive indicates the image is ready for use.
	ImageStatusActive ImageStatus = "ACTIVE"
	// ImageStatusError indicates image creation failed.
	ImageStatusError ImageStatus = "ERROR"
	// ImageStatusDeleting indicates the image is being removed.
	ImageStatusDeleting ImageStatus = "DELETING"
)

// Image represents a bootable operating system template (ISO, QCOW2, etc.).
type Image struct {
	ID          uuid.UUID   `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	OS          string      `json:"os"`        // e.g. "ubuntu", "centos"
	Version     string      `json:"version"`   // e.g. "22.04"
	SizeGB      int         `json:"size_gb"`   // Minimum disk size required
	FilePath    string      `json:"file_path"` // Path in object storage/filesystem
	Format      string      `json:"format"`    // e.g. "qcow2", "iso"
	IsPublic    bool        `json:"is_public"` // If true, available to all users
	UserID      uuid.UUID   `json:"user_id"`   // Owner (nil for system images if handled)
	Status      ImageStatus `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}
