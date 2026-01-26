package ports

import (
	"context"
)

// EncryptionService handles encryption and decryption of data streams.
type EncryptionService interface {
	// Encrypt returns an encrypted version of the key and an initialization vector.
	// For object storage, it might return a key ID or wrapper.
	Encrypt(ctx context.Context, bucket string, data []byte) ([]byte, error)
	Decrypt(ctx context.Context, bucket string, encryptedData []byte) ([]byte, error)

	// CreateKey creates a new encryption key for a bucket
	CreateKey(ctx context.Context, bucket string) (string, error)
	// RotateKey rotates the key for a bucket (re-encryption is separate process)
	RotateKey(ctx context.Context, bucket string) (string, error)
}

type EncryptionKey struct {
	ID           string
	BucketName   string
	EncryptedKey []byte
	Algorithm    string
}

type EncryptionRepository interface {
	SaveKey(ctx context.Context, key EncryptionKey) error
	GetKey(ctx context.Context, bucketName string) (*EncryptionKey, error)
}
