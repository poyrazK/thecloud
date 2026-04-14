// Package postgres provides Postgres-backed repository implementations.
package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	cerr "github.com/poyrazk/thecloud/internal/errors"
)

// VolumeEncryptionKey stores an encrypted DEK for a volume.
type VolumeEncryptionKey struct {
	VolumeID     uuid.UUID
	EncryptedDEK []byte
	KmsKeyID     string
	Algorithm    string
	CreatedAt    interface{} // time.Time but may be pgx return type
}

// VolumeEncryptionRepository stores encrypted volume DEKs in Postgres.
type VolumeEncryptionRepository struct {
	db DB
}

// NewVolumeEncryptionRepository constructs a VolumeEncryptionRepository.
func NewVolumeEncryptionRepository(db DB) *VolumeEncryptionRepository {
	return &VolumeEncryptionRepository{db: db}
}

// SaveKey stores an encrypted DEK for a volume.
func (r *VolumeEncryptionRepository) SaveKey(ctx context.Context, volID uuid.UUID, kmsKeyID string, encryptedDEK []byte, algorithm string) error {
	query := `
		INSERT INTO volume_encryption_keys (volume_id, encrypted_dek, kms_key_id, algorithm)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (volume_id) DO UPDATE SET
			encrypted_dek = EXCLUDED.encrypted_dek,
			kms_key_id = EXCLUDED.kms_key_id,
			algorithm = EXCLUDED.algorithm
	`
	_, err := r.db.Exec(ctx, query, volID, encryptedDEK, kmsKeyID, algorithm)
	if err != nil {
		return cerr.Wrap(cerr.Internal, "failed to save volume encryption key", err)
	}
	return nil
}

// GetKey retrieves an encrypted DEK and KMS key ID for a volume.
func (r *VolumeEncryptionRepository) GetKey(ctx context.Context, volID uuid.UUID) ([]byte, string, error) {
	query := `SELECT encrypted_dek, kms_key_id FROM volume_encryption_keys WHERE volume_id = $1`
	var encryptedDEK []byte
	var kmsKeyID string
	err := r.db.QueryRow(ctx, query, volID).Scan(&encryptedDEK, &kmsKeyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", cerr.New(cerr.NotFound, "volume encryption key not found")
		}
		return nil, "", cerr.Wrap(cerr.Internal, "failed to get volume encryption key", err)
	}
	return encryptedDEK, kmsKeyID, nil
}

// DeleteKey removes the encrypted DEK for a volume.
func (r *VolumeEncryptionRepository) DeleteKey(ctx context.Context, volID uuid.UUID) error {
	query := `DELETE FROM volume_encryption_keys WHERE volume_id = $1`
	_, err := r.db.Exec(ctx, query, volID)
	if err != nil {
		return cerr.Wrap(cerr.Internal, "failed to delete volume encryption key", err)
	}
	return nil
}