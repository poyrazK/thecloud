// Package ports defines service and repository interfaces.
package ports

import (
	"context"

	"github.com/google/uuid"
)

// VolumeEncryptionService manages encryption keys for database volumes.
type VolumeEncryptionService interface {
	// CreateVolumeKey creates a new encrypted DEK for a volume.
	CreateVolumeKey(ctx context.Context, volumeID uuid.UUID, kmsKeyID string) error
	// GetVolumeDEK retrieves and decrypts the DEK for a volume.
	GetVolumeDEK(ctx context.Context, volumeID uuid.UUID) ([]byte, error)
	// DeleteVolumeKey removes the DEK for a volume.
	DeleteVolumeKey(ctx context.Context, volumeID uuid.UUID) error
	// IsVolumeEncrypted returns whether a volume has encryption enabled.
	IsVolumeEncrypted(ctx context.Context, volumeID uuid.UUID) (bool, error)
}

// VolumeEncryptionRepository persists encrypted volume DEKs.
type VolumeEncryptionRepository interface {
	// SaveKey stores an encrypted DEK for a volume.
	SaveKey(ctx context.Context, volID uuid.UUID, kmsKeyID string, encryptedDEK []byte, algorithm string) error
	// GetKey retrieves an encrypted DEK for a volume.
	GetKey(ctx context.Context, volID uuid.UUID) ([]byte, string, error)
	// DeleteKey removes the encrypted DEK for a volume.
	DeleteKey(ctx context.Context, volID uuid.UUID) error
}
