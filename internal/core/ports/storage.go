// Package ports defines service and repository interfaces.
package ports

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
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
	// DeleteVersion permanently deletes a specific version's metadata.
	DeleteVersion(ctx context.Context, bucket, key, versionID string) error
	// GetMetaByVersion retrieves metadata for a specific version of an object.
	GetMetaByVersion(ctx context.Context, bucket, key, versionID string) (*domain.Object, error)
	// ListVersions returns all versions of a specific object.
	ListVersions(ctx context.Context, bucket, key string) ([]*domain.Object, error)

	// Bucket operations
	CreateBucket(ctx context.Context, bucket *domain.Bucket) error
	GetBucket(ctx context.Context, name string) (*domain.Bucket, error)
	DeleteBucket(ctx context.Context, name string) error
	ListBuckets(ctx context.Context, userID string) ([]*domain.Bucket, error)
	// SetBucketVersioning enables or disables versioning for a bucket.
	SetBucketVersioning(ctx context.Context, name string, enabled bool) error

	// Multipart operations
	SaveMultipartUpload(ctx context.Context, upload *domain.MultipartUpload) error
	GetMultipartUpload(ctx context.Context, uploadID uuid.UUID) (*domain.MultipartUpload, error)
	DeleteMultipartUpload(ctx context.Context, uploadID uuid.UUID) error
	SavePart(ctx context.Context, part *domain.Part) error
	ListParts(ctx context.Context, uploadID uuid.UUID) ([]*domain.Part, error)
}

// FileStore abstracts the low-level binary data operations for object storage (e.g., Local disk, S3).
type FileStore interface {
	// Write streams binary data from a reader into storage at the specified location.
	Write(ctx context.Context, bucket, key string, r io.Reader) (int64, error)
	// Read opens a stream to the binary data of an object in storage.
	Read(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	// Delete permanently removes binary data from the storage backend.
	Delete(ctx context.Context, bucket, key string) error
	// GetClusterStatus returns the current state of the storage cluster.
	GetClusterStatus(ctx context.Context) (*domain.StorageCluster, error)
	// Assemble combines multiple parts into a single object and removes the parts.
	Assemble(ctx context.Context, bucket, key string, parts []string) (int64, error)
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
	// DownloadVersion retrieves both the binary content and metadata for a specific version of an object.
	DownloadVersion(ctx context.Context, bucket, key, versionID string) (io.ReadCloser, *domain.Object, error)
	// ListVersions returns all versions for a specific object.
	ListVersions(ctx context.Context, bucket, key string) ([]*domain.Object, error)
	// DeleteVersion removes a specific version of an object.
	DeleteVersion(ctx context.Context, bucket, key, versionID string) error

	// Bucket operations
	CreateBucket(ctx context.Context, name string, isPublic bool) (*domain.Bucket, error)
	GetBucket(ctx context.Context, name string) (*domain.Bucket, error)
	DeleteBucket(ctx context.Context, name string) error
	ListBuckets(ctx context.Context) ([]*domain.Bucket, error)
	// SetBucketVersioning enables or disables versioning for a bucket.
	SetBucketVersioning(ctx context.Context, name string, enabled bool) error
	// GetClusterStatus returns the current state of the storage cluster.
	GetClusterStatus(ctx context.Context) (*domain.StorageCluster, error)

	// Multipart operations
	CreateMultipartUpload(ctx context.Context, bucket, key string) (*domain.MultipartUpload, error)
	UploadPart(ctx context.Context, uploadID uuid.UUID, partNumber int, r io.Reader) (*domain.Part, error)
	CompleteMultipartUpload(ctx context.Context, uploadID uuid.UUID) (*domain.Object, error)
	AbortMultipartUpload(ctx context.Context, uploadID uuid.UUID) error

	// Presigned URLs
	GeneratePresignedURL(ctx context.Context, bucket, key, method string, expiry time.Duration) (*domain.PresignedURL, error)
}
