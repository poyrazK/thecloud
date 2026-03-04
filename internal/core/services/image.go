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
	rbacSvc    ports.RBACService
	fileStore  ports.FileStore
	bucketName string
	logger     *slog.Logger
}

// NewImageService constructs the image service for managing custom images.
func NewImageService(repo ports.ImageRepository, rbacSvc ports.RBACService, fileStore ports.FileStore, logger *slog.Logger) ports.ImageService {
	return &imageService{
		repo:       repo,
		rbacSvc:    rbacSvc,
		fileStore:  fileStore,
		bucketName: "images",
		logger:     logger,
	}
}

func (s *imageService) RegisterImage(ctx context.Context, name, description, os, version string, isPublic bool) (*domain.Image, error) {
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionImageCreate); err != nil {
		return nil, err
	}

	img := &domain.Image{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		OS:          os,
		Version:     version,
		IsPublic:    isPublic,
		UserID:      userID,
		TenantID:    &tenantID,
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
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionImageCreate); err != nil {
		return err
	}

	img, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Double check ownership via RBAC or simple ID match if non-admin
	if img.UserID != userID {
		// Try authorizing as admin for override if needed, but usually upload is owner-only
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
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionImageRead); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, id)
}

func (s *imageService) ListImages(ctx context.Context, userID uuid.UUID, includePublic bool) ([]*domain.Image, error) {
	uID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, uID, tenantID, domain.PermissionImageRead); err != nil {
		return nil, err
	}

	return s.repo.List(ctx, userID, includePublic)
}

func (s *imageService) DeleteImage(ctx context.Context, id uuid.UUID) error {
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionImageDelete); err != nil {
		return err
	}

	img, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Permission check
	if img.UserID != userID {
		// Log or handle unauthorized delete attempt if needed
	}

	if img.FilePath != "" {
		_ = s.fileStore.Delete(ctx, s.bucketName, img.FilePath)
	}

	return s.repo.Delete(ctx, id)
}
