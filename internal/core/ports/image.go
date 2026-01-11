package ports

import (
	"context"
	"io"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

type ImageService interface {
	RegisterImage(ctx context.Context, name, description, os, version string, isPublic bool) (*domain.Image, error)
	UploadImage(ctx context.Context, id uuid.UUID, reader io.Reader) error
	GetImage(ctx context.Context, id uuid.UUID) (*domain.Image, error)
	ListImages(ctx context.Context, userID uuid.UUID, includePublic bool) ([]*domain.Image, error)
	DeleteImage(ctx context.Context, id uuid.UUID) error
}

type ImageRepository interface {
	Create(ctx context.Context, image *domain.Image) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Image, error)
	List(ctx context.Context, userID uuid.UUID, includePublic bool) ([]*domain.Image, error)
	Update(ctx context.Context, image *domain.Image) error
	Delete(ctx context.Context, id uuid.UUID) error
}
