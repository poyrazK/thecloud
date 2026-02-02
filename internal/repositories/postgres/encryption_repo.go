// Package postgres provides Postgres-backed repository implementations.
package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/poyrazk/thecloud/internal/core/ports"
	cerr "github.com/poyrazk/thecloud/internal/errors"
)

// EncryptionRepository stores encryption keys in Postgres.
type EncryptionRepository struct {
	db DB
}

// NewEncryptionRepository constructs an EncryptionRepository.
func NewEncryptionRepository(db DB) *EncryptionRepository {
	return &EncryptionRepository{db: db}
}

func (r *EncryptionRepository) SaveKey(ctx context.Context, key ports.EncryptionKey) error {
	query := `
		INSERT INTO encryption_keys (id, bucket_name, encrypted_key, algorithm)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (bucket_name) DO UPDATE SET
			id = EXCLUDED.id,
			encrypted_key = EXCLUDED.encrypted_key,
			algorithm = EXCLUDED.algorithm
	`
	_, err := r.db.Exec(ctx, query, key.ID, key.BucketName, key.EncryptedKey, key.Algorithm)
	if err != nil {
		return cerr.Wrap(cerr.Internal, "failed to save encryption key", err)
	}
	return nil
}

func (r *EncryptionRepository) GetKey(ctx context.Context, bucketName string) (*ports.EncryptionKey, error) {
	query := `SELECT id, bucket_name, encrypted_key, algorithm FROM encryption_keys WHERE bucket_name = $1`
	var k ports.EncryptionKey
	err := r.db.QueryRow(ctx, query, bucketName).Scan(&k.ID, &k.BucketName, &k.EncryptedKey, &k.Algorithm)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, cerr.New(cerr.NotFound, "encryption key not found")
		}
		return nil, cerr.Wrap(cerr.Internal, "failed to get encryption key", err)
	}
	return &k, nil
}
