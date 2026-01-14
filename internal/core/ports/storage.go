// Package ports defines service and repository interfaces.
package ports

import (
	"context"
	"io"

	"github.com/poyrazk/thecloud/internal/core/domain"
)

// StorageRepository manages the persistence of object storage metadata (e.g., File name, size, ETag).
type StorageRepository interface {
	// SaveMeta persists or updates metadata for a storage object.
	SaveMeta(ctx context.Context, obj *domain.Object) error
	// GetMeta retrieves metadata for a specific object in a bucket.
	GetMeta(ctx context.Context, bucket, key string) (*domain.Object, error)
	// List returns metadata for all objects currently stored in a specific bucket.
	List(ctx context.Context, bucket string) ([]*domain.Object, error)
	// SoftDelete marks an object as deleted without immediately removing its underlying binary data.
	SoftDelete(ctx context.Context, bucket, key string) error
}

// FileStore abstracts the low-level binary data operations for object storage (e.g., Local disk, S3).
type FileStore interface {
	// Write streams binary data from a reader into storage at the specified location.
	Write(ctx context.Context, bucket, key string, r io.Reader) (int64, error)
	// Read opens a stream to the binary data of an object in storage.
	Read(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	// Delete permanently removes binary data from the storage backend.
	Delete(ctx context.Context, bucket, key string) error
}

// StorageService provides business logic for managing bucket-based object storage resources (e.g., Cloud Storage).
type StorageService interface {
	// Upload manages the metadata registration and binary data transfer of a new object.
	Upload(ctx context.Context, bucket, key string, r io.Reader) (*domain.Object, error)
	// Download retrieves both the binary content and metadata for a specified object.
	Download(ctx context.Context, bucket, key string) (io.ReadCloser, *domain.Object, error)
	// ListObjects returns metadata for all accessible objects in a bucket.
	ListObjects(ctx context.Context, bucket string) ([]*domain.Object, error)
	// DeleteObject manages the coordinated removal of an object's metadata and its binary data.
	DeleteObject(ctx context.Context, bucket, key string) error
}
