package services_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupEncryptionServiceTest(t *testing.T) (*services.EncryptionService, *postgres.EncryptionRepository, context.Context) {
	t.Helper()
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	masterKeyHex := "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"
	repo := postgres.NewEncryptionRepository(db)
	svc, err := services.NewEncryptionService(repo, masterKeyHex)
	require.NoError(t, err)

	return svc, repo, ctx
}

func TestEncryptionService(t *testing.T) {
	svc, repo, ctx := setupEncryptionServiceTest(t)
	bucket := "test-bucket"

	t.Run("CreateKey", func(t *testing.T) {
		keyID, err := svc.CreateKey(ctx, bucket)
		require.NoError(t, err)
		assert.NotEmpty(t, keyID)

		// Verify in DB
		fetched, err := repo.GetKey(ctx, bucket)
		require.NoError(t, err)
		assert.Equal(t, keyID, fetched.ID)
	})

	t.Run("RotateKey", func(t *testing.T) {
		oldID, _ := svc.CreateKey(ctx, bucket)

		newID, err := svc.RotateKey(ctx, bucket)
		require.NoError(t, err)
		assert.NotEqual(t, oldID, newID)

		// Verify in DB
		fetched, _ := repo.GetKey(ctx, bucket)
		assert.Equal(t, newID, fetched.ID)
	})

	t.Run("EncryptDecrypt", func(t *testing.T) {
		plaintext := []byte("top secret data")
		_, _ = svc.CreateKey(ctx, bucket)

		cipherReader, err := svc.Encrypt(ctx, bucket, bytes.NewReader(plaintext))
		require.NoError(t, err)

		ciphertext, err := io.ReadAll(cipherReader)
		require.NoError(t, err)
		assert.NotEqual(t, plaintext, ciphertext)

		plainReader, err := svc.Decrypt(ctx, bucket, bytes.NewReader(ciphertext))
		require.NoError(t, err)

		decrypted, err := io.ReadAll(plainReader)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("DecryptInvalidData", func(t *testing.T) {
		_, _ = svc.CreateKey(ctx, bucket)
		_, err := svc.Decrypt(ctx, bucket, bytes.NewReader([]byte("short")))
		require.Error(t, err)
	})

	t.Run("KeyNotFound", func(t *testing.T) {
		_, err := svc.Encrypt(ctx, "non-existent", bytes.NewReader([]byte("data")))
		require.Error(t, err)
	})
}

