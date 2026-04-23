// Package services implements core business workflows.
package services

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
)

// ImageServiceParams defines the dependencies for ImageService.
type ImageServiceParams struct {
	Repo       ports.ImageRepository
	RBACSvc    ports.RBACService
	FileStore  ports.FileStore
	Logger     *slog.Logger
	BucketName string
}

type imageService struct {
	repo       ports.ImageRepository
	rbacSvc    ports.RBACService
	fileStore  ports.FileStore
	bucketName string
	logger     *slog.Logger
}

// NewImageService constructs the image service for managing custom images.
func NewImageService(params ImageServiceParams) ports.ImageService {
	bucketName := params.BucketName
	if bucketName == "" {
		bucketName = "images"
	}
	return &imageService{
		repo:       params.Repo,
		rbacSvc:    params.RBACSvc,
		fileStore:  params.FileStore,
		bucketName: bucketName,
		logger:     params.Logger,
	}
}

func (s *imageService) RegisterImage(ctx context.Context, name, description, os, version string, isPublic bool) (*domain.Image, error) {
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionImageCreate, "*"); err != nil {
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

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionImageCreate, id.String()); err != nil {
		return err
	}

	img, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Tenant boundary check
	if img.TenantID != nil && *img.TenantID != tenantID {
		return errors.New(errors.NotFound, "image not found")
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

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionImageRead, id.String()); err != nil {
		return nil, err
	}

	img, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Tenant boundary check
	if img.TenantID != nil && *img.TenantID != tenantID {
		return nil, errors.New(errors.NotFound, "image not found")
	}

	return img, nil
}

func (s *imageService) ListImages(ctx context.Context, userID uuid.UUID, includePublic bool) ([]*domain.Image, error) {
	uID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, uID, tenantID, domain.PermissionImageRead, "*"); err != nil {
		return nil, err
	}

	// Horizontal access check: if requesting images for another user, need elevated permission
	if userID != uID {
		if err := s.rbacSvc.Authorize(ctx, uID, tenantID, domain.PermissionImageReadAll, "*"); err != nil {
			return nil, errors.New(errors.Forbidden, "cannot list images for another user")
		}
	}

	return s.repo.List(ctx, userID, includePublic)
}

func (s *imageService) DeleteImage(ctx context.Context, id uuid.UUID) error {
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionImageDelete, id.String()); err != nil {
		return err
	}

	img, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Tenant boundary check
	if img.TenantID != nil && *img.TenantID != tenantID {
		return errors.New(errors.NotFound, "image not found")
	}

	// Permission check
	if img.UserID != userID {
		return errors.New(errors.Forbidden, "cannot delete someone else's image")
	}

	if img.FilePath != "" {
		if err := s.fileStore.Delete(ctx, s.bucketName, img.FilePath); err != nil {
			s.logger.Error("failed to delete image file from storage", "file_path", img.FilePath, "bucket", s.bucketName, "error", err)
			return fmt.Errorf("failed to delete image file: %w", err)
		}
	}

	return s.repo.Delete(ctx, id)
}

func (s *imageService) ImportImage(ctx context.Context, name, imageURL, description, os, version string, isPublic bool) (*domain.Image, error) {
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionImageCreate, "*"); err != nil {
		return nil, err
	}

	// Validate URL before creating any record (CodeQL: go/request-forgery)
	parsedURL, err := url.Parse(imageURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("invalid URL scheme: only http and https are allowed")
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
		SourceURL:   imageURL,
		Format:      formatFromURL(imageURL),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, img); err != nil {
		return nil, err
	}

	// Download and stream to storage
	if err := s.importFromURL(ctx, img, imageURL); err != nil {
		img.Status = domain.ImageStatusError
		_ = s.repo.Update(ctx, img)
		return nil, err
	}

	img.Status = domain.ImageStatusActive
	if err := s.repo.Update(ctx, img); err != nil {
		return nil, err
	}

	return img, nil
}

func formatFromURL(url string) string {
	ext := strings.ToLower(filepath.Ext(url))
	switch ext {
	case ".qcow2":
		return "qcow2"
	case ".img":
		return "img"
	case ".raw":
		return "raw"
	case ".iso":
		return "iso"
	default:
		return "qcow2"
	}
}

func (s *imageService) importFromURL(ctx context.Context, img *domain.Image, imageURL string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch image: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("remote returned status %d", resp.StatusCode)
	}

	key := fmt.Sprintf("%s.%s", img.ID.String(), img.Format)
	size, err := s.fileStore.Write(ctx, s.bucketName, key, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to store image: %w", err)
	}

	img.SizeGB = int(size / (1024 * 1024 * 1024))
	if img.SizeGB == 0 && size > 0 {
		img.SizeGB = 1
	}
	img.FilePath = key
	return nil
}
