//go:build integration

package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockKMSForIntegration is a mock KMS that simulates Vault Transit behavior in tests.
type mockKMSForIntegration struct {
	keys map[string][]byte // keyID -> encrypted DEK
}

func newMockKMSForIntegration() *mockKMSForIntegration {
	return &mockKMSForIntegration{keys: make(map[string][]byte)}
}

func (m *mockKMSForIntegration) Encrypt(ctx context.Context, keyID string, plaintext []byte) ([]byte, error) {
	// Simulate failure only for explicit "fail-key" sentinel
	if keyID == "fail-key" {
		return nil, errors.New("kms encrypt failure")
	}
	// Simulate Vault Transit encrypt: append "-encrypted" to mark as encrypted
	encrypted := plaintext // copy to avoid modifying input
	m.keys[keyID] = append(encrypted, []byte("-encrypted")...)
	return m.keys[keyID], nil
}

func (m *mockKMSForIntegration) Decrypt(ctx context.Context, keyID string, ciphertext []byte) ([]byte, error) {
	// Simulate Vault Transit decrypt
	_, ok := m.keys[keyID]
	if !ok {
		return nil, nil
	}
	// Remove the "-encrypted" suffix we added
	plaintext := ciphertext[:len(ciphertext)-10]
	return plaintext, nil
}

func (m *mockKMSForIntegration) GenerateKey(ctx context.Context, keyID string) ([]byte, error) {
	return []byte("generated-key"), nil
}

func TestVolumeEncryptionService_Integration(t *testing.T) {
	// Setup real postgres database
	db := setupDB(t)
	cleanDB(t, db)

	volEncRepo := postgres.NewVolumeEncryptionRepository(db)
	mockKMS := newMockKMSForIntegration()
	svc, err := services.NewVolumeEncryptionService(volEncRepo, mockKMS)
	require.NoError(t, err)

	volID := uuid.New()
	kmsKeyID := "vault:transit/test-key"

	t.Run("CreateVolumeKey_and_GetVolumeDEK", func(t *testing.T) {
		// Create an encrypted DEK for a volume
		err := svc.CreateVolumeKey(context.Background(), volID, kmsKeyID)
		require.NoError(t, err)

		// Retrieve and decrypt the DEK
		dek, err := svc.GetVolumeDEK(context.Background(), volID)
		require.NoError(t, err)
		assert.NotEmpty(t, dek)
	})

	t.Run("IsVolumeEncrypted_True", func(t *testing.T) {
		encrypted, err := svc.IsVolumeEncrypted(context.Background(), volID)
		require.NoError(t, err)
		assert.True(t, encrypted)
	})

	t.Run("DeleteVolumeKey", func(t *testing.T) {
		deleteVolID := uuid.New()
		err := svc.CreateVolumeKey(context.Background(), deleteVolID, kmsKeyID)
		require.NoError(t, err)

		err = svc.DeleteVolumeKey(context.Background(), deleteVolID)
		require.NoError(t, err)

		encrypted, err := svc.IsVolumeEncrypted(context.Background(), deleteVolID)
		require.NoError(t, err)
		assert.False(t, encrypted)
	})

	t.Run("DeleteVolumeKey_NonExistent", func(t *testing.T) {
		err := svc.DeleteVolumeKey(context.Background(), uuid.New())
		assert.NoError(t, err) // Should not error on non-existent key
	})

	t.Run("CreateVolumeKey_KMSFailure", func(t *testing.T) {
		// failKMS has empty keys map, so Encrypt will return error for any key
		failKMS := &mockKMSForIntegration{keys: make(map[string][]byte)}
		failSvc, svcErr := services.NewVolumeEncryptionService(volEncRepo, failKMS)
		require.NoError(t, svcErr)

		err := failSvc.CreateVolumeKey(context.Background(), uuid.New(), "fail-key")
		assert.Error(t, err)
	})
}

func TestVolumeEncryptionRepository_Integration(t *testing.T) {
	// Setup real postgres database
	db := setupDB(t)
	cleanDB(t, db)

	repo := postgres.NewVolumeEncryptionRepository(db)
	volID := uuid.New()

	t.Run("SaveKey_and_GetKey", func(t *testing.T) {
		encryptedDEK := []byte("encrypted-dek-data-123")
		err := repo.SaveKey(context.Background(), volID, "vault:transit/test-key", encryptedDEK, "AES-256-GCM")
		require.NoError(t, err)

		retrievedDEK, kmsKeyID, err := repo.GetKey(context.Background(), volID)
		require.NoError(t, err)
		assert.Equal(t, encryptedDEK, retrievedDEK)
		assert.Equal(t, "vault:transit/test-key", kmsKeyID)
	})

	t.Run("GetKey_NotFound", func(t *testing.T) {
		_, _, err := repo.GetKey(context.Background(), uuid.New())
		assert.Error(t, err)
	})

	t.Run("DeleteKey", func(t *testing.T) {
		deleteVolID := uuid.New()
		err := repo.SaveKey(context.Background(), deleteVolID, "vault:transit/delete-test", []byte("encrypted-dek"), "AES-256-GCM")
		require.NoError(t, err)

		err = repo.DeleteKey(context.Background(), deleteVolID)
		require.NoError(t, err)

		_, _, err = repo.GetKey(context.Background(), deleteVolID)
		assert.Error(t, err)
	})

	t.Run("UpdateKey", func(t *testing.T) {
		updateVolID := uuid.New()
		err := repo.SaveKey(context.Background(), updateVolID, "vault:transit/original", []byte("original-dek"), "AES-256-GCM")
		require.NoError(t, err)

		err = repo.SaveKey(context.Background(), updateVolID, "vault:transit/updated", []byte("updated-dek"), "AES-256-GCM")
		require.NoError(t, err)

		encryptedDEK, kmsKeyID, err := repo.GetKey(context.Background(), updateVolID)
		require.NoError(t, err)
		assert.Equal(t, "vault:transit/updated", kmsKeyID)
		assert.Equal(t, []byte("updated-dek"), encryptedDEK)
	})
}

// Ensure mockKMSForIntegration implements ports.KMSClient
var _ ports.KMSClient = (*mockKMSForIntegration)(nil)