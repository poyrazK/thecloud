package services_test

import (
	"context"
	"testing"

	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupEncryptionServiceTest(t *testing.T) (*services.EncryptionService, *postgres.EncryptionRepository, context.Context) {
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
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID)

		// Verify in DB
		fetched, err := repo.GetKey(ctx, bucket)
		assert.NoError(t, err)
		assert.Equal(t, keyID, fetched.ID)
	})

	t.Run("RotateKey", func(t *testing.T) {
		oldID, _ := svc.CreateKey(ctx, bucket)

		newID, err := svc.RotateKey(ctx, bucket)
		assert.NoError(t, err)
		assert.NotEqual(t, oldID, newID)

		// Verify in DB
		fetched, _ := repo.GetKey(ctx, bucket)
		assert.Equal(t, newID, fetched.ID)
	})

	t.Run("EncryptDecrypt", func(t *testing.T) {
		plaintext := []byte("top secret data")
		_, _ = svc.CreateKey(ctx, bucket)

		ciphertext, err := svc.Encrypt(ctx, bucket, plaintext)
		assert.NoError(t, err)
		assert.NotEqual(t, plaintext, ciphertext)

		decrypted, err := svc.Decrypt(ctx, bucket, ciphertext)
		assert.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("DecryptInvalidData", func(t *testing.T) {
		_, _ = svc.CreateKey(ctx, bucket)
		_, err := svc.Decrypt(ctx, bucket, []byte("short"))
		assert.Error(t, err)
	})

	t.Run("KeyNotFound", func(t *testing.T) {
		_, err := svc.Encrypt(ctx, "non-existent", []byte("data"))
		assert.Error(t, err)
	})
}
