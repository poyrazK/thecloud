// Package ports defines interfaces for adapters and services.
package ports

import (
	"context"
	"io"
)

// EncryptionService handles encryption and decryption of data streams.
type EncryptionService interface {
	// Encrypt returns a transformed io.Reader that encrypts the data from r on-the-fly.
	Encrypt(ctx context.Context, bucket string, r io.Reader) (io.Reader, error)
	// Decrypt returns a transformed io.Reader that decrypts the data from r on-the-fly.
	Decrypt(ctx context.Context, bucket string, r io.Reader) (io.Reader, error)

	// CreateKey creates a new encryption key for a bucket
	CreateKey(ctx context.Context, bucket string) (string, error)
	// RotateKey rotates the key for a bucket (re-encryption is separate process)
	RotateKey(ctx context.Context, bucket string) (string, error)
}

// EncryptionKey stores metadata for a bucket encryption key.
type EncryptionKey struct {
	ID           string
	BucketName   string
	EncryptedKey []byte
	Algorithm    string
}

// EncryptionRepository persists encryption keys for buckets.
type EncryptionRepository interface {
	SaveKey(ctx context.Context, key EncryptionKey) error
	GetKey(ctx context.Context, bucketName string) (*EncryptionKey, error)
}
