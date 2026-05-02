// Package services_test provides tests for services.
package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/stretchr/testify/mock"
)

// mockKMSClient is a mock KMS client for testing volume encryption.
type mockKMSClient struct {
	mock.Mock
}

func (m *mockKMSClient) Encrypt(ctx context.Context, keyID string, plaintext []byte) ([]byte, error) {
	args := m.Called(ctx, keyID, plaintext)
	var result []byte
	if args.Get(0) != nil {
		result = args.Get(0).([]byte)
	}
	return result, args.Error(1)
}

func (m *mockKMSClient) Decrypt(ctx context.Context, keyID string, ciphertext []byte) ([]byte, error) {
	args := m.Called(ctx, keyID, ciphertext)
	var result []byte
	if args.Get(0) != nil {
		result = args.Get(0).([]byte)
	}
	return result, args.Error(1)
}

func (m *mockKMSClient) GenerateKey(ctx context.Context, keyID string) ([]byte, error) {
	args := m.Called(ctx, keyID)
	var result []byte
	if args.Get(0) != nil {
		result = args.Get(0).([]byte)
	}
	return result, args.Error(1)
}

func TestVolumeEncryptionService_CreateVolumeKey(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		repo := new(mockVolumeEncryptionRepo)
		kms := new(mockKMSClient)
		svc, err := services.NewVolumeEncryptionService(repo, kms)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		volID := uuid.New()
		kmsKeyID := "vault:transit/my-key"

		// Mock DEK generation and KMS encryption
		kms.On("Encrypt", mock.Anything, kmsKeyID, mock.Anything).Return([]byte("encrypted-dek"), nil)
		repo.On("SaveKey", mock.Anything, volID, kmsKeyID, mock.Anything, "AES-256-GCM").Return(nil)

		err = svc.CreateVolumeKey(context.Background(), volID, kmsKeyID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		repo.AssertExpectations(t)
		kms.AssertExpectations(t)
	})

	t.Run("kms encrypt failure", func(t *testing.T) {
		repo := new(mockVolumeEncryptionRepo)
		kms := new(mockKMSClient)
		svc, err := services.NewVolumeEncryptionService(repo, kms)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		volID := uuid.New()
		kmsKeyID := "vault:transit/my-key"

		kms.On("Encrypt", mock.Anything, kmsKeyID, mock.Anything).Return(nil, context.DeadlineExceeded)

		err = svc.CreateVolumeKey(context.Background(), volID, kmsKeyID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		repo.AssertExpectations(t)
		kms.AssertExpectations(t)
	})
}

func TestVolumeEncryptionService_GetVolumeDEK(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		repo := new(mockVolumeEncryptionRepo)
		kms := new(mockKMSClient)
		svc, err := services.NewVolumeEncryptionService(repo, kms)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		volID := uuid.New()
		kmsKeyID := "vault:transit/my-key"
		encryptedDEK := []byte("encrypted-dek")
		decryptedDEK := []byte("decrypted-dek")

		repo.On("GetKey", mock.Anything, volID).Return(encryptedDEK, kmsKeyID, nil)
		kms.On("Decrypt", mock.Anything, kmsKeyID, encryptedDEK).Return(decryptedDEK, nil)

		dek, err := svc.GetVolumeDEK(context.Background(), volID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if string(dek) != string(decryptedDEK) {
			t.Fatalf("expected DEK %v, got %v", decryptedDEK, dek)
		}

		repo.AssertExpectations(t)
		kms.AssertExpectations(t)
	})
}

func TestVolumeEncryptionService_IsVolumeEncrypted(t *testing.T) {
	t.Parallel()

	t.Run("encrypted", func(t *testing.T) {
		repo := new(mockVolumeEncryptionRepo)
		kms := new(mockKMSClient)
		svc, err := services.NewVolumeEncryptionService(repo, kms)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		volID := uuid.New()
		repo.On("GetKey", mock.Anything, volID).Return([]byte("encrypted-dek"), "vault:transit/my-key", nil)

		encrypted, err := svc.IsVolumeEncrypted(context.Background(), volID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !encrypted {
			t.Fatal("expected volume to be encrypted")
		}

		repo.AssertExpectations(t)
	})

	t.Run("not encrypted", func(t *testing.T) {
		repo := new(mockVolumeEncryptionRepo)
		kms := new(mockKMSClient)
		svc, err := services.NewVolumeEncryptionService(repo, kms)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		volID := uuid.New()
		repo.On("GetKey", mock.Anything, volID).Return(nil, "", errors.New(errors.NotFound, "volume encryption key not found"))

		encrypted, err := svc.IsVolumeEncrypted(context.Background(), volID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if encrypted {
			t.Fatal("expected volume to not be encrypted")
		}

		repo.AssertExpectations(t)
	})
}

// mockVolumeEncryptionRepo is a mock volume encryption repository.
type mockVolumeEncryptionRepo struct {
	mock.Mock
}

func (m *mockVolumeEncryptionRepo) SaveKey(ctx context.Context, volID uuid.UUID, kmsKeyID string, encryptedDEK []byte, algorithm string) error {
	args := m.Called(ctx, volID, kmsKeyID, encryptedDEK, algorithm)
	return args.Error(0)
}

func (m *mockVolumeEncryptionRepo) GetKey(ctx context.Context, volID uuid.UUID) ([]byte, string, error) {
	args := m.Called(ctx, volID)
	var encryptedDEK []byte
	if args.Get(0) != nil {
		encryptedDEK = args.Get(0).([]byte)
	}
	return encryptedDEK, args.String(1), args.Error(2)
}

func (m *mockVolumeEncryptionRepo) DeleteKey(ctx context.Context, volID uuid.UUID) error {
	args := m.Called(ctx, volID)
	return args.Error(0)
}
