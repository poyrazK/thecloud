package services

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"io"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/poyrazk/thecloud/internal/repositories/postgres" // for model, ideally move to domain
)

// EncryptionService implements ports.EncryptionService
type EncryptionService struct {
	repo      *postgres.EncryptionRepository // Should rely on interface in real injection but for now struct
	masterKey []byte                         // 32-bytes
}

func NewEncryptionService(repo *postgres.EncryptionRepository, masterKeyHex string) (*EncryptionService, error) {
	key, err := hex.DecodeString(masterKeyHex)
	if err != nil {
		return nil, errors.New(errors.Internal, "invalid master key hex")
	}
	if len(key) != 32 {
		return nil, errors.New(errors.Internal, "master key must be 32 bytes (AES-256)")
	}

	return &EncryptionService{
		repo:      repo,
		masterKey: key,
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
	err = s.repo.SaveKey(ctx, postgres.EncryptionKey{
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

func (s *EncryptionService) RotateKey(ctx context.Context, bucket string) (string, error) {
	return s.CreateKey(ctx, bucket)
}

// Encrypt encrypts data using the bucket's active key
func (s *EncryptionService) Encrypt(ctx context.Context, bucket string, data []byte) ([]byte, error) {
	dek, err := s.getAndDecryptDEK(ctx, bucket)
	if err != nil {
		return nil, err
	}
	return s.encryptData(dek, data)
}

// Decrypt decrypts data using the bucket's active key
func (s *EncryptionService) Decrypt(ctx context.Context, bucket string, encryptedData []byte) ([]byte, error) {
	dek, err := s.getAndDecryptDEK(ctx, bucket)
	if err != nil {
		return nil, err
	}
	return s.decryptData(dek, encryptedData)
}

// getAndDecryptDEK fetches the encrypted DEK from DB and decrypts it with Master Key
func (s *EncryptionService) getAndDecryptDEK(ctx context.Context, bucket string) ([]byte, error) {
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
