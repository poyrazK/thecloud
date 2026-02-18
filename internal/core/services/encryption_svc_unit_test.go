package services_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockEncryptionRepo struct {
	mock.Mock
}

func (m *MockEncryptionRepo) SaveKey(ctx context.Context, key ports.EncryptionKey) error {
	return m.Called(ctx, key).Error(0)
}

func (m *MockEncryptionRepo) GetKey(ctx context.Context, bucketName string) (*ports.EncryptionKey, error) {
	args := m.Called(ctx, bucketName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ports.EncryptionKey), args.Error(1)
}

func TestEncryptionService_Unit(t *testing.T) {
	mockRepo := new(MockEncryptionRepo)
	
	// Create a valid 32-byte hex master key
	masterKey := make([]byte, 32)
	_, err := rand.Read(masterKey)
	assert.NoError(t, err)
	
	masterKeyHex := hex.EncodeToString(masterKey)

	svc, err := services.NewEncryptionService(mockRepo, masterKeyHex)
	assert.NoError(t, err)

	ctx := context.Background()
	bucket := "my-bucket"

	t.Run("CreateKey", func(t *testing.T) {
		mockRepo.On("SaveKey", mock.Anything, mock.MatchedBy(func(k ports.EncryptionKey) bool {
			return k.BucketName == bucket && len(k.EncryptedKey) > 0
		})).Return(nil).Once()

		keyID, err := svc.CreateKey(ctx, bucket)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Encrypt_Decrypt", func(t *testing.T) {
		// Helper: create an encrypted DEK payload using the service's logic
		// Since encryptData is private, we can use CreateKey's logic:
		// We mock SaveKey to capture the argument passed to it
		var savedKey ports.EncryptionKey
		mockRepo.On("SaveKey", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			savedKey = args.Get(1).(ports.EncryptionKey)
		}).Return(nil).Once()

		_, err := svc.CreateKey(ctx, bucket)
		assert.NoError(t, err)

		// Now mock GetKey to return this valid key
		mockRepo.On("GetKey", mock.Anything, bucket).Return(&savedKey, nil)

		data := []byte("secret message")
		encrypted, err := svc.Encrypt(ctx, bucket, data)
		assert.NoError(t, err)
		assert.NotEmpty(t, encrypted)
		assert.NotEqual(t, data, encrypted)

		decrypted, err := svc.Decrypt(ctx, bucket, encrypted)
		assert.NoError(t, err)
		assert.Equal(t, data, decrypted)
	})

	t.Run("RotateKey", func(t *testing.T) {
		mockRepo.On("SaveKey", mock.Anything, mock.MatchedBy(func(k ports.EncryptionKey) bool {
			return k.BucketName == bucket
		})).Return(nil).Once()

		keyID, err := svc.RotateKey(ctx, bucket)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID)
	})
}
