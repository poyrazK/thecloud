// Package services implements core business logic.
package services

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
)

// EncryptionService implements ports.EncryptionService
type EncryptionService struct {
	repo      ports.EncryptionRepository
	masterKey []byte // 32-bytes
}

// NewEncryptionService constructs an EncryptionService from the repository and master key.
func NewEncryptionService(repo ports.EncryptionRepository, masterKeyHex string) (*EncryptionService, error) {
	key, err := hex.DecodeString(masterKeyHex)
	if err == nil && len(key) == 32 {
		return &EncryptionService{
			repo:      repo,
			masterKey: key,
		}, nil
	}

	// Try as raw string if it's explicitly 32 bytes
	if len(masterKeyHex) == 32 {
		return &EncryptionService{
			repo:      repo,
			masterKey: []byte(masterKeyHex),
		}, nil
	}

	// Fallback: Use SHA256 to derive a 32-byte key from the passphrase/input
	// This ensures we always have a valid AES-256 key regardless of input length
	hash := sha256.Sum256([]byte(masterKeyHex))
	return &EncryptionService{
		repo:      repo,
		masterKey: hash[:],
	}, nil
}

func (s *EncryptionService) CreateKey(ctx context.Context, bucket string) (string, error) {
	// Generate new random data key (DEK)
	dek := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, dek); err != nil {
		return "", errors.Wrap(errors.Internal, "failed to generate random key", err)
	}

	// Encrypt DEK with Master Key (KEK) using AES-GCM
	encryptedDEK, err := s.encryptData(s.masterKey, dek)
	if err != nil {
		return "", err
	}

	keyID := uuid.New().String()
	err = s.repo.SaveKey(ctx, ports.EncryptionKey{
		ID:           keyID,
		BucketName:   bucket,
		EncryptedKey: encryptedDEK,
		Algorithm:    "AES-256-GCM",
	})
	if err != nil {
		return "", err
	}
	return keyID, nil
}

// RotateKey generates a new DEK for the bucket.
// It relies on the repository to handle versioning (INSERT without ON CONFLICT).
func (s *EncryptionService) RotateKey(ctx context.Context, bucket string) (string, error) {
	return s.CreateKey(ctx, bucket)
}

// Encrypt encrypts data stream using the bucket's active key
func (s *EncryptionService) Encrypt(ctx context.Context, bucket string, r io.Reader) (io.Reader, error) {
	// For encryption, always use the latest key (versionID="")
	dek, err := s.getAndDecryptDEK(ctx, bucket, "")
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to create cipher", err)
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to generate IV", err)
	}

	stream := cipher.NewCTR(block, iv)
	
	// The output stream should be: IV + EncryptedData
	return io.MultiReader(
		bytes.NewReader(iv),
		&cipher.StreamReader{S: stream, R: r},
	), nil
}

// Decrypt decrypts data stream using the bucket's active key
func (s *EncryptionService) Decrypt(ctx context.Context, bucket string, r io.Reader) (io.Reader, error) {
	dek, err := s.getAndDecryptDEK(ctx, bucket, "")
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to create cipher", err)
	}

	// Read IV from the start of the stream
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(r, iv); err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to read IV", err)
	}

	stream := cipher.NewCTR(block, iv)

	return &cipher.StreamReader{S: stream, R: r}, nil
}

// getAndDecryptDEK fetches the encrypted DEK from DB and decrypts it with Master Key
func (s *EncryptionService) getAndDecryptDEK(ctx context.Context, bucket, _ string) ([]byte, error) {
	// TODO: Update repo.GetKey to support optional versionID filters if needed.
	// Currently repo.GetKey likely returns the latest.
	// If versionID is provided, we should fetch that specific key.
	// But without changing repo interface in 'ports' (which I can't see), I assume GetKey returns the latest "active" key.
	// If I wanted a specific key I'd need GetKeyByID(ctx, keyID).
	keyRecord, err := s.repo.GetKey(ctx, bucket)
	if err != nil {
		return nil, err
	}

	return s.decryptData(s.masterKey, keyRecord.EncryptedKey)
}

// encryptData performs AES-GCM encryption
func (s *EncryptionService) encryptData(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to create cipher", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to create GCM", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to generate nonce", err)
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

// decryptData performs AES-GCM decryption
func (s *EncryptionService) decryptData(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to create cipher", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to create GCM", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New(errors.Internal, "invalid ciphertext size")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "decryption failed", err)
	}
	return plaintext, nil
}
