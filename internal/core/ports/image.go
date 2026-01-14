// Package ports defines service and repository interfaces.
package ports

import (
	"context"
	"io"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

// ImageService provides business logic for managing bootable machine images and templates.
type ImageService interface {
	// RegisterImage creates metadata for a new system image.
	RegisterImage(ctx context.Context, name, description, os, version string, isPublic bool) (*domain.Image, error)
	// UploadImage streams the binary image data to the underlying storage backend.
	UploadImage(ctx context.Context, id uuid.UUID, reader io.Reader) error
	// GetImage retrieves details for a specific image by its UUID.
	GetImage(ctx context.Context, id uuid.UUID) (*domain.Image, error)
	// ListImages returns images available to the user, including private and optionally public ones.
	ListImages(ctx context.Context, userID uuid.UUID, includePublic bool) ([]*domain.Image, error)
	// DeleteImage decommission an image and removes its associated storage artifacts.
	DeleteImage(ctx context.Context, id uuid.UUID) error
}

// ImageRepository handles the persistence of image metadata.
type ImageRepository interface {
	// Create saves a new image record.
	Create(ctx context.Context, image *domain.Image) error
	// GetByID fetches an image by its unique ID.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Image, error)
	// List retrieves images based on ownership and visibility flags.
	List(ctx context.Context, userID uuid.UUID, includePublic bool) ([]*domain.Image, error)
	// Update modifies an existing image's metadata or status.
	Update(ctx context.Context, image *domain.Image) error
	// Delete removes an image record from storage.
	Delete(ctx context.Context, id uuid.UUID) error
}
