// Package services implements core business logic.
package services

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
)

const (
	// encryptionChunkSize defines the size of plaintext chunks for AEAD framing.
	encryptionChunkSize = 64 * 1024 // 64KB
	// nonceSize is the standard size for AES-GCM nonces.
	nonceSize = 12
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

// Encrypt returns a transformed io.Reader that encrypts the data from r using chunked AES-GCM.
func (s *EncryptionService) Encrypt(ctx context.Context, bucket string, r io.Reader) (io.Reader, error) {
	dek, err := s.getAndDecryptDEK(ctx, bucket, "")
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to create cipher", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to create GCM", err)
	}

	baseNonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, baseNonce); err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to generate base nonce", err)
	}

	return &chunkedGCMReader{
		r:         r,
		aead:      aead,
		baseNonce: baseNonce,
	}, nil
}

type chunkedGCMReader struct {
	r         io.Reader
	aead      cipher.AEAD
	baseNonce []byte
	counter   uint64
	headerSent bool
	buf       bytes.Buffer
}

func (c *chunkedGCMReader) Read(p []byte) (n int, err error) {
	if !c.headerSent {
		c.buf.Write(c.baseNonce)
		c.headerSent = true
	}

	for c.buf.Len() < len(p) {
		plaintext := make([]byte, encryptionChunkSize)
		nr, readErr := io.ReadFull(c.r, plaintext)
		if nr > 0 {
			nonce := make([]byte, nonceSize)
			copy(nonce, c.baseNonce)
			// Simple counter-based nonce derivation: XOR baseNonce with counter
			binary.BigEndian.PutUint64(nonce[nonceSize-8:], binary.BigEndian.Uint64(nonce[nonceSize-8:])^c.counter)
			c.counter++

			ciphertext := c.aead.Seal(nil, nonce, plaintext[:nr], nil)
			
			// Record framing: [Length (4 bytes)] [Ciphertext + Tag]
			// G115: add safety check for int -> uint32 conversion
			if uint64(len(ciphertext)) > 0xFFFFFFFF {
				return 0, fmt.Errorf("ciphertext too large: %d", len(ciphertext))
			}
			lenBuf := make([]byte, 4)
			binary.BigEndian.PutUint32(lenBuf, uint32(len(ciphertext)))
			c.buf.Write(lenBuf)
			c.buf.Write(ciphertext)
		}
		if readErr != nil {
			if readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
				if c.buf.Len() == 0 {
					return 0, io.EOF
				}
				break
			}
			return 0, readErr
		}
	}

	return c.buf.Read(p)
}

// Decrypt returns a transformed io.Reader that decrypts the data from r using chunked AES-GCM.
func (s *EncryptionService) Decrypt(ctx context.Context, bucket string, r io.Reader) (io.Reader, error) {
	dek, err := s.getAndDecryptDEK(ctx, bucket, "")
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to create cipher", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to create GCM", err)
	}

	baseNonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(r, baseNonce); err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to read base nonce", err)
	}

	return &chunkedGCMDecryptReader{
		r:         r,
		aead:      aead,
		baseNonce: baseNonce,
	}, nil
}

type chunkedGCMDecryptReader struct {
	r         io.Reader
	aead      cipher.AEAD
	baseNonce []byte
	counter   uint64
	buf       bytes.Buffer
}

func (c *chunkedGCMDecryptReader) Read(p []byte) (n int, err error) {
	for c.buf.Len() < len(p) {
		lenBuf := make([]byte, 4)
		_, readErr := io.ReadFull(c.r, lenBuf)
		if readErr != nil {
			if readErr == io.EOF {
				if c.buf.Len() == 0 {
					return 0, io.EOF
				}
				break
			}
			return 0, readErr
		}

		chunkLen := binary.BigEndian.Uint32(lenBuf)
		ciphertext := make([]byte, chunkLen)
		if _, readErr := io.ReadFull(c.r, ciphertext); readErr != nil {
			return 0, readErr
		}

		nonce := make([]byte, nonceSize)
		copy(nonce, c.baseNonce)
		binary.BigEndian.PutUint64(nonce[nonceSize-8:], binary.BigEndian.Uint64(nonce[nonceSize-8:])^c.counter)
		c.counter++

		plaintext, openErr := c.aead.Open(nil, nonce, ciphertext, nil)
		if openErr != nil {
			return 0, errors.Wrap(errors.Internal, "decryption/authentication failed", openErr)
		}
		c.buf.Write(plaintext)
	}

	return c.buf.Read(p)
}

// getAndDecryptDEK fetches the encrypted DEK from DB and decrypts it with Master Key
func (s *EncryptionService) getAndDecryptDEK(ctx context.Context, bucket, _ string) ([]byte, error) {
	keyRecord, err := s.repo.GetKey(ctx, bucket)
	if err != nil {
		return nil, err
	}

	return s.decryptData(s.masterKey, keyRecord.EncryptedKey)
}

// encryptData performs AES-GCM encryption for small data (e.g. DEKs)
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

// decryptData performs AES-GCM decryption for small data
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
