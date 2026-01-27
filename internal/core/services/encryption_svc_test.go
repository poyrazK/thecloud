package services_test

import (
	"context"
	"testing"

	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEncryptionService(t *testing.T) {
	masterKeyHex := "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"
	bucket := "test-bucket"
	ctx := context.Background()

	newService := func(t *testing.T) (*services.EncryptionService, *MockEncryptionRepository) {
		repo := new(MockEncryptionRepository)
		svc, err := services.NewEncryptionService(repo, masterKeyHex)
		assert.NoError(t, err)
		return svc, repo
	}

	t.Run("CreateKey", func(t *testing.T) {
		svc, mockRepo := newService(t)
		mockRepo.On("SaveKey", ctx, mock.MatchedBy(func(key ports.EncryptionKey) bool {
			return key.BucketName == bucket && len(key.EncryptedKey) > 0
		})).Return(nil).Once()

		keyID, err := svc.CreateKey(ctx, bucket)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("RotateKey", func(t *testing.T) {
		svc, mockRepo := newService(t)
		mockRepo.On("SaveKey", ctx, mock.MatchedBy(func(key ports.EncryptionKey) bool {
			return key.BucketName == bucket && len(key.EncryptedKey) > 0
		})).Return(nil).Once()

		keyID, err := svc.RotateKey(ctx, bucket)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("EncryptDecrypt", func(t *testing.T) {
		svc, mockRepo := newService(t)
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

	t.Run("EncryptRepoError", func(t *testing.T) {
		svc, mockRepo := newService(t)
		mockRepo.On("GetKey", ctx, bucket).Return(nil, errors.New(errors.Internal, "repo error")).Once()

		_, err := svc.Encrypt(ctx, bucket, []byte("data"))
		assert.Error(t, err)
	})

	t.Run("DecryptRepoError", func(t *testing.T) {
		svc, mockRepo := newService(t)
		mockRepo.On("GetKey", ctx, bucket).Return(nil, errors.New(errors.Internal, "repo error")).Once()

		_, err := svc.Decrypt(ctx, bucket, []byte("data"))
		assert.Error(t, err)
	})

	t.Run("DecryptInvalidData", func(t *testing.T) {
		svc, mockRepo := newService(t)
		mockRepo.On("GetKey", ctx, bucket).Return(&ports.EncryptionKey{EncryptedKey: []byte("short")}, nil).Once()

		_, err := svc.Decrypt(ctx, bucket, []byte("data"))
		assert.Error(t, err)
	})

	t.Run("CreateKeyRepoError", func(t *testing.T) {
		svc, mockRepo := newService(t)
		mockRepo.On("SaveKey", ctx, mock.Anything).Return(errors.New(errors.Internal, "save failed")).Once()

		_, err := svc.CreateKey(ctx, bucket)
		assert.Error(t, err)
	})

	t.Run("InvalidConstructor", func(t *testing.T) {
		mockRepo := new(MockEncryptionRepository)
		_, err := services.NewEncryptionService(mockRepo, "invalid hex")
		assert.Error(t, err)

		shortKey := "000102"
		_, err = services.NewEncryptionService(mockRepo, shortKey)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "32 bytes")
	})
}
