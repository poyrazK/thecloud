package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

func TestImageRepository_Create(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewImageRepository(mock)
	img := &domain.Image{
		ID:          uuid.New(),
		Name:        "ubuntu-22.04",
		Description: "Ubuntu 22.04 LTS",
		OS:          "linux",
		Version:     "22.04",
		SizeGB:      10,
		FilePath:    "/images/ubuntu.qcow2",
		Format:      "qcow2",
		IsPublic:    true,
		UserID:      uuid.New(),
		Status:      domain.ImageStatusActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mock.ExpectExec("INSERT INTO images").
		WithArgs(img.ID, img.Name, img.Description, img.OS, img.Version, img.SizeGB, img.FilePath, img.Format, img.IsPublic, img.UserID, img.Status, img.CreatedAt, img.UpdatedAt).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.Create(context.Background(), img)
	assert.NoError(t, err)
}

func TestImageRepository_GetByID(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewImageRepository(mock)
	id := uuid.New()
	now := time.Now()

	mock.ExpectQuery("SELECT id, name, description, os, version, size_gb, file_path, format, is_public, user_id, status, created_at, updated_at FROM images WHERE id = \\$1").
		WithArgs(id).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "description", "os", "version", "size_gb", "file_path", "format", "is_public", "user_id", "status", "created_at", "updated_at"}).
			AddRow(id, "ubuntu", "Ubuntu", "linux", "22.04", 10, "/path", "qcow2", true, uuid.New(), domain.ImageStatusActive, now, now))

	img, err := repo.GetByID(context.Background(), id)
	assert.NoError(t, err)
	assert.NotNil(t, img)
	assert.Equal(t, id, img.ID)
}

func TestImageRepository_List(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewImageRepository(mock)
	userID := uuid.New()
	now := time.Now()

	mock.ExpectQuery("SELECT id, name, description, os, version, size_gb, file_path, format, is_public, user_id, status, created_at, updated_at FROM images").
		WithArgs(userID, true).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "description", "os", "version", "size_gb", "file_path", "format", "is_public", "user_id", "status", "created_at", "updated_at"}).
			AddRow(uuid.New(), "ubuntu", "Ubuntu", "linux", "22.04", 10, "/path", "qcow2", true, userID, domain.ImageStatusActive, now, now))

	images, err := repo.List(context.Background(), userID, true)
	assert.NoError(t, err)
	assert.Len(t, images, 1)
}

func TestImageRepository_Update(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewImageRepository(mock)
	img := &domain.Image{
		ID:          uuid.New(),
		Name:        "updated-name",
		Description: "Updated desc",
		OS:          "linux",
		Version:     "22.04",
		SizeGB:      20,
		FilePath:    "/new/path",
		Format:      "raw",
		IsPublic:    false,
		Status:      domain.ImageStatusActive,
		UpdatedAt:   time.Now(),
	}

	mock.ExpectExec("UPDATE images SET").
		WithArgs(img.Name, img.Description, img.OS, img.Version, img.SizeGB, img.FilePath, img.Format, img.IsPublic, img.Status, pgxmock.AnyArg(), img.ID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err = repo.Update(context.Background(), img)
	assert.NoError(t, err)
}

func TestImageRepository_Delete(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewImageRepository(mock)
	id := uuid.New()

	mock.ExpectExec("DELETE FROM images WHERE id = \\$1").
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = repo.Delete(context.Background(), id)
	assert.NoError(t, err)
}
