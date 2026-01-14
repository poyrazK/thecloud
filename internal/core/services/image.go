// Package services implements core business workflows.
package services

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
)

type imageService struct {
	repo       ports.ImageRepository
	fileStore  ports.FileStore
	bucketName string
	logger     *slog.Logger
}

// NewImageService constructs the image service for managing custom images.
func NewImageService(repo ports.ImageRepository, fileStore ports.FileStore, logger *slog.Logger) ports.ImageService {
	return &imageService{
		repo:       repo,
		fileStore:  fileStore,
		bucketName: "images",
		logger:     logger,
	}
}

func (s *imageService) RegisterImage(ctx context.Context, name, description, os, version string, isPublic bool) (*domain.Image, error) {
	userID := appcontext.UserIDFromContext(ctx)
	if userID == uuid.Nil {
		return nil, errors.New(errors.Unauthorized, "user not found in context")
	}

	img := &domain.Image{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		OS:          os,
		Version:     version,
		IsPublic:    isPublic,
		UserID:      userID,
		Status:      domain.ImageStatusPending,
		Format:      "qcow2",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, img); err != nil {
		return nil, err
	}

	return img, nil
}

func (s *imageService) UploadImage(ctx context.Context, id uuid.UUID, reader io.Reader) error {
	img, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Permission check (only owner can upload)
	userID := appcontext.UserIDFromContext(ctx)
	if img.UserID != userID {
		return errors.New(errors.Forbidden, "cannot upload to someone else's image")
	}

	key := fmt.Sprintf("%s.qcow2", id.String())
	size, err := s.fileStore.Write(ctx, s.bucketName, key, reader)
	if err != nil {
		img.Status = domain.ImageStatusError
		_ = s.repo.Update(ctx, img)
		return err
	}

	img.SizeGB = int(size / (1024 * 1024 * 1024))
	if img.SizeGB == 0 && size > 0 {
		img.SizeGB = 1 // Min 1GB for display
	}
	img.FilePath = key
	img.Status = domain.ImageStatusActive

	return s.repo.Update(ctx, img)
}

func (s *imageService) GetImage(ctx context.Context, id uuid.UUID) (*domain.Image, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *imageService) ListImages(ctx context.Context, userID uuid.UUID, includePublic bool) ([]*domain.Image, error) {
	return s.repo.List(ctx, userID, includePublic)
}

func (s *imageService) DeleteImage(ctx context.Context, id uuid.UUID) error {
	img, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Permission check
	userID := appcontext.UserIDFromContext(ctx)
	if img.UserID != userID {
		return errors.New(errors.Forbidden, "cannot delete someone else's image")
	}

	if img.FilePath != "" {
		_ = s.fileStore.Delete(ctx, s.bucketName, img.FilePath)
	}

	return s.repo.Delete(ctx, id)
}
