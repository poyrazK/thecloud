package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
)

type imageRepository struct {
	db DB
}

func NewImageRepository(db DB) ports.ImageRepository {
	return &imageRepository{db: db}
}

func (r *imageRepository) Create(ctx context.Context, image *domain.Image) error {
	query := `
		INSERT INTO images (id, name, description, os, version, size_gb, file_path, format, is_public, user_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.db.Exec(ctx, query,
		image.ID, image.Name, image.Description, image.OS, image.Version,
		image.SizeGB, image.FilePath, image.Format, image.IsPublic, image.UserID,
		image.Status, image.CreatedAt, image.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create image: %w", err)
	}
	return nil
}

func (r *imageRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Image, error) {
	query := `
		SELECT id, name, description, os, version, size_gb, file_path, format, is_public, user_id, status, created_at, updated_at
		FROM images WHERE id = $1
	`
	var img domain.Image
	err := r.db.QueryRow(ctx, query, id).Scan(
		&img.ID, &img.Name, &img.Description, &img.OS, &img.Version,
		&img.SizeGB, &img.FilePath, &img.Format, &img.IsPublic, &img.UserID,
		&img.Status, &img.CreatedAt, &img.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New(errors.NotFound, "image not found")
		}
		return nil, fmt.Errorf("failed to get image: %w", err)
	}
	return &img, nil
}

func (r *imageRepository) List(ctx context.Context, userID uuid.UUID, includePublic bool) ([]*domain.Image, error) {
	query := `
		SELECT id, name, description, os, version, size_gb, file_path, format, is_public, user_id, status, created_at, updated_at
		FROM images 
		WHERE user_id = $1 OR ($2 = true AND is_public = true)
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID, includePublic)
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}
	defer rows.Close()

	var images []*domain.Image
	for rows.Next() {
		var img domain.Image
		err := rows.Scan(
			&img.ID, &img.Name, &img.Description, &img.OS, &img.Version,
			&img.SizeGB, &img.FilePath, &img.Format, &img.IsPublic, &img.UserID,
			&img.Status, &img.CreatedAt, &img.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan image: %w", err)
		}
		images = append(images, &img)
	}
	return images, nil
}

func (r *imageRepository) Update(ctx context.Context, img *domain.Image) error {
	img.UpdatedAt = time.Now()
	query := `
		UPDATE images 
		SET name = $1, description = $2, os = $3, version = $4, size_gb = $5, 
		    file_path = $6, format = $7, is_public = $8, status = $9, updated_at = $10
		WHERE id = $11
	`
	_, err := r.db.Exec(ctx, query,
		img.Name, img.Description, img.OS, img.Version, img.SizeGB,
		img.FilePath, img.Format, img.IsPublic, img.Status, img.UpdatedAt,
		img.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update image: %w", err)
	}
	return nil
}

func (r *imageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM images WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}
	return nil
}
