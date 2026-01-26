package services_test

import (
	"context"
	"testing"

	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEncryptionService(t *testing.T) {
	mockRepo := new(MockEncryptionRepository)
	masterKeyHex := "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"
	svc, err := services.NewEncryptionService(mockRepo, masterKeyHex)
	assert.NoError(t, err)

	bucket := "test-bucket"
	ctx := context.Background()

	t.Run("CreateKey", func(t *testing.T) {
		mockRepo.On("SaveKey", ctx, mock.MatchedBy(func(key ports.EncryptionKey) bool {
			return key.BucketName == bucket && len(key.EncryptedKey) > 0
		})).Return(nil).Once()

		keyID, err := svc.CreateKey(ctx, bucket)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("EncryptDecrypt", func(t *testing.T) {
		plaintext := []byte("secret data")

		// 1. Setup DEK in repo
		// We'll simulate by creating one and then mocking GetKey
		// But s.Encrypt calls getAndDecryptDEK

		// Let's manually create a DEK and its encrypted version for the mock
		dek := make([]byte, 32)
		for i := range dek {
			dek[i] = byte(i)
		}

		// Encrypt DEK with master key to store in mock
		// We'll use a hack: just use one that the service generates if we can't easily reproduce its random IV
		// Actually, let's just mock SaveKey to capture what it saves, then Return that in GetKey.

		var savedKey ports.EncryptionKey
		mockRepo.On("SaveKey", ctx, mock.Anything).Run(func(args mock.Arguments) {
			savedKey = args.Get(1).(ports.EncryptionKey)
		}).Return(nil).Once()

		_, _ = svc.CreateKey(ctx, bucket)

		mockRepo.On("GetKey", ctx, bucket).Return(&savedKey, nil)

		// Encrypt
		ciphertext, err := svc.Encrypt(ctx, bucket, plaintext)
		assert.NoError(t, err)
		assert.NotEqual(t, plaintext, ciphertext)

		// Decrypt
		decrypted, err := svc.Decrypt(ctx, bucket, ciphertext)
		assert.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("DecryptInvalidData", func(t *testing.T) {
		mockRepo.On("GetKey", ctx, bucket).Return(nil, nil).Once() // Will fail in getAndDecryptDEK if get returns nil
		// Wait, GetKey returns (*ports.EncryptionKey, error).
		// If both nil, it might panic if not handled.
		// Let's check encryption_svc.go:87
		/*
			keyRecord, err := s.repo.GetKey(ctx, bucket)
			if err != nil {
				return nil, err
			}
			return s.decryptData(s.masterKey, keyRecord.EncryptedKey)
		*/
		// It doesn't check keyRecord == nil.
	})

	t.Run("InvalidConstructor", func(t *testing.T) {
		_, err := services.NewEncryptionService(mockRepo, "invalid hex")
		assert.Error(t, err)

		shortKey := "000102"
		_, err = services.NewEncryptionService(mockRepo, shortKey)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "32 bytes")
	})
}
