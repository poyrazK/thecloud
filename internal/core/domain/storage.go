// Package domain defines core business entities.
package domain

import (
	"io"
	"time"

	"github.com/google/uuid"
)

// Object represents stored object metadata in the storage subsystem.
type Object struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	ARN         string     `json:"arn"`
	Bucket      string     `json:"bucket"`
	Key         string     `json:"key"`
	VersionID   string     `json:"version_id"`
	IsLatest    bool       `json:"is_latest"`
	SizeBytes   int64      `json:"size_bytes"`
	ContentType string     `json:"content_type"`
	CreatedAt   time.Time  `json:"created_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	Data        io.Reader  `json:"-"` // Stream for reading/writing
}

// Bucket represents a storage bucket configuration and metadata.
type Bucket struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	UserID            uuid.UUID `json:"user_id"`
	IsPublic          bool      `json:"is_public"`
	VersioningEnabled bool      `json:"versioning_enabled"`
	EncryptionEnabled bool      `json:"encryption_enabled"`
	EncryptionKeyID   string    `json:"encryption_key_id,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

// StorageNode describes a node in the storage cluster.
type StorageNode struct {
	ID       string    `json:"id"`
	Address  string    `json:"address"` // host:port
	DataDir  string    `json:"data_dir"`
	Status   string    `json:"status"`
	LastSeen time.Time `json:"last_seen"`
}

// StorageCluster aggregates storage nodes for cluster-level operations.
type StorageCluster struct {
	Nodes []StorageNode `json:"nodes"`
}

// MultipartUpload represents an in-progress multipart upload.
type MultipartUpload struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Bucket    string    `json:"bucket"`
	Key       string    `json:"key"`
	CreatedAt time.Time `json:"created_at"`
}

// Part represents a single part of a multipart upload.
type Part struct {
	UploadID   uuid.UUID `json:"upload_id"`
	PartNumber int       `json:"part_number"`
	SizeBytes  int64     `json:"size_bytes"`
	ETag       string    `json:"etag"`
}

// PresignedURL represents a generated signed URL for temporary access.
type PresignedURL struct {
	URL       string    `json:"url"`
	Method    string    `json:"method"`     // GET or PUT
	ExpiresAt time.Time `json:"expires_at"` // Timestamp when URL expires
}
