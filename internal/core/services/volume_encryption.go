// Package services implements core business logic.
package services

import (
	"context"
	"crypto/rand"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
)

const (
	dekKeySize         = 32            // 256-bit DEK
	dekCipherAlgorithm = "AES-256-GCM" // DEK encryption algorithm
)

// VolumeEncryptionServiceImpl implements ports.VolumeEncryptionService.
type VolumeEncryptionServiceImpl struct {
	repo ports.VolumeEncryptionRepository
	kms  ports.KMSClient
}

// NewVolumeEncryptionService creates a new VolumeEncryptionService.
func NewVolumeEncryptionService(repo ports.VolumeEncryptionRepository, kms ports.KMSClient) (*VolumeEncryptionServiceImpl, error) {
	if repo == nil {
		return nil, errors.New(errors.Internal, "nil repo")
	}
	if kms == nil {
		return nil, errors.New(errors.Internal, "nil kms")
	}
	return &VolumeEncryptionServiceImpl{
		repo: repo,
		kms:  kms,
	}, nil
}

// CreateVolumeKey generates a new DEK, encrypts it with KMS, and stores it.
func (s *VolumeEncryptionServiceImpl) CreateVolumeKey(ctx context.Context, volumeID uuid.UUID, kmsKeyID string) error {
	// Generate random 256-bit DEK
	dek := make([]byte, dekKeySize)
	if _, err := rand.Read(dek); err != nil {
		return errors.Wrap(errors.Internal, "failed to generate DEK", err)
	}

	// Encrypt DEK with KMS (Vault Transit)
	encryptedDEK, err := s.kms.Encrypt(ctx, kmsKeyID, dek)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to encrypt DEK", err)
	}

	// Store encrypted DEK in database
	err = s.repo.SaveKey(ctx, volumeID, kmsKeyID, encryptedDEK, dekCipherAlgorithm)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to store encrypted DEK", err)
	}

	return nil
}

// GetVolumeDEK retrieves the encrypted DEK and decrypts it with KMS.
func (s *VolumeEncryptionServiceImpl) GetVolumeDEK(ctx context.Context, volumeID uuid.UUID) ([]byte, error) {
	// Get encrypted DEK and KMS key ID from database
	encryptedDEK, kmsKeyID, err := s.repo.GetKey(ctx, volumeID)
	if err != nil {
		if errors.Is(err, errors.NotFound) {
			return nil, errors.New(errors.NotFound, "volume DEK not found")
		}
		return nil, errors.Wrap(errors.Internal, "failed to fetch volume DEK", err)
	}

	// Decrypt DEK with KMS
	dek, err := s.kms.Decrypt(ctx, kmsKeyID, encryptedDEK)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to decrypt DEK", err)
	}

	return dek, nil
}

// DeleteVolumeKey removes the DEK for a volume.
func (s *VolumeEncryptionServiceImpl) DeleteVolumeKey(ctx context.Context, volumeID uuid.UUID) error {
	err := s.repo.DeleteKey(ctx, volumeID)
	if err != nil {
		if errors.Is(err, errors.NotFound) {
			return nil // Deleting non-existent key is a no-op
		}
		return errors.Wrap(errors.Internal, "failed to delete volume DEK", err)
	}
	return nil
}

// IsVolumeEncrypted checks whether a volume has an encryption key.
func (s *VolumeEncryptionServiceImpl) IsVolumeEncrypted(ctx context.Context, volumeID uuid.UUID) (bool, error) {
	_, _, err := s.repo.GetKey(ctx, volumeID)
	if err != nil {
		if errors.Is(err, errors.NotFound) {
			return false, nil
		}
		return false, errors.Wrap(errors.Internal, "failed to check volume encryption", err)
	}
	return true, nil
}

// Ensure VolumeEncryptionServiceImpl implements ports.VolumeEncryptionService
var _ ports.VolumeEncryptionService = (*VolumeEncryptionServiceImpl)(nil)
